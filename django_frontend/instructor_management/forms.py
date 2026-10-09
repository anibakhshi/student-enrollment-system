from __future__ import annotations

import re
from typing import Any

from django import forms


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


class InstructorForm(forms.Form):
    STATUS_ACTIVE = "active"
    STATUS_INACTIVE = "inactive"

    STATUS_CHOICES = (
        (STATUS_ACTIVE, "فعال"),
        (STATUS_INACTIVE, "غیرفعال"),
    )

    first_name = forms.CharField(
        label="نام",
        max_length=100,
        widget=forms.TextInput(
            attrs={
                "class": "form-control",
                "placeholder": "مثلاً پرهام",
                "autocomplete": "given-name",
                "autofocus": True,
            }
        ),
    )

    last_name = forms.CharField(
        label="نام خانوادگی",
        max_length=100,
        widget=forms.TextInput(
            attrs={
                "class": "form-control",
                "placeholder": "مثلاً درویشی",
                "autocomplete": "family-name",
            }
        ),
    )

    email = forms.EmailField(
        label="آدرس ایمیل",
        max_length=254,
        widget=forms.EmailInput(
            attrs={
                "class": "form-control",
                "placeholder": "instructor@example.com",
                "autocomplete": "email",
                "dir": "ltr",
            }
        ),
    )

    phone = forms.CharField(
        label="شماره تماس",
        max_length=20,
        required=False,
        widget=forms.TextInput(
            attrs={
                "class": "form-control",
                "placeholder": "09123456789",
                "autocomplete": "tel",
                "inputmode": "tel",
                "dir": "ltr",
            }
        ),
    )

    expertise = forms.CharField(
        label="حوزه تخصص",
        max_length=200,
        required=False,
        widget=forms.TextInput(
            attrs={
                "class": "form-control",
                "placeholder": (
                    "مثلاً توسعه نرم‌افزار، هوش مصنوعی و Golang"
                ),
            }
        ),
    )

    bio = forms.CharField(
        label="معرفی استاد",
        required=False,
        max_length=3000,
        widget=forms.Textarea(
            attrs={
                "class": "form-control form-control--textarea",
                "placeholder": (
                    "سوابق حرفه‌ای، تجربیات آموزشی و "
                    "زمینه‌های تخصصی استاد را وارد کنید."
                ),
                "rows": 7,
            }
        ),
    )

    status = forms.ChoiceField(
        label="وضعیت همکاری",
        choices=STATUS_CHOICES,
        initial=STATUS_ACTIVE,
        required=False,
        widget=forms.Select(
            attrs={
                "class": "form-control form-select",
            }
        ),
    )

    def __init__(
        self,
        *args: Any,
        include_status: bool = False,
        **kwargs: Any,
    ) -> None:
        super().__init__(*args, **kwargs)

        self.include_status = include_status

        if not include_status:
            self.fields.pop("status")

        for field_name, field in self.fields.items():
            described_by = f"id_{field_name}_errors"
            field.widget.attrs.setdefault(
                "aria-describedby",
                described_by,
            )

            if field.required:
                field.widget.attrs.setdefault(
                    "aria-required",
                    "true",
                )

    def clean_first_name(self) -> str:
        value = normalize_text(
            self.cleaned_data["first_name"]
        )

        if len(value) < 2:
            raise forms.ValidationError(
                "نام باید حداقل ۲ نویسه داشته باشد."
            )

        return value

    def clean_last_name(self) -> str:
        value = normalize_text(
            self.cleaned_data["last_name"]
        )

        if len(value) < 2:
            raise forms.ValidationError(
                "نام خانوادگی باید حداقل ۲ نویسه داشته باشد."
            )

        return value

    def clean_email(self) -> str:
        return self.cleaned_data["email"].strip().lower()

    def clean_phone(self) -> str:
        value = normalize_digits(
            self.cleaned_data.get("phone", "")
        )

        value = re.sub(
            r"[\s\-\(\)]",
            "",
            value,
        )

        if not value:
            return ""

        if value.startswith("+98"):
            value = f"0{value[3:]}"
        elif value.startswith("0098"):
            value = f"0{value[4:]}"
        elif value.startswith("98") and len(value) == 12:
            value = f"0{value[2:]}"

        if not re.fullmatch(r"09\d{9}", value):
            raise forms.ValidationError(
                "شماره تماس باید یک شماره موبایل معتبر "
                "مانند 09123456789 باشد."
            )

        return value

    def clean_expertise(self) -> str:
        return normalize_text(
            self.cleaned_data.get("expertise", "")
        )

    def clean_bio(self) -> str:
        value = self.cleaned_data.get("bio", "").strip()

        if value and len(value) < 10:
            raise forms.ValidationError(
                "متن معرفی استاد باید حداقل ۱۰ نویسه باشد."
            )

        return value

    def api_payload(self) -> dict[str, str]:
        payload = {
            "first_name": self.cleaned_data["first_name"],
            "last_name": self.cleaned_data["last_name"],
            "email": self.cleaned_data["email"],
            "phone": self.cleaned_data.get("phone", ""),
            "bio": self.cleaned_data.get("bio", ""),
            "expertise": self.cleaned_data.get(
                "expertise",
                "",
            ),
        }

        if self.include_status:
            payload["status"] = self.cleaned_data["status"]

        return payload

    def add_api_errors(self, payload: Any) -> None:
        if not isinstance(payload, dict):
            self.add_error(
                None,
                "ذخیره اطلاعات استاد انجام نشد.",
            )
            return

        error_data = payload.get("data")

        if not isinstance(error_data, dict):
            error_data = payload.get("errors")

        if not isinstance(error_data, dict):
            self.add_error(
                None,
                payload.get(
                    "message",
                    "ذخیره اطلاعات استاد انجام نشد.",
                ),
            )
            return

        field_aliases = {
            "first_name": "first_name",
            "last_name": "last_name",
            "email": "email",
            "phone": "phone",
            "bio": "bio",
            "expertise": "expertise",
            "status": "status",
        }

        added_error = False

        for api_field, message in error_data.items():
            form_field = field_aliases.get(api_field)

            if form_field not in self.fields:
                continue

            if isinstance(message, list):
                for item in message:
                    self.add_error(
                        form_field,
                        str(item),
                    )
            else:
                self.add_error(
                    form_field,
                    str(message),
                )

            added_error = True

        if not added_error:
            self.add_error(
                None,
                payload.get(
                    "message",
                    "اطلاعات ارسال‌شده معتبر نیست.",
                ),
            )


class InstructorSearchForm(forms.Form):
    q = forms.CharField(
        label="جست‌وجوی استاد",
        required=False,
        max_length=120,
        widget=forms.SearchInput(
            attrs={
                "class": "search-input",
                "placeholder": (
                    "جست‌وجو بر اساس نام، ایمیل یا تخصص"
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
            *InstructorForm.STATUS_CHOICES,
        ),
        widget=forms.Select(
            attrs={
                "class": "filter-select",
            }
        ),
    )

    def clean_q(self) -> str:
        return normalize_text(
            self.cleaned_data.get("q", "")
        )