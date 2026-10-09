from django.urls import path

from .views import (
    InstructorCreateView,
    InstructorDeleteView,
    InstructorDetailView,
    InstructorListView,
    InstructorUpdateView,
)


app_name = "instructor_management"


urlpatterns = [
    path(
        "",
        InstructorListView.as_view(),
        name="list",
    ),
    path(
        "create/",
        InstructorCreateView.as_view(),
        name="create",
    ),
    path(
        "<int:instructor_id>/",
        InstructorDetailView.as_view(),
        name="detail",
    ),
    path(
        "<int:instructor_id>/edit/",
        InstructorUpdateView.as_view(),
        name="update",
    ),
    path(
        "<int:instructor_id>/delete/",
        InstructorDeleteView.as_view(),
        name="delete",
    ),
]