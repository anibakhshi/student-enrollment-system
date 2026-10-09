from __future__ import annotations

from typing import Any

import requests
from django.conf import settings


class GoAPIError(Exception):
    """
    خطای قابل‌کنترل هنگام ارتباط با Go REST API.

    این کلاس علاوه بر پیام خطا، کد وضعیت HTTP و پاسخ خام API
    را نگهداری می‌کند تا فرم‌ها و Viewها بتوانند خطاهای اعتبارسنجی
    را به‌شکل مناسب به کاربر نمایش دهند.
    """

    def __init__(
        self,
        message: str,
        *,
        status_code: int | None = None,
        payload: Any = None,
    ) -> None:
        super().__init__(message)

        self.message = message
        self.status_code = status_code
        self.payload = payload

    def __str__(self) -> str:
        return self.message


class GoAPIClient:
    """
    کلاینت مشترک برای ارتباط Django GUI با Go REST API.
    """

    def __init__(
        self,
        base_url: str | None = None,
        timeout: float | None = None,
    ) -> None:
        self.base_url = (
            base_url
            or settings.GO_API_BASE_URL
        ).rstrip("/")

        self.timeout = (
            timeout
            if timeout is not None
            else settings.GO_API_TIMEOUT
        )

    def build_url(self, path: str) -> str:
        """
        یک مسیر نسبی API را به URL کامل تبدیل می‌کند.
        """

        normalized_path = f"/{path.lstrip('/')}"
        return f"{self.base_url}{normalized_path}"

    @staticmethod
    def extract_error_message(
        payload: Any,
        default_message: str,
    ) -> str:
        """
        بهترین پیام خطای موجود در پاسخ API را استخراج می‌کند.
        """

        if not isinstance(payload, dict):
            return default_message

        for key in (
            "message",
            "error",
            "detail",
        ):
            value = payload.get(key)

            if isinstance(value, str) and value.strip():
                return value.strip()

        data = payload.get("data")

        if isinstance(data, dict):
            for key in (
                "message",
                "error",
                "detail",
            ):
                value = data.get(key)

                if isinstance(value, str) and value.strip():
                    return value.strip()

        errors = payload.get("errors")

        if isinstance(errors, dict):
            messages: list[str] = []

            for field_errors in errors.values():
                if isinstance(field_errors, list):
                    messages.extend(
                        str(item)
                        for item in field_errors
                        if str(item).strip()
                    )
                elif field_errors:
                    messages.append(str(field_errors))

            if messages:
                return " ".join(messages)

        if isinstance(errors, list):
            messages = [
                str(item)
                for item in errors
                if str(item).strip()
            ]

            if messages:
                return " ".join(messages)

        return default_message

    def request(
        self,
        method: str,
        path: str,
        *,
        params: dict[str, Any] | None = None,
        json: Any = None,
        data: Any = None,
        files: Any = None,
        headers: dict[str, str] | None = None,
    ) -> dict[str, Any]:
        """
        یک درخواست HTTP به Go API ارسال می‌کند.

        پاسخ‌های JSON عادی، پاسخ خالی 204 و خطاهای شبکه یا
        اعتبارسنجی API در این متد به‌صورت یکپارچه مدیریت می‌شوند.
        """

        url = self.build_url(path)

        try:
            response = requests.request(
                method=method.upper(),
                url=url,
                params=params,
                json=json,
                data=data,
                files=files,
                headers=headers,
                timeout=self.timeout,
            )
        except requests.Timeout as exc:
            raise GoAPIError(
                "زمان انتظار برای دریافت پاسخ از سرویس اصلی "
                "به پایان رسید.",
            ) from exc
        except requests.ConnectionError as exc:
            raise GoAPIError(
                "ارتباط با سرویس اصلی برقرار نشد.",
            ) from exc
        except requests.RequestException as exc:
            raise GoAPIError(
                "هنگام ارتباط با سرویس اصلی خطایی رخ داد.",
            ) from exc

        if response.status_code == 204:
            return {
                "success": True,
                "message": "",
                "data": None,
            }

        payload: Any = None

        if response.content:
            try:
                payload = response.json()
            except ValueError:
                payload = None

        if not response.ok:
            default_message = (
                "درخواست توسط سرویس اصلی پردازش نشد."
            )

            message = self.extract_error_message(
                payload,
                default_message,
            )

            raise GoAPIError(
                message,
                status_code=response.status_code,
                payload=payload,
            )

        if payload is None:
            return {
                "success": True,
                "message": "",
                "data": None,
            }

        if not isinstance(payload, dict):
            raise GoAPIError(
                "ساختار پاسخ سرویس اصلی معتبر نیست.",
                status_code=response.status_code,
                payload=payload,
            )

        return payload

    def get(
        self,
        path: str,
        *,
        params: dict[str, Any] | None = None,
    ) -> dict[str, Any]:
        return self.request(
            "GET",
            path,
            params=params,
        )

    def post(
        self,
        path: str,
        *,
        json: Any = None,
        data: Any = None,
        files: Any = None,
    ) -> dict[str, Any]:
        return self.request(
            "POST",
            path,
            json=json,
            data=data,
            files=files,
        )

    def put(
        self,
        path: str,
        *,
        json: Any = None,
        data: Any = None,
        files: Any = None,
    ) -> dict[str, Any]:
        return self.request(
            "PUT",
            path,
            json=json,
            data=data,
            files=files,
        )

    def patch(
        self,
        path: str,
        *,
        json: Any = None,
        data: Any = None,
        files: Any = None,
    ) -> dict[str, Any]:
        return self.request(
            "PATCH",
            path,
            json=json,
            data=data,
            files=files,
        )

    def delete(
        self,
        path: str,
    ) -> dict[str, Any]:
        return self.request(
            "DELETE",
            path,
        )


go_api = GoAPIClient()