from django.conf import settings

from dashboard.api_client import GoAPIError, go_api


class CatalogContextMixin:
    page_title = "کاتالوگ آموزشی"
    active_page = "catalog"

    def get_context_data(self, **kwargs):
        context = super().get_context_data(**kwargs)

        api_available = False
        api_health = {}
        errors = []

        try:
            response = go_api.get("/health")
            api_health = response.get("data", {})
            api_available = True
        except GoAPIError as exc:
            errors.append(exc.message)

        context.update(
            {
                "page_title": self.page_title,
                "active_page": self.active_page,
                "api_available": api_available,
                "api_health": api_health,
                "api_base_url": settings.GO_API_BASE_URL,
                "errors": errors,
            }
        )

        return context