from unittest.mock import patch

from django.contrib.auth import get_user_model
from django.test import TestCase
from django.urls import reverse

from dashboard.api_client import GoAPIError


class EnrollmentManagementViewTests(TestCase):
    def setUp(self):
        self.user = get_user_model().objects.create_superuser(
            username="enrollment-admin",
            email="enrollment-admin@example.com",
            password="StrongPassword123!",
            first_name="آنیتا",
            last_name="بخشی",
        )

        self.client.force_login(self.user)

        self.list_url = reverse(
            "enrollment_management:list"
        )
        self.create_url = reverse(
            "enrollment_management:create"
        )
        self.detail_url = reverse(
            "enrollment_management:detail",
            args=[4],
        )
        self.update_url = reverse(
            "enrollment_management:update",
            args=[4],
        )
        self.delete_url = reverse(
            "enrollment_management:delete",
            args=[4],
        )

        self.students = [
            {
                "id": 1,
                "first_name": "Anita",
                "last_name": "Bakhshi",
                "age": 20,
                "national_code": "0012345678",
                "email": "anita@example.com",
                "phone": "09121234567",
            },
            {
                "id": 2,
                "first_name": "Ali",
                "last_name": "Ahmadi",
                "age": 22,
                "national_code": "0023456789",
                "email": "ali@example.com",
                "phone": "09129876543",
            },
        ]

        self.courses = [
            {
                "id": 1,
                "instructor_id": 1,
                "code": "DS-101",
                "title": "Data Science Fundamentals",
                "description": (
                    "Introduction to data science"
                ),
                "price": 15000000,
                "capacity": 25,
                "duration_hours": 60,
                "start_date": "2027-10-01T00:00:00Z",
                "end_date": "2027-12-15T00:00:00Z",
                "status": "open",
            },
            {
                "id": 2,
                "instructor_id": 2,
                "code": "GO-201",
                "title": "REST API Development with Go",
                "description": (
                    "Building production-ready APIs"
                ),
                "price": 18000000,
                "capacity": 20,
                "duration_hours": 72,
                "start_date": "2027-11-01T00:00:00Z",
                "end_date": "2028-01-10T00:00:00Z",
                "status": "draft",
            },
        ]

        self.enrollments = [
            {
                "id": 4,
                "student_id": 1,
                "course_id": 1,
                "status": "pending",
                "enrolled_at": (
                    "2026-09-18T15:22:40Z"
                ),
                "notes": "Waiting for confirmation",
                "created_at": (
                    "2026-09-18T15:22:40Z"
                ),
                "updated_at": (
                    "2026-09-18T15:22:40Z"
                ),
            },
            {
                "id": 5,
                "student_id": 2,
                "course_id": 2,
                "status": "completed",
                "enrolled_at": (
                    "2026-08-01T10:00:00Z"
                ),
                "confirmed_at": (
                    "2026-08-02T10:00:00Z"
                ),
                "completed_at": (
                    "2026-09-01T10:00:00Z"
                ),
                "notes": "Course completed",
                "created_at": (
                    "2026-08-01T10:00:00Z"
                ),
                "updated_at": (
                    "2026-09-01T10:00:00Z"
                ),
            },
        ]

        self.enrollment_details = {
            "enrollment": self.enrollments[0],
            "student": self.students[0],
            "course": self.courses[0],
        }

    def api_get_response(self, path):
        responses = {
            "/api/v1/students": {
                "success": True,
                "data": self.students,
            },
            "/api/v1/courses": {
                "success": True,
                "data": self.courses,
            },
            "/api/v1/enrollments": {
                "success": True,
                "data": self.enrollments,
            },
            "/api/v1/enrollments/4": {
                "success": True,
                "data": self.enrollment_details,
            },
        }

        if path not in responses:
            raise AssertionError(
                f"Unexpected mocked API path: {path}"
            )

        return responses[path]

    def test_anonymous_user_is_redirected_to_login(self):
        self.client.logout()

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
        student_user = (
            get_user_model().objects.create_user(
                username="enrollment-student",
                password="StrongPassword123!",
            )
        )

        student_user.profile.role = "student"
        student_user.profile.save(
            update_fields=["role"]
        )

        self.client.force_login(student_user)

        response = self.client.get(self.list_url)

        self.assertEqual(response.status_code, 403)

    @patch("enrollment_management.views.go_api.get")
    def test_enrollment_list_displays_data(
        self,
        mock_get,
    ):
        mock_get.side_effect = self.api_get_response

        response = self.client.get(self.list_url)

        self.assertEqual(response.status_code, 200)

        self.assertTemplateUsed(
            response,
            (
                "enrollments_management/"
                "enrollment_list.html"
            ),
        )

        self.assertEqual(
            response.context["total_enrollments"],
            2,
        )

        self.assertTrue(
            response.context["api_available"]
        )

        self.assertContains(response, "Anita Bakhshi")

        self.assertContains(
            response,
            "Data Science Fundamentals",
        )

        self.assertContains(
            response,
            "در انتظار تأیید",
        )

    @patch("enrollment_management.views.go_api.get")
    def test_enrollment_list_filters_status(
        self,
        mock_get,
    ):
        mock_get.side_effect = self.api_get_response

        response = self.client.get(
            self.list_url,
            {"status": "completed"},
        )

        self.assertEqual(response.status_code, 200)

        self.assertEqual(
            response.context["total_enrollments"],
            1,
        )

        self.assertContains(response, "Ali Ahmadi")

        enrollments = response.context["enrollments"]

        self.assertEqual(len(enrollments), 1)
        self.assertEqual(
            enrollments[0]["student_name"],
            "Ali Ahmadi",
        )
        self.assertEqual(
            enrollments[0]["status"],
            "completed",
        )

    @patch("enrollment_management.views.go_api.get")
    def test_enrollment_list_filters_student_and_course(
        self,
        mock_get,
    ):
        mock_get.side_effect = self.api_get_response

        response = self.client.get(
            self.list_url,
            {
                "student": "1",
                "course": "1",
            },
        )

        self.assertEqual(response.status_code, 200)

        self.assertEqual(
            response.context["total_enrollments"],
            1,
        )

        self.assertContains(response, "Anita Bakhshi")

    @patch("enrollment_management.views.go_api.get")
    def test_enrollment_list_searches_notes(
        self,
        mock_get,
    ):
        mock_get.side_effect = self.api_get_response

        response = self.client.get(
            self.list_url,
            {"q": "Waiting"},
        )

        self.assertEqual(response.status_code, 200)

        self.assertEqual(
            response.context["total_enrollments"],
            1,
        )

        self.assertContains(response, "Anita Bakhshi")

    @patch("enrollment_management.views.go_api.get")
    def test_enrollment_list_handles_api_failure(
        self,
        mock_get,
    ):
        mock_get.side_effect = GoAPIError(
            "ارتباط با سرویس اصلی برقرار نشد."
        )

        response = self.client.get(self.list_url)

        self.assertEqual(response.status_code, 200)

        self.assertFalse(
            response.context["api_available"]
        )

        self.assertEqual(
            response.context["total_enrollments"],
            0,
        )

        self.assertContains(
            response,
            "ارتباط با سرویس اصلی برقرار نشد.",
        )

    @patch("enrollment_management.views.go_api.get")
    def test_enrollment_detail_displays_information(
        self,
        mock_get,
    ):
        mock_get.side_effect = self.api_get_response

        response = self.client.get(self.detail_url)

        self.assertEqual(response.status_code, 200)

        self.assertTemplateUsed(
            response,
            (
                "enrollments_management/"
                "enrollment_detail.html"
            ),
        )

        enrollment = response.context["enrollment"]

        self.assertEqual(enrollment["id"], 4)

        self.assertEqual(
            enrollment["student_name"],
            "Anita Bakhshi",
        )

        self.assertEqual(
            enrollment["course_name"],
            "DS-101 — Data Science Fundamentals",
        )

        self.assertEqual(
            enrollment["status_label"],
            "در انتظار تأیید",
        )

    @patch("enrollment_management.views.go_api.get")
    def test_missing_enrollment_returns_not_found(
        self,
        mock_get,
    ):
        mock_get.side_effect = GoAPIError(
            "Enrollment not found",
            status_code=404,
        )

        response = self.client.get(
            reverse(
                "enrollment_management:detail",
                args=[999],
            )
        )

        self.assertEqual(response.status_code, 404)

    @patch("enrollment_management.views.go_api.get")
    def test_create_page_lists_students_and_open_courses(
        self,
        mock_get,
    ):
        mock_get.side_effect = self.api_get_response

        response = self.client.get(self.create_url)

        self.assertEqual(response.status_code, 200)

        form = response.context["form"]

        student_choices = {
            str(value)
            for value, _label
            in form.fields["student_id"].choices
            if value
        }

        course_choices = {
            str(value)
            for value, _label
            in form.fields["course_id"].choices
            if value
        }

        self.assertEqual(
            student_choices,
            {"1", "2"},
        )

        self.assertEqual(
            course_choices,
            {"1"},
        )

    @patch("enrollment_management.views.go_api.post")
    @patch("enrollment_management.views.go_api.get")
    def test_create_enrollment_posts_payload(
        self,
        mock_get,
        mock_post,
    ):
        mock_get.side_effect = self.api_get_response

        mock_post.return_value = {
            "success": True,
            "data": {
                "id": 6,
                "student_id": 2,
                "course_id": 1,
                "status": "pending",
                "notes": "New enrollment",
            },
        }

        response = self.client.post(
            self.create_url,
            {
                "student_id": "2",
                "course_id": "1",
                "notes": " New enrollment ",
            },
        )

        self.assertRedirects(
            response,
            reverse(
                "enrollment_management:detail",
                args=[6],
            ),
            fetch_redirect_response=False,
        )

        mock_post.assert_called_once_with(
            "/api/v1/enrollments",
            json={
                "student_id": 2,
                "course_id": 1,
                "notes": "New enrollment",
            },
        )

    @patch("enrollment_management.views.go_api.post")
    @patch("enrollment_management.views.go_api.get")
    def test_create_displays_api_field_error(
        self,
        mock_get,
        mock_post,
    ):
        mock_get.side_effect = self.api_get_response

        mock_post.side_effect = GoAPIError(
            "Validation failed",
            status_code=400,
            payload={
                "success": False,
                "message": "Validation failed",
                "errors": {
                    "course_id": (
                        "Course capacity has been reached"
                    ),
                },
            },
        )

        response = self.client.post(
            self.create_url,
            {
                "student_id": "1",
                "course_id": "1",
                "notes": "",
            },
        )

        self.assertEqual(response.status_code, 400)

        form = response.context["form"]

        self.assertIn("course_id", form.errors)

        self.assertIn(
            "Course capacity has been reached",
            form.errors["course_id"],
        )

    @patch("enrollment_management.views.go_api.get")
    def test_update_page_has_valid_transitions(
        self,
        mock_get,
    ):
        mock_get.side_effect = self.api_get_response

        response = self.client.get(self.update_url)

        self.assertEqual(response.status_code, 200)

        form = response.context["form"]

        choices = {
            value
            for value, _label
            in form.fields["status"].choices
        }

        self.assertEqual(
            choices,
            {
                "pending",
                "confirmed",
                "cancelled",
            },
        )

        self.assertEqual(
            form.initial["status"],
            "pending",
        )

    @patch("enrollment_management.views.go_api.put")
    @patch("enrollment_management.views.go_api.get")
    def test_update_enrollment_sends_payload(
        self,
        mock_get,
        mock_put,
    ):
        mock_get.side_effect = self.api_get_response

        mock_put.return_value = {
            "success": True,
            "data": {
                **self.enrollments[0],
                "status": "confirmed",
                "notes": "Enrollment confirmed",
            },
        }

        response = self.client.post(
            self.update_url,
            {
                "status": "confirmed",
                "notes": " Enrollment confirmed ",
            },
        )

        self.assertRedirects(
            response,
            self.detail_url,
            fetch_redirect_response=False,
        )

        mock_put.assert_called_once_with(
            "/api/v1/enrollments/4",
            json={
                "status": "confirmed",
                "notes": "Enrollment confirmed",
            },
        )

    @patch("enrollment_management.views.go_api.put")
    @patch("enrollment_management.views.go_api.get")
    def test_terminal_enrollment_cannot_be_updated(
        self,
        mock_get,
        mock_put,
    ):
        terminal_enrollment = {
            **self.enrollments[0],
            "status": "completed",
            "completed_at": (
                "2026-09-20T10:00:00Z"
            ),
        }

        terminal_details = {
            "enrollment": terminal_enrollment,
            "student": self.students[0],
            "course": self.courses[0],
        }

        mock_get.return_value = {
            "success": True,
            "data": terminal_details,
        }

        response = self.client.post(
            self.update_url,
            {
                "status": "completed",
                "notes": "Cannot change",
            },
        )

        self.assertRedirects(
            response,
            self.detail_url,
            fetch_redirect_response=False,
        )

        mock_put.assert_not_called()

    @patch("enrollment_management.views.go_api.get")
    def test_delete_confirmation_page(
        self,
        mock_get,
    ):
        mock_get.side_effect = self.api_get_response

        response = self.client.get(self.delete_url)

        self.assertEqual(response.status_code, 200)

        self.assertTemplateUsed(
            response,
            (
                "enrollments_management/"
                "enrollment_confirm_delete.html"
            ),
        )

        self.assertContains(
            response,
            "آیا از حذف این ثبت‌نام مطمئن هستید؟",
        )

        self.assertContains(response, "Anita Bakhshi")

    @patch("enrollment_management.views.go_api.delete")
    @patch("enrollment_management.views.go_api.get")
    def test_delete_enrollment_calls_api(
        self,
        mock_get,
        mock_delete,
    ):
        mock_get.side_effect = self.api_get_response

        mock_delete.return_value = {
            "success": True,
            "data": None,
        }

        response = self.client.post(self.delete_url)

        self.assertRedirects(
            response,
            self.list_url,
            fetch_redirect_response=False,
        )

        mock_delete.assert_called_once_with(
            "/api/v1/enrollments/4"
        )

    @patch("enrollment_management.views.go_api.delete")
    @patch("enrollment_management.views.go_api.get")
    def test_delete_failure_redirects_to_detail(
        self,
        mock_get,
        mock_delete,
    ):
        mock_get.side_effect = self.api_get_response

        mock_delete.side_effect = GoAPIError(
            "حذف ثبت‌نام انجام نشد.",
            status_code=409,
        )

        response = self.client.post(self.delete_url)

        self.assertRedirects(
            response,
            self.detail_url,
            fetch_redirect_response=False,
        )