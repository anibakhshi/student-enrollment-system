from django.contrib.auth import get_user_model
from django.db import IntegrityError, transaction
from django.test import TestCase

from .models import UserProfile


User = get_user_model()


class UserProfileTests(TestCase):
    def test_profile_is_created_for_new_user(self):
        user = User.objects.create_user(
            username="student",
            email="student@example.com",
            password="StrongPassword123!",
            first_name="سارا",
            last_name="محمدی",
        )

        self.assertTrue(
            UserProfile.objects.filter(user=user).exists()
        )
        self.assertEqual(
            user.profile.role,
            UserProfile.Role.STUDENT,
        )
        self.assertEqual(
            user.profile.display_name,
            "سارا محمدی",
        )

    def test_superuser_receives_admin_role(self):
        user = User.objects.create_superuser(
            username="administrator",
            email="admin@example.com",
            password="StrongPassword123!",
        )

        self.assertEqual(
            user.profile.role,
            UserProfile.Role.ADMIN,
        )
        self.assertTrue(user.profile.is_admin)
        self.assertTrue(user.profile.can_manage_education)

    def test_education_staff_can_manage_education(self):
        user = User.objects.create_user(
            username="education",
            password="StrongPassword123!",
        )
        user.profile.role = UserProfile.Role.EDUCATION_STAFF
        user.profile.save()

        self.assertTrue(user.profile.can_manage_education)
        self.assertFalse(user.profile.is_admin)

    def test_student_cannot_manage_education(self):
        user = User.objects.create_user(
            username="regular-student",
            password="StrongPassword123!",
        )

        self.assertFalse(user.profile.can_manage_education)
        self.assertFalse(user.profile.is_admin)

    def test_nonempty_national_code_must_be_unique(self):
        first_user = User.objects.create_user(
            username="first-student",
            password="StrongPassword123!",
        )
        first_user.profile.national_code = "0012345678"
        first_user.profile.save()

        second_user = User.objects.create_user(
            username="second-student",
            password="StrongPassword123!",
        )
        second_user.profile.national_code = "0012345678"

        with self.assertRaises(IntegrityError):
            with transaction.atomic():
                second_user.profile.save()

    def test_empty_national_codes_can_repeat(self):
        User.objects.create_user(
            username="student-one",
            password="StrongPassword123!",
        )
        User.objects.create_user(
            username="student-two",
            password="StrongPassword123!",
        )

        self.assertEqual(
            UserProfile.objects.filter(national_code="").count(),
            2,
        )