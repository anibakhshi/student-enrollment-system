from __future__ import annotations

from datetime import date
from typing import Any, Iterable

from django import forms
from django.core.exceptions import ValidationError


PERSIAN_DIGITS = "۰۱۲۳۴۵۶۷۸۹"
ARABIC_DIGITS = "٠١٢٣٤٥٦٧٨٩"
ENGLISH_DIGITS = "0123456789"

DIGIT_TRANSLATION = str.maketrans(
    PERSIAN_DIGITS + ARABIC_DIGITS,
    ENGLISH_DIGITS + ENGLISH_DIGITS,
)


def normalize_digits(value: str) -> str:
    return value.translate(DIGIT_TRANSLATION)


def normalize_text(value: str) -> str:
    return " ".join(value.strip().split())


class NormalizedIntegerField(forms.IntegerField):
    def to_python(self, value: Any) -> int | None:
        if isinstance(value, str):
            value = normalize_digits(value)
            value = (
                value.replace(",", "")
                .replace("٬", "")
                .replace("،", "")
                .strip()
            )

        return super().to_python(value)


class CourseForm(forms.Form):
    STATUS_CHOICES = (
        ("draft", "پیش‌نویس"),
        ("open", "باز برای ثبت‌نام"),
        ("closed", "بسته"),
        ("completed", "تکمیل‌شده"),
        ("cancelled", "لغوشده"),
    )

    instructor_id = forms.TypedChoiceField(
        label="استاد دوره",
        choices=(),
        coerce=int,
        empty_value=None,
        widget=forms.Select(
            attrs={
                "class": "form-control form-select",
            }
        ),
    )

    code = forms.CharField(
        label="کد دوره",
        min_length=2,
        max_length=30,
        widget=forms.TextInput(
            attrs={
                "class": "form-control",
                "placeholder": "مثلاً GO-201",
                "autocomplete": "off",
                "dir": "ltr",
            }
        ),
    )

    title = forms.CharField(
        label="عنوان دوره",
        min_length=2,
        max_length=200,
        widget=forms.TextInput(
            attrs={
                "class": "form-control",
                "placeholder": (
                    "مثلاً توسعه REST API با Golang"
                ),
                "autocomplete": "off",
            }
        ),
    )

    description = forms.CharField(
        label="توضیحات دوره",
        max_length=5000,
        required=False,
        widget=forms.Textarea(
            attrs={
                "class": (
                    "form-control form-control--textarea"
                ),
                "placeholder": (
                    "اهداف، محتوای آموزشی و مخاطبان "
                    "دوره را توضیح دهید."
                ),
                "rows": 8,
            }
        ),
    )

    price = NormalizedIntegerField(
        label="قیمت دوره",
        min_value=0,
        max_value=999999999999,
        widget=forms.NumberInput(
            attrs={
                "class": "form-control",
                "placeholder": "مثلاً 18000000",
                "min": "0",
                "step": "1",
                "inputmode": "numeric",
                "dir": "ltr",
            }
        ),
    )

    capacity = NormalizedIntegerField(
        label="ظرفیت دوره",
        min_value=1,
        max_value=1000,
        widget=forms.NumberInput(
            attrs={
                "class": "form-control",
                "placeholder": "مثلاً 25",
                "min": "1",
                "max": "1000",
                "step": "1",
                "inputmode": "numeric",
                "dir": "ltr",
            }
        ),
    )

    duration_hours = NormalizedIntegerField(
        label="مدت دوره به ساعت",
        min_value=1,
        max_value=5000,
        widget=forms.NumberInput(
            attrs={
                "class": "form-control",
                "placeholder": "مثلاً 72",
                "min": "1",
                "max": "5000",
                "step": "1",
                "inputmode": "numeric",
                "dir": "ltr",
            }
        ),
    )

    start_date = forms.DateField(
        label="تاریخ شروع",
        input_formats=["%Y-%m-%d"],
        widget=forms.DateInput(
            format="%Y-%m-%d",
            attrs={
                "class": "form-control",
                "type": "date",
                "dir": "ltr",
            },
        ),
    )

    end_date = forms.DateField(
        label="تاریخ پایان",
        input_formats=["%Y-%m-%d"],
        widget=forms.DateInput(
            format="%Y-%m-%d",
            attrs={
                "class": "form-control",
                "type": "date",
                "dir": "ltr",
            },
        ),
    )

    status = forms.ChoiceField(
        label="وضعیت دوره",
        choices=STATUS_CHOICES,
        initial="draft",
        widget=forms.Select(
            attrs={
                "class": "form-control form-select",
            }
        ),
    )

    def __init__(
        self,
        *args: Any,
        instructors: Iterable[dict[str, Any]] | None = None,
        include_status: bool = False,
        **kwargs: Any,
    ) -> None:
        super().__init__(*args, **kwargs)

        self.include_status = include_status

        instructor_choices: list[tuple[int | str, str]] = [
            ("", "یک استاد فعال انتخاب کنید")
        ]

        for instructor in instructors or []:
            if not isinstance(instructor, dict):
                continue

            if instructor.get("status") != "active":
                continue

            instructor_id = instructor.get("id")

            if not instructor_id:
                continue

            first_name = str(
                instructor.get("first_name") or ""
            ).strip()
            last_name = str(
                instructor.get("last_name") or ""
            ).strip()
            expertise = str(
                instructor.get("expertise") or ""
            ).strip()

            full_name = (
                f"{first_name} {last_name}".strip()
                or f"استاد شماره {instructor_id}"
            )

            label = full_name

            if expertise:
                label = f"{full_name} — {expertise}"

            instructor_choices.append(
                (int(instructor_id), label)
            )

        self.fields["instructor_id"].choices = (
            instructor_choices
        )

        if not include_status:
            self.fields.pop("status")

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

    def clean_code(self) -> str:
        value = normalize_text(
            self.cleaned_data["code"]
        ).upper()

        if len(value) < 2:
            raise ValidationError(
                "کد دوره باید حداقل ۲ نویسه داشته باشد."
            )

        return value

    def clean_title(self) -> str:
        value = normalize_text(
            self.cleaned_data["title"]
        )

        if len(value) < 2:
            raise ValidationError(
                "عنوان دوره باید حداقل ۲ نویسه داشته باشد."
            )

        return value

    def clean_description(self) -> str:
        return self.cleaned_data.get(
            "description",
            "",
        ).strip()

    def clean(self) -> dict[str, Any]:
        cleaned_data = super().clean()

        start_date = cleaned_data.get("start_date")
        end_date = cleaned_data.get("end_date")

        if (
            isinstance(start_date, date)
            and isinstance(end_date, date)
            and end_date < start_date
        ):
            self.add_error(
                "end_date",
                (
                    "تاریخ پایان نمی‌تواند قبل از "
                    "تاریخ شروع باشد."
                ),
            )

        return cleaned_data

    def api_payload(self) -> dict[str, Any]:
        payload: dict[str, Any] = {
            "instructor_id": self.cleaned_data[
                "instructor_id"
            ],
            "code": self.cleaned_data["code"],
            "title": self.cleaned_data["title"],
            "description": self.cleaned_data.get(
                "description",
                "",
            ),
            "price": self.cleaned_data["price"],
            "capacity": self.cleaned_data["capacity"],
            "duration_hours": self.cleaned_data[
                "duration_hours"
            ],
            "start_date": self.cleaned_data[
                "start_date"
            ].isoformat(),
            "end_date": self.cleaned_data[
                "end_date"
            ].isoformat(),
        }

        if self.include_status:
            payload["status"] = self.cleaned_data[
                "status"
            ]

        return payload

    def add_api_errors(self, payload: Any) -> None:
        if not isinstance(payload, dict):
            self.add_error(
                None,
                "ذخیره اطلاعات دوره انجام نشد.",
            )
            return

        errors = payload.get("errors")

        if not isinstance(errors, dict):
            data = payload.get("data")

            if isinstance(data, dict):
                errors = data

        if not isinstance(errors, dict):
            self.add_error(
                None,
                payload.get(
                    "message",
                    "ذخیره اطلاعات دوره انجام نشد.",
                ),
            )
            return

        error_added = False

        for field_name, error_message in errors.items():
            if field_name not in self.fields:
                continue

            messages = (
                error_message
                if isinstance(error_message, list)
                else [error_message]
            )

            for message in messages:
                self.add_error(
                    field_name,
                    str(message),
                )
                error_added = True

        if not error_added:
            self.add_error(
                None,
                payload.get(
                    "message",
                    "اطلاعات دوره معتبر نیست.",
                ),
            )


