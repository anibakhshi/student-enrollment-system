from django import forms
from django.contrib.auth.forms import AuthenticationForm


class PersianAuthenticationForm(AuthenticationForm):
    username = forms.CharField(
        label="نام کاربری",
        widget=forms.TextInput(
            attrs={
                "class": "auth-input",
                "placeholder": "نام کاربری خود را وارد کنید",
                "autocomplete": "username",
                "autofocus": True,
            }
        ),
    )
    password = forms.CharField(
        label="رمز عبور",
        strip=False,
        widget=forms.PasswordInput(
            attrs={
                "class": "auth-input",
                "placeholder": "رمز عبور خود را وارد کنید",
                "autocomplete": "current-password",
            }
        ),
    )

    error_messages = {
        "invalid_login": (
            "نام کاربری یا رمز عبور صحیح نیست. "
            "لطفاً اطلاعات خود را دوباره بررسی کنید."
        ),
        "inactive": "این حساب کاربری غیرفعال شده است.",
    }