from __future__ import annotations

from pathlib import Path

from django import forms
from django.core.exceptions import ValidationError


MAX_PHOTO_SIZE = 5 * 1024 * 1024

ALLOWED_PHOTO_CONTENT_TYPES = {
    "image/jpeg",
    "image/png",
    "image/webp",
}

ALLOWED_PHOTO_EXTENSIONS = {
    ".jpg",
    ".jpeg",
    ".png",
    ".webp",
}


def normalize_digits(value: str) -> str:
    """
    ارقام فارسی و عربی را به ارقام انگلیسی تبدیل می‌کند.
    """

    translation_table = str.maketrans(
        "۰۱۲۳۴۵۶۷۸۹٠١٢٣٤٥٦٧٨٩",
        "01234567890123456789",
    )

    return value.translate(translation_table)


class StudentForm(forms.Form):
    first_name = forms.CharField(
        label="نام",
        min_length=2,
        max_length=100,
        widget=forms.TextInput(
            attrs={
                "class": "student-field",
                "placeholder": "برای مثال: آنیتا",
                "autocomplete": "given-name",
            }
        ),
    )

    last_name = forms.CharField(
        label="نام خانوادگی",
        min_length=2,
        max_length=100,
        widget=forms.TextInput(
            attrs={
                "class": "student-field",
                "placeholder": "برای مثال: بخشی",
                "autocomplete": "family-name",
            }
        ),
    )

    age = forms.IntegerField(
        label="سن",
        min_value=1,
        max_value=120,
        widget=forms.NumberInput(
            attrs={
                "class": "student-field",
                "placeholder": "برای مثال: ۲۰",
                "min": "1",
                "max": "120",
                "inputmode": "numeric",
            }
        ),
    )

    national_code = forms.CharField(
        label="کد ملی",
        min_length=10,
        max_length=10,
        widget=forms.TextInput(
            attrs={
                "class": "student-field",
                "placeholder": "کد ملی ۱۰ رقمی",
                "autocomplete": "off",
                "inputmode": "numeric",
                "maxlength": "10",
                "dir": "ltr",
            }
        ),
    )

    email = forms.EmailField(
        label="ایمیل",
        max_length=254,
        widget=forms.EmailInput(
            attrs={
                "class": "student-field",
                "placeholder": "student@example.com",
                "autocomplete": "email",
                "dir": "ltr",
            }
        ),
    )

    phone = forms.CharField(
        label="شماره موبایل",
        min_length=11,
        max_length=11,
        widget=forms.TextInput(
            attrs={
                "class": "student-field",
                "placeholder": "09121234567",
                "autocomplete": "tel",
                "inputmode": "tel",
                "maxlength": "11",
                "dir": "ltr",
            }
        ),
    )

    photo = forms.ImageField(
        label="تصویر پروفایل",
        required=False,
        widget=forms.ClearableFileInput(
            attrs={
                "class": "student-file-input",
                "accept": ".jpg,.jpeg,.png,.webp,"
                "image/jpeg,image/png,image/webp",
            }
        ),
        help_text=(
            "فرمت‌های مجاز JPG، PNG و WebP؛ "
            "حداکثر حجم فایل ۵ مگابایت."
        ),
    )

    def clean_first_name(self) -> str:
        value = " ".join(
            self.cleaned_data["first_name"].split()
        )

        if not value:
            raise ValidationError(
                "نام دانشجو را وارد کنید."
            )

        return value

    def clean_last_name(self) -> str:
        value = " ".join(
            self.cleaned_data["last_name"].split()
        )

        if not value:
            raise ValidationError(
                "نام خانوادگی دانشجو را وارد کنید."
            )

        return value

    def clean_national_code(self) -> str:
        value = normalize_digits(
            self.cleaned_data["national_code"].strip()
        )

        if not value.isdigit():
            raise ValidationError(
                "کد ملی فقط باید شامل عدد باشد."
            )

        if len(value) != 10:
            raise ValidationError(
                "کد ملی باید دقیقاً ۱۰ رقم باشد."
            )

        if len(set(value)) == 1:
            raise ValidationError(
                "کد ملی واردشده معتبر نیست."
            )

        return value

    def clean_phone(self) -> str:
        value = normalize_digits(
            self.cleaned_data["phone"].strip()
        )

        value = (
            value.replace(" ", "")
            .replace("-", "")
        )

        if not value.isdigit():
            raise ValidationError(
                "شماره موبایل فقط باید شامل عدد باشد."
            )

        if not value.startswith("09"):
            raise ValidationError(
                "شماره موبایل باید با 09 شروع شود."
            )

        if len(value) != 11:
            raise ValidationError(
                "شماره موبایل باید دقیقاً ۱۱ رقم باشد."
            )

        return value

    def clean_email(self) -> str:
        return self.cleaned_data["email"].strip().lower()

    def clean_photo(self):
        photo = self.cleaned_data.get("photo")

        if not photo:
            return None

        if photo.size > MAX_PHOTO_SIZE:
            raise ValidationError(
                "حجم تصویر نباید بیشتر از ۵ مگابایت باشد."
            )

        content_type = getattr(
            photo,
            "content_type",
            "",
        ).lower()

        extension = Path(photo.name).suffix.lower()

        if content_type not in ALLOWED_PHOTO_CONTENT_TYPES:
            raise ValidationError(
                "نوع فایل تصویر مجاز نیست."
            )

        if extension not in ALLOWED_PHOTO_EXTENSIONS:
            raise ValidationError(
                "پسوند فایل تصویر مجاز نیست."
            )

        return photo

    def api_payload(self) -> dict[str, object]:
        """
        داده‌های مناسب برای ارسال JSON به Go API.
        """

        if not self.is_valid():
            raise ValueError(
                "فرم باید پیش از ساخت payload معتبر باشد."
            )

        return {
            "first_name": self.cleaned_data["first_name"],
            "last_name": self.cleaned_data["last_name"],
            "age": self.cleaned_data["age"],
            "national_code": self.cleaned_data[
                "national_code"
            ],
            "email": self.cleaned_data["email"],
            "phone": self.cleaned_data["phone"],
        }


class StudentSearchForm(forms.Form):
    q = forms.CharField(
        label="جست‌وجوی دانشجو",
        required=False,
        max_length=120,
        widget=forms.SearchInput(
            attrs={
                "class": "student-search-input",
                "placeholder": (
                    "جست‌وجو با نام، کد ملی، ایمیل "
                    "یا شماره موبایل..."
                ),
                "autocomplete": "off",
            }
        ),
    )

    def clean_q(self) -> str:
        value = self.cleaned_data.get("q", "")
        return " ".join(value.split())