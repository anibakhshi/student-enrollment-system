from io import StringIO

from django.core.management import call_command
from django.test import TestCase

from catalog.models import Category, InstructorProfile


class SeedCatalogCommandTests(TestCase):
    def run_seed_command(self):
        output = StringIO()

        call_command(
            "seed_catalog",
            stdout=output,
        )

        return output.getvalue()

    def test_seed_catalog_creates_initial_data(self):
        output = self.run_seed_command()

        self.assertEqual(Category.objects.count(), 6)
        self.assertEqual(InstructorProfile.objects.count(), 4)

        self.assertTrue(
            Category.objects.filter(
                slug="software-development",
                name="برنامه‌نویسی و توسعه نرم‌افزار",
            ).exists()
        )

        self.assertTrue(
            Category.objects.filter(
                slug="devops-and-cloud",
                name="DevOps و رایانش ابری",
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
            InstructorProfile.objects.filter(
                slug="arash-foroughi",
                first_name="آرش",
                last_name="فروغی",
                is_featured=True,
                is_active=True,
            ).exists()
        )

        self.assertIn(
            "Catalog seed completed successfully",
            output,
        )

    def test_seed_catalog_is_idempotent(self):
        self.run_seed_command()
        second_output = self.run_seed_command()

        self.assertEqual(Category.objects.count(), 6)
        self.assertEqual(InstructorProfile.objects.count(), 4)

        self.assertIn(
            "Updated category",
            second_output,
        )
        self.assertIn(
            "Updated instructor",
            second_output,
        )

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

        self.run_seed_command()

        category.refresh_from_db()
        instructor.refresh_from_db()

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
            "توسعه نرم‌افزار و برنامه‌نویسی",
        )
        self.assertTrue(instructor.is_featured)
        self.assertTrue(instructor.is_active)