from django.contrib.auth.mixins import (
    LoginRequiredMixin,
    UserPassesTestMixin,
)
from django.core.exceptions import PermissionDenied


class EducationManagementRequiredMixin(
    LoginRequiredMixin,
    UserPassesTestMixin,
):
    """Allow only administrators and education staff."""

    raise_exception = False

    def test_func(self):
        user = self.request.user

        if not user.is_authenticated:
            return False

        try:
            return user.profile.can_manage_education
        except AttributeError:
            return user.is_superuser

    def handle_no_permission(self):
        if not self.request.user.is_authenticated:
            return super().handle_no_permission()

        raise PermissionDenied(
            "شما اجازه دسترسی به این بخش را ندارید."
        )