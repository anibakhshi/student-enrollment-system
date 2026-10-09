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

from .forms import (
    EnrollmentCreateForm,
    EnrollmentSearchForm,
    EnrollmentUpdateForm,
    course_display_name,
    student_full_name,
)


ENROLLMENTS_API_PATH = "/api/v1/enrollments"
STUDENTS_API_PATH = "/api/v1/students"
COURSES_API_PATH = "/api/v1/courses"
ENROLLMENTS_PER_PAGE = 12

STATUS_LABELS = {
    "pending": "در انتظار تأیید",
    "confirmed": "تأییدشده",
    "cancelled": "لغوشده",
    "completed": "تکمیل‌شده",
}


def get_api_data(
    response: dict[str, Any],
    default: Any = None,
) -> Any:
    if not isinstance(response, dict):
        return default

    data = response.get("data")
    return default if data is None else data


def build_lookup(
    items: list[dict[str, Any]],
) -> dict[int, dict[str, Any]]:
    lookup: dict[int, dict[str, Any]] = {}

    for item in items:
        item_id = item.get("id")

        try:
            normalized_id = int(item_id)
        except (TypeError, ValueError):
            continue

        lookup[normalized_id] = item

    return lookup


def prepare_enrollment(
    enrollment: dict[str, Any],
    students_by_id: dict[int, dict[str, Any]] | None = None,
    courses_by_id: dict[int, dict[str, Any]] | None = None,
) -> dict[str, Any]:
    prepared = dict(enrollment)

    status = str(
        prepared.get("status") or "pending"
    ).strip().lower()

    prepared["status"] = status
    prepared["status_label"] = STATUS_LABELS.get(
        status,
        status,
    )
    prepared["is_pending"] = status == "pending"
    prepared["is_confirmed"] = status == "confirmed"
    prepared["is_cancelled"] = status == "cancelled"
    prepared["is_completed"] = status == "completed"
    prepared["is_terminal"] = status in {
        "cancelled",
        "completed",
    }

    student_id = prepared.get("student_id")
    course_id = prepared.get("course_id")

    try:
        normalized_student_id = int(student_id)
    except (TypeError, ValueError):
        normalized_student_id = 0

    try:
        normalized_course_id = int(course_id)
    except (TypeError, ValueError):
        normalized_course_id = 0

    student = (
        students_by_id.get(normalized_student_id)
        if students_by_id
        else None
    )

    course = (
        courses_by_id.get(normalized_course_id)
        if courses_by_id
        else None
    )

    prepared["student"] = student
    prepared["course"] = course

    prepared["student_name"] = (
        student_full_name(student)
        if isinstance(student, dict)
        else f"دانشجوی شماره {student_id or '—'}"
    )

    prepared["course_name"] = (
        course_display_name(course)
        if isinstance(course, dict)
        else f"دوره شماره {course_id or '—'}"
    )

    for field_name in (
        "enrolled_at",
        "confirmed_at",
        "cancelled_at",
        "completed_at",
        "created_at",
        "updated_at",
    ):
        value = str(
            prepared.get(field_name) or ""
        ).strip()

        prepared[f"{field_name}_value"] = value
        prepared[f"{field_name}_date"] = (
            value[:10] if value else ""
        )

    return prepared


def add_api_error_to_form(
    form: EnrollmentCreateForm | EnrollmentUpdateForm,
    error: GoAPIError,
) -> None:
    form.add_api_errors(error.payload)

    if not form.errors:
        form.add_error(None, str(error))


