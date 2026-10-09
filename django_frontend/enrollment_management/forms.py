from __future__ import annotations

from typing import Any, Iterable

from django import forms


STATUS_CHOICES = (
    ("pending", "در انتظار تأیید"),
    ("confirmed", "تأییدشده"),
    ("cancelled", "لغوشده"),
    ("completed", "تکمیل‌شده"),
)

STATUS_TRANSITIONS = {
    "pending": (
        ("pending", "در انتظار تأیید"),
        ("confirmed", "تأییدشده"),
        ("cancelled", "لغوشده"),
    ),
    "confirmed": (
        ("confirmed", "تأییدشده"),
        ("completed", "تکمیل‌شده"),
        ("cancelled", "لغوشده"),
    ),
    "cancelled": (),
    "completed": (),
}


def normalize_text(value: str) -> str:
    return " ".join(value.strip().split())


def student_full_name(
    student: dict[str, Any],
) -> str:
    full_name = " ".join(
        [
            str(student.get("first_name") or "").strip(),
            str(student.get("last_name") or "").strip(),
        ]
    ).strip()

    return full_name or (
        f"دانشجوی شماره {student.get('id', '—')}"
    )


def course_display_name(
    course: dict[str, Any],
) -> str:
    code = str(course.get("code") or "").strip()
    title = str(course.get("title") or "").strip()

    if code and title:
        return f"{code} — {title}"

    return title or code or (
        f"دوره شماره {course.get('id', '—')}"
    )


class EnrollmentCreateForm(forms.Form):
    student_id = forms.TypedChoiceField(
        label="دانشجو",
        choices=(),
        coerce=int,
        empty_value=None,
        widget=forms.Select(
            attrs={
                "class": "form-control form-select",
            }
        ),
    )

    course_id = forms.TypedChoiceField(
        label="دوره",
        choices=(),
        coerce=int,
        empty_value=None,
        widget=forms.Select(
            attrs={
                "class": "form-control form-select",
            }
        ),
    )

    notes = forms.CharField(
        label="یادداشت ثبت‌نام",
        required=False,
        max_length=2000,
        widget=forms.Textarea(
            attrs={
                "class": (
                    "form-control form-control--textarea"
                ),
                "placeholder": (
                    "توضیحات تکمیلی یا نکات مربوط "
                    "به ثبت‌نام را وارد کنید."
                ),
                "rows": 7,
            }
        ),
    )

    def __init__(
        self,
        *args: Any,
        students: Iterable[dict[str, Any]] | None = None,
        courses: Iterable[dict[str, Any]] | None = None,
        **kwargs: Any,
    ) -> None:
        super().__init__(*args, **kwargs)

        student_choices: list[
            tuple[int | str, str]
        ] = [
            ("", "یک دانشجو انتخاب کنید"),
        ]

        for student in students or []:
            if not isinstance(student, dict):
                continue

            student_id = student.get("id")

            if not student_id:
                continue

            full_name = student_full_name(student)
            national_code = str(
                student.get("national_code") or ""
            ).strip()

            label = full_name

            if national_code:
                label = (
                    f"{full_name} — کد ملی: "
                    f"{national_code}"
                )

            student_choices.append(
                (int(student_id), label)
            )

        course_choices: list[
            tuple[int | str, str]
        ] = [
            ("", "یک دوره باز انتخاب کنید"),
        ]

        for course in courses or []:
            if not isinstance(course, dict):
                continue

            course_id = course.get("id")

            if not course_id:
                continue

            if course.get("status") != "open":
                continue

            label = course_display_name(course)
            capacity = course.get("capacity")

            if capacity not in (None, ""):
                label = (
                    f"{label} — ظرفیت: {capacity} نفر"
                )

            course_choices.append(
                (int(course_id), label)
            )

        self.fields["student_id"].choices = (
            student_choices
        )
        self.fields["course_id"].choices = (
            course_choices
        )

        for field_name, field in self.fields.items():
            field.widget.attrs.setdefault(
                "aria-describedby",
                f"id_{field_name}_errors",
            )

            if field.required:
                field.widget.attrs.setdefault(
                    "aria-required",
                    "true",
                )

    def clean_notes(self) -> str:
        value = self.cleaned_data.get(
            "notes",
            "",
        ).strip()

        if len(value) > 2000:
            raise forms.ValidationError(
                "یادداشت ثبت‌نام نباید بیشتر از "
                "۲۰۰۰ نویسه باشد."
            )

        return value

    def api_payload(self) -> dict[str, Any]:
        return {
            "student_id": self.cleaned_data[
                "student_id"
            ],
            "course_id": self.cleaned_data[
                "course_id"
            ],
            "notes": self.cleaned_data.get(
                "notes",
                "",
            ),
        }

    def add_api_errors(self, payload: Any) -> None:
        add_api_errors_to_form(self, payload)


