from io import StringIO
from unittest.mock import patch

from bs4 import BeautifulSoup
from django.core.management import call_command
from django.test import TestCase

from catalog.management.commands.sync_sematec_catalog import RemotePage
from catalog.models import (
    Category,
    CourseCatalog,
    CourseOffering,
    InstructorProfile,
)


class SyncSematecCatalogCommandTests(TestCase):
    base_url = "https://sematec.test"

    @classmethod
    def setUpTestData(cls):
        categories = [
            ("software-development", "برنامه‌نویسی"),
            ("data-and-ai", "داده و هوش مصنوعی"),
            ("network-and-infrastructure", "شبکه"),
            ("devops-and-cloud", "DevOps"),
            ("cyber-security", "امنیت"),
            ("business-and-management", "مدیریت"),
        ]
        for order, (slug, name) in enumerate(categories, start=1):
            Category.objects.create(
                slug=slug,
                name=name,
                display_order=order,
            )

    def fake_page(self, client, path_or_url):
        url = path_or_url
        if not url.startswith("http"):
            url = f"{self.base_url}{path_or_url}"

        if "/teacher/page/" in url or url.endswith("/teacher/"):
            html = """
                <html><body>
                    <a href="/teacher/parham-darvishi/">
                        پرهام درویشی
                    </a>
                </body></html>
            """
        elif url.endswith("/teacher/parham-darvishi/"):
            html = """
                <html><head>
                    <meta property="og:image"
                          content="/media/parham.jpg">
                </head><body>
                    <main>
                        <h1>پرهام درویشی</h1>
                        <p>توسعه و معماری نرم‌افزار</p>
                        <h2>پرهام درویشی کیست؟</h2>
                        <p>مدرس و متخصص توسعه نرم‌افزار.</p>
                        <h2>دوره‌های استاد</h2>
                        <h3>
                            <a href="/course/golang/">
                                دوره مقدماتی Golang
                            </a>
                        </h3>
                    </main>
                </body></html>
            """
        elif url.endswith("/course/"):
            html = """
                <html><body>
                    <a href="/course/golang/">Golang</a>
                </body></html>
            """
        elif url.endswith("/course/golang/"):
            html = """
                <html><head>
                    <meta name="description"
                          content="آموزش زبان Go برای توسعه Backend">
                    <meta property="og:image"
                          content="/media/golang.jpg">
                </head><body>
                    <h1>دوره مقدماتی Golang</h1>
                    <p>5,500,000 تومان</p>
                    <p>زمان دوره | 15 ساعت</p>
                    <p>
                        پیش نیاز | آشنایی با فلوچارت و الگوریتم
                        کلاس‌های فعال
                    </p>
                </body></html>
            """
        else:
            self.fail(f"Unexpected URL requested: {url}")

        return RemotePage(
            url=url,
            soup=BeautifulSoup(html, "html.parser"),
        )

    def run_sync(self, *extra_arguments):
        output = StringIO()
        error_output = StringIO()

        with patch(
            "catalog.management.commands.sync_sematec_catalog."
            "SematecClient.get_page",
            autospec=True,
            side_effect=self.fake_page,
        ):
            call_command(
                "sync_sematec_catalog",
                "--base-url",
                self.base_url,
                "--delay",
                "0",
                *extra_arguments,
                stdout=output,
                stderr=error_output,
            )

        return output.getvalue()

    def test_sync_creates_instructor_course_and_relation(self):
        output = self.run_sync()

        self.assertEqual(InstructorProfile.objects.count(), 1)
        self.assertEqual(CourseCatalog.objects.count(), 1)
        self.assertEqual(CourseOffering.objects.count(), 1)

        instructor = InstructorProfile.objects.get()
        course = CourseCatalog.objects.get()
        offering = CourseOffering.objects.get()

        self.assertEqual(instructor.full_name, "پرهام درویشی")
        self.assertEqual(
            instructor.source_url,
            f"{self.base_url}/teacher/parham-darvishi/",
        )
        self.assertEqual(course.slug, "golang")
        self.assertEqual(course.price_toman, 5_500_000)
        self.assertEqual(course.duration_hours, 15)
        self.assertEqual(
            course.level,
            CourseCatalog.Level.BEGINNER,
        )
        self.assertEqual(
            course.category.slug,
            "software-development",
        )
        self.assertEqual(offering.course, course)
        self.assertEqual(offering.instructor, instructor)
        self.assertEqual(
            offering.status,
            CourseOffering.Status.UPCOMING,
        )
        self.assertIn("Sync completed", output)

    def test_sync_is_idempotent(self):
        self.run_sync()
        second_output = self.run_sync()

        self.assertEqual(InstructorProfile.objects.count(), 1)
        self.assertEqual(CourseCatalog.objects.count(), 1)
        self.assertEqual(CourseOffering.objects.count(), 1)
        self.assertIn("1 instructors", second_output)
        self.assertIn("1 courses", second_output)

    def test_dry_run_rolls_back_changes(self):
        output = self.run_sync("--dry-run")

        self.assertEqual(InstructorProfile.objects.count(), 0)
        self.assertEqual(CourseCatalog.objects.count(), 0)
        self.assertEqual(CourseOffering.objects.count(), 0)
        self.assertIn("Dry run completed", output)
