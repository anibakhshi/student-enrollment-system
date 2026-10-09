from __future__ import annotations

from typing import Any

from django.conf import settings
from django.contrib import messages
from django.core.paginator import Paginator
from django.http import Http404, HttpRequest, HttpResponse
from django.shortcuts import redirect, render
from django.urls import reverse
from django.views import View

from accounts.mixins import EducationManagementRequiredMixin
from dashboard.api_client import GoAPIError, go_api

from .forms import StudentForm, StudentSearchForm


STUDENTS_API_PATH = "/api/v1/students"
STUDENTS_PER_PAGE = 12


def get_api_data(
    response: dict[str, Any],
    default: Any = None,
) -> Any:
    """
    بخش data پاسخ استاندارد Go API را برمی‌گرداند.
    """

    if not isinstance(response, dict):
        return default

    data = response.get("data")
    return default if data is None else data


def build_profile_image_url(
    student: dict[str, Any],
) -> str:
    """
    آدرس داخلی تصویر API را به آدرس قابل‌دسترسی مرورگر تبدیل می‌کند.
    """

    image_path = str(
        student.get("profile_image_path") or ""
    ).strip()

    if not image_path:
        return ""

    if image_path.startswith(
        (
            "http://",
            "https://",
        )
    ):
        return image_path

    public_base_url = getattr(
        settings,
        "GO_API_PUBLIC_URL",
        settings.GO_API_BASE_URL,
    ).rstrip("/")

    return f"{public_base_url}/{image_path.lstrip('/')}"


def prepare_student(
    student: dict[str, Any],
) -> dict[str, Any]:
    """
    اطلاعات نمایشی تکمیلی دانشجو را تولید می‌کند.
    """

    prepared = dict(student)

    first_name = str(
        prepared.get("first_name") or ""
    ).strip()

    last_name = str(
        prepared.get("last_name") or ""
    ).strip()

    full_name = f"{first_name} {last_name}".strip()

    prepared["full_name"] = (
        full_name or "دانشجوی بدون نام"
    )

    initials = "".join(
        part[0]
        for part in (
            first_name,
            last_name,
        )
        if part
    )

    prepared["initials"] = initials[:2] or "د"
    prepared["profile_image_url"] = (
        build_profile_image_url(prepared)
    )

    return prepared


def upload_student_photo(
    student_id: int,
    photo: Any,
) -> None:
    """
    تصویر دانشجو را به endpoint اختصاصی Go API ارسال می‌کند.
    """

    if not photo:
        return

    photo.seek(0)

    go_api.post(
        f"{STUDENTS_API_PATH}/{student_id}/photo",
        files={
            "photo": (
                photo.name,
                photo,
                getattr(
                    photo,
                    "content_type",
                    "application/octet-stream",
                ),
            )
        },
    )


def add_api_error_to_form(
    form: StudentForm,
    error: GoAPIError,
) -> None:
    """
    خطای Go API را به فرم Django متصل می‌کند.
    """

    field_mapping = {
        "first_name": "first_name",
        "last_name": "last_name",
        "age": "age",
        "national_code": "national_code",
        "email": "email",
        "phone": "phone",
        "photo": "photo",
    }

    payload = error.payload

    if isinstance(payload, dict):
        errors = payload.get("errors")

        if isinstance(errors, dict):
            error_added = False

            for api_field, field_errors in errors.items():
                form_field = field_mapping.get(api_field)

                if form_field not in form.fields:
                    continue

                if isinstance(field_errors, list):
                    for field_error in field_errors:
                        form.add_error(
                            form_field,
                            str(field_error),
                        )
                        error_added = True

                elif field_errors:
                    form.add_error(
                        form_field,
                        str(field_errors),
                    )
                    error_added = True

            if error_added:
                return

    form.add_error(
        None,
        str(error),
    )


class StudentBaseView(
    EducationManagementRequiredMixin,
    View,
):
    """
    تنظیمات و context مشترک صفحات مدیریت دانشجویان.
    """

    active_page = "students"

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

    def get_student(
        self,
        student_id: int,
    ) -> dict[str, Any]:
        try:
            response = go_api.get(
                f"{STUDENTS_API_PATH}/{student_id}"
            )
        except GoAPIError as error:
            if error.status_code == 404:
                raise Http404(
                    "دانشجوی موردنظر پیدا نشد."
                ) from error

            raise

        student = get_api_data(response)

        if not isinstance(student, dict):
            raise Http404(
                "اطلاعات دانشجوی موردنظر موجود نیست."
            )

        return prepare_student(student)


