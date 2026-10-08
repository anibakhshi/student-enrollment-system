from django.db import models


class Category(models.Model):
    name = models.CharField(
        max_length=120,
        unique=True,
        verbose_name="نام دسته‌بندی",
    )
    slug = models.SlugField(
        max_length=140,
        unique=True,
        verbose_name="شناسه",
    )
    description = models.TextField(
        blank=True,
        verbose_name="توضیحات",
    )
    accent_color = models.CharField(
        max_length=20,
        default="#6D5DFB",
        verbose_name="رنگ شاخص",
    )
    icon_name = models.CharField(
        max_length=60,
        blank=True,
        verbose_name="نام آیکن",
    )
    display_order = models.PositiveIntegerField(
        default=0,
        verbose_name="ترتیب نمایش",
    )
    is_active = models.BooleanField(
        default=True,
        verbose_name="فعال",
    )
    created_at = models.DateTimeField(
        auto_now_add=True,
        verbose_name="زمان ایجاد",
    )
    updated_at = models.DateTimeField(
        auto_now=True,
        verbose_name="آخرین تغییر",
    )

    class Meta:
        ordering = ["display_order", "name"]
        verbose_name = "دسته‌بندی"
        verbose_name_plural = "دسته‌بندی‌ها"

    def __str__(self):
        return self.name


class InstructorProfile(models.Model):
    api_instructor_id = models.PositiveBigIntegerField(
        null=True,
        blank=True,
        unique=True,
        verbose_name="شناسه استاد در Go API",
    )
    first_name = models.CharField(
        max_length=100,
        verbose_name="نام",
    )
    last_name = models.CharField(
        max_length=100,
        verbose_name="نام خانوادگی",
    )
    slug = models.SlugField(
        max_length=220,
        unique=True,
        verbose_name="شناسه",
    )
    bio = models.TextField(
        blank=True,
        verbose_name="درباره استاد",
    )
    expertise = models.CharField(
        max_length=300,
        blank=True,
        verbose_name="تخصص",
    )
    email = models.EmailField(
        blank=True,
        verbose_name="ایمیل",
    )
    phone = models.CharField(
        max_length=30,
        blank=True,
        verbose_name="شماره تماس",
    )
    image_url = models.URLField(
        max_length=500,
        blank=True,
        verbose_name="آدرس تصویر",
    )
    source_url = models.URLField(
        max_length=500,
        blank=True,
        verbose_name="منبع اطلاعات",
    )
    is_featured = models.BooleanField(
        default=False,
        verbose_name="استاد منتخب",
    )
    is_active = models.BooleanField(
        default=True,
        verbose_name="فعال",
    )
    created_at = models.DateTimeField(
        auto_now_add=True,
        verbose_name="زمان ایجاد",
    )
    updated_at = models.DateTimeField(
        auto_now=True,
        verbose_name="آخرین تغییر",
    )

    class Meta:
        ordering = ["first_name", "last_name"]
        verbose_name = "استاد"
        verbose_name_plural = "استادان"

    @property
    def full_name(self):
        return f"{self.first_name} {self.last_name}".strip()

    def __str__(self):
        return self.full_name


