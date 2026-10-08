from django.core.management.base import BaseCommand
from django.db import transaction
from django.utils import timezone

from catalog.models import (
    Category,
    CourseCatalog,
    CourseOffering,
    InstructorProfile,
)


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
        "expertise": "توسعه و معماری نرم‌افزار",
        "bio": (
            "مدرس و متخصص توسعه نرم‌افزار با بیش از ۱۲ سال تجربه "
            "در تدریس، برنامه‌نویسی، مشاوره فنی و هدایت تیم‌های نرم‌افزاری."
        ),
        "source_url": "https://sematec-co.com/teacher/parham-darvishi/",
        "is_featured": True,
    },
    {
        "first_name": "آرمین",
        "last_name": "یعقوبی",
        "slug": "armin-yaghoubi",
        "expertise": "برنامه‌نویسی، Python، .NET و SQL Server",
        "bio": (
            "برنامه‌نویس و مدرس حوزه نرم‌افزار با سابقه فعالیت حرفه‌ای "
            "از سال ۱۳۹۶ و تجربه توسعه و مدیریت پروژه‌های سازمانی و ملی."
        ),
        "source_url": "https://sematec-co.com/teacher/arminyaghoubi/",
        "is_featured": True,
    },
    {
        "first_name": "ساناز",
        "last_name": "عباس‌زاده",
        "slug": "sanaz-abbaszadeh",
        "expertise": "پایگاه داده، SQL Server و بهینه‌سازی Query",
        "bio": (
            "متخصص و مدرس پایگاه داده و SQL Server با تجربه طراحی و توسعه "
            "سامانه‌های بانکی، انبار داده و نرم‌افزارهای سازمانی."
        ),
        "source_url": "https://sematec-co.com/teacher/sanaz-abbaszadeh/",
        "is_featured": True,
    },
    {
        "first_name": "آرش",
        "last_name": "فروغی",
        "slug": "arash-foroughi",
        "expertise": "Cloud، DevOps، Linux و زیرساخت",
        "bio": (
            "متخصص Cloud و DevOps با بیش از ۱۲ سال تجربه در AWS، "
            "Kubernetes، Linux، Terraform، Ansible، Docker و مانیتورینگ."
        ),
        "source_url": "https://sematec-co.com/teacher/arash-furuqi/",
        "is_featured": True,
    },
]


