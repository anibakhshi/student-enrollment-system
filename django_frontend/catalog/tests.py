from io import StringIO

from django.core.management import call_command
from django.test import TestCase

from catalog.models import (
    Category,
    CourseCatalog,
    CourseOffering,
    InstructorProfile,
)


class SeedCatalogCommandTests(TestCase):
    def run_seed_command(self):
        output = StringIO()
        call_command("seed_catalog", stdout=output)
        return output.getvalue()

    def test_seed_catalog_creates_initial_data(self):
        output = self.run_seed_command()

        self.assertEqual(Category.objects.count(), 6)
        self.assertEqual(InstructorProfile.objects.count(), 4)
        self.assertEqual(CourseCatalog.objects.count(), 10)
        self.assertEqual(CourseOffering.objects.count(), 12)

        self.assertTrue(
            Category.objects.filter(
                slug="software-development",
                name="برنامه‌نویسی و توسعه نرم‌افزار",
            ).exists()
        )
        self.assertTrue(
            InstructorProfile.objects.filter(
                slug="parham-darvishi",
                first_name="پرهام",
                last_name="درویشی",
                is_featured=True,
                is_active=True,
            ).exists()
        )
        self.assertTrue(
            CourseCatalog.objects.filter(
                slug="golang",
                price_toman=5_500_000,
                duration_hours=15,
                level=CourseCatalog.Level.BEGINNER,
                is_active=True,
            ).exists()
        )
        self.assertTrue(
            CourseOffering.objects.filter(
                course__slug="golang",
                instructor__slug="parham-darvishi",
                status=CourseOffering.Status.OPEN,
                delivery_mode=CourseOffering.DeliveryMode.HYBRID,
            ).exists()
        )
        self.assertIn("10 courses", output)
        self.assertIn("12 offerings", output)

    def test_seed_catalog_is_idempotent(self):
        self.run_seed_command()
        second_output = self.run_seed_command()

        self.assertEqual(Category.objects.count(), 6)
        self.assertEqual(InstructorProfile.objects.count(), 4)
        self.assertEqual(CourseCatalog.objects.count(), 10)
        self.assertEqual(CourseOffering.objects.count(), 12)

        self.assertIn("Updated category", second_output)
        self.assertIn("Updated instructor", second_output)
        self.assertIn("Updated course", second_output)
        self.assertIn("Updated offering", second_output)

    def test_seed_catalog_updates_existing_records(self):
        category = Category.objects.create(
            name="نام قدیمی",
            slug="software-development",
            description="توضیحات قدیمی",
            accent_color="#000000",
            display_order=999,
            is_active=False,
        )
        instructor = InstructorProfile.objects.create(
            first_name="نام قدیمی",
            last_name="نام خانوادگی قدیمی",
            slug="parham-darvishi",
            expertise="تخصص قدیمی",
            is_featured=False,
            is_active=False,
        )
        course = CourseCatalog.objects.create(
            category=category,
            title="عنوان قدیمی",
            slug="golang",
            price_toman=1,
            duration_hours=1,
            is_featured=False,
            is_active=False,
        )

        self.run_seed_command()

        category.refresh_from_db()
        instructor.refresh_from_db()
        course.refresh_from_db()

        self.assertEqual(
            category.name,
            "برنامه‌نویسی و توسعه نرم‌افزار",
        )
        self.assertEqual(category.accent_color, "#6D5DFB")
        self.assertEqual(category.display_order, 10)
        self.assertTrue(category.is_active)

        self.assertEqual(instructor.first_name, "پرهام")
        self.assertEqual(instructor.last_name, "درویشی")
        self.assertEqual(
            instructor.expertise,
            "توسعه و معماری نرم‌افزار",
        )
        self.assertTrue(instructor.is_featured)
        self.assertTrue(instructor.is_active)

        self.assertEqual(course.title, "دوره مقدماتی Golang")
        self.assertEqual(course.price_toman, 5_500_000)
        self.assertEqual(course.duration_hours, 15)
        self.assertEqual(
            course.category.slug,
            "software-development",
        )
        self.assertTrue(course.is_featured)
        self.assertTrue(course.is_active)

    def test_seed_catalog_links_required_instructors_to_courses(self):
        self.run_seed_command()

        expected_course_counts = {
            "parham-darvishi": 3,
            "armin-yaghoubi": 3,
            "sanaz-abbaszadeh": 2,
            "arash-foroughi": 3,
        }

        for instructor_slug, expected_count in expected_course_counts.items():
            actual_count = (
                CourseOffering.objects.filter(
                    instructor__slug=instructor_slug,
                    is_active=True,
                )
                .values("course_id")
                .distinct()
                .count()
            )
            self.assertEqual(
                actual_count,
                expected_count,
                msg=f"Unexpected course count for {instructor_slug}",
            )

    def test_seed_catalog_keeps_unannounced_classes_upcoming(self):
        self.run_seed_command()

        offering = CourseOffering.objects.get(
            course__slug="advanced-python",
            instructor__slug="armin-yaghoubi",
        )

        self.assertEqual(
            offering.status,
            CourseOffering.Status.UPCOMING,
        )
        self.assertEqual(
            offering.start_date_text,
            "در انتظار اعلام",
        )
        self.assertEqual(
            offering.registration_url,
            offering.course.source_url,
        )