class StudentListView(StudentBaseView):
    template_name = "students/student_list.html"

    def get(
        self,
        request: HttpRequest,
    ) -> HttpResponse:
        search_form = StudentSearchForm(
            request.GET or None
        )

        students: list[dict[str, Any]] = []
        api_error = ""

        try:
            response = go_api.get(
                STUDENTS_API_PATH
            )

            raw_students = get_api_data(
                response,
                default=[],
            )

            if isinstance(raw_students, list):
                students = [
                    prepare_student(student)
                    for student in raw_students
                    if isinstance(student, dict)
                ]

        except GoAPIError as error:
            api_error = str(error)

        query = ""

        if search_form.is_valid():
            query = search_form.cleaned_data["q"]

        if query:
            normalized_query = query.casefold()

            students = [
                student
                for student in students
                if normalized_query
                in " ".join(
                    [
                        str(
                            student.get(
                                "first_name",
                                "",
                            )
                        ),
                        str(
                            student.get(
                                "last_name",
                                "",
                            )
                        ),
                        str(
                            student.get(
                                "national_code",
                                "",
                            )
                        ),
                        str(
                            student.get(
                                "email",
                                "",
                            )
                        ),
                        str(
                            student.get(
                                "phone",
                                "",
                            )
                        ),
                    ]
                ).casefold()
            ]

        students.sort(
            key=lambda item: (
                str(
                    item.get(
                        "last_name",
                        "",
                    )
                ).casefold(),
                str(
                    item.get(
                        "first_name",
                        "",
                    )
                ).casefold(),
            )
        )

        paginator = Paginator(
            students,
            STUDENTS_PER_PAGE,
        )

        page_obj = paginator.get_page(
            request.GET.get("page")
        )

        context = self.base_context(
            page_title="مدیریت دانشجویان",
            students=page_obj.object_list,
            page_obj=page_obj,
            paginator=paginator,
            search_form=search_form,
            current_query=query,
            total_students=len(students),
            api_error=api_error,
            api_available=not bool(api_error),
        )

        return render(
            request,
            self.template_name,
            context,
        )


class StudentDetailView(StudentBaseView):
    template_name = "students/student_detail.html"

    def get(
        self,
        request: HttpRequest,
        student_id: int,
    ) -> HttpResponse:
        try:
            student = self.get_student(student_id)
        except GoAPIError as error:
            context = self.base_context(
                page_title="جزئیات دانشجو",
                student=None,
                api_error=str(error),
                api_available=False,
            )

            return render(
                request,
                self.template_name,
                context,
                status=503,
            )

        context = self.base_context(
            page_title=student["full_name"],
            student=student,
            api_error="",
        )

        return render(
            request,
            self.template_name,
            context,
        )


class StudentCreateView(StudentBaseView):
    template_name = "students/student_form.html"

    def get(
        self,
        request: HttpRequest,
    ) -> HttpResponse:
        form = StudentForm()

        return render(
            request,
            self.template_name,
            self.base_context(
                page_title="ثبت دانشجوی جدید",
                form=form,
                form_mode="create",
                student=None,
                submit_label="ثبت دانشجو",
            ),
        )

    def post(
        self,
        request: HttpRequest,
    ) -> HttpResponse:
        form = StudentForm(
            request.POST,
            request.FILES,
        )

        if form.is_valid():
            try:
                response = go_api.post(
                    STUDENTS_API_PATH,
                    json=form.api_payload(),
                )

                student = get_api_data(response)

                if not isinstance(student, dict):
                    raise GoAPIError(
                        "اطلاعات دانشجوی ثبت‌شده "
                        "از سرویس دریافت نشد."
                    )

                student_id = student.get("id")

                if not student_id:
                    raise GoAPIError(
                        "شناسه دانشجوی ثبت‌شده "
                        "از سرویس دریافت نشد."
                    )

                photo = form.cleaned_data.get("photo")

                if photo:
                    try:
                        upload_student_photo(
                            int(student_id),
                            photo,
                        )
                    except GoAPIError as photo_error:
                        messages.warning(
                            request,
                            "دانشجو ثبت شد؛ اما تصویر پروفایل "
                            f"بارگذاری نشد: {photo_error}",
                        )

                messages.success(
                    request,
                    "دانشجوی جدید با موفقیت ثبت شد.",
                )

                return redirect(
                    "students:detail",
                    student_id=student_id,
                )

            except GoAPIError as error:
                add_api_error_to_form(
                    form,
                    error,
                )

        return render(
            request,
            self.template_name,
            self.base_context(
                page_title="ثبت دانشجوی جدید",
                form=form,
                form_mode="create",
                student=None,
                submit_label="ثبت دانشجو",
            ),
            status=400,
        )


