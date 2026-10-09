from django.conf import settings
from django.db import models


class UserProfile(models.Model):
    class Role(models.TextChoices):
        ADMIN = "admin", "مدیر سیستم"
        EDUCATION_STAFF = "education_staff", "کارشناس آموزش"
        INSTRUCTOR = "instructor", "استاد"
        STUDENT = "student", "دانشجو"

    user = models.OneToOneField(
        settings.AUTH_USER_MODEL,
        on_delete=models.CASCADE,
        related_name="profile",
        verbose_name="کاربر",
    )
    role = models.CharField(
        max_length=32,
        choices=Role.choices,
        default=Role.STUDENT,
        db_index=True,
        verbose_name="نقش",
    )
    phone = models.CharField(
        max_length=20,
        blank=True,
        verbose_name="شماره تماس",
    )
    national_code = models.CharField(
        max_length=20,
        blank=True,
        db_index=True,
        verbose_name="کد ملی",
    )
    avatar = models.ImageField(
        upload_to="accounts/avatars/%Y/%m/",
        blank=True,
        null=True,
        verbose_name="تصویر پروفایل",
    )
    bio = models.TextField(
        blank=True,
        verbose_name="درباره کاربر",
    )
    is_verified = models.BooleanField(
        default=False,
        verbose_name="تأیید شده",
    )
    created_at = models.DateTimeField(
        auto_now_add=True,
        verbose_name="تاریخ ایجاد",
    )
    updated_at = models.DateTimeField(
        auto_now=True,
        verbose_name="آخرین تغییر",
    )

    class Meta:
        verbose_name = "پروفایل کاربر"
        verbose_name_plural = "پروفایل‌های کاربران"
        ordering = ("user__first_name", "user__last_name", "user__username")
        constraints = [
            models.UniqueConstraint(
                fields=("national_code",),
                condition=~models.Q(national_code=""),
                name="accounts_unique_nonempty_national_code",
            ),
        ]

    def __str__(self):
        return f"{self.display_name} — {self.get_role_display()}"

    @property
    def display_name(self):
        full_name = self.user.get_full_name().strip()
        return full_name or self.user.username

    @property
    def is_admin(self):
        return (
            self.role == self.Role.ADMIN
            or self.user.is_superuser
        )

    @property
    def can_manage_education(self):
        return (
            self.is_admin
            or self.role == self.Role.EDUCATION_STAFF
        )