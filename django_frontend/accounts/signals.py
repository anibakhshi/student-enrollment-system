from django.contrib.auth import get_user_model
from django.db.models.signals import post_save
from django.dispatch import receiver

from .models import UserProfile


User = get_user_model()


@receiver(post_save, sender=User)
def ensure_user_profile(sender, instance, **kwargs):
    default_role = UserProfile.Role.STUDENT

    if instance.is_superuser or instance.is_staff:
        default_role = UserProfile.Role.ADMIN

    profile, created = UserProfile.objects.get_or_create(
        user=instance,
        defaults={"role": default_role},
    )

    if (
        not created
        and (instance.is_superuser or instance.is_staff)
        and profile.role != UserProfile.Role.ADMIN
    ):
        profile.role = UserProfile.Role.ADMIN
        profile.save(update_fields=("role", "updated_at"))