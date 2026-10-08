import re
import time
from dataclasses import dataclass
from decimal import Decimal, InvalidOperation
from urllib.parse import urljoin, urlparse

import requests
from bs4 import BeautifulSoup
from django.conf import settings
from django.core.management.base import BaseCommand, CommandError
from django.db import transaction
from django.utils import timezone
from django.utils.text import slugify

from catalog.models import (
    Category,
    CourseCatalog,
    CourseOffering,
    InstructorProfile,
)


DEFAULT_BASE_URL = "https://sematec-co.com"
PERSIAN_DIGITS = str.maketrans("۰۱۲۳۴۵۶۷۸۹٬", "0123456789,")
ARABIC_DIGITS = str.maketrans("٠١٢٣٤٥٦٧٨٩", "0123456789")


@dataclass(frozen=True)
class RemotePage:
    url: str
    soup: BeautifulSoup


class SematecClient:
    def __init__(self, base_url, timeout, delay, stdout):
        self.base_url = base_url.rstrip("/")
        self.timeout = timeout
        self.delay = delay
        self.stdout = stdout
        self.session = requests.Session()
        self.session.headers.update(
            {
                "User-Agent": (
                    "StudentEnrollmentCatalogSync/1.1 "
                    "(+educational-project; polite-crawler)"
                ),
                "Accept-Language": "fa-IR,fa;q=0.9,en;q=0.7",
            }
        )

    def get_page(self, path_or_url):
        url = urljoin(f"{self.base_url}/", path_or_url)

        try:
            response = self.session.get(url, timeout=self.timeout)
            response.raise_for_status()
        except requests.RequestException as exc:
            raise CommandError(f"Could not download {url}: {exc}") from exc

        if self.delay:
            time.sleep(self.delay)

        return RemotePage(
            url=response.url,
            soup=BeautifulSoup(response.text, "html.parser"),
        )


