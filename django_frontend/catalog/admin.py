from django.contrib import admin

from .models import (
    Category,
    CourseCatalog,
    CourseOffering,
    InstructorProfile,
)


@admin.register(Category)
class CategoryAdmin(admin.ModelAdmin):
    list_display = (
        "name",
        "display_order",
        "is_active",
        "updated_at",
    )
    list_filter = ("is_active",)
    search_fields = ("name", "description")
    prepopulated_fields = {"slug": ("name",)}
    ordering = ("display_order", "name")


@admin.register(InstructorProfile)
class InstructorProfileAdmin(admin.ModelAdmin):
    list_display = (
        "full_name",
        "expertise",
        "api_instructor_id",
        "is_featured",
        "is_active",
    )
    list_filter = (
        "is_featured",
        "is_active",
    )
    search_fields = (
        "first_name",
        "last_name",
        "expertise",
        "email",
    )
    prepopulated_fields = {
        "slug": (
            "first_name",
            "last_name",
        ),
    }


@admin.register(CourseCatalog)
class CourseCatalogAdmin(admin.ModelAdmin):
    list_display = (
        "title",
        "category",
        "price_toman",
        "duration_hours",
        "level",
        "is_featured",
        "is_active",
    )
    list_filter = (
        "category",
        "level",
        "is_featured",
        "is_active",
    )
    search_fields = (
        "title",
        "short_description",
        "description",
    )
    prepopulated_fields = {"slug": ("title",)}
    autocomplete_fields = ("category",)


@admin.register(CourseOffering)
class CourseOfferingAdmin(admin.ModelAdmin):
    list_display = (
        "course",
        "instructor",
        "start_date_text",
        "delivery_mode",
        "status",
        "is_active",
    )
    list_filter = (
        "delivery_mode",
        "status",
        "is_active",
    )
    search_fields = (
        "course__title",
        "instructor__first_name",
        "instructor__last_name",
    )
    autocomplete_fields = (
        "course",
        "instructor",
    )