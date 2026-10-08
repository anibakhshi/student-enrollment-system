from django.db.models import Count, Prefetch, Q
from django.views.generic import DetailView, ListView

from .mixins import CatalogContextMixin
from .models import (
    Category,
    CourseCatalog,
    CourseOffering,
    InstructorProfile,
)


class CourseListView(
    CatalogContextMixin,
    ListView,
):
    model = CourseCatalog
    template_name = "catalog/course_list.html"
    context_object_name = "courses"
    paginate_by = 9

    page_title = "دوره‌های آموزشی"
    active_page = "courses"

    def get_queryset(self):
        active_offerings = (
            CourseOffering.objects.filter(
                is_active=True,
                instructor__is_active=True,
            )
            .select_related("instructor")
            .order_by(
                "status",
                "start_date_text",
                "id",
            )
        )

        queryset = (
            CourseCatalog.objects.filter(
                is_active=True,
                category__is_active=True,
            )
            .select_related("category")
            .prefetch_related(
                Prefetch(
                    "offerings",
                    queryset=active_offerings,
                )
            )
            .order_by(
                "-is_featured",
                "title",
                "id",
            )
        )

        query = self.get_filter_value("q")
        category_slug = self.get_filter_value("category")
        instructor_slug = self.get_filter_value("instructor")

        if query:
            queryset = queryset.filter(
                Q(title__icontains=query)
                | Q(short_description__icontains=query)
                | Q(description__icontains=query)
                | Q(prerequisites__icontains=query)
                | Q(category__name__icontains=query)
                | Q(
                    offerings__instructor__first_name__icontains=query
                )
                | Q(
                    offerings__instructor__last_name__icontains=query
                )
                | Q(
                    offerings__instructor__expertise__icontains=query
                )
            )

        if category_slug:
            queryset = queryset.filter(
                category__slug=category_slug,
            )

        if instructor_slug:
            queryset = queryset.filter(
                offerings__is_active=True,
                offerings__instructor__is_active=True,
                offerings__instructor__slug=instructor_slug,
            )

        return queryset.distinct()

    def get_filter_value(self, name):
        return self.request.GET.get(name, "").strip()

    def get_context_data(self, **kwargs):
        context = super().get_context_data(**kwargs)

        context["categories"] = (
            Category.objects.filter(is_active=True)
            .annotate(
                active_course_count=Count(
                    "courses",
                    filter=Q(courses__is_active=True),
                    distinct=True,
                )
            )
            .order_by(
                "display_order",
                "name",
            )
        )

        context["instructors"] = (
            InstructorProfile.objects.filter(is_active=True)
            .annotate(
                active_course_count=Count(
                    "offerings__course",
                    filter=Q(
                        offerings__is_active=True,
                        offerings__course__is_active=True,
                    ),
                    distinct=True,
                )
            )
            .order_by(
                "-is_featured",
                "first_name",
                "last_name",
            )
        )

        context["current_query"] = self.get_filter_value("q")
        context["current_category"] = self.get_filter_value(
            "category"
        )
        context["current_instructor"] = self.get_filter_value(
            "instructor"
        )

        context["filters_are_active"] = any(
            [
                context["current_query"],
                context["current_category"],
                context["current_instructor"],
            ]
        )

        return context


class CourseDetailView(
    CatalogContextMixin,
    DetailView,
):
    model = CourseCatalog
    template_name = "catalog/course_detail.html"
    context_object_name = "course"
    slug_url_kwarg = "slug"

    page_title = "جزئیات دوره"
    active_page = "courses"

    def get_queryset(self):
        active_offerings = (
            CourseOffering.objects.filter(
                is_active=True,
                instructor__is_active=True,
            )
            .select_related("instructor")
            .order_by(
                "status",
                "start_date_text",
                "schedule_text",
                "id",
            )
        )

        return (
            CourseCatalog.objects.filter(
                is_active=True,
                category__is_active=True,
            )
            .select_related("category")
            .prefetch_related(
                Prefetch(
                    "offerings",
                    queryset=active_offerings,
                )
            )
        )

    def get_context_data(self, **kwargs):
        context = super().get_context_data(**kwargs)

        context["page_title"] = self.object.title

        context["related_courses"] = (
            CourseCatalog.objects.filter(
                is_active=True,
                category__is_active=True,
                category=self.object.category,
            )
            .exclude(pk=self.object.pk)
            .select_related("category")
            .prefetch_related(
                Prefetch(
                    "offerings",
                    queryset=(
                        CourseOffering.objects.filter(
                            is_active=True,
                            instructor__is_active=True,
                        )
                        .select_related("instructor")
                        .order_by("id")
                    ),
                )
            )
            .order_by(
                "-is_featured",
                "title",
            )[:3]
        )

        return context


class InstructorListView(
    CatalogContextMixin,
    ListView,
):
    model = InstructorProfile
    template_name = "catalog/instructor_list.html"
    context_object_name = "instructors"
    paginate_by = 12

    page_title = "استادان آکادمی"
    active_page = "instructors"

    def get_queryset(self):
        query = self.request.GET.get("q", "").strip()

        queryset = (
            InstructorProfile.objects.filter(is_active=True)
            .annotate(
                active_course_count=Count(
                    "offerings__course",
                    filter=Q(
                        offerings__is_active=True,
                        offerings__course__is_active=True,
                    ),
                    distinct=True,
                )
            )
            .order_by(
                "-is_featured",
                "first_name",
                "last_name",
                "id",
            )
        )

        if query:
            queryset = queryset.filter(
                Q(first_name__icontains=query)
                | Q(last_name__icontains=query)
                | Q(expertise__icontains=query)
                | Q(bio__icontains=query)
                | Q(
                    offerings__course__title__icontains=query
                )
            ).distinct()

        return queryset

    def get_context_data(self, **kwargs):
        context = super().get_context_data(**kwargs)

        context["current_query"] = self.request.GET.get(
            "q",
            "",
        ).strip()

        context["filters_are_active"] = bool(
            context["current_query"]
        )

        return context


class InstructorDetailView(
    CatalogContextMixin,
    DetailView,
):
    model = InstructorProfile
    template_name = "catalog/instructor_detail.html"
    context_object_name = "instructor"
    slug_url_kwarg = "slug"

    page_title = "پروفایل استاد"
    active_page = "instructors"

    def get_queryset(self):
        active_offerings = (
            CourseOffering.objects.filter(
                is_active=True,
                course__is_active=True,
                course__category__is_active=True,
            )
            .select_related(
                "course",
                "course__category",
            )
            .order_by(
                "status",
                "course__title",
                "start_date_text",
                "id",
            )
        )

        return (
            InstructorProfile.objects.filter(is_active=True)
            .prefetch_related(
                Prefetch(
                    "offerings",
                    queryset=active_offerings,
                )
            )
        )

    def get_context_data(self, **kwargs):
        context = super().get_context_data(**kwargs)

        context["page_title"] = self.object.full_name

        active_offerings = list(
            self.object.offerings.all()
        )

        context["active_offerings"] = active_offerings
        context["active_course_count"] = len(
            {
                offering.course_id
                for offering in active_offerings
            }
        )

        return context