class EnrollmentBaseView(
    EducationManagementRequiredMixin,
    View,
):
    active_page = "enrollment_management"

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

    def get_students(self) -> list[dict[str, Any]]:
        response = go_api.get(STUDENTS_API_PATH)
        students = get_api_data(response, default=[])

        if not isinstance(students, list):
            return []

        return [
            student
            for student in students
            if isinstance(student, dict)
        ]

    def get_courses(self) -> list[dict[str, Any]]:
        response = go_api.get(COURSES_API_PATH)
        courses = get_api_data(response, default=[])

        if not isinstance(courses, list):
            return []

        return [
            course
            for course in courses
            if isinstance(course, dict)
        ]

    def get_enrollment_details(
        self,
        enrollment_id: int,
    ) -> dict[str, Any]:
        try:
            response = go_api.get(
                f"{ENROLLMENTS_API_PATH}/{enrollment_id}"
            )
        except GoAPIError as error:
            if error.status_code == 404:
                raise Http404(
                    "ثبت‌نام موردنظر پیدا نشد."
                ) from error

            raise

        details = get_api_data(response)

        if not isinstance(details, dict):
            raise Http404(
                "اطلاعات ثبت‌نام موردنظر موجود نیست."
            )

        enrollment = details.get("enrollment")
        student = details.get("student")
        course = details.get("course")

        if not isinstance(enrollment, dict):
            raise Http404(
                "اطلاعات ثبت‌نام موردنظر معتبر نیست."
            )

        students_by_id: dict[int, dict[str, Any]] = {}
        courses_by_id: dict[int, dict[str, Any]] = {}

        if isinstance(student, dict) and student.get("id"):
            students_by_id[int(student["id"])] = student

        if isinstance(course, dict) and course.get("id"):
            courses_by_id[int(course["id"])] = course

        prepared = prepare_enrollment(
            enrollment,
            students_by_id,
            courses_by_id,
        )

        prepared["student"] = (
            student if isinstance(student, dict) else None
        )
        prepared["course"] = (
            course if isinstance(course, dict) else None
        )

        return prepared


class EnrollmentListView(EnrollmentBaseView):
    template_name = (
        "enrollments_management/enrollment_list.html"
    )

    def get(
        self,
        request: HttpRequest,
    ) -> HttpResponse:
        students: list[dict[str, Any]] = []
        courses: list[dict[str, Any]] = []
        enrollments: list[dict[str, Any]] = []
        api_errors: list[str] = []

        try:
            students = self.get_students()
        except GoAPIError as error:
            api_errors.append(str(error))

        try:
            courses = self.get_courses()
        except GoAPIError as error:
            api_errors.append(str(error))

        students_by_id = build_lookup(students)
        courses_by_id = build_lookup(courses)

        try:
            response = go_api.get(ENROLLMENTS_API_PATH)
            raw_enrollments = get_api_data(
                response,
                default=[],
            )

            if isinstance(raw_enrollments, list):
                enrollments = [
                    prepare_enrollment(
                        enrollment,
                        students_by_id,
                        courses_by_id,
                    )
                    for enrollment in raw_enrollments
                    if isinstance(enrollment, dict)
                ]

        except GoAPIError as error:
            api_errors.append(str(error))

        search_form = EnrollmentSearchForm(
            request.GET or None,
            students=students,
            courses=courses,
        )

        query = ""
        selected_status = ""
        selected_student = ""
        selected_course = ""

        if search_form.is_valid():
            query = search_form.cleaned_data["q"]
            selected_status = (
                search_form.cleaned_data["status"]
            )
            selected_student = (
                search_form.cleaned_data["student"]
            )
            selected_course = (
                search_form.cleaned_data["course"]
            )

        if query:
            normalized_query = query.casefold()

            enrollments = [
                enrollment
                for enrollment in enrollments
                if normalized_query
                in " ".join(
                    [
                        str(
                            enrollment.get(
                                "student_name",
                                "",
                            )
                        ),
                        str(
                            enrollment.get(
                                "course_name",
                                "",
                            )
                        ),
                        str(
                            enrollment.get(
                                "notes",
                                "",
                            )
                        ),
                    ]
                ).casefold()
            ]

        if selected_status:
            enrollments = [
                enrollment
                for enrollment in enrollments
                if enrollment.get("status")
                == selected_status
            ]

        if selected_student:
            enrollments = [
                enrollment
                for enrollment in enrollments
                if str(enrollment.get("student_id"))
                == selected_student
            ]

        if selected_course:
            enrollments = [
                enrollment
                for enrollment in enrollments
                if str(enrollment.get("course_id"))
                == selected_course
            ]

        enrollments.sort(
            key=lambda item: (
                str(item.get("enrolled_at_value", "")),
                int(item.get("id") or 0),
            ),
            reverse=True,
        )

        status_counts = {
            status: sum(
                1
                for enrollment in enrollments
                if enrollment.get("status") == status
            )
            for status in STATUS_LABELS
        }

        paginator = Paginator(
            enrollments,
            ENROLLMENTS_PER_PAGE,
        )

        page_obj = paginator.get_page(
            request.GET.get("page")
        )

        return render(
            request,
            self.template_name,
            self.base_context(
                page_title="مدیریت ثبت‌نام‌ها",
                enrollments=page_obj.object_list,
                students=students,
                courses=courses,
                page_obj=page_obj,
                paginator=paginator,
                search_form=search_form,
                current_query=query,
                current_status=selected_status,
                current_student=selected_student,
                current_course=selected_course,
                total_enrollments=len(enrollments),
                status_counts=status_counts,
                api_error=" | ".join(api_errors),
                api_available=not bool(api_errors),
            ),
        )


