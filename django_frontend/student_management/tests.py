from unittest.mock import patch

from django.contrib.auth import get_user_model
from django.test import TestCase, override_settings
from django.urls import reverse

from accounts.models import UserProfile
from dashboard.api_client import GoAPIError


@override_settings(
    STORAGES={
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
)
class StudentManagementViewTests(TestCase):
    def setUp(self):
        user_model = get_user_model()

        self.manager = user_model.objects.create_user(
            username="education-manager",
            password="StrongPassword123!",
            first_name="آنیتا",
            last_name="بخشی",
            is_staff=True,
        )

        self.manager.profile.role = (
            UserProfile.Role.EDUCATION_STAFF
        )
        self.manager.profile.save(
            update_fields=["role"]
        )

        self.student_user = user_model.objects.create_user(
            username="student-user",
            password="StrongPassword123!",
            first_name="سارا",
            last_name="محمدی",
        )

        self.student_user.profile.role = (
            UserProfile.Role.STUDENT
        )
        self.student_user.profile.save(
            update_fields=["role"]
        )

        self.student_data = {
            "id": 1,
            "first_name": "Anita",
            "last_name": "Bakhshi",
            "age": 20,
            "national_code": "0012345678",
            "email": "anita@example.com",
            "phone": "09121234567",
            "profile_image_path": "",
            "created_at": "2026-10-01T10:00:00Z",
            "updated_at": "2026-10-02T11:00:00Z",
        }

        self.valid_form_data = {
            "first_name": "سارا",
            "last_name": "محمدی",
            "age": 21,
            "national_code": "1234567890",
            "email": "sara@example.com",
            "phone": "09123456789",
        }

        self.list_url = reverse("students:list")
        self.create_url = reverse("students:create")

    def login_as_manager(self):
        self.client.force_login(self.manager)

    def test_anonymous_user_is_redirected_to_login(self):
        response = self.client.get(self.list_url)

        expected_url = (
            f"{reverse('accounts:login')}"
            f"?next={self.list_url}"
        )

        self.assertRedirects(
            response,
            expected_url,
            fetch_redirect_response=False,
        )

    def test_student_role_cannot_access_management(self):
        self.client.force_login(self.student_user)

        response = self.client.get(self.list_url)

        self.assertEqual(
            response.status_code,
            403,
        )

    @patch("student_management.views.go_api.get")
    def test_manager_can_view_student_list(
        self,
        mock_get,
    ):
        self.login_as_manager()

        mock_get.return_value = {
            "success": True,
            "message": "Students retrieved successfully",
            "data": [
                self.student_data,
            ],
        }

        response = self.client.get(self.list_url)

        self.assertEqual(
            response.status_code,
            200,
        )

        self.assertTemplateUsed(
            response,
            "students/student_list.html",
        )

        self.assertContains(
            response,
            "Anita Bakhshi",
        )

        self.assertContains(
            response,
            "0012345678",
        )

        self.assertEqual(
            response.context["total_students"],
            1,
        )

        mock_get.assert_called_once_with(
            "/api/v1/students"
        )

    @patch("student_management.views.go_api.get")
    def test_student_list_search_filters_results(
        self,
        mock_get,
    ):
        self.login_as_manager()

        second_student = {
            **self.student_data,
            "id": 2,
            "first_name": "Ali",
            "last_name": "Ahmadi",
            "national_code": "0023456789",
            "email": "ali@example.com",
            "phone": "09121111111",
        }

        mock_get.return_value = {
            "success": True,
            "data": [
                self.student_data,
                second_student,
            ],
        }

        response = self.client.get(
            self.list_url,
            {
                "q": "Ali",
            },
        )

        self.assertEqual(
            response.status_code,
            200,
        )

        self.assertContains(
            response,
            "Ali Ahmadi",
        )

        self.assertNotContains(
            response,
            "Anita Bakhshi",
        )

        self.assertEqual(
            response.context["total_students"],
            1,
        )

    @patch("student_management.views.go_api.get")
    def test_list_page_handles_api_failure(
        self,
        mock_get,
    ):
        self.login_as_manager()

        mock_get.side_effect = GoAPIError(
            "ارتباط با سرویس اصلی برقرار نشد."
        )

        response = self.client.get(self.list_url)

        self.assertEqual(
            response.status_code,
            200,
        )

        self.assertFalse(
            response.context["api_available"]
        )

        self.assertContains(
            response,
            "دریافت اطلاعات دانشجویان ممکن نشد",
        )

        self.assertContains(
            response,
            "ارتباط با سرویس اصلی برقرار نشد.",
        )

    @patch("student_management.views.go_api.get")
    def test_manager_can_view_student_detail(
        self,
        mock_get,
    ):
        self.login_as_manager()

        mock_get.return_value = {
            "success": True,
            "data": self.student_data,
        }

        url = reverse(
            "students:detail",
            args=[1],
        )

        response = self.client.get(url)

        self.assertEqual(
            response.status_code,
            200,
        )

        self.assertTemplateUsed(
            response,
            "students/student_detail.html",
        )

        self.assertContains(
            response,
            "Anita Bakhshi",
        )

        self.assertContains(
            response,
            "anita@example.com",
        )

        mock_get.assert_called_once_with(
            "/api/v1/students/1"
        )

    @patch("student_management.views.go_api.get")
    def test_missing_student_returns_not_found(
        self,
        mock_get,
    ):
        self.login_as_manager()

        mock_get.side_effect = GoAPIError(
            "Student not found",
            status_code=404,
        )

        url = reverse(
            "students:detail",
            args=[404],
        )

        response = self.client.get(url)

        self.assertEqual(
            response.status_code,
            404,
        )

    def test_create_page_is_available_to_manager(self):
        self.login_as_manager()

        response = self.client.get(
            self.create_url
        )

        self.assertEqual(
            response.status_code,
            200,
        )

        self.assertTemplateUsed(
            response,
            "students/student_form.html",
        )

        self.assertContains(
            response,
            "ثبت دانشجوی جدید",
        )

    @patch("student_management.views.go_api.post")
    def test_manager_can_create_student(
        self,
        mock_post,
    ):
        self.login_as_manager()

        mock_post.return_value = {
            "success": True,
            "message": "Student created successfully",
            "data": {
                **self.student_data,
                **self.valid_form_data,
                "id": 3,
            },
        }

        response = self.client.post(
            self.create_url,
            self.valid_form_data,
        )

        self.assertRedirects(
            response,
            reverse(
                "students:detail",
                args=[3],
            ),
            fetch_redirect_response=False,
        )

        mock_post.assert_called_once_with(
            "/api/v1/students",
            json={
                "first_name": "سارا",
                "last_name": "محمدی",
                "age": 21,
                "national_code": "1234567890",
                "email": "sara@example.com",
                "phone": "09123456789",
            },
        )

    @patch("student_management.views.go_api.post")
    def test_create_form_displays_api_error(
        self,
        mock_post,
    ):
        self.login_as_manager()

        mock_post.side_effect = GoAPIError(
            "کد ملی قبلاً ثبت شده است.",
            status_code=409,
        )

        response = self.client.post(
            self.create_url,
            self.valid_form_data,
        )

        self.assertEqual(
            response.status_code,
            400,
        )

        self.assertContains(
            response,
            "کد ملی قبلاً ثبت شده است.",
            status_code=400,
        )

    @patch("student_management.views.go_api.put")
    @patch("student_management.views.go_api.get")
    def test_manager_can_update_student(
        self,
        mock_get,
        mock_put,
    ):
        self.login_as_manager()

        mock_get.return_value = {
            "success": True,
            "data": self.student_data,
        }

        mock_put.return_value = {
            "success": True,
            "message": "Student updated successfully",
            "data": {
                **self.student_data,
                **self.valid_form_data,
            },
        }

        url = reverse(
            "students:update",
            args=[1],
        )

        response = self.client.post(
            url,
            self.valid_form_data,
        )

        self.assertRedirects(
            response,
            reverse(
                "students:detail",
                args=[1],
            ),
            fetch_redirect_response=False,
        )

        mock_put.assert_called_once_with(
            "/api/v1/students/1",
            json={
                "first_name": "سارا",
                "last_name": "محمدی",
                "age": 21,
                "national_code": "1234567890",
                "email": "sara@example.com",
                "phone": "09123456789",
            },
        )

    @patch("student_management.views.go_api.delete")
    @patch("student_management.views.go_api.get")
    def test_manager_can_delete_student(
        self,
        mock_get,
        mock_delete,
    ):
        self.login_as_manager()

        mock_get.return_value = {
            "success": True,
            "data": self.student_data,
        }

        mock_delete.return_value = {
            "success": True,
            "message": "",
            "data": None,
        }

        url = reverse(
            "students:delete",
            args=[1],
        )

        response = self.client.post(url)

        self.assertRedirects(
            response,
            self.list_url,
            fetch_redirect_response=False,
        )

        mock_delete.assert_called_once_with(
            "/api/v1/students/1"
        )

    @patch("student_management.views.go_api.get")
    def test_delete_confirmation_page_is_available(
        self,
        mock_get,
    ):
        self.login_as_manager()

        mock_get.return_value = {
            "success": True,
            "data": self.student_data,
        }

        url = reverse(
            "students:delete",
            args=[1],
        )

        response = self.client.get(url)

        self.assertEqual(
            response.status_code,
            200,
        )

        self.assertTemplateUsed(
            response,
            "students/student_confirm_delete.html",
        )

        self.assertContains(
            response,
            "Anita Bakhshi",
        )