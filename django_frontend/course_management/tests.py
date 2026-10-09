from unittest.mock import patch

from django.contrib.auth import get_user_model
from django.test import TestCase
from django.urls import reverse

from dashboard.api_client import GoAPIError


class CourseManagementViewTests(TestCase):
    def setUp(self):
        self.user = get_user_model().objects.create_superuser(
            username="course-admin",
            email="course-admin@example.com",
            password="StrongPassword123!",
            first_name="آنیتا",
            last_name="بخشی",
        )

        self.client.force_login(self.user)

        self.list_url = reverse("course_management:list")
        self.create_url = reverse("course_management:create")
        self.detail_url = reverse(
            "course_management:detail",
            args=[1],
        )
        self.update_url = reverse(
            "course_management:update",
            args=[1],
        )
        self.delete_url = reverse(
            "course_management:delete",
            args=[1],
        )

        self.instructors = [
            {
                "id": 1,
                "first_name": "Parham",
                "last_name": "Darvishi",
                "email": "parham@example.com",
                "expertise": "Software Engineering",
                "status": "active",
            },
            {
                "id": 2,
                "first_name": "Maryam",
                "last_name": "Rahimi",
                "email": "maryam@example.com",
                "expertise": "Data Science",
                "status": "active",
            },
            {
                "id": 3,
                "first_name": "Inactive",
                "last_name": "Instructor",
                "email": "inactive@example.com",
                "expertise": "Networks",
                "status": "inactive",
            },
        ]

        self.courses = [
            {
                "id": 1,
                "instructor_id": 1,
                "code": "GO-201",
                "title": "REST API Development with Go",
                "description": (
                    "Building production-ready REST APIs "
                    "with Go and PostgreSQL"
                ),
                "price": 18000000,
                "capacity": 20,
                "duration_hours": 72,
                "start_date": "2026-10-10T00:00:00Z",
                "end_date": "2027-01-10T00:00:00Z",
                "status": "open",
            },
            {
                "id": 2,
                "instructor_id": 2,
                "code": "DS-101",
                "title": "Data Science Fundamentals",
                "description": (
                    "Introduction to practical data science"
                ),
                "price": 15000000,
                "capacity": 25,
                "duration_hours": 60,
                "start_date": "2026-10-01T00:00:00Z",
                "end_date": "2026-12-15T00:00:00Z",
                "status": "draft",
            },
        ]

        self.valid_form_data = {
            "instructor_id": "1",
            "code": "go-301",
            "title": "Advanced Go Programming",
            "description": "Advanced backend programming.",
            "price": "20000000",
            "capacity": "18",
            "duration_hours": "80",
            "start_date": "2026-11-01",
            "end_date": "2027-02-01",
        }

    def api_get_response(self, path):
        if path == "/api/v1/instructors":
            return {
                "success": True,
                "data": self.instructors,
            }

        if path == "/api/v1/courses":
            return {
                "success": True,
                "data": self.courses,
            }

        if path == "/api/v1/courses/1":
            return {
                "success": True,
                "data": self.courses[0],
            }

        raise AssertionError(
            f"Unexpected mocked API path: {path}"
        )

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
        student = get_user_model().objects.create_user(
            username="course-student",
            password="StrongPassword123!",
        )

        student.profile.role = "student"
        student.profile.save(update_fields=["role"])

        self.client.force_login(student)

        response = self.client.get(self.list_url)

        self.assertEqual(response.status_code, 403)

    @patch("course_management.views.go_api.get")
    def test_course_list_displays_courses(
        self,
        mock_get,
    ):
        mock_get.side_effect = self.api_get_response

        response = self.client.get(self.list_url)

        self.assertEqual(response.status_code, 200)

        self.assertTemplateUsed(
            response,
            "courses_management/course_list.html",
        )

        self.assertEqual(
            response.context["total_courses"],
            2,
        )

        self.assertContains(
            response,
            "REST API Development with Go",
        )

        self.assertContains(
            response,
            "Data Science Fundamentals",
        )

        self.assertContains(
            response,
            "Parham Darvishi",
        )

        self.assertTrue(
            response.context["api_available"]
        )

    @patch("course_management.views.go_api.get")
    def test_course_list_searches_by_title_and_instructor(
        self,
        mock_get,
    ):
        mock_get.side_effect = self.api_get_response

        title_response = self.client.get(
            self.list_url,
            {"q": "Advanced Go"},
        )

        self.assertEqual(title_response.status_code, 200)
        self.assertEqual(
            title_response.context["total_courses"],
            0,
        )

        instructor_response = self.client.get(
            self.list_url,
            {"q": "Parham"},
        )

        self.assertEqual(
            instructor_response.context["total_courses"],
            1,
        )

        self.assertContains(
            instructor_response,
            "REST API Development with Go",
        )

    @patch("course_management.views.go_api.get")
    def test_course_list_filters_by_status_and_instructor(
        self,
        mock_get,
    ):
        mock_get.side_effect = self.api_get_response

        response = self.client.get(
            self.list_url,
            {
                "status": "open",
                "instructor": "1",
            },
        )

        self.assertEqual(response.status_code, 200)

        self.assertEqual(
            response.context["total_courses"],
            1,
        )

        self.assertContains(
            response,
            "REST API Development with Go",
        )

        self.assertNotContains(
            response,
            "Data Science Fundamentals",
        )

    @patch("course_management.views.go_api.get")
    def test_course_list_handles_api_failure(
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
            response.context["total_courses"],
            0,
        )

        self.assertContains(
            response,
            "ارتباط با سرویس اصلی برقرار نشد.",
        )

    @patch("course_management.views.go_api.get")
    def test_course_detail_displays_course_information(
        self,
        mock_get,
    ):
        mock_get.side_effect = self.api_get_response

        response = self.client.get(self.detail_url)

        self.assertEqual(response.status_code, 200)

        self.assertTemplateUsed(
            response,
            "courses_management/course_detail.html",
        )

        self.assertEqual(
            response.context["course"]["id"],
            1,
        )

        self.assertEqual(
            response.context["course"]["status_label"],
            "باز برای ثبت‌نام",
        )

        self.assertEqual(
            response.context["course"]["instructor_name"],
            "Parham Darvishi",
        )

        self.assertContains(
            response,
            "REST API Development with Go",
        )

        self.assertContains(
            response,
            "GO-201",
        )

    @patch("course_management.views.go_api.get")
    def test_missing_course_returns_not_found(
        self,
        mock_get,
    ):
        def side_effect(path):
            if path == "/api/v1/instructors":
                return {
                    "success": True,
                    "data": self.instructors,
                }

            raise GoAPIError(
                "دوره موردنظر پیدا نشد.",
                status_code=404,
            )

        mock_get.side_effect = side_effect

        response = self.client.get(
            reverse(
                "course_management:detail",
                args=[999],
            )
        )

        self.assertEqual(response.status_code, 404)

    @patch("course_management.views.go_api.get")
    def test_create_page_only_lists_active_instructors(
        self,
        mock_get,
    ):
        mock_get.side_effect = self.api_get_response

        response = self.client.get(self.create_url)

        self.assertEqual(response.status_code, 200)

        form = response.context["form"]
        choices = list(
            form.fields["instructor_id"].choices
        )

        choice_values = {
            str(value)
            for value, _label in choices
            if value
        }

        self.assertEqual(choice_values, {"1", "2"})
        self.assertNotIn("status", form.fields)

    @patch("course_management.views.go_api.post")
    @patch("course_management.views.go_api.get")
    def test_create_course_sends_expected_payload(
        self,
        mock_get,
        mock_post,
    ):
        mock_get.side_effect = self.api_get_response

        mock_post.return_value = {
            "success": True,
            "data": {
                "id": 3,
                **self.valid_form_data,
            },
        }

        response = self.client.post(
            self.create_url,
            self.valid_form_data,
        )

        self.assertRedirects(
            response,
            reverse(
                "course_management:detail",
                args=[3],
            ),
            fetch_redirect_response=False,
        )

        mock_post.assert_called_once_with(
            "/api/v1/courses",
            json={
                "instructor_id": 1,
                "code": "GO-301",
                "title": "Advanced Go Programming",
                "description": (
                    "Advanced backend programming."
                ),
                "price": 20000000,
                "capacity": 18,
                "duration_hours": 80,
                "start_date": "2026-11-01",
                "end_date": "2027-02-01",
            },
        )

    @patch("course_management.views.go_api.post")
    @patch("course_management.views.go_api.get")
    def test_create_course_rejects_invalid_date_range(
        self,
        mock_get,
        mock_post,
    ):
        mock_get.side_effect = self.api_get_response

        form_data = {
            **self.valid_form_data,
            "start_date": "2027-02-01",
            "end_date": "2026-11-01",
        }

        response = self.client.post(
            self.create_url,
            form_data,
        )

        self.assertEqual(response.status_code, 400)

        form = response.context["form"]

        self.assertIn("end_date", form.errors)

        self.assertIn(
            "تاریخ پایان نمی‌تواند قبل از تاریخ شروع باشد.",
            form.errors["end_date"],
        )

        mock_post.assert_not_called()

    @patch("course_management.views.go_api.post")
    @patch("course_management.views.go_api.get")
    def test_create_course_displays_api_field_errors(
        self,
        mock_get,
        mock_post,
    ):
        mock_get.side_effect = self.api_get_response

        mock_post.side_effect = GoAPIError(
            "اطلاعات دوره معتبر نیست.",
            status_code=422,
            payload={
                "message": "اطلاعات دوره معتبر نیست.",
                "errors": {
                    "code": [
                        "کد دوره قبلاً استفاده شده است."
                    ],
                },
            },
        )

        response = self.client.post(
            self.create_url,
            self.valid_form_data,
        )

        self.assertEqual(response.status_code, 400)

        form = response.context["form"]

        self.assertIn("code", form.errors)

        self.assertIn(
            "کد دوره قبلاً استفاده شده است.",
            form.errors["code"],
        )

    @patch("course_management.views.go_api.get")
    def test_update_page_is_populated_with_course_data(
        self,
        mock_get,
    ):
        mock_get.side_effect = self.api_get_response

        response = self.client.get(self.update_url)

        self.assertEqual(response.status_code, 200)

        form = response.context["form"]

        self.assertEqual(
            form.initial["code"],
            "GO-201",
        )

        self.assertEqual(
            form.initial["instructor_id"],
            1,
        )

        self.assertEqual(
            form.initial["start_date"],
            "2026-10-10",
        )

        self.assertEqual(
            form.initial["status"],
            "open",
        )

        self.assertIn("status", form.fields)

    @patch("course_management.views.go_api.put")
    @patch("course_management.views.go_api.get")
    def test_update_course_sends_status(
        self,
        mock_get,
        mock_put,
    ):
        mock_get.side_effect = self.api_get_response

        update_data = {
            **self.valid_form_data,
            "status": "closed",
        }

        response = self.client.post(
            self.update_url,
            update_data,
        )

        self.assertRedirects(
            response,
            self.detail_url,
            fetch_redirect_response=False,
        )

        mock_put.assert_called_once_with(
            "/api/v1/courses/1",
            json={
                "instructor_id": 1,
                "code": "GO-301",
                "title": "Advanced Go Programming",
                "description": (
                    "Advanced backend programming."
                ),
                "price": 20000000,
                "capacity": 18,
                "duration_hours": 80,
                "start_date": "2026-11-01",
                "end_date": "2027-02-01",
                "status": "closed",
            },
        )

    @patch("course_management.views.go_api.get")
    def test_delete_confirmation_displays_course(
        self,
        mock_get,
    ):
        mock_get.side_effect = self.api_get_response

        response = self.client.get(self.delete_url)

        self.assertEqual(response.status_code, 200)

        self.assertTemplateUsed(
            response,
            (
                "courses_management/"
                "course_confirm_delete.html"
            ),
        )

        self.assertContains(
            response,
            "REST API Development with Go",
        )

        self.assertContains(
            response,
            "آیا از حذف این دوره مطمئن هستید؟",
        )

    @patch("course_management.views.go_api.delete")
    @patch("course_management.views.go_api.get")
    def test_delete_course_calls_api(
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
            "/api/v1/courses/1"
        )

    @patch("course_management.views.go_api.delete")
    @patch("course_management.views.go_api.get")
    def test_delete_failure_redirects_to_detail(
        self,
        mock_get,
        mock_delete,
    ):
        mock_get.side_effect = self.api_get_response

        mock_delete.side_effect = GoAPIError(
            "این دوره دارای ثبت‌نام فعال است.",
            status_code=409,
        )

        response = self.client.post(self.delete_url)

        self.assertRedirects(
            response,
            self.detail_url,
            fetch_redirect_response=False,
        )