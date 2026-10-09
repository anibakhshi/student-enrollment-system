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

from .forms import InstructorForm, InstructorSearchForm


INSTRUCTORS_API_PATH = "/api/v1/instructors"
INSTRUCTORS_PER_PAGE = 12


def get_api_data(
    response: dict[str, Any],
    default: Any = None,
) -> Any:
    if not isinstance(response, dict):
        return default

    data = response.get("data")
    return default if data is None else data


def prepare_instructor(
    instructor: dict[str, Any],
) -> dict[str, Any]:
    prepared = dict(instructor)

    first_name = str(
        prepared.get("first_name") or ""
    ).strip()
    last_name = str(
        prepared.get("last_name") or ""
    ).strip()

    full_name = f"{first_name} {last_name}".strip()

    prepared["full_name"] = (
        full_name or "استاد بدون نام"
    )

    initials = "".join(
        part[0]
        for part in (first_name, last_name)
        if part
    )

    prepared["initials"] = initials[:2] or "ا"

    status = str(
        prepared.get("status") or "inactive"
    ).lower()

    prepared["status"] = status
    prepared["is_active"] = status == "active"
    prepared["status_label"] = (
        "فعال" if prepared["is_active"] else "غیرفعال"
    )

    return prepared


def add_api_error_to_form(
    form: InstructorForm,
    error: GoAPIError,
) -> None:
    payload = error.payload

    if isinstance(payload, dict):
        errors = payload.get("errors")

        if isinstance(errors, dict):
            payload = {
                "errors": errors,
                "message": payload.get(
                    "message",
                    str(error),
                ),
            }

    form.add_api_errors(payload)

    if not form.errors:
        form.add_error(None, str(error))


class InstructorBaseView(
    EducationManagementRequiredMixin,
    View,
):
    active_page = "instructor_management"

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

    def get_instructor(
        self,
        instructor_id: int,
    ) -> dict[str, Any]:
        try:
            response = go_api.get(
                f"{INSTRUCTORS_API_PATH}/{instructor_id}"
            )
        except GoAPIError as error:
            if error.status_code == 404:
                raise Http404(
                    "استاد موردنظر پیدا نشد."
                ) from error
            raise

        instructor = get_api_data(response)

        if not isinstance(instructor, dict):
            raise Http404(
                "اطلاعات استاد موردنظر موجود نیست."
            )

        return prepare_instructor(instructor)


class InstructorListView(InstructorBaseView):
    template_name = (
        "instructors_management/instructor_list.html"
    )

    def get(
        self,
        request: HttpRequest,
    ) -> HttpResponse:
        search_form = InstructorSearchForm(
            request.GET or None
        )

        instructors: list[dict[str, Any]] = []
        api_error = ""

        try:
            response = go_api.get(
                INSTRUCTORS_API_PATH
            )
            raw_instructors = get_api_data(
                response,
                default=[],
            )

            if isinstance(raw_instructors, list):
                instructors = [
                    prepare_instructor(instructor)
                    for instructor in raw_instructors
                    if isinstance(instructor, dict)
                ]

        except GoAPIError as error:
            api_error = str(error)

        query = ""
        selected_status = ""

        if search_form.is_valid():
            query = search_form.cleaned_data["q"]
            selected_status = (
                search_form.cleaned_data["status"]
            )

        if query:
            normalized_query = query.casefold()

            instructors = [
                instructor
                for instructor in instructors
                if normalized_query
                in " ".join(
                    [
                        str(instructor.get("first_name", "")),
                        str(instructor.get("last_name", "")),
                        str(instructor.get("email", "")),
                        str(instructor.get("phone", "")),
                        str(instructor.get("expertise", "")),
                        str(instructor.get("bio", "")),
                    ]
                ).casefold()
            ]

        if selected_status:
            instructors = [
                instructor
                for instructor in instructors
                if instructor.get("status")
                == selected_status
            ]

        instructors.sort(
            key=lambda item: (
                not bool(item.get("is_active")),
                str(item.get("last_name", "")).casefold(),
                str(item.get("first_name", "")).casefold(),
            )
        )

        active_count = sum(
            1
            for instructor in instructors
            if instructor.get("is_active")
        )

        paginator = Paginator(
            instructors,
            INSTRUCTORS_PER_PAGE,
        )
        page_obj = paginator.get_page(
            request.GET.get("page")
        )

        return render(
            request,
            self.template_name,
            self.base_context(
                page_title="مدیریت استادان",
                instructors=page_obj.object_list,
                page_obj=page_obj,
                paginator=paginator,
                search_form=search_form,
                current_query=query,
                current_status=selected_status,
                total_instructors=len(instructors),
                active_instructors=active_count,
                inactive_instructors=(
                    len(instructors) - active_count
                ),
                api_error=api_error,
                api_available=not bool(api_error),
            ),
        )


class InstructorDetailView(InstructorBaseView):
    template_name = (
        "instructors_management/instructor_detail.html"
    )

    def get(
        self,
        request: HttpRequest,
        instructor_id: int,
    ) -> HttpResponse:
        try:
            instructor = self.get_instructor(
                instructor_id
            )
        except GoAPIError as error:
            return render(
                request,
                self.template_name,
                self.base_context(
                    page_title="جزئیات استاد",
                    instructor=None,
                    api_error=str(error),
                    api_available=False,
                ),
                status=503,
            )

        return render(
            request,
            self.template_name,
            self.base_context(
                page_title=instructor["full_name"],
                instructor=instructor,
                api_error="",
            ),
        )


