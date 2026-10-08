from django.contrib.auth.mixins import LoginRequiredMixin
from django.views.generic import TemplateView


class ProfileView(LoginRequiredMixin, TemplateView):
    template_name = "accounts/profile.html"

    def get_context_data(self, **kwargs):
        context = super().get_context_data(**kwargs)
        context["active_page"] = "profile"
        context["page_title"] = "حساب کاربری"
        return context