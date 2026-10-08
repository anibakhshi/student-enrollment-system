from unittest.mock import patch

from django.test import TestCase
from django.urls import reverse

from catalog.models import (
    Category,
    CourseCatalog,
    CourseOffering,
    InstructorProfile,
)


class CatalogViewTests(TestCase):
    def setUp(self):
        self.health_patcher = patch(
            "catalog.mixins.go_api.get",
            return_value={
                "success": True,
                "data": {
                    "status": "healthy",
                    "version": "1.0.0",
                },
            },
        )
        self.health_mock = self.health_patcher.start()
        self.addCleanup(self.health_patcher.stop)

        self.programming_category = Category.objects.create(
            name="برنامه‌نویسی",
            slug="programming",
            description="دوره‌های توسعه نرم‌افزار",
            accent_color="#6366F1",
            display_order=10,
            is_active=True,
        )

        self.devops_category = Category.objects.create(
            name="DevOps",
            slug="devops",
            description="دوره‌های زیرساخت و DevOps",
            accent_color="#22D3EE",
            display_order=20,
            is_active=True,
        )

        self.parham = InstructorProfile.objects.create(
            first_name="پرهام",
            last_name="درویشی",
            slug="parham-darvishi",
            expertise="برنامه‌نویسی و توسعه نرم‌افزار",
            is_featured=True,
            is_active=True,
        )

        self.arash = InstructorProfile.objects.create(
            first_name="آرش",
            last_name="فروغی",
            slug="arash-foroughi",
            expertise="DevOps و زیرساخت",
            is_featured=True,
            is_active=True,
        )

        self.go_course = CourseCatalog.objects.create(
            category=self.programming_category,
            title="توسعه REST API با Go",
            slug="go-rest-api",
            short_description=(
                "آموزش طراحی REST API حرفه‌ای با زبان Go"
            ),
            description=(
                "در این دوره طراحی سرویس‌های Backend "
                "با زبان Go آموزش داده می‌شود."
            ),
            price_toman=18_000_000,
            duration_hours=72,
            level=CourseCatalog.Level.INTERMEDIATE,
            is_featured=True,
            is_active=True,
        )

        self.devops_course = CourseCatalog.objects.create(
            category=self.devops_category,
            title="مهندسی DevOps",
            slug="devops-engineering",
            short_description="آموزش ابزارها و فرایندهای DevOps",
            description=(
                "دوره جامع زیرساخت، Docker و فرایندهای CI/CD"
            ),
            price_toman=19_500_000,
            duration_hours=100,
            level=CourseCatalog.Level.ADVANCED,
            is_active=True,
        )

        CourseOffering.objects.create(
            course=self.go_course,
            instructor=self.parham,
            start_date_text="آبان ۱۴۰۵",
            schedule_text="پنجشنبه‌ها ساعت ۱۴ تا ۱۸",
            delivery_mode=CourseOffering.DeliveryMode.HYBRID,
            status=CourseOffering.Status.OPEN,
            capacity=20,
            is_active=True,
        )

        CourseOffering.objects.create(
            course=self.devops_course,
            instructor=self.arash,
            start_date_text="زمستان ۱۴۰۵",
            schedule_text="پنجشنبه‌ها ساعت ۱۴ تا ۱۹",
            delivery_mode=CourseOffering.DeliveryMode.HYBRID,
            status=CourseOffering.Status.UPCOMING,
            capacity=18,
            is_active=True,
        )

    def test_course_list_page_displays_active_courses(self):
        response = self.client.get(
            reverse("catalog:course_list")
        )

        self.assertEqual(response.status_code, 200)
        self.assertTemplateUsed(
            response,
            "catalog/course_list.html",
        )
        self.assertContains(
            response,
            "توسعه REST API با Go",
        )
        self.assertContains(
            response,
            "مهندسی DevOps",
        )
        self.assertEqual(
            response.context["active_page"],
            "courses",
        )
        self.assertTrue(
            response.context["api_available"]
        )

    def test_course_search_finds_matching_course(self):
        response = self.client.get(
            reverse("catalog:course_list"),
            {"q": "Go"},
        )

        self.assertEqual(response.status_code, 200)
        self.assertContains(
            response,
            "توسعه REST API با Go",
        )
        self.assertNotContains(
            response,
            "مهندسی DevOps",
        )

    def test_course_filters_by_category_and_instructor(self):
        category_response = self.client.get(
            reverse("catalog:course_list"),
            {"category": "devops"},
        )

        self.assertContains(
            category_response,
            "مهندسی DevOps",
        )
        self.assertNotContains(
            category_response,
            "توسعه REST API با Go",
        )

        instructor_response = self.client.get(
            reverse("catalog:course_list"),
            {"instructor": "parham-darvishi"},
        )

        self.assertContains(
            instructor_response,
            "توسعه REST API با Go",
        )
        self.assertNotContains(
            instructor_response,
            "مهندسی DevOps",
        )

    def test_course_detail_displays_course_information(self):
        response = self.client.get(
            reverse(
                "catalog:course_detail",
                args=[self.go_course.slug],
            )
        )

        self.assertEqual(response.status_code, 200)
        self.assertTemplateUsed(
            response,
            "catalog/course_detail.html",
        )
        self.assertContains(
            response,
            self.go_course.title,
        )
        self.assertEqual(
            response.context["page_title"],
            self.go_course.title,
        )

        related_courses = list(
            response.context["related_courses"]
        )

        self.assertNotIn(
            self.go_course,
            related_courses,
        )

    def test_inactive_course_returns_not_found(self):
        self.go_course.is_active = False
        self.go_course.save(update_fields=["is_active"])

        response = self.client.get(
            reverse(
                "catalog:course_detail",
                args=[self.go_course.slug],
            )
        )

        self.assertEqual(response.status_code, 404)

    def test_instructor_list_and_detail_pages(self):
        list_response = self.client.get(
            reverse("catalog:instructor_list")
        )

        self.assertEqual(list_response.status_code, 200)
        self.assertTemplateUsed(
            list_response,
            "catalog/instructor_list.html",
        )
        self.assertContains(
            list_response,
            "پرهام درویشی",
        )
        self.assertContains(
            list_response,
            "آرش فروغی",
        )

        detail_response = self.client.get(
            reverse(
                "catalog:instructor_detail",
                args=[self.parham.slug],
            )
        )

        self.assertEqual(detail_response.status_code, 200)
        self.assertTemplateUsed(
            detail_response,
            "catalog/instructor_detail.html",
        )
        self.assertContains(
            detail_response,
            "پرهام درویشی",
        )
        self.assertEqual(
            detail_response.context["active_course_count"],
            1,
        )
        self.assertEqual(
            detail_response.context["active_page"],
            "instructors",
        )