class EnrollmentDetailView(EnrollmentBaseView):
    template_name = (
        "enrollments_management/enrollment_detail.html"
    )

    def get(
        self,
        request: HttpRequest,
        enrollment_id: int,
    ) -> HttpResponse:
        try:
            enrollment = self.get_enrollment_details(
                enrollment_id
            )
        except GoAPIError as error:
            return render(
                request,
                self.template_name,
                self.base_context(
                    page_title="جزئیات ثبت‌نام",
                    enrollment=None,
                    api_error=str(error),
                    api_available=False,
                ),
                status=503,
            )

        return render(
            request,
            self.template_name,
            self.base_context(
                page_title=(
                    f"ثبت‌نام شماره {enrollment_id}"
                ),
                enrollment=enrollment,
                api_error="",
            ),
        )


class EnrollmentCreateView(EnrollmentBaseView):
    template_name = (
        "enrollments_management/enrollment_form.html"
    )

    def load_form_dependencies(
        self,
    ) -> tuple[
        list[dict[str, Any]],
        list[dict[str, Any]],
        str,
    ]:
        students: list[dict[str, Any]] = []
        courses: list[dict[str, Any]] = []
        errors: list[str] = []

        try:
            students = self.get_students()
        except GoAPIError as error:
            errors.append(str(error))

        try:
            courses = self.get_courses()
        except GoAPIError as error:
            errors.append(str(error))

        return students, courses, " | ".join(errors)

    def get(
        self,
        request: HttpRequest,
    ) -> HttpResponse:
        students, courses, api_error = (
            self.load_form_dependencies()
        )

        form = EnrollmentCreateForm(
            students=students,
            courses=courses,
        )

        return render(
            request,
            self.template_name,
            self.base_context(
                page_title="ثبت‌نام جدید",
                form=form,
                form_mode="create",
                enrollment=None,
                api_error=api_error,
                api_available=not bool(api_error),
            ),
        )

    def post(
        self,
        request: HttpRequest,
    ) -> HttpResponse:
        students, courses, api_error = (
            self.load_form_dependencies()
        )

        form = EnrollmentCreateForm(
            request.POST,
            students=students,
            courses=courses,
        )

        if form.is_valid():
            try:
                response = go_api.post(
                    ENROLLMENTS_API_PATH,
                    json=form.api_payload(),
                )

                enrollment = get_api_data(response)

                if not isinstance(enrollment, dict):
                    raise GoAPIError(
                        "اطلاعات ثبت‌نام ایجادشده "
                        "از سرویس دریافت نشد."
                    )

                enrollment_id = enrollment.get("id")

                if not enrollment_id:
                    raise GoAPIError(
                        "شناسه ثبت‌نام ایجادشده "
                        "از سرویس دریافت نشد."
                    )

                messages.success(
                    request,
                    "ثبت‌نام دانشجو با موفقیت انجام شد.",
                )

                return redirect(
                    "enrollment_management:detail",
                    enrollment_id=enrollment_id,
                )

            except GoAPIError as error:
                add_api_error_to_form(form, error)

        return render(
            request,
            self.template_name,
            self.base_context(
                page_title="ثبت‌نام جدید",
                form=form,
                form_mode="create",
                enrollment=None,
                api_error=api_error,
                api_available=not bool(api_error),
            ),
            status=400,
        )