class CourseSearchForm(forms.Form):
    q = forms.CharField(
        label="جست‌وجوی دوره",
        required=False,
        max_length=200,
        widget=forms.SearchInput(
            attrs={
                "class": "search-input",
                "placeholder": (
                    "جست‌وجو بر اساس عنوان، کد یا استاد"
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
            *CourseForm.STATUS_CHOICES,
        ),
        widget=forms.Select(
            attrs={
                "class": "filter-select",
            }
        ),
    )

    instructor = forms.ChoiceField(
        label="استاد",
        required=False,
        choices=(("", "همه استادان"),),
        widget=forms.Select(
            attrs={
                "class": "filter-select",
            }
        ),
    )

    def __init__(
        self,
        *args: Any,
        instructors: Iterable[dict[str, Any]] | None = None,
        **kwargs: Any,
    ) -> None:
        super().__init__(*args, **kwargs)

        choices: list[tuple[str, str]] = [
            ("", "همه استادان")
        ]

        for instructor in instructors or []:
            if not isinstance(instructor, dict):
                continue

            instructor_id = instructor.get("id")

            if not instructor_id:
                continue

            full_name = " ".join(
                [
                    str(
                        instructor.get(
                            "first_name",
                            "",
                        )
                    ).strip(),
                    str(
                        instructor.get(
                            "last_name",
                            "",
                        )
                    ).strip(),
                ]
            ).strip()

            choices.append(
                (
                    str(instructor_id),
                    full_name
                    or f"استاد شماره {instructor_id}",
                )
            )

        self.fields["instructor"].choices = choices

    def clean_q(self) -> str:
        return normalize_text(
            self.cleaned_data.get("q", "")
        )