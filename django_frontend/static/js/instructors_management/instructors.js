"use strict";

document.addEventListener("DOMContentLoaded", () => {
    const photoInput = document.querySelector(
        ".student-file-input"
    );

    const previewImage = document.getElementById(
        "studentPhotoPreviewImage"
    );

    const placeholder = document.getElementById(
        "studentPhotoPlaceholder"
    );

    if (!photoInput || !previewImage) {
        return;
    }

    const originalSource = previewImage.getAttribute("src") || "";
    const maxFileSize = 5 * 1024 * 1024;

    const allowedTypes = new Set([
        "image/jpeg",
        "image/png",
        "image/webp",
    ]);

    photoInput.addEventListener("change", () => {
        const file = photoInput.files?.[0];

        if (!file) {
            if (originalSource) {
                previewImage.src = originalSource;
                previewImage.hidden = false;
            } else {
                previewImage.src = "";
                previewImage.hidden = true;

                if (placeholder) {
                    placeholder.hidden = false;
                }
            }

            return;
        }

        if (!allowedTypes.has(file.type)) {
            window.alert(
                "فرمت تصویر باید JPG، PNG یا WebP باشد."
            );

            photoInput.value = "";
            return;
        }

        if (file.size > maxFileSize) {
            window.alert(
                "حجم تصویر نباید بیشتر از ۵ مگابایت باشد."
            );

            photoInput.value = "";
            return;
        }

        const reader = new FileReader();

        reader.addEventListener("load", (event) => {
            previewImage.src = event.target.result;
            previewImage.hidden = false;

            if (placeholder) {
                placeholder.hidden = true;
            }
        });

        reader.readAsDataURL(file);
    });

    const deleteForm = document.querySelector(
        ".student-delete-card form"
    );

    if (deleteForm) {
        deleteForm.addEventListener("submit", (event) => {
            const confirmed = window.confirm(
                "آیا از حذف این دانشجو اطمینان دارید؟"
            );

            if (!confirmed) {
                event.preventDefault();
            }
        });
    }
});