class InstructorCreateView(InstructorBaseView):
    template_name = (
        "instructors_management/instructor_form.html"
    )

    def get(
        self,
        request: HttpRequest,
    ) -> HttpResponse:
        form = InstructorForm()

        return render(
            request,
            self.template_name,
            self.base_context(
                page_title="ثبت استاد جدید",
                form=form,
                form_mode="create",
                instructor=None,
                submit_label="ثبت استاد",
            ),
        )

    def post(
        self,
        request: HttpRequest,
    ) -> HttpResponse:
        form = InstructorForm(request.POST)

        if form.is_valid():
            try:
                response = go_api.post(
                    INSTRUCTORS_API_PATH,
                    json=form.api_payload(),
                )

                instructor = get_api_data(response)

                if not isinstance(instructor, dict):
                    raise GoAPIError(
                        "اطلاعات استاد ثبت‌شده "
                        "از سرویس دریافت نشد."
                    )

                instructor_id = instructor.get("id")

                if not instructor_id:
                    raise GoAPIError(
                        "شناسه استاد ثبت‌شده "
                        "از سرویس دریافت نشد."
                    )

                messages.success(
                    request,
                    "استاد جدید با موفقیت ثبت شد.",
                )

                return redirect(
                    "instructor_management:detail",
                    instructor_id=instructor_id,
                )

            except GoAPIError as error:
                add_api_error_to_form(form, error)

        return render(
            request,
            self.template_name,
            self.base_context(
                page_title="ثبت استاد جدید",
                form=form,
                form_mode="create",
                instructor=None,
                submit_label="ثبت استاد",
            ),
            status=400,
        )


class InstructorUpdateView(InstructorBaseView):
    template_name = (
        "instructors_management/instructor_form.html"
    )

    def get(
        self,
        request: HttpRequest,
        instructor_id: int,
    ) -> HttpResponse:
        try:
            instructor = self.get_instructor(
                instructor_id
            )
        except GoAPIError as error:
            messages.error(request, str(error))
            return redirect(
                "instructor_management:list"
            )

        form = InstructorForm(
            include_status=True,
            initial={
                "first_name": instructor.get(
                    "first_name",
                    "",
                ),
                "last_name": instructor.get(
                    "last_name",
                    "",
                ),
                "email": instructor.get("email", ""),
                "phone": instructor.get("phone", ""),
                "expertise": instructor.get(
                    "expertise",
                    "",
                ),
                "bio": instructor.get("bio", ""),
                "status": instructor.get(
                    "status",
                    "active",
                ),
            },
        )

        return render(
            request,
            self.template_name,
            self.base_context(
                page_title="ویرایش استاد",
                form=form,
                form_mode="update",
                instructor=instructor,
                submit_label="ذخیره تغییرات",
            ),
        )

    def post(
        self,
        request: HttpRequest,
        instructor_id: int,
    ) -> HttpResponse:
        try:
            instructor = self.get_instructor(
                instructor_id
            )
        except GoAPIError as error:
            messages.error(request, str(error))
            return redirect(
                "instructor_management:list"
            )

        form = InstructorForm(
            request.POST,
            include_status=True,
        )

        if form.is_valid():
            try:
                go_api.put(
                    f"{INSTRUCTORS_API_PATH}/{instructor_id}",
                    json=form.api_payload(),
                )

                messages.success(
                    request,
                    "اطلاعات استاد با موفقیت ویرایش شد.",
                )

                return redirect(
                    "instructor_management:detail",
                    instructor_id=instructor_id,
                )

            except GoAPIError as error:
                add_api_error_to_form(form, error)

        return render(
            request,
            self.template_name,
            self.base_context(
                page_title="ویرایش استاد",
                form=form,
                form_mode="update",
                instructor=instructor,
                submit_label="ذخیره تغییرات",
            ),
            status=400,
        )


class InstructorDeleteView(InstructorBaseView):
    template_name = (
        "instructors_management/"
        "instructor_confirm_delete.html"
    )

    def get(
        self,
        request: HttpRequest,
        instructor_id: int,
    ) -> HttpResponse:
        try:
            instructor = self.get_instructor(
                instructor_id
            )
        except GoAPIError as error:
            messages.error(request, str(error))
            return redirect(
                "instructor_management:list"
            )

        return render(
            request,
            self.template_name,
            self.base_context(
                page_title="حذف استاد",
                instructor=instructor,
            ),
        )

    def post(
        self,
        request: HttpRequest,
        instructor_id: int,
    ) -> HttpResponse:
        try:
            instructor = self.get_instructor(
                instructor_id
            )

            go_api.delete(
                f"{INSTRUCTORS_API_PATH}/{instructor_id}"
            )

        except GoAPIError as error:
            if error.status_code == 404:
                raise Http404(
                    "استاد موردنظر پیدا نشد."
                ) from error

            messages.error(
                request,
                f"حذف استاد انجام نشد: {error}",
            )

            return redirect(
                "instructor_management:detail",
                instructor_id=instructor_id,
            )

        messages.success(
            request,
            (
                f"استاد «{instructor['full_name']}» "
                "با موفقیت حذف شد."
            ),
        )

        return redirect(
            "instructor_management:list"
        )