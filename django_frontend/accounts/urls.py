from django.contrib.auth.views import LoginView, LogoutView
from django.urls import path

from .forms import PersianAuthenticationForm
from .views import ProfileView


app_name = "accounts"


urlpatterns = [
    path(
        "login/",
        LoginView.as_view(
            template_name="accounts/login.html",
            authentication_form=PersianAuthenticationForm,
            redirect_authenticated_user=True,
        ),
        name="login",
    ),
    path(
        "logout/",
        LogoutView.as_view(),
        name="logout",
    ),
    path(
        "profile/",
        ProfileView.as_view(),
        name="profile",
    ),
]