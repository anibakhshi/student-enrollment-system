from django.urls import path

from .views import (
    CourseDetailView,
    CourseListView,
    InstructorDetailView,
    InstructorListView,
)


app_name = "catalog"


urlpatterns = [
    path(
        "courses/",
        CourseListView.as_view(),
        name="course_list",
    ),
    path(
        "courses/<slug:slug>/",
        CourseDetailView.as_view(),
        name="course_detail",
    ),
    path(
        "instructors/",
        InstructorListView.as_view(),
        name="instructor_list",
    ),
    path(
        "instructors/<slug:slug>/",
        InstructorDetailView.as_view(),
        name="instructor_detail",
    ),
]