class CourseCatalog(models.Model):
    class Level(models.TextChoices):
        BEGINNER = "beginner", "مقدماتی"
        INTERMEDIATE = "intermediate", "متوسط"
        ADVANCED = "advanced", "پیشرفته"
        ALL_LEVELS = "all_levels", "مناسب همه سطوح"

    api_course_id = models.PositiveBigIntegerField(
        null=True,
        blank=True,
        unique=True,
        verbose_name="شناسه دوره در Go API",
    )
    category = models.ForeignKey(
        Category,
        on_delete=models.PROTECT,
        related_name="courses",
        verbose_name="دسته‌بندی",
    )
    title = models.CharField(
        max_length=220,
        verbose_name="عنوان دوره",
    )
    slug = models.SlugField(
        max_length=240,
        unique=True,
        verbose_name="شناسه",
    )
    short_description = models.CharField(
        max_length=350,
        blank=True,
        verbose_name="توضیح کوتاه",
    )
    description = models.TextField(
        blank=True,
        verbose_name="توضیحات کامل",
    )
    price_toman = models.PositiveBigIntegerField(
        default=0,
        verbose_name="قیمت به تومان",
    )
    duration_hours = models.PositiveIntegerField(
        default=0,
        verbose_name="مدت دوره به ساعت",
    )
    level = models.CharField(
        max_length=20,
        choices=Level.choices,
        default=Level.ALL_LEVELS,
        verbose_name="سطح دوره",
    )
    prerequisites = models.TextField(
        blank=True,
        verbose_name="پیش‌نیازها",
    )
    cover_image_url = models.URLField(
        max_length=500,
        blank=True,
        verbose_name="تصویر دوره",
    )
    source_url = models.URLField(
        max_length=500,
        blank=True,
        verbose_name="صفحه رسمی دوره",
    )
    is_featured = models.BooleanField(
        default=False,
        verbose_name="دوره منتخب",
    )
    is_active = models.BooleanField(
        default=True,
        verbose_name="فعال",
    )
    source_checked_at = models.DateTimeField(
        null=True,
        blank=True,
        verbose_name="زمان بررسی منبع",
    )
    created_at = models.DateTimeField(
        auto_now_add=True,
        verbose_name="زمان ایجاد",
    )
    updated_at = models.DateTimeField(
        auto_now=True,
        verbose_name="آخرین تغییر",
    )

    class Meta:
        ordering = ["-is_featured", "title"]
        verbose_name = "دوره"
        verbose_name_plural = "دوره‌ها"

    def __str__(self):
        return self.title


class CourseOffering(models.Model):
    class DeliveryMode(models.TextChoices):
        IN_PERSON = "in_person", "حضوری"
        ONLINE = "online", "آنلاین"
        HYBRID = "hybrid", "حضوری و آنلاین"

    class Status(models.TextChoices):
        UPCOMING = "upcoming", "به‌زودی"
        OPEN = "open", "در حال ثبت‌نام"
        CLOSED = "closed", "تکمیل ظرفیت"
        FINISHED = "finished", "پایان‌یافته"

    course = models.ForeignKey(
        CourseCatalog,
        on_delete=models.CASCADE,
        related_name="offerings",
        verbose_name="دوره",
    )
    instructor = models.ForeignKey(
        InstructorProfile,
        on_delete=models.PROTECT,
        related_name="offerings",
        verbose_name="استاد",
    )
    start_date_text = models.CharField(
        max_length=120,
        blank=True,
        verbose_name="تاریخ شروع",
    )
    schedule_text = models.CharField(
        max_length=200,
        blank=True,
        verbose_name="برنامه برگزاری",
    )
    delivery_mode = models.CharField(
        max_length=20,
        choices=DeliveryMode.choices,
        default=DeliveryMode.IN_PERSON,
        verbose_name="نوع برگزاری",
    )
    status = models.CharField(
        max_length=20,
        choices=Status.choices,
        default=Status.UPCOMING,
        verbose_name="وضعیت",
    )
    capacity = models.PositiveIntegerField(
        null=True,
        blank=True,
        verbose_name="ظرفیت",
    )
    price_override_toman = models.PositiveBigIntegerField(
        null=True,
        blank=True,
        verbose_name="قیمت اختصاصی این کلاس",
    )
    registration_url = models.URLField(
        max_length=500,
        blank=True,
        verbose_name="لینک ثبت‌نام",
    )
    is_active = models.BooleanField(
        default=True,
        verbose_name="فعال",
    )
    created_at = models.DateTimeField(
        auto_now_add=True,
        verbose_name="زمان ایجاد",
    )
    updated_at = models.DateTimeField(
        auto_now=True,
        verbose_name="آخرین تغییر",
    )

    class Meta:
        ordering = ["status", "course__title"]
        verbose_name = "کلاس در حال برگزاری"
        verbose_name_plural = "کلاس‌های در حال برگزاری"
        constraints = [
            models.UniqueConstraint(
                fields=[
                    "course",
                    "instructor",
                    "start_date_text",
                    "schedule_text",
                ],
                name="unique_catalog_course_offering",
            ),
        ]

    @property
    def effective_price_toman(self):
        if self.price_override_toman is not None:
            return self.price_override_toman

        return self.course.price_toman

    def __str__(self):
        return f"{self.course.title} - {self.instructor.full_name}"