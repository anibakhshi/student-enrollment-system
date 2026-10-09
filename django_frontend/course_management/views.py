from __future__ import annotations

from typing import Any

from django.conf import settings
from django.contrib import messages
from django.core.paginator import Paginator
from django.http import Http404, HttpRequest, HttpResponse
from django.shortcuts import redirect, render
from django.views import View

from accounts.mixins import EducationManagementRequiredMixin
from dashboard.api_client import GoAPIError, go_api

from .forms import CourseForm, CourseSearchForm


COURSES_API_PATH = "/api/v1/courses"
INSTRUCTORS_API_PATH = "/api/v1/instructors"
COURSES_PER_PAGE = 12

STATUS_LABELS = {
    "draft": "پیش‌نویس",
    "open": "باز برای ثبت‌نام",
    "closed": "بسته",
    "completed": "تکمیل‌شده",
    "cancelled": "لغوشده",
}


def get_api_data(
    response: dict[str, Any],
    default: Any = None,
) -> Any:
    if not isinstance(response, dict):
        return default

    data = response.get("data")
    return default if data is None else data


def instructor_full_name(
    instructor: dict[str, Any],
) -> str:
    full_name = " ".join(
        [
            str(instructor.get("first_name") or "").strip(),
            str(instructor.get("last_name") or "").strip(),
        ]
    ).strip()

    return full_name or (
        f"استاد شماره {instructor.get('id', '—')}"
    )


def prepare_course(
    course: dict[str, Any],
    instructors_by_id: dict[int, dict[str, Any]] | None = None,
) -> dict[str, Any]:
    prepared = dict(course)

    status = str(
        prepared.get("status") or "draft"
    ).lower()

    prepared["status"] = status
    prepared["status_label"] = STATUS_LABELS.get(
        status,
        status,
    )
    prepared["is_open"] = status == "open"

    price = prepared.get("price")

    try:
        prepared["formatted_price"] = (
            f"{int(price):,}"
        )
    except (TypeError, ValueError):
        prepared["formatted_price"] = "—"

    for field_name in ("start_date", "end_date"):
        value = str(
            prepared.get(field_name) or ""
        ).strip()

        prepared[f"{field_name}_value"] = (
            value[:10] if value else ""
        )

    instructor_id = prepared.get("instructor_id")
    instructor = None

    try:
        normalized_instructor_id = int(instructor_id)
    except (TypeError, ValueError):
        normalized_instructor_id = 0

    if instructors_by_id:
        instructor = instructors_by_id.get(
            normalized_instructor_id
        )

    prepared["instructor"] = instructor
    prepared["instructor_name"] = (
        instructor_full_name(instructor)
        if isinstance(instructor, dict)
        else f"استاد شماره {instructor_id or '—'}"
    )

    return prepared


def add_api_error_to_form(
    form: CourseForm,
    error: GoAPIError,
) -> None:
    form.add_api_errors(error.payload)

    if not form.errors:
        form.add_error(None, str(error))


class CourseBaseView(
    EducationManagementRequiredMixin,
    View,
):
    active_page = "course_management"

    def base_context(
        self,
        **kwargs: Any,
    ) -> dict[str, Any]:
        context = {
            "active_page": self.active_page,
            "api_available": True,
            "api_base_url": settings.GO_API_BASE_URL,
        }

        context.update(kwargs)
        return context

    def get_instructors(
        self,
    ) -> list[dict[str, Any]]:
        response = go_api.get(INSTRUCTORS_API_PATH)
        instructors = get_api_data(response, default=[])

        if not isinstance(instructors, list):
            return []

        return [
            instructor
            for instructor in instructors
            if isinstance(instructor, dict)
        ]

    def get_course(
        self,
        course_id: int,
        instructors: list[dict[str, Any]] | None = None,
    ) -> dict[str, Any]:
        try:
            response = go_api.get(
                f"{COURSES_API_PATH}/{course_id}"
            )
        except GoAPIError as error:
            if error.status_code == 404:
                raise Http404(
                    "دوره موردنظر پیدا نشد."
                ) from error
            raise

        course = get_api_data(response)

        if not isinstance(course, dict):
            raise Http404(
                "اطلاعات دوره موردنظر موجود نیست."
            )

        instructors_by_id = {
            int(instructor["id"]): instructor
            for instructor in instructors or []
            if instructor.get("id")
        }

        return prepare_course(
            course,
            instructors_by_id,
        )