COURSES = [
    {
        "category_slug": "software-development",
        "title": "آموزش برنامه‌نویسی مقدماتی C#",
        "slug": "ms-net-fundamentals",
        "short_description": (
            "شروع برنامه‌نویسی با C#، مفاهیم پایه، کنترل جریان و شی‌گرایی"
        ),
        "description": (
            "دوره‌ای پروژه‌محور برای یادگیری مبانی C#، متغیرها، شرط‌ها، "
            "حلقه‌ها، توابع، مدیریت خطا و آشنایی با برنامه‌نویسی شی‌گرا."
        ),
        "price_toman": 3_900_000,
        "duration_hours": 24,
        "level": CourseCatalog.Level.BEGINNER,
        "prerequisites": "ندارد",
        "source_url": "https://sematec-co.com/course/ms-net-fundamentals/",
        "is_featured": True,
    },
    {
        "category_slug": "software-development",
        "title": "Fullstack Web Development with .NET 10 and Next.js",
        "slug": "fullstack-web-development",
        "short_description": (
            "توسعه Fullstack با .NET 10، Next.js، REST، gRPC و GraphQL"
        ),
        "description": (
            "آموزش طراحی و توسعه برنامه‌های Fullstack و سبک‌های معماری API "
            "با استفاده از فناوری‌های مدرن اکوسیستم .NET و Next.js."
        ),
        "price_toman": 9_900_000,
        "duration_hours": 51,
        "level": CourseCatalog.Level.ADVANCED,
        "prerequisites": "Programming in C# 1",
        "source_url": "https://sematec-co.com/course/fullstack-web-development/",
        "is_featured": True,
    },
    {
        "category_slug": "software-development",
        "title": "دوره مقدماتی Golang",
        "slug": "golang",
        "short_description": (
            "آموزش مقدماتی زبان Go برای ورود به توسعه Backend"
        ),
        "description": (
            "آشنایی عملی با زبان Go، ساختار برنامه، انواع داده، توابع، "
            "مدیریت خطا و مفاهیم لازم برای شروع توسعه نرم‌افزار با Golang."
        ),
        "price_toman": 5_500_000,
        "duration_hours": 15,
        "level": CourseCatalog.Level.BEGINNER,
        "prerequisites": "آشنایی با فلوچارت و الگوریتم",
        "source_url": "https://sematec-co.com/course/golang/",
        "is_featured": True,
    },
    {
        "category_slug": "data-and-ai",
        "title": "SQL Performance & Advanced Query Optimization",
        "slug": "sql-performance-advanced-query-optimization",
        "short_description": (
            "بهینه‌سازی Query و تحلیل عملکرد SQL Server"
        ),
        "description": (
            "آموزش کوئری‌های پیشرفته، تکنیک‌های بهینه‌سازی عملکرد و کار با "
            "ابزارهای مانیتورینگ و تحلیل کارایی SQL Server."
        ),
        "price_toman": 6_900_000,
        "duration_hours": 24,
        "level": CourseCatalog.Level.ADVANCED,
        "prerequisites": "SQL Server",
        "source_url": (
            "https://sematec-co.com/course/"
            "sql-performance-advanced-query-optimization/"
        ),
        "is_featured": True,
    },
    {
        "category_slug": "data-and-ai",
        "title": "SQL Server 2025 Database Implementation",
        "slug": "sql-server-2025-database-implementation",
        "short_description": (
            "طراحی و پیاده‌سازی حرفه‌ای پایگاه داده با SQL Server"
        ),
        "description": (
            "دوره جامع پیاده‌سازی پایگاه داده برای یادگیری ساختار، مدیریت "
            "و توسعه بانک‌های اطلاعاتی مبتنی بر SQL Server."
        ),
        "price_toman": 8_700_000,
        "duration_hours": 52,
        "level": CourseCatalog.Level.INTERMEDIATE,
        "prerequisites": "SQL Server و آشنایی با مفاهیم بانک‌های اطلاعاتی",
        "source_url": (
            "https://sematec-co.com/course/"
            "sql-server-2025-database-implementation/"
        ),
        "is_featured": True,
    },
    {
        "category_slug": "software-development",
        "title": "برنامه‌نویسی با Python",
        "slug": "programing-python",
        "short_description": (
            "آموزش کاربردی پایتون از پایه با تمرین و پروژه"
        ),
        "description": (
            "مسیر جامع ورود به برنامه‌نویسی Python شامل مفاهیم پایه، "
            "ساختارهای کنترلی، توابع، ساختمان داده و تمرین‌های کاربردی."
        ),
        "price_toman": 7_500_000,
        "duration_hours": 40,
        "level": CourseCatalog.Level.BEGINNER,
        "prerequisites": "ندارد",
        "source_url": "https://sematec-co.com/course/programing-python/",
        "is_featured": True,
    },
    {
        "category_slug": "software-development",
        "title": "Advanced Python",
        "slug": "advanced-python",
        "short_description": (
            "مفاهیم پیشرفته و الگوهای حرفه‌ای برنامه‌نویسی Python"
        ),
        "description": (
            "ادامه مسیر Python برای تسلط بر امکانات پیشرفته زبان، "
            "ساختارهای حرفه‌ای و توسعه پروژه‌های قابل نگهداری."
        ),
        "price_toman": 8_900_000,
        "duration_hours": 40,
        "level": CourseCatalog.Level.ADVANCED,
        "prerequisites": "برنامه‌نویسی با پایتون",
        "source_url": "https://sematec-co.com/course/advanced-python/",
        "is_featured": False,
    },
    {
        "category_slug": "devops-and-cloud",
        "title": "AWS + Terraform",
        "slug": "aws-cloud-practitioner-essentials",
        "short_description": (
            "زیرساخت ابری AWS و Infrastructure as Code با Terraform"
        ),
        "description": (
            "آموزش سرویس‌های ابری AWS و پیاده‌سازی و مدیریت زیرساخت "
            "به‌صورت کد با Terraform."
        ),
        "price_toman": 11_900_000,
        "duration_hours": 48,
        "level": CourseCatalog.Level.INTERMEDIATE,
        "prerequisites": "DevOps",
        "source_url": (
            "https://sematec-co.com/course/"
            "aws-cloud-practitioner-essentials/"
        ),
        "is_featured": True,
    },
    {
        "category_slug": "cyber-security",
        "title": "DevSecOps",
        "slug": "devsecops",
        "short_description": (
            "ادغام امنیت در چرخه توسعه و عملیات نرم‌افزار"
        ),
        "description": (
            "دوره تخصصی DevSecOps برای پیاده‌سازی امنیت در کنترل نسخه، "
            "CI/CD، کانتینرها، زیرساخت و چرخه تحویل نرم‌افزار."
        ),
        "price_toman": 19_500_000,
        "duration_hours": 100,
        "level": CourseCatalog.Level.ADVANCED,
        "prerequisites": "DevOps",
        "source_url": "https://sematec-co.com/course/devsecops/",
        "is_featured": True,
    },
    {
        "category_slug": "devops-and-cloud",
        "title": "DevOps Administration Pack",
        "slug": "devops-administration-pack",
        "short_description": (
            "مسیر جامع مدیریت DevOps، کانتینر، اتوماسیون و زیرساخت"
        ),
        "description": (
            "یک مسیر تخصصی و پروژه‌محور برای یادگیری ابزارها و مهارت‌های "
            "عملی موردنیاز مدیریت DevOps و ورود به بازار کار."
        ),
        "price_toman": 18_900_000,
        "duration_hours": 120,
        "level": CourseCatalog.Level.ADVANCED,
        "prerequisites": "Network+",
        "source_url": (
            "https://sematec-co.com/course/devops-administration-pack/"
        ),
        "is_featured": True,
    },
]


