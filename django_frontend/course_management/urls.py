from django.urls import path

from .views import (
    CourseCreateView,
    CourseDeleteView,
    CourseDetailView,
    CourseListView,
    CourseUpdateView,
)


app_name = "course_management"


urlpatterns = [
    path(
        "",
        CourseListView.as_view(),
        name="list",
    ),
    path(
        "create/",
        CourseCreateView.as_view(),
        name="create",
    ),
    path(
        "<int:course_id>/",
        CourseDetailView.as_view(),
        name="detail",
    ),
    path(
        "<int:course_id>/edit/",
        CourseUpdateView.as_view(),
        name="update",
    ),
    path(
        "<int:course_id>/delete/",
        CourseDeleteView.as_view(),
        name="delete",
    ),
]