class Command(BaseCommand):
    help = "Synchronize instructors and courses from the official Sematec site."

    def add_arguments(self, parser):
        parser.add_argument(
            "--base-url",
            default=getattr(
                settings,
                "SEMATEC_BASE_URL",
                DEFAULT_BASE_URL,
            ),
        )
        parser.add_argument(
            "--timeout",
            type=float,
            default=getattr(settings, "SEMATEC_SYNC_TIMEOUT", 20.0),
        )
        parser.add_argument(
            "--delay",
            type=float,
            default=getattr(settings, "SEMATEC_SYNC_DELAY", 0.15),
            help="Delay in seconds between remote requests.",
        )
        parser.add_argument(
            "--limit",
            type=int,
            default=None,
            help="Only process the first N instructors and courses.",
        )
        parser.add_argument(
            "--dry-run",
            action="store_true",
            help="Download and parse data, then roll back database changes.",
        )

    def handle(self, *args, **options):
        if options["limit"] is not None and options["limit"] < 1:
            raise CommandError("--limit must be a positive integer.")

        self.client = SematecClient(
            base_url=options["base_url"],
            timeout=options["timeout"],
            delay=max(options["delay"], 0),
            stdout=self.stdout,
        )
        self.base_url = options["base_url"].rstrip("/")
        self.checked_at = timezone.now()

        with transaction.atomic():
            statistics = self.synchronize(options["limit"])

            if options["dry_run"]:
                transaction.set_rollback(True)

        mode = "Dry run completed" if options["dry_run"] else "Sync completed"
        self.stdout.write(
            self.style.SUCCESS(
                f"{mode}: {statistics['instructors']} instructors, "
                f"{statistics['courses']} courses and "
                f"{statistics['relations']} instructor-course relations "
                "processed."
            )
        )

    def synchronize(self, limit):
        instructor_urls = self.collect_instructor_urls()
        course_urls = self.collect_course_urls()

        if limit is not None:
            instructor_urls = instructor_urls[:limit]
            course_urls = course_urls[:limit]

        self.stdout.write(
            f"Discovered {len(instructor_urls)} instructors and "
            f"{len(course_urls)} courses."
        )

        instructors = {}
        relations = set()

        for index, url in enumerate(instructor_urls, start=1):
            try:
                instructor, course_links = self.sync_instructor(url)
            except CommandError as exc:
                self.stderr.write(self.style.WARNING(f"Skipped instructor: {exc}"))
                continue

            instructors[url] = instructor
            relations.update(
                (instructor.slug, course_url)
                for course_url in course_links
            )
            self.stdout.write(
                f"[{index}/{len(instructor_urls)}] Instructor: "
                f"{instructor.full_name} ({len(course_links)} courses)"
            )

        # Always retain courses referenced by processed instructors. Applying
        # the limit again here used to silently discard valid relations.
        all_course_urls = sorted(
            set(course_urls)
            | {course_url for _, course_url in relations}
        )

        courses_by_url = {}
        for index, url in enumerate(all_course_urls, start=1):
            try:
                course = self.sync_course(url)
            except CommandError as exc:
                self.stderr.write(self.style.WARNING(f"Skipped course: {exc}"))
                continue

            courses_by_url[url] = course
            self.stdout.write(
                f"[{index}/{len(all_course_urls)}] Course: {course.title}"
            )

        relation_count = self.sync_relations(relations, courses_by_url)

        return {
            "instructors": len(instructors),
            "courses": len(courses_by_url),
            "relations": relation_count,
        }

    def collect_instructor_urls(self):
        urls = set()

        for page_number in range(1, 4):
            path = (
                "/teacher/"
                if page_number == 1
                else f"/teacher/page/{page_number}/"
            )
            try:
                page = self.client.get_page(path)
            except CommandError as exc:
                if page_number == 1:
                    raise
                self.stderr.write(self.style.WARNING(str(exc)))
                break

            for anchor in page.soup.select('a[href*="/teacher/"]'):
                normalized = self.normalize_content_url(
                    anchor.get("href"),
                    "/teacher/",
                )
                if normalized:
                    urls.add(normalized)

        return sorted(urls)

    def collect_course_urls(self):
        page = self.client.get_page("/course/")
        urls = set()

        for anchor in page.soup.select('a[href*="/course/"]'):
            normalized = self.normalize_content_url(
                anchor.get("href"),
                "/course/",
            )
            if normalized:
                urls.add(normalized)

        return sorted(urls)

    def normalize_content_url(self, href, prefix):
        if not href:
            return None

        absolute = urljoin(f"{self.base_url}/", href)
        parsed = urlparse(absolute)
        base_host = urlparse(self.base_url).netloc.lower()
        path = parsed.path.rstrip("/") + "/"

        if parsed.netloc.lower() != base_host:
            return None
        if not path.startswith(prefix):
            return None
        if path == prefix or "/page/" in path:
            return None

        return f"{parsed.scheme}://{parsed.netloc}{path}"

    def extract_instructor_course_urls(self, soup):
        """Return only courses shown in the instructor's course section."""
        heading = soup.select_one("h2.jtc-teacher-courses-title")

        if heading is None:
            heading = next(
                (
                    item
                    for item in soup.select("h2, h3")
                    if "دوره‌های استاد" in self.clean_text(
                        item.get_text(" ", strip=True)
                    )
                    or "دوره های استاد" in self.clean_text(
                        item.get_text(" ", strip=True)
                    )
                ),
                None,
            )

        if heading is None or heading.parent is None:
            return []

        course_urls = set()
        for anchor in heading.parent.select('a[href*="/course/"]'):
            normalized = self.normalize_content_url(
                anchor.get("href"),
                "/course/",
            )
            if normalized:
                course_urls.add(normalized)

        return sorted(course_urls)

    def sync_instructor(self, url):
        page = self.client.get_page(url)
        title = self.first_text(page.soup, "h1")
        if not title:
            raise CommandError(f"Instructor title was not found: {url}")

        first_name, last_name = self.split_person_name(title)
        slug = self.url_slug(page.url) or slugify(
            title,
            allow_unicode=True,
        )
        expertise = self.extract_instructor_expertise(page.soup)
        bio = self.extract_instructor_bio(page.soup, title)
        image_url = self.extract_image_url(page.soup, page.url)

        instructor = InstructorProfile.objects.filter(
            source_url=page.url,
        ).order_by("id").first()

        defaults = {
            "first_name": first_name,
            "last_name": last_name,
            "expertise": expertise[:300],
            "bio": bio,
            "image_url": image_url,
            "source_url": page.url,
            "is_active": True,
        }

        if instructor is None:
            instructor, _ = InstructorProfile.objects.update_or_create(
                slug=slug,
                defaults=defaults,
            )
        else:
            for field_name, value in defaults.items():
                setattr(instructor, field_name, value)
            instructor.save()

        course_links = self.extract_instructor_course_urls(page.soup)
        return instructor, course_links

    def sync_course(self, url):
        page = self.client.get_page(url)
        title = self.first_text(page.soup, "h1")
        if not title:
            raise CommandError(f"Course title was not found: {url}")

        slug = self.url_slug(page.url) or slugify(
            title,
            allow_unicode=True,
        )
        page_text = self.clean_text(page.soup.get_text(" ", strip=True))
        description = self.extract_meta_content(
            page.soup,
            "meta[name='description']",
        )
        if not description:
            description = self.extract_first_description(page.soup, title)

        category = self.resolve_category(title, description)
        price = self.extract_price(page.soup)
        duration = self.extract_duration(page_text)
        prerequisites = self.extract_prerequisites(page_text)
        image_url = self.extract_image_url(page.soup, page.url)
        unavailable = "ناموجود" in page_text and price == 0

        course, _ = CourseCatalog.objects.update_or_create(
            slug=slug,
            defaults={
                "category": category,
                "title": title[:220],
                "short_description": description[:350],
                "description": description,
                "price_toman": price,
                "duration_hours": duration,
                "level": self.infer_level(title),
                "prerequisites": prerequisites,
                "cover_image_url": image_url,
                "source_url": page.url,
                "is_featured": False,
                "is_active": not unavailable,
                "source_checked_at": self.checked_at,
            },
        )

        return course

    def sync_relations(self, relations, courses_by_url):
        processed = 0

        for instructor_slug, course_url in sorted(relations):
            course = courses_by_url.get(course_url)
            if course is None:
                continue

            instructor = InstructorProfile.objects.filter(
                slug=instructor_slug,
            ).first()
            if instructor is None:
                continue

            existing = CourseOffering.objects.filter(
                course=course,
                instructor=instructor,
                is_active=True,
            ).exists()

            if not existing:
                CourseOffering.objects.create(
                    course=course,
                    instructor=instructor,
                    start_date_text="در انتظار اعلام",
                    schedule_text="زمان‌بندی جدید به‌زودی اعلام می‌شود",
                    delivery_mode=CourseOffering.DeliveryMode.HYBRID,
                    status=CourseOffering.Status.UPCOMING,
                    registration_url=course.source_url,
                    is_active=True,
                )

            processed += 1

        return processed

    def resolve_category(self, title, description):
        text = f"{title} {description}".lower()
        rules = [
            (
                "cyber-security",
                (
                    "security",
                    "امنیت",
                    "fortinet",
                    "fortigate",
                    "ceh",
                    "devsecops",
                    "penetration",
                ),
            ),
            (
                "devops-and-cloud",
                (
                    "devops",
                    "docker",
                    "kubernetes",
                    "cloud",
                    "aws",
                    "terraform",
                    "ansible",
                    "openshift",
                ),
            ),
            (
                "network-and-infrastructure",
                (
                    "network",
                    "شبکه",
                    "linux",
                    "لینوکس",
                    "cisco",
                    "mikrotik",
                    "vmware",
                    "veeam",
                    "data center",
                ),
            ),
            (
                "data-and-ai",
                (
                    "sql",
                    "database",
                    "دیتابیس",
                    "پایگاه داده",
                    "data science",
                    "علم داده",
                    "هوش مصنوعی",
                    "machine learning",
                    "power bi",
                ),
            ),
            (
                "business-and-management",
                (
                    "business",
                    "مدیریت",
                    "pmp",
                    "scrum",
                    "agile",
                    "بانکداری",
                    "itil",
                ),
            ),
        ]

        for slug, keywords in rules:
            if any(keyword in text for keyword in keywords):
                category = Category.objects.filter(slug=slug).first()
                if category is not None:
                    return category

        category = Category.objects.filter(
            slug="software-development"
        ).first()
        if category is None:
            category = Category.objects.filter(is_active=True).first()
        if category is None:
            raise CommandError(
                "No active catalog category exists. Run seed_catalog first."
            )
        return category

    @staticmethod
    def split_person_name(full_name):
        parts = full_name.split()
        if len(parts) == 1:
            return parts[0], ""
        return parts[0], " ".join(parts[1:])

    @staticmethod
    def url_slug(url):
        return urlparse(url).path.strip("/").split("/")[-1]

    @staticmethod
    def clean_text(value):
        return re.sub(r"\s+", " ", value or "").strip()

    def first_text(self, soup, selector):
        element = soup.select_one(selector)
        if element is None:
            return ""
        return self.clean_text(element.get_text(" ", strip=True))

    def extract_meta_content(self, soup, selector):
        element = soup.select_one(selector)
        if element is None:
            return ""
        return self.clean_text(element.get("content", ""))

    def extract_image_url(self, soup, page_url):
        image_url = self.extract_meta_content(
            soup,
            "meta[property='og:image']",
        )
        return urljoin(page_url, image_url) if image_url else ""

    def extract_instructor_expertise(self, soup):
        heading = soup.select_one("h1")
        if heading is not None:
            candidate = heading.find_next(
                lambda tag: tag.name in {"p", "div"}
                and self.clean_text(tag.get_text(" ", strip=True))
            )
            if candidate is not None:
                text = self.clean_text(candidate.get_text(" ", strip=True))
                if len(text) <= 300:
                    return text

        return "مدرس دوره‌های تخصصی فناوری اطلاعات"

    def extract_instructor_bio(self, soup, title):
        for heading in soup.select("h2, h3"):
            heading_text = self.clean_text(
                heading.get_text(" ", strip=True)
            )
            if "کیست" not in heading_text and "درباره" not in heading_text:
                continue

            paragraphs = []
            for sibling in heading.find_all_next(["p", "h2"], limit=6):
                if sibling.name == "h2":
                    break
                text = self.clean_text(sibling.get_text(" ", strip=True))
                if text and text != title:
                    paragraphs.append(text)

            if paragraphs:
                return "\n\n".join(paragraphs)[:4000]

        return f"معرفی و سوابق آموزشی {title} در سماتک."

    def extract_first_description(self, soup, title):
        heading = soup.select_one("h1")
        if heading is not None:
            paragraph = heading.find_next("p")
            if paragraph is not None:
                text = self.clean_text(paragraph.get_text(" ", strip=True))
                if text and text != title:
                    return text[:4000]

        return f"اطلاعات دوره {title} براساس صفحه رسمی سماتک."

    @staticmethod
    def normalize_number(value):
        normalized = value.translate(PERSIAN_DIGITS).translate(
            ARABIC_DIGITS
        )
        return normalized.replace(",", "").replace("٬", "")

    def extract_price_values(self, value):
        matches = re.findall(
            r"([۰-۹٠-٩\d][۰-۹٠-٩\d,٬]*)\s*تومان",
            self.clean_text(value),
        )
        values = []

        for match in matches:
            try:
                amount = int(Decimal(self.normalize_number(match)))
            except (InvalidOperation, ValueError):
                continue
            if amount > 0:
                values.append(amount)

        return values

    def extract_price(self, source):
        """Extract the visible product price before broad page fallbacks."""
        if isinstance(source, str):
            return max(self.extract_price_values(source), default=0)

        preferred_selectors = (
            ".summary p.price",
            ".summary .price",
            "main p.price",
            ".product-info p.price",
            "p.price",
        )
        for selector in preferred_selectors:
            element = source.select_one(selector)
            if element is None:
                continue
            values = self.extract_price_values(
                element.get_text(" ", strip=True)
            )
            if values:
                return max(values)

        for selector in (
            "meta[property='product:price:amount']",
            "meta[itemprop='price']",
        ):
            element = source.select_one(selector)
            if element is None:
                continue
            raw_value = element.get("content", "")
            try:
                value = int(Decimal(self.normalize_number(raw_value)))
            except (InvalidOperation, ValueError):
                continue
            if value > 0:
                return value

        main = source.select_one("main")
        page_text = (main or source).get_text(" ", strip=True)
        return max(self.extract_price_values(page_text), default=0)

    def extract_duration(self, page_text):
        match = re.search(
            r"زمان دوره\s*[|:]?\s*([۰-۹٠-٩\d]+)\s*ساعت",
            page_text,
            flags=re.IGNORECASE,
        )
        if match is None:
            return 0
        return int(self.normalize_number(match.group(1)))

    def extract_prerequisites(self, page_text):
        match = re.search(
            r"پیش\s*نیاز\s*[|:]?\s*(.{1,180}?)"
            r"(?=کلاس‌های فعال|سرفصل|افزودن به علاقه|$)",
            page_text,
        )
        if match is None:
            return ""
        return self.clean_text(match.group(1))[:500]

    @staticmethod
    def infer_level(title):
        lowered = title.lower()
        if "پیشرفته" in title or "advanced" in lowered:
            return CourseCatalog.Level.ADVANCED
        if "مقدماتی" in title or "fundamental" in lowered:
            return CourseCatalog.Level.BEGINNER
        return CourseCatalog.Level.ALL_LEVELS