OFFERINGS = [
    {
        "course_slug": "ms-net-fundamentals",
        "instructor_slug": "parham-darvishi",
        "start_date_text": "۱۴۰۵/۰۴/۱۹",
        "schedule_text": "جمعه‌ها، ساعت ۱۴ تا ۱۸",
        "delivery_mode": CourseOffering.DeliveryMode.HYBRID,
        "status": CourseOffering.Status.OPEN,
    },
    {
        "course_slug": "fullstack-web-development",
        "instructor_slug": "parham-darvishi",
        "start_date_text": "در انتظار اعلام",
        "schedule_text": "زمان‌بندی جدید به‌زودی اعلام می‌شود",
        "delivery_mode": CourseOffering.DeliveryMode.HYBRID,
        "status": CourseOffering.Status.UPCOMING,
    },
    {
        "course_slug": "golang",
        "instructor_slug": "parham-darvishi",
        "start_date_text": "۱۴۰۵/۰۹/۰۶",
        "schedule_text": "جمعه‌ها، ساعت ۹ تا ۱۳",
        "delivery_mode": CourseOffering.DeliveryMode.HYBRID,
        "status": CourseOffering.Status.OPEN,
    },
    {
        "course_slug": "sql-performance-advanced-query-optimization",
        "instructor_slug": "sanaz-abbaszadeh",
        "start_date_text": "۱۴۰۵/۰۹/۱۲",
        "schedule_text": "پنجشنبه‌ها، ساعت ۱۴ تا ۱۸",
        "delivery_mode": CourseOffering.DeliveryMode.HYBRID,
        "status": CourseOffering.Status.OPEN,
    },
    {
        "course_slug": "sql-server-2025-database-implementation",
        "instructor_slug": "sanaz-abbaszadeh",
        "start_date_text": "۱۴۰۵/۰۷/۰۷",
        "schedule_text": "یکشنبه و سه‌شنبه‌ها، ساعت ۱۶:۳۰ تا ۲۰:۳۰",
        "delivery_mode": CourseOffering.DeliveryMode.HYBRID,
        "status": CourseOffering.Status.OPEN,
    },
    {
        "course_slug": "sql-server-2025-database-implementation",
        "instructor_slug": "armin-yaghoubi",
        "start_date_text": "در انتظار اعلام",
        "schedule_text": "زمان‌بندی جدید به‌زودی اعلام می‌شود",
        "delivery_mode": CourseOffering.DeliveryMode.HYBRID,
        "status": CourseOffering.Status.UPCOMING,
    },
    {
        "course_slug": "programing-python",
        "instructor_slug": "armin-yaghoubi",
        "start_date_text": "۱۴۰۵/۰۸/۰۸",
        "schedule_text": "جمعه‌ها، ساعت ۱۴ تا ۱۸",
        "delivery_mode": CourseOffering.DeliveryMode.HYBRID,
        "status": CourseOffering.Status.OPEN,
    },
    {
        "course_slug": "advanced-python",
        "instructor_slug": "armin-yaghoubi",
        "start_date_text": "در انتظار اعلام",
        "schedule_text": "زمان‌بندی جدید به‌زودی اعلام می‌شود",
        "delivery_mode": CourseOffering.DeliveryMode.HYBRID,
        "status": CourseOffering.Status.UPCOMING,
    },
    {
        "course_slug": "aws-cloud-practitioner-essentials",
        "instructor_slug": "arash-foroughi",
        "start_date_text": "۱۴۰۵/۰۸/۱۴",
        "schedule_text": "پنجشنبه‌ها، ساعت ۱۴ تا ۱۹",
        "delivery_mode": CourseOffering.DeliveryMode.HYBRID,
        "status": CourseOffering.Status.OPEN,
    },
    {
        "course_slug": "devsecops",
        "instructor_slug": "arash-foroughi",
        "start_date_text": "۱۴۰۵/۱۱/۰۸",
        "schedule_text": "پنجشنبه‌ها، ساعت ۱۴ تا ۱۹",
        "delivery_mode": CourseOffering.DeliveryMode.HYBRID,
        "status": CourseOffering.Status.OPEN,
    },
    {
        "course_slug": "devops-administration-pack",
        "instructor_slug": "arash-foroughi",
        "start_date_text": "۱۴۰۵/۰۸/۱۵",
        "schedule_text": "جمعه‌ها، ساعت ۸:۳۰ تا ۱۳:۳۰",
        "delivery_mode": CourseOffering.DeliveryMode.ONLINE,
        "status": CourseOffering.Status.OPEN,
    },
    {
        "course_slug": "devops-administration-pack",
        "instructor_slug": "arash-foroughi",
        "start_date_text": "۱۴۰۵/۰۹/۱۵",
        "schedule_text": "یکشنبه‌ها، ساعت ۱۶:۳۰ تا ۲۰:۳۰",
        "delivery_mode": CourseOffering.DeliveryMode.ONLINE,
        "status": CourseOffering.Status.OPEN,
    },
]


