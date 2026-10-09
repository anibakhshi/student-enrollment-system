from django.contrib import admin

from .models import UserProfile


@admin.register(UserProfile)
class UserProfileAdmin(admin.ModelAdmin):
    list_display = (
        "display_name",
        "role",
        "phone",
        "national_code",
        "is_verified",
        "created_at",
    )
    list_filter = (
        "role",
        "is_verified",
        "created_at",
    )
    search_fields = (
        "user__username",
        "user__first_name",
        "user__last_name",
        "user__email",
        "phone",
        "national_code",
    )
    list_select_related = ("user",)
    readonly_fields = (
        "created_at",
        "updated_at",
    )
    ordering = (
        "user__first_name",
        "user__last_name",
        "user__username",
    )

    @admin.display(description="نام کاربر")
    def display_name(self, obj):
        return obj.display_name