class StudentUpdateView(StudentBaseView):
    template_name = "students/student_form.html"

    def get(
        self,
        request: HttpRequest,
        student_id: int,
    ) -> HttpResponse:
        try:
            student = self.get_student(student_id)
        except GoAPIError as error:
            messages.error(
                request,
                str(error),
            )

            return redirect(
                "students:list"
            )

        form = StudentForm(
            initial={
                "first_name": student.get(
                    "first_name",
                    "",
                ),
                "last_name": student.get(
                    "last_name",
                    "",
                ),
                "age": student.get(
                    "age",
                    "",
                ),
                "national_code": student.get(
                    "national_code",
                    "",
                ),
                "email": student.get(
                    "email",
                    "",
                ),
                "phone": student.get(
                    "phone",
                    "",
                ),
            }
        )

        return render(
            request,
            self.template_name,
            self.base_context(
                page_title="ویرایش دانشجو",
                form=form,
                form_mode="update",
                student=student,
                submit_label="ذخیره تغییرات",
            ),
        )

    def post(
        self,
        request: HttpRequest,
        student_id: int,
    ) -> HttpResponse:
        try:
            student = self.get_student(student_id)
        except GoAPIError as error:
            messages.error(
                request,
                str(error),
            )

            return redirect(
                "students:list"
            )

        form = StudentForm(
            request.POST,
            request.FILES,
        )

        if form.is_valid():
            try:
                go_api.put(
                    f"{STUDENTS_API_PATH}/{student_id}",
                    json=form.api_payload(),
                )

                photo = form.cleaned_data.get("photo")

                if photo:
                    try:
                        upload_student_photo(
                            student_id,
                            photo,
                        )
                    except GoAPIError as photo_error:
                        messages.warning(
                            request,
                            "اطلاعات دانشجو ویرایش شد؛ "
                            "اما تصویر پروفایل بارگذاری نشد: "
                            f"{photo_error}",
                        )

                messages.success(
                    request,
                    "اطلاعات دانشجو با موفقیت ویرایش شد.",
                )

                return redirect(
                    "students:detail",
                    student_id=student_id,
                )

            except GoAPIError as error:
                add_api_error_to_form(
                    form,
                    error,
                )

        return render(
            request,
            self.template_name,
            self.base_context(
                page_title="ویرایش دانشجو",
                form=form,
                form_mode="update",
                student=student,
                submit_label="ذخیره تغییرات",
            ),
            status=400,
        )


class StudentDeleteView(StudentBaseView):
    template_name = (
        "students/student_confirm_delete.html"
    )

    def get(
        self,
        request: HttpRequest,
        student_id: int,
    ) -> HttpResponse:
        try:
            student = self.get_student(student_id)
        except GoAPIError as error:
            messages.error(
                request,
                str(error),
            )

            return redirect(
                "students:list"
            )

        return render(
            request,
            self.template_name,
            self.base_context(
                page_title="حذف دانشجو",
                student=student,
            ),
        )

    def post(
        self,
        request: HttpRequest,
        student_id: int,
    ) -> HttpResponse:
        try:
            student = self.get_student(student_id)

            go_api.delete(
                f"{STUDENTS_API_PATH}/{student_id}"
            )

        except GoAPIError as error:
            if error.status_code == 404:
                raise Http404(
                    "دانشجوی موردنظر پیدا نشد."
                ) from error

            messages.error(
                request,
                f"حذف دانشجو انجام نشد: {error}",
            )

            return redirect(
                "students:detail",
                student_id=student_id,
            )

        messages.success(
            request,
            (
                f"دانشجو «{student['full_name']}» "
                "با موفقیت حذف شد."
            ),
        )

        return redirect(
            "students:list"
        )