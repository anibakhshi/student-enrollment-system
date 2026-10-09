from django.urls import path

from .views import (
    StudentCreateView,
    StudentDeleteView,
    StudentDetailView,
    StudentListView,
    StudentUpdateView,
)


app_name = "students"


urlpatterns = [
    path(
        "",
        StudentListView.as_view(),
        name="list",
    ),
    path(
        "create/",
        StudentCreateView.as_view(),
        name="create",
    ),
    path(
        "<int:student_id>/",
        StudentDetailView.as_view(),
        name="detail",
    ),
    path(
        "<int:student_id>/edit/",
        StudentUpdateView.as_view(),
        name="update",
    ),
    path(
        "<int:student_id>/delete/",
        StudentDeleteView.as_view(),
        name="delete",
    ),
]