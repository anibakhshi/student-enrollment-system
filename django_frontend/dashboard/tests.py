from unittest.mock import patch

from django.contrib.auth import get_user_model
from django.test import TestCase
from django.urls import reverse

from .api_client import GoAPIError


class DashboardViewTests(TestCase):
    def setUp(self):
        self.url = reverse("dashboard:index")

        self.user = get_user_model().objects.create_user(
            username="dashboard-admin",
            password="StrongPassword123!",
            first_name="آنیتا",
            last_name="بخشی",
            is_staff=True,
        )

        self.client.force_login(self.user)

    def test_dashboard_requires_authentication(self):
        self.client.logout()

        response = self.client.get(self.url)

        expected_url = (
            f"{reverse('accounts:login')}"
            f"?next={self.url}"
        )

        self.assertRedirects(
            response,
            expected_url,
            fetch_redirect_response=False,
        )

    @patch("dashboard.views.go_api.get")
    def test_dashboard_displays_api_data(
        self,
        mock_get,
    ):
        responses = {
            "/health": {
                "success": True,
                "data": {
                    "service": "student-enrollment-api",
                    "status": "healthy",
                    "version": "1.0.0",
                },
            },
            "/api/v1/students": {
                "success": True,
                "data": [
                    {
                        "id": 1,
                        "first_name": "Anita",
                        "last_name": "Bakhshi",
                    },
                    {
                        "id": 2,
                        "first_name": "Ali",
                        "last_name": "Ahmadi",
                    },
                ],
            },
            "/api/v1/instructors": {
                "success": True,
                "data": [
                    {
                        "id": 1,
                        "first_name": "Parham",
                        "last_name": "Darvishi",
                    }
                ],
            },
            "/api/v1/courses": {
                "success": True,
                "data": [
                    {
                        "id": 1,
                        "title": (
                            "REST API Development with Go"
                        ),
                    }
                ],
            },
            "/api/v1/enrollments": {
                "success": True,
                "data": [
                    {
                        "id": 1,
                        "student_id": 1,
                        "course_id": 1,
                        "status": "confirmed",
                        "created_at": (
                            "2026-10-03T10:00:00Z"
                        ),
                    }
                ],
            },
            "/api/v1/payments": {
                "success": True,
                "data": [
                    {
                        "id": 1,
                        "enrollment_id": 1,
                        "amount": 15000000,
                        "status": "succeeded",
                        "created_at": (
                            "2026-10-03T10:05:00Z"
                        ),
                    }
                ],
            },
        }

        mock_get.side_effect = (
            lambda path: responses[path]
        )

        response = self.client.get(self.url)

        self.assertEqual(response.status_code, 200)

        self.assertTemplateUsed(
            response,
            "dashboard/index.html",
        )

        self.assertTrue(
            response.context["api_available"]
        )

        self.assertEqual(
            response.context["stats"]["students"],
            2,
        )

        self.assertEqual(
            response.context["stats"]["instructors"],
            1,
        )

        self.assertEqual(
            response.context["stats"]["courses"],
            1,
        )

        self.assertEqual(
            response.context["stats"]["enrollments"],
            1,
        )

        self.assertEqual(
            response.context["stats"]["payments"],
            1,
        )

        self.assertEqual(
            response.context["stats"]["total_revenue"],
            "15,000,000",
        )

        self.assertContains(
            response,
            "Anita Bakhshi",
        )

        self.assertContains(
            response,
            "REST API Development with Go",
        )

        self.assertContains(
            response,
            "ارتباط با سرویس برقرار است",
        )

    @patch("dashboard.views.go_api.get")
    def test_dashboard_handles_api_failure(
        self,
        mock_get,
    ):
        mock_get.side_effect = GoAPIError(
            "ارتباط با سرویس اصلی برقرار نشد."
        )

        response = self.client.get(self.url)

        self.assertEqual(response.status_code, 200)

        self.assertFalse(
            response.context["api_available"]
        )

        self.assertEqual(
            response.context["stats"]["students"],
            0,
        )

        self.assertContains(
            response,
            "Go API قطع",
        )

        self.assertContains(
            response,
            "بخشی از اطلاعات دریافت نشد",
        )