from unittest.mock import patch

from django.contrib.auth import get_user_model
from django.test import TestCase, override_settings
from django.urls import reverse

from accounts.models import UserProfile
from dashboard.api_client import GoAPIError


TEST_STORAGES = {
    "default": {
        "BACKEND": (
            "django.core.files.storage."
            "FileSystemStorage"
        ),
    },
    "staticfiles": {
        "BACKEND": (
            "django.contrib.staticfiles.storage."
            "StaticFilesStorage"
        ),
    },
}


@override_settings(STORAGES=TEST_STORAGES)
class InstructorManagementViewTests(TestCase):
    def setUp(self):
        self.user = get_user_model().objects.create_user(
            username="education-manager",
            password="StrongPassword123!",
            first_name="آنیتا",
            last_name="بخشی",
            is_staff=True,
        )

        self.user.profile.role = UserProfile.Role.ADMIN
        self.user.profile.save(update_fields=["role"])

        self.client.force_login(self.user)

        self.instructor = {
            "id": 1,
            "first_name": "پرهام",
            "last_name": "درویشی",
            "email": "parham@example.com",
            "phone": "09123456789",
            "bio": (
                "مدرس و توسعه‌دهنده نرم‌افزار "
                "با تجربه تخصصی در Golang."
            ),
            "expertise": (
                "توسعه نرم‌افزار و برنامه‌نویسی"
            ),
            "status": "active",
            "created_at": "2026-10-01T10:00:00Z",
            "updated_at": "2026-10-02T11:00:00Z",
        }

    def api_response(self, data):
        return {
            "success": True,
            "message": "Operation completed",
            "data": data,
        }

    def test_anonymous_user_is_redirected(self):
        self.client.logout()

        url = reverse("instructor_management:list")
        response = self.client.get(url)

        self.assertEqual(response.status_code, 302)
        self.assertIn(
            reverse("accounts:login"),
            response.url,
        )

    def test_student_role_cannot_access_management(self):
        self.user.profile.role = UserProfile.Role.STUDENT
        self.user.profile.save(update_fields=["role"])

        response = self.client.get(
            reverse("instructor_management:list")
        )

        self.assertEqual(response.status_code, 403)

    @patch("instructor_management.views.go_api.get")
    def test_instructor_list_displays_api_data(
        self,
        mock_get,
    ):
        mock_get.return_value = self.api_response(
            [self.instructor]
        )

        response = self.client.get(
            reverse("instructor_management:list")
        )

        self.assertEqual(response.status_code, 200)
        self.assertTemplateUsed(
            response,
            (
                "instructors_management/"
                "instructor_list.html"
            ),
        )
        self.assertContains(response, "پرهام درویشی")
        self.assertContains(
            response,
            "توسعه نرم‌افزار و برنامه‌نویسی",
        )
        self.assertEqual(
            response.context["total_instructors"],
            1,
        )
        self.assertEqual(
            response.context["active_instructors"],
            1,
        )

    @patch("instructor_management.views.go_api.get")
    def test_instructor_list_searches_by_expertise(
        self,
        mock_get,
    ):
        second_instructor = {
            **self.instructor,
            "id": 2,
            "first_name": "ساناز",
            "last_name": "عباس‌زاده",
            "email": "sanaz@example.com",
            "expertise": "SQL Server",
        }

        mock_get.return_value = self.api_response(
            [
                self.instructor,
                second_instructor,
            ]
        )

        response = self.client.get(
            reverse("instructor_management:list"),
            {"q": "SQL Server"},
        )

        self.assertEqual(response.status_code, 200)
        self.assertContains(
            response,
            "ساناز عباس‌زاده",
        )
        self.assertNotContains(
            response,
            "پرهام درویشی",
        )

    @patch("instructor_management.views.go_api.get")
    def test_instructor_list_filters_status(
        self,
        mock_get,
    ):
        inactive_instructor = {
            **self.instructor,
            "id": 2,
            "first_name": "آرش",
            "last_name": "فروغی",
            "email": "arash@example.com",
            "status": "inactive",
        }

        mock_get.return_value = self.api_response(
            [
                self.instructor,
                inactive_instructor,
            ]
        )

        response = self.client.get(
            reverse("instructor_management:list"),
            {"status": "inactive"},
        )

        self.assertEqual(response.status_code, 200)
        self.assertContains(response, "آرش فروغی")
        self.assertNotContains(
            response,
            "پرهام درویشی",
        )

    @patch("instructor_management.views.go_api.get")
    def test_list_handles_api_failure(
        self,
        mock_get,
    ):
        mock_get.side_effect = GoAPIError(
            "ارتباط با سرویس اصلی برقرار نشد."
        )

        response = self.client.get(
            reverse("instructor_management:list")
        )

        self.assertEqual(response.status_code, 200)
        self.assertFalse(
            response.context["api_available"]
        )
        self.assertContains(
            response,
            "ارتباط با سرویس اصلی برقرار نشد.",
        )

    @patch("instructor_management.views.go_api.get")
    def test_instructor_detail_displays_information(
        self,
        mock_get,
    ):
        mock_get.return_value = self.api_response(
            self.instructor
        )

        response = self.client.get(
            reverse(
                "instructor_management:detail",
                args=[1],
            )
        )

        self.assertEqual(response.status_code, 200)
        self.assertTemplateUsed(
            response,
            (
                "instructors_management/"
                "instructor_detail.html"
            ),
        )
        self.assertContains(response, "پرهام درویشی")
        self.assertContains(
            response,
            "parham@example.com",
        )

    @patch("instructor_management.views.go_api.get")
    def test_missing_instructor_returns_404(
        self,
        mock_get,
    ):
        mock_get.side_effect = GoAPIError(
            "Instructor not found",
            status_code=404,
        )

        response = self.client.get(
            reverse(
                "instructor_management:detail",
                args=[999],
            )
        )

        self.assertEqual(response.status_code, 404)

    def test_create_page_is_available(self):
        response = self.client.get(
            reverse("instructor_management:create")
        )

        self.assertEqual(response.status_code, 200)
        self.assertContains(response, "ثبت استاد جدید")

    @patch("instructor_management.views.go_api.post")
    def test_create_instructor_posts_valid_payload(
        self,
        mock_post,
    ):
        mock_post.return_value = self.api_response(
            self.instructor
        )

        response = self.client.post(
            reverse("instructor_management:create"),
            {
                "first_name": "پرهام",
                "last_name": "درویشی",
                "email": "PARHAM@example.com",
                "phone": "۰۹۱۲۳۴۵۶۷۸۹",
                "expertise": (
                    "توسعه نرم‌افزار و برنامه‌نویسی"
                ),
                "bio": (
                    "مدرس و توسعه‌دهنده نرم‌افزار "
                    "با تجربه تخصصی در Golang."
                ),
            },
        )

        self.assertRedirects(
            response,
            reverse(
                "instructor_management:detail",
                args=[1],
            ),
            fetch_redirect_response=False,
        )

        mock_post.assert_called_once_with(
            "/api/v1/instructors",
            json={
                "first_name": "پرهام",
                "last_name": "درویشی",
                "email": "parham@example.com",
                "phone": "09123456789",
                "expertise": (
                    "توسعه نرم‌افزار و برنامه‌نویسی"
                ),
                "bio": (
                    "مدرس و توسعه‌دهنده نرم‌افزار "
                    "با تجربه تخصصی در Golang."
                ),
            },
        )

    @patch("instructor_management.views.go_api.post")
    def test_create_displays_api_validation_error(
        self,
        mock_post,
    ):
        mock_post.side_effect = GoAPIError(
            "Instructor email already exists",
            status_code=409,
            payload={
                "success": False,
                "message": (
                    "Instructor email already exists"
                ),
                "errors": {
                    "email": (
                        "این ایمیل قبلاً ثبت شده است."
                    ),
                },
            },
        )

        response = self.client.post(
            reverse("instructor_management:create"),
            {
                "first_name": "پرهام",
                "last_name": "درویشی",
                "email": "parham@example.com",
                "phone": "09123456789",
                "expertise": "Golang",
                "bio": (
                    "مدرس حرفه‌ای توسعه نرم‌افزار."
                ),
            },
        )

        self.assertEqual(response.status_code, 400)
        self.assertContains(
            response,
            "این ایمیل قبلاً ثبت شده است.",
            status_code=400,
        )

    @patch("instructor_management.views.go_api.put")
    @patch("instructor_management.views.go_api.get")
    def test_update_instructor(
        self,
        mock_get,
        mock_put,
    ):
        mock_get.return_value = self.api_response(
            self.instructor
        )
        mock_put.return_value = self.api_response(
            {
                **self.instructor,
                "expertise": "Golang و معماری نرم‌افزار",
            }
        )

        response = self.client.post(
            reverse(
                "instructor_management:update",
                args=[1],
            ),
            {
                "first_name": "پرهام",
                "last_name": "درویشی",
                "email": "parham@example.com",
                "phone": "09123456789",
                "expertise": (
                    "Golang و معماری نرم‌افزار"
                ),
                "bio": (
                    "مدرس و توسعه‌دهنده نرم‌افزار "
                    "با تجربه تخصصی در Golang."
                ),
                "status": "active",
            },
        )

        self.assertRedirects(
            response,
            reverse(
                "instructor_management:detail",
                args=[1],
            ),
            fetch_redirect_response=False,
        )

        mock_put.assert_called_once()
        payload = mock_put.call_args.kwargs["json"]

        self.assertEqual(payload["status"], "active")
        self.assertEqual(
            payload["expertise"],
            "Golang و معماری نرم‌افزار",
        )

    @patch("instructor_management.views.go_api.get")
    def test_delete_confirmation_page(
        self,
        mock_get,
    ):
        mock_get.return_value = self.api_response(
            self.instructor
        )

        response = self.client.get(
            reverse(
                "instructor_management:delete",
                args=[1],
            )
        )

        self.assertEqual(response.status_code, 200)
        self.assertContains(response, "حذف استاد")
        self.assertContains(response, "پرهام درویشی")

    @patch("instructor_management.views.go_api.delete")
    @patch("instructor_management.views.go_api.get")
    def test_delete_instructor(
        self,
        mock_get,
        mock_delete,
    ):
        mock_get.return_value = self.api_response(
            self.instructor
        )
        mock_delete.return_value = None

        response = self.client.post(
            reverse(
                "instructor_management:delete",
                args=[1],
            )
        )

        self.assertRedirects(
            response,
            reverse("instructor_management:list"),
            fetch_redirect_response=False,
        )

        mock_delete.assert_called_once_with(
            "/api/v1/instructors/1"
        )