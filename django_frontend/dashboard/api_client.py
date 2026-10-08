import logging
from typing import Any

import requests
from django.conf import settings


logger = logging.getLogger(__name__)


class GoAPIError(Exception):
    """Raised when communication with the Go API fails."""

    def __init__(
        self,
        message: str,
        status_code: int | None = None,
        errors: dict[str, Any] | None = None,
    ):
        super().__init__(message)
        self.message = message
        self.status_code = status_code
        self.errors = errors or {}


class GoAPIClient:
    """HTTP client for the Student Enrollment Go REST API."""

    def __init__(self) -> None:
        self.base_url = settings.GO_API_BASE_URL
        self.timeout = settings.GO_API_TIMEOUT
        self.session = requests.Session()

        self.session.headers.update(
            {
                "Accept": "application/json",
                "User-Agent": "student-enrollment-django-gui/1.1.0",
            }
        )

    def request(
        self,
        method: str,
        path: str,
        **kwargs: Any,
    ) -> dict[str, Any]:
        url = f"{self.base_url}/{path.lstrip('/')}"

        try:
            response = self.session.request(
                method=method,
                url=url,
                timeout=self.timeout,
                **kwargs,
            )
        except requests.Timeout as exc:
            logger.warning(
                "Go API request timed out",
                extra={"url": url, "method": method},
            )

            raise GoAPIError(
                "زمان پاسخ‌گویی سرویس به پایان رسید."
            ) from exc

        except requests.RequestException as exc:
            logger.error(
                "Could not connect to Go API",
                extra={"url": url, "method": method},
            )

            raise GoAPIError(
                "ارتباط با سرویس اصلی برقرار نشد."
            ) from exc

        try:
            payload = response.json()
        except ValueError as exc:
            raise GoAPIError(
                "پاسخ دریافتی از سرویس معتبر نیست.",
                status_code=response.status_code,
            ) from exc

        if not response.ok or not payload.get("success", False):
            raise GoAPIError(
                message=payload.get(
                    "message",
                    "درخواست با خطا مواجه شد.",
                ),
                status_code=response.status_code,
                errors=payload.get("errors", {}),
            )

        return payload

    def get(
        self,
        path: str,
        **kwargs: Any,
    ) -> dict[str, Any]:
        return self.request("GET", path, **kwargs)

    def post(
        self,
        path: str,
        **kwargs: Any,
    ) -> dict[str, Any]:
        return self.request("POST", path, **kwargs)

    def put(
        self,
        path: str,
        **kwargs: Any,
    ) -> dict[str, Any]:
        return self.request("PUT", path, **kwargs)

    def delete(
        self,
        path: str,
        **kwargs: Any,
    ) -> dict[str, Any]:
        return self.request("DELETE", path, **kwargs)


go_api = GoAPIClient()