from django.contrib.auth import get_user_model
from django.db import IntegrityError, transaction
from django.test import TestCase, override_settings
from django.urls import reverse
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

        from django.urls import reverse

@override_settings(
    STORAGES={
        "default": {
            "BACKEND": (
                "django.core.files.storage."
                "FileSystemStorage"
            ),
        },
        "staticfiles": {
            "BACKEND": (
                "django.contrib.staticfiles.storage."
                "StaticFilesStorage"
            ),
        },
    }
)

class AuthenticationViewTests(TestCase):
    @classmethod
    def setUpTestData(cls):
        cls.password = "StrongPassword123!"
        cls.user = User.objects.create_user(
            username="anita",
            email="anita@example.com",
            password=cls.password,
            first_name="آنیتا",
            last_name="بخشی",
        )

    def test_login_page_is_available(self):
        response = self.client.get(
            reverse("accounts:login")
        )

        self.assertEqual(response.status_code, 200)
        self.assertContains(response, "ورود به حساب کاربری")
        self.assertContains(response, "نام کاربری")
        self.assertContains(response, "رمز عبور")

    def test_invalid_login_displays_persian_error(self):
        response = self.client.post(
            reverse("accounts:login"),
            {
                "username": self.user.username,
                "password": "IncorrectPassword!",
            },
        )

        self.assertEqual(response.status_code, 200)
        self.assertContains(
            response,
            "نام کاربری یا رمز عبور صحیح نیست",
        )
        self.assertFalse(
            response.wsgi_request.user.is_authenticated
        )

    def test_valid_login_redirects_to_dashboard(self):
        response = self.client.post(
            reverse("accounts:login"),
            {
                "username": self.user.username,
                "password": self.password,
            },
        )

        self.assertRedirects(
            response,
            reverse("dashboard:index"),
            fetch_redirect_response=False,
        )

    def test_anonymous_user_is_redirected_from_profile(self):
        profile_url = reverse("accounts:profile")

        response = self.client.get(profile_url)

        expected_url = (
            f"{reverse('accounts:login')}"
            f"?next={profile_url}"
        )

        self.assertRedirects(
            response,
            expected_url,
            fetch_redirect_response=False,
        )

    def test_authenticated_user_can_view_profile(self):
        self.client.force_login(self.user)

        response = self.client.get(
            reverse("accounts:profile")
        )

        self.assertEqual(response.status_code, 200)
        self.assertContains(response, "آنیتا بخشی")
        self.assertContains(response, "دانشجو")
        self.assertContains(response, self.user.email)

    def test_logout_requires_post_and_ends_session(self):
        self.client.force_login(self.user)

        get_response = self.client.get(
            reverse("accounts:logout")
        )
        self.assertEqual(get_response.status_code, 405)

        post_response = self.client.post(
            reverse("accounts:logout")
        )

        self.assertRedirects(
            post_response,
            reverse("accounts:login"),
            fetch_redirect_response=False,
        )

        profile_response = self.client.get(
            reverse("accounts:profile")
        )
        self.assertEqual(profile_response.status_code, 302)