class CourseListView(CourseBaseView):
    template_name = (
        "courses_management/course_list.html"
    )

    def get(
        self,
        request: HttpRequest,
    ) -> HttpResponse:
        courses: list[dict[str, Any]] = []
        instructors: list[dict[str, Any]] = []
        api_errors: list[str] = []

        try:
            instructors = self.get_instructors()
        except GoAPIError as error:
            api_errors.append(str(error))

        instructors_by_id = {
            int(instructor["id"]): instructor
            for instructor in instructors
            if instructor.get("id")
        }

        try:
            response = go_api.get(COURSES_API_PATH)
            raw_courses = get_api_data(
                response,
                default=[],
            )

            if isinstance(raw_courses, list):
                courses = [
                    prepare_course(
                        course,
                        instructors_by_id,
                    )
                    for course in raw_courses
                    if isinstance(course, dict)
                ]

        except GoAPIError as error:
            api_errors.append(str(error))

        search_form = CourseSearchForm(
            request.GET or None,
            instructors=instructors,
        )

        query = ""
        selected_status = ""
        selected_instructor = ""

        if search_form.is_valid():
            query = search_form.cleaned_data["q"]
            selected_status = (
                search_form.cleaned_data["status"]
            )
            selected_instructor = (
                search_form.cleaned_data["instructor"]
            )

        if query:
            normalized_query = query.casefold()

            courses = [
                course
                for course in courses
                if normalized_query
                in " ".join(
                    [
                        str(course.get("code", "")),
                        str(course.get("title", "")),
                        str(
                            course.get(
                                "description",
                                "",
                            )
                        ),
                        str(
                            course.get(
                                "instructor_name",
                                "",
                            )
                        ),
                    ]
                ).casefold()
            ]

        if selected_status:
            courses = [
                course
                for course in courses
                if course.get("status")
                == selected_status
            ]

        if selected_instructor:
            courses = [
                course
                for course in courses
                if str(course.get("instructor_id"))
                == selected_instructor
            ]

        courses.sort(
            key=lambda item: (
                str(
                    item.get("start_date_value", "")
                ),
                str(item.get("title", "")).casefold(),
            ),
            reverse=True,
        )

        status_counts = {
            status: sum(
                1
                for course in courses
                if course.get("status") == status
            )
            for status in STATUS_LABELS
        }

        paginator = Paginator(
            courses,
            COURSES_PER_PAGE,
        )
        page_obj = paginator.get_page(
            request.GET.get("page")
        )

        return render(
            request,
            self.template_name,
            self.base_context(
                page_title="مدیریت دوره‌ها",
                courses=page_obj.object_list,
                instructors=instructors,
                page_obj=page_obj,
                paginator=paginator,
                search_form=search_form,
                current_query=query,
                current_status=selected_status,
                current_instructor=selected_instructor,
                total_courses=len(courses),
                status_counts=status_counts,
                api_error=" | ".join(api_errors),
                api_available=not bool(api_errors),
            ),
        )


class CourseDetailView(CourseBaseView):
    template_name = (
        "courses_management/course_detail.html"
    )

    def get(
        self,
        request: HttpRequest,
        course_id: int,
    ) -> HttpResponse:
        try:
            instructors = self.get_instructors()
            course = self.get_course(
                course_id,
                instructors,
            )
        except GoAPIError as error:
            return render(
                request,
                self.template_name,
                self.base_context(
                    page_title="جزئیات دوره",
                    course=None,
                    api_error=str(error),
                    api_available=False,
                ),
                status=503,
            )

        return render(
            request,
            self.template_name,
            self.base_context(
                page_title=course.get(
                    "title",
                    "جزئیات دوره",
                ),
                course=course,
                api_error="",
            ),
        )


class CourseCreateView(CourseBaseView):
    template_name = (
        "courses_management/course_form.html"
    )

    def get(
        self,
        request: HttpRequest,
    ) -> HttpResponse:
        try:
            instructors = self.get_instructors()
            api_error = ""
        except GoAPIError as error:
            instructors = []
            api_error = str(error)

        form = CourseForm(
            instructors=instructors,
        )

        return render(
            request,
            self.template_name,
            self.base_context(
                page_title="ثبت دوره جدید",
                form=form,
                form_mode="create",
                course=None,
                submit_label="ثبت دوره",
                api_error=api_error,
                api_available=not bool(api_error),
            ),
        )

    def post(
        self,
        request: HttpRequest,
    ) -> HttpResponse:
        try:
            instructors = self.get_instructors()
            api_error = ""
        except GoAPIError as error:
            instructors = []
            api_error = str(error)

        form = CourseForm(
            request.POST,
            instructors=instructors,
        )

        if form.is_valid():
            try:
                response = go_api.post(
                    COURSES_API_PATH,
                    json=form.api_payload(),
                )
                course = get_api_data(response)

                if not isinstance(course, dict):
                    raise GoAPIError(
                        "اطلاعات دوره ثبت‌شده "
                        "از سرویس دریافت نشد."
                    )

                course_id = course.get("id")

                if not course_id:
                    raise GoAPIError(
                        "شناسه دوره ثبت‌شده "
                        "از سرویس دریافت نشد."
                    )

                messages.success(
                    request,
                    "دوره جدید با موفقیت ثبت شد.",
                )

                return redirect(
                    "course_management:detail",
                    course_id=course_id,
                )

            except GoAPIError as error:
                add_api_error_to_form(form, error)

        return render(
            request,
            self.template_name,
            self.base_context(
                page_title="ثبت دوره جدید",
                form=form,
                form_mode="create",
                course=None,
                submit_label="ثبت دوره",
                api_error=api_error,
                api_available=not bool(api_error),
            ),
            status=400,
        )