class Command(BaseCommand):
    help = "Create or update the initial course catalog data."

    @transaction.atomic
    def handle(self, *args, **options):
        categories = self.seed_categories()
        instructors = self.seed_instructors()
        courses = self.seed_courses(categories)
        offering_count = self.seed_offerings(courses, instructors)

        self.stdout.write(
            self.style.SUCCESS(
                "Catalog seed completed successfully: "
                f"{len(categories)} categories, "
                f"{len(instructors)} instructors, "
                f"{len(courses)} courses and "
                f"{offering_count} offerings processed."
            )
        )

    def seed_categories(self):
        categories = {}

        for category_data in CATEGORIES:
            slug = category_data["slug"]
            defaults = {
                key: value
                for key, value in category_data.items()
                if key != "slug"
            }
            category, created = Category.objects.update_or_create(
                slug=slug,
                defaults={**defaults, "is_active": True},
            )
            categories[slug] = category
            action = "Created" if created else "Updated"
            self.stdout.write(f"{action} category: {category.name}")

        return categories

    def seed_instructors(self):
        instructors = {}

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
                    defaults={**defaults, "is_active": True},
                )
            )
            instructors[slug] = instructor
            action = "Created" if created else "Updated"
            self.stdout.write(
                f"{action} instructor: {instructor.full_name}"
            )

        return instructors

    def seed_courses(self, categories):
        courses = {}
        checked_at = timezone.now()

        for course_data in COURSES:
            slug = course_data["slug"]
            category_slug = course_data["category_slug"]
            defaults = {
                key: value
                for key, value in course_data.items()
                if key not in {"slug", "category_slug"}
            }
            course, created = CourseCatalog.objects.update_or_create(
                slug=slug,
                defaults={
                    **defaults,
                    "category": categories[category_slug],
                    "source_checked_at": checked_at,
                    "is_active": True,
                },
            )
            courses[slug] = course
            action = "Created" if created else "Updated"
            self.stdout.write(f"{action} course: {course.title}")

        return courses

    def seed_offerings(self, courses, instructors):
        for offering_data in OFFERINGS:
            course = courses[offering_data["course_slug"]]
            instructor = instructors[offering_data["instructor_slug"]]
            start_date_text = offering_data["start_date_text"]
            schedule_text = offering_data["schedule_text"]
            defaults = {
                key: value
                for key, value in offering_data.items()
                if key
                not in {
                    "course_slug",
                    "instructor_slug",
                    "start_date_text",
                    "schedule_text",
                }
            }
            offering, created = CourseOffering.objects.update_or_create(
                course=course,
                instructor=instructor,
                start_date_text=start_date_text,
                schedule_text=schedule_text,
                defaults={
                    **defaults,
                    "registration_url": course.source_url,
                    "is_active": True,
                },
            )
            action = "Created" if created else "Updated"
            self.stdout.write(f"{action} offering: {offering}")

        return len(OFFERINGS)
