from django.urls import path

from .views import (
    EnrollmentCreateView,
    EnrollmentDeleteView,
    EnrollmentDetailView,
    EnrollmentListView,
    EnrollmentUpdateView,
)


app_name = "enrollment_management"


urlpatterns = [
    path(
        "",
        EnrollmentListView.as_view(),
        name="list",
    ),
    path(
        "create/",
        EnrollmentCreateView.as_view(),
        name="create",
    ),
    path(
        "<int:enrollment_id>/",
        EnrollmentDetailView.as_view(),
        name="detail",
    ),
    path(
        "<int:enrollment_id>/edit/",
        EnrollmentUpdateView.as_view(),
        name="update",
    ),
    path(
        "<int:enrollment_id>/delete/",
        EnrollmentDeleteView.as_view(),
        name="delete",
    ),
]