class CourseUpdateView(CourseBaseView):
    template_name = (
        "courses_management/course_form.html"
    )

    def get(
        self,
        request: HttpRequest,
        course_id: int,
    ) -> HttpResponse:
        try:
            instructors = self.get_instructors()
            course = self.get_course(
                course_id,
                instructors,
            )
        except GoAPIError as error:
            messages.error(request, str(error))
            return redirect("course_management:list")

        form = CourseForm(
            instructors=instructors,
            include_status=True,
            initial={
                "instructor_id": course.get(
                    "instructor_id"
                ),
                "code": course.get("code", ""),
                "title": course.get("title", ""),
                "description": course.get(
                    "description",
                    "",
                ),
                "price": course.get("price", 0),
                "capacity": course.get("capacity", 1),
                "duration_hours": course.get(
                    "duration_hours",
                    1,
                ),
                "start_date": course.get(
                    "start_date_value",
                    "",
                ),
                "end_date": course.get(
                    "end_date_value",
                    "",
                ),
                "status": course.get(
                    "status",
                    "draft",
                ),
            },
        )

        return render(
            request,
            self.template_name,
            self.base_context(
                page_title="ویرایش دوره",
                form=form,
                form_mode="update",
                course=course,
                submit_label="ذخیره تغییرات",
                api_error="",
            ),
        )

    def post(
        self,
        request: HttpRequest,
        course_id: int,
    ) -> HttpResponse:
        try:
            instructors = self.get_instructors()
            course = self.get_course(
                course_id,
                instructors,
            )
        except GoAPIError as error:
            messages.error(request, str(error))
            return redirect("course_management:list")

        form = CourseForm(
            request.POST,
            instructors=instructors,
            include_status=True,
        )

        if form.is_valid():
            try:
                go_api.put(
                    f"{COURSES_API_PATH}/{course_id}",
                    json=form.api_payload(),
                )

                messages.success(
                    request,
                    "اطلاعات دوره با موفقیت ویرایش شد.",
                )

                return redirect(
                    "course_management:detail",
                    course_id=course_id,
                )

            except GoAPIError as error:
                add_api_error_to_form(form, error)

        return render(
            request,
            self.template_name,
            self.base_context(
                page_title="ویرایش دوره",
                form=form,
                form_mode="update",
                course=course,
                submit_label="ذخیره تغییرات",
                api_error="",
            ),
            status=400,
        )


class CourseDeleteView(CourseBaseView):
    template_name = (
        "courses_management/course_confirm_delete.html"
    )

    def get(
        self,
        request: HttpRequest,
        course_id: int,
    ) -> HttpResponse:
        try:
            instructors = self.get_instructors()
            course = self.get_course(
                course_id,
                instructors,
            )
        except GoAPIError as error:
            messages.error(request, str(error))
            return redirect("course_management:list")

        return render(
            request,
            self.template_name,
            self.base_context(
                page_title="حذف دوره",
                course=course,
            ),
        )

    def post(
        self,
        request: HttpRequest,
        course_id: int,
    ) -> HttpResponse:
        try:
            instructors = self.get_instructors()
            course = self.get_course(
                course_id,
                instructors,
            )

            go_api.delete(
                f"{COURSES_API_PATH}/{course_id}"
            )

        except GoAPIError as error:
            if error.status_code == 404:
                raise Http404(
                    "دوره موردنظر پیدا نشد."
                ) from error

            messages.error(
                request,
                f"حذف دوره انجام نشد: {error}",
            )

            return redirect(
                "course_management:detail",
                course_id=course_id,
            )

        messages.success(
            request,
            (
                f"دوره «{course.get('title', '')}» "
                "با موفقیت حذف شد."
            ),
        )

        return redirect("course_management:list")