class EnrollmentUpdateView(EnrollmentBaseView):
    template_name = (
        "enrollments_management/enrollment_form.html"
    )

    def get(
        self,
        request: HttpRequest,
        enrollment_id: int,
    ) -> HttpResponse:
        try:
            enrollment = self.get_enrollment_details(
                enrollment_id
            )
        except GoAPIError as error:
            messages.error(request, str(error))
            return redirect("enrollment_management:list")

        form = EnrollmentUpdateForm(
            current_status=enrollment["status"],
            initial={
                "status": enrollment["status"],
                "notes": enrollment.get("notes", ""),
            },
        )

        return render(
            request,
            self.template_name,
            self.base_context(
                page_title="ویرایش ثبت‌نام",
                form=form,
                form_mode="update",
                enrollment=enrollment,
                api_error="",
            ),
        )

    def post(
        self,
        request: HttpRequest,
        enrollment_id: int,
    ) -> HttpResponse:
        try:
            enrollment = self.get_enrollment_details(
                enrollment_id
            )
        except GoAPIError as error:
            messages.error(request, str(error))
            return redirect("enrollment_management:list")

        if enrollment["is_terminal"]:
            messages.error(
                request,
                "ثبت‌نام نهایی‌شده قابل ویرایش نیست.",
            )

            return redirect(
                "enrollment_management:detail",
                enrollment_id=enrollment_id,
            )

        form = EnrollmentUpdateForm(
            request.POST,
            current_status=enrollment["status"],
        )

        if form.is_valid():
            try:
                go_api.put(
                    (
                        f"{ENROLLMENTS_API_PATH}/"
                        f"{enrollment_id}"
                    ),
                    json=form.api_payload(),
                )

                messages.success(
                    request,
                    "اطلاعات ثبت‌نام با موفقیت ویرایش شد.",
                )

                return redirect(
                    "enrollment_management:detail",
                    enrollment_id=enrollment_id,
                )

            except GoAPIError as error:
                add_api_error_to_form(form, error)

        return render(
            request,
            self.template_name,
            self.base_context(
                page_title="ویرایش ثبت‌نام",
                form=form,
                form_mode="update",
                enrollment=enrollment,
                api_error="",
            ),
            status=400,
        )


class EnrollmentDeleteView(EnrollmentBaseView):
    template_name = (
        "enrollments_management/"
        "enrollment_confirm_delete.html"
    )

    def get(
        self,
        request: HttpRequest,
        enrollment_id: int,
    ) -> HttpResponse:
        try:
            enrollment = self.get_enrollment_details(
                enrollment_id
            )
        except GoAPIError as error:
            messages.error(request, str(error))
            return redirect("enrollment_management:list")

        return render(
            request,
            self.template_name,
            self.base_context(
                page_title="حذف ثبت‌نام",
                enrollment=enrollment,
            ),
        )

    def post(
        self,
        request: HttpRequest,
        enrollment_id: int,
    ) -> HttpResponse:
        try:
            self.get_enrollment_details(enrollment_id)

            go_api.delete(
                f"{ENROLLMENTS_API_PATH}/{enrollment_id}"
            )

        except GoAPIError as error:
            if error.status_code == 404:
                raise Http404(
                    "ثبت‌نام موردنظر پیدا نشد."
                ) from error

            messages.error(
                request,
                f"حذف ثبت‌نام انجام نشد: {error}",
            )

            return redirect(
                "enrollment_management:detail",
                enrollment_id=enrollment_id,
            )

        messages.success(
            request,
            "ثبت‌نام با موفقیت حذف شد.",
        )

        return redirect("enrollment_management:list")