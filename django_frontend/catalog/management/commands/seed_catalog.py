from django.core.management.base import BaseCommand
from django.db import transaction

from catalog.models import Category, InstructorProfile


CATEGORIES = [
    {
        "name": "برنامه‌نویسی و توسعه نرم‌افزار",
        "slug": "software-development",
        "description": (
            "دوره‌های برنامه‌نویسی، توسعه وب، طراحی نرم‌افزار "
            "و مهندسی Backend و Frontend"
        ),
        "accent_color": "#6D5DFB",
        "icon_name": "code",
        "display_order": 10,
    },
    {
        "name": "داده و هوش مصنوعی",
        "slug": "data-and-ai",
        "description": (
            "دوره‌های علم داده، یادگیری ماشین، هوش مصنوعی، "
            "تحلیل داده و هوش تجاری"
        ),
        "accent_color": "#00B8A9",
        "icon_name": "database",
        "display_order": 20,
    },
    {
        "name": "شبکه و زیرساخت",
        "slug": "network-and-infrastructure",
        "description": (
            "دوره‌های شبکه، لینوکس، مدیریت زیرساخت "
            "و سرویس‌های سازمانی"
        ),
        "accent_color": "#FF8A3D",
        "icon_name": "network",
        "display_order": 30,
    },
    {
        "name": "DevOps و رایانش ابری",
        "slug": "devops-and-cloud",
        "description": (
            "دوره‌های DevOps، Docker، Kubernetes، CI/CD "
            "و فناوری‌های Cloud"
        ),
        "accent_color": "#2388FF",
        "icon_name": "cloud",
        "display_order": 40,
    },
    {
        "name": "امنیت سایبری",
        "slug": "cyber-security",
        "description": (
            "دوره‌های امنیت شبکه، تست نفوذ، امنیت نرم‌افزار "
            "و مدیریت امنیت اطلاعات"
        ),
        "accent_color": "#EC4899",
        "icon_name": "shield",
        "display_order": 50,
    },
    {
        "name": "مدیریت و تحلیل کسب‌وکار",
        "slug": "business-and-management",
        "description": (
            "دوره‌های مدیریت پروژه، تحلیل کسب‌وکار، "
            "مهارت‌های مدیریتی و سازمانی"
        ),
        "accent_color": "#EAB308",
        "icon_name": "briefcase",
        "display_order": 60,
    },
]


INSTRUCTORS = [
    {
        "first_name": "پرهام",
        "last_name": "درویشی",
        "slug": "parham-darvishi",
        "expertise": "توسعه نرم‌افزار و برنامه‌نویسی",
        "is_featured": True,
    },
    {
        "first_name": "آرمین",
        "last_name": "یعقوبی",
        "slug": "armin-yaghoubi",
        "expertise": "فناوری اطلاعات و آموزش‌های تخصصی",
        "is_featured": True,
    },
    {
        "first_name": "ساناز",
        "last_name": "عباس‌زاده",
        "slug": "sanaz-abbaszadeh",
        "expertise": "پایگاه داده و SQL Server",
        "is_featured": True,
    },
    {
        "first_name": "آرش",
        "last_name": "فروغی",
        "slug": "arash-foroughi",
        "expertise": "DevOps، زیرساخت و رایانش ابری",
        "is_featured": True,
    },
]


class Command(BaseCommand):
    help = "Create or update the initial course catalog data."

    @transaction.atomic
    def handle(self, *args, **options):
        category_count = self.seed_categories()
        instructor_count = self.seed_instructors()

        self.stdout.write(
            self.style.SUCCESS(
                "Catalog seed completed successfully: "
                f"{category_count} categories and "
                f"{instructor_count} instructors processed."
            )
        )

    def seed_categories(self):
        for category_data in CATEGORIES:
            slug = category_data["slug"]
            defaults = {
                key: value
                for key, value in category_data.items()
                if key != "slug"
            }

            category, created = Category.objects.update_or_create(
                slug=slug,
                defaults={
                    **defaults,
                    "is_active": True,
                },
            )

            action = "Created" if created else "Updated"

            self.stdout.write(
                f"{action} category: {category.name}"
            )

        return len(CATEGORIES)

    def seed_instructors(self):
        for instructor_data in INSTRUCTORS:
            slug = instructor_data["slug"]
            defaults = {
                key: value
                for key, value in instructor_data.items()
                if key != "slug"
            }

            instructor, created = (
                InstructorProfile.objects.update_or_create(
                    slug=slug,
                    defaults={
                        **defaults,
                        "source_url": "https://sematec-co.com/",
                        "is_active": True,
                    },
                )
            )

            action = "Created" if created else "Updated"

            self.stdout.write(
                f"{action} instructor: {instructor.full_name}"
            )

        return len(INSTRUCTORS)