class EnrollmentUpdateForm(forms.Form):
    status = forms.ChoiceField(
        label="وضعیت ثبت‌نام",
        choices=(),
        widget=forms.Select(
            attrs={
                "class": "form-control form-select",
            }
        ),
    )

    notes = forms.CharField(
        label="یادداشت ثبت‌نام",
        required=False,
        max_length=2000,
        widget=forms.Textarea(
            attrs={
                "class": (
                    "form-control form-control--textarea"
                ),
                "placeholder": (
                    "توضیحات تکمیلی ثبت‌نام را "
                    "وارد کنید."
                ),
                "rows": 7,
            }
        ),
    )

    def __init__(
        self,
        *args: Any,
        current_status: str = "pending",
        **kwargs: Any,
    ) -> None:
        super().__init__(*args, **kwargs)

        normalized_status = str(
            current_status or "pending"
        ).strip().lower()

        self.current_status = normalized_status

        choices = STATUS_TRANSITIONS.get(
            normalized_status,
            (),
        )

        self.fields["status"].choices = choices

        if not choices:
            self.fields["status"].disabled = True
            self.fields["status"].required = False

        for field_name, field in self.fields.items():
            field.widget.attrs.setdefault(
                "aria-describedby",
                f"id_{field_name}_errors",
            )

            if field.required:
                field.widget.attrs.setdefault(
                    "aria-required",
                    "true",
                )

    def clean_status(self) -> str:
        status = self.cleaned_data.get("status")

        if not self.fields["status"].choices:
            raise forms.ValidationError(
                "این ثبت‌نام در وضعیت نهایی قرار دارد "
                "و قابل ویرایش نیست."
            )

        return str(status)

    def clean_notes(self) -> str:
        return self.cleaned_data.get(
            "notes",
            "",
        ).strip()

    def api_payload(self) -> dict[str, Any]:
        return {
            "status": self.cleaned_data["status"],
            "notes": self.cleaned_data.get(
                "notes",
                "",
            ),
        }

    def add_api_errors(self, payload: Any) -> None:
        add_api_errors_to_form(self, payload)


class EnrollmentSearchForm(forms.Form):
    q = forms.CharField(
        label="جست‌وجوی ثبت‌نام",
        required=False,
        max_length=200,
        widget=forms.SearchInput(
            attrs={
                "class": "search-input",
                "placeholder": (
                    "جست‌وجو بر اساس دانشجو، دوره "
                    "یا یادداشت"
                ),
                "autocomplete": "off",
            }
        ),
    )

    status = forms.ChoiceField(
        label="وضعیت",
        required=False,
        choices=(
            ("", "همه وضعیت‌ها"),
            *STATUS_CHOICES,
        ),
        widget=forms.Select(
            attrs={
                "class": "filter-select",
            }
        ),
    )

    student = forms.ChoiceField(
        label="دانشجو",
        required=False,
        choices=(("", "همه دانشجویان"),),
        widget=forms.Select(
            attrs={
                "class": "filter-select",
            }
        ),
    )

    course = forms.ChoiceField(
        label="دوره",
        required=False,
        choices=(("", "همه دوره‌ها"),),
        widget=forms.Select(
            attrs={
                "class": "filter-select",
            }
        ),
    )

    def __init__(
        self,
        *args: Any,
        students: Iterable[dict[str, Any]] | None = None,
        courses: Iterable[dict[str, Any]] | None = None,
        **kwargs: Any,
    ) -> None:
        super().__init__(*args, **kwargs)

        student_choices: list[tuple[str, str]] = [
            ("", "همه دانشجویان"),
        ]

        for student in students or []:
            if not isinstance(student, dict):
                continue

            student_id = student.get("id")

            if not student_id:
                continue

            student_choices.append(
                (
                    str(student_id),
                    student_full_name(student),
                )
            )

        course_choices: list[tuple[str, str]] = [
            ("", "همه دوره‌ها"),
        ]

        for course in courses or []:
            if not isinstance(course, dict):
                continue

            course_id = course.get("id")

            if not course_id:
                continue

            course_choices.append(
                (
                    str(course_id),
                    course_display_name(course),
                )
            )

        self.fields["student"].choices = (
            student_choices
        )
        self.fields["course"].choices = (
            course_choices
        )


def add_api_errors_to_form(
    form: forms.Form,
    payload: Any,
) -> None:
    if not isinstance(payload, dict):
        form.add_error(
            None,
            "پردازش اطلاعات ثبت‌نام انجام نشد.",
        )
        return

    errors = payload.get("errors")

    if not isinstance(errors, dict):
        data = payload.get("data")

        if isinstance(data, dict):
            errors = data

    if not isinstance(errors, dict):
        form.add_error(
            None,
            payload.get(
                "message",
                "پردازش اطلاعات ثبت‌نام انجام نشد.",
            ),
        )
        return

    field_aliases = {
        "student": "student_id",
        "course": "course_id",
    }

    error_added = False

    for field_name, error_message in errors.items():
        normalized_field = field_aliases.get(
            field_name,
            field_name,
        )

        if normalized_field not in form.fields:
            continue

        messages = (
            error_message
            if isinstance(error_message, list)
            else [error_message]
        )

        for message in messages:
            form.add_error(
                normalized_field,
                str(message),
            )
            error_added = True

    if not error_added:
        form.add_error(
            None,
            payload.get(
                "message",
                "اطلاعات ثبت‌نام معتبر نیست.",
            ),
        )