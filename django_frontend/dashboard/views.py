from collections import Counter
from datetime import datetime
from typing import Any

from django.conf import settings
from django.shortcuts import render

from .api_client import GoAPIError, go_api


def _get_collection(
    path: str,
    errors: list[str],
) -> list[dict[str, Any]]:
    try:
        response = go_api.get(path)
        data = response.get("data", [])

        if isinstance(data, list):
            return data

        return []
    except GoAPIError as exc:
        errors.append(exc.message)
        return []


def _parse_datetime(value: str | None) -> datetime:
    if not value:
        return datetime.min

    try:
        return datetime.fromisoformat(
            value.replace("Z", "+00:00")
        )
    except ValueError:
        return datetime.min


def _status_label(status: str) -> str:
    labels = {
        "pending": "در انتظار",
        "confirmed": "تأییدشده",
        "completed": "تکمیل‌شده",
        "cancelled": "لغوشده",
        "active": "فعال",
        "inactive": "غیرفعال",
        "open": "باز",
        "closed": "بسته",
        "succeeded": "موفق",
        "failed": "ناموفق",
        "refunded": "بازگشت وجه",
    }

    return labels.get(status, status)


def _status_class(status: str) -> str:
    classes = {
        "pending": "warning",
        "confirmed": "info",
        "completed": "success",
        "cancelled": "danger",
        "active": "success",
        "inactive": "muted",
        "open": "success",
        "closed": "danger",
        "succeeded": "success",
        "failed": "danger",
        "refunded": "info",
    }

    return classes.get(status, "muted")


def index(request):
    errors: list[str] = []

    api_available = False
    api_health: dict[str, Any] = {}

    try:
        health_response = go_api.get("/health")
        api_health = health_response.get("data", {})
        api_available = True
    except GoAPIError as exc:
        errors.append(exc.message)

    students = _get_collection(
        "/api/v1/students",
        errors,
    )
    instructors = _get_collection(
        "/api/v1/instructors",
        errors,
    )
    courses = _get_collection(
        "/api/v1/courses",
        errors,
    )
    enrollments = _get_collection(
        "/api/v1/enrollments",
        errors,
    )
    payments = _get_collection(
        "/api/v1/payments",
        errors,
    )

    student_lookup = {
        student.get("id"): (
            f"{student.get('first_name', '')} "
            f"{student.get('last_name', '')}"
        ).strip()
        for student in students
    }

    course_lookup = {
        course.get("id"): course.get(
            "title",
            "دوره بدون عنوان",
        )
        for course in courses
    }

    recent_enrollments = sorted(
        enrollments,
        key=lambda item: _parse_datetime(
            item.get("created_at")
        ),
        reverse=True,
    )[:5]

    for enrollment in recent_enrollments:
        enrollment["student_name"] = student_lookup.get(
            enrollment.get("student_id"),
            "دانشجوی نامشخص",
        )
        enrollment["course_title"] = course_lookup.get(
            enrollment.get("course_id"),
            "دوره نامشخص",
        )
        enrollment["status_label"] = _status_label(
            enrollment.get("status", "")
        )
        enrollment["status_class"] = _status_class(
            enrollment.get("status", "")
        )

    recent_payments = sorted(
        payments,
        key=lambda item: _parse_datetime(
            item.get("created_at")
        ),
        reverse=True,
    )[:5]

    enrollment_lookup = {
        enrollment.get("id"): enrollment
        for enrollment in enrollments
    }

    for payment in recent_payments:
        enrollment = enrollment_lookup.get(
            payment.get("enrollment_id"),
            {},
        )

        payment["student_name"] = student_lookup.get(
            enrollment.get("student_id"),
            "دانشجوی نامشخص",
        )
        payment["status_label"] = _status_label(
            payment.get("status", "")
        )
        payment["status_class"] = _status_class(
            payment.get("status", "")
        )
        payment["amount_display"] = (
            f"{float(payment.get('amount', 0)):,.0f}"
        )

    successful_payments = [
        payment
        for payment in payments
        if payment.get("status") == "succeeded"
    ]

    total_revenue = sum(
        float(payment.get("amount", 0))
        for payment in successful_payments
    )

    enrollment_status_counts = Counter(
        enrollment.get("status", "unknown")
        for enrollment in enrollments
    )

    chart_items = [
        {
            "label": "در انتظار",
            "count": enrollment_status_counts.get(
                "pending",
                0,
            ),
            "class": "warning",
        },
        {
            "label": "تأییدشده",
            "count": enrollment_status_counts.get(
                "confirmed",
                0,
            ),
            "class": "info",
        },
        {
            "label": "تکمیل‌شده",
            "count": enrollment_status_counts.get(
                "completed",
                0,
            ),
            "class": "success",
        },
        {
            "label": "لغوشده",
            "count": enrollment_status_counts.get(
                "cancelled",
                0,
            ),
            "class": "danger",
        },
    ]

    largest_chart_value = max(
        [item["count"] for item in chart_items] + [1]
    )

    for item in chart_items:
        item["percentage"] = round(
            item["count"] / largest_chart_value * 100
        )

    context = {
        "page_title": "داشبورد مدیریت",
        "active_page": "dashboard",
        "api_available": api_available,
        "api_health": api_health,
        "api_base_url": settings.GO_API_BASE_URL,
        "errors": list(dict.fromkeys(errors)),
        "stats": {
            "students": len(students),
            "instructors": len(instructors),
            "courses": len(courses),
            "enrollments": len(enrollments),
            "payments": len(payments),
            "successful_payments": len(
                successful_payments
            ),
            "total_revenue": f"{total_revenue:,.0f}",
        },
        "chart_items": chart_items,
        "recent_enrollments": recent_enrollments,
        "recent_payments": recent_payments,
    }

    return render(
        request,
        "dashboard/index.html",
        context,
    )