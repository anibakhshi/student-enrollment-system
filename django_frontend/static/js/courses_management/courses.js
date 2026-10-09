(() => {
    "use strict";

    const form = document.querySelector(
        "[data-course-form]"
    );

    if (!form) {
        return;
    }

    const submitButton = form.querySelector(
        "[data-submit-button]"
    );

    const startDate = form.querySelector(
        "#id_start_date"
    );

    const endDate = form.querySelector(
        "#id_end_date"
    );

    let changed = false;
    let submitted = false;

    const synchronizeEndDate = () => {
        if (!startDate || !endDate) {
            return;
        }

        endDate.min = startDate.value;

        if (
            startDate.value &&
            endDate.value &&
            endDate.value < startDate.value
        ) {
            endDate.value = startDate.value;
        }
    };

    startDate?.addEventListener(
        "change",
        synchronizeEndDate
    );

    synchronizeEndDate();

    form.addEventListener("input", () => {
        changed = true;
    });

    form.addEventListener("change", () => {
        changed = true;
    });

    form.addEventListener("submit", () => {
        submitted = true;

        if (submitButton) {
            submitButton.disabled = true;
            submitButton.classList.add("is-loading");

            const label = submitButton.querySelector(
                "span:last-child"
            );

            if (label) {
                label.textContent = "در حال ذخیره...";
            }
        }
    });

    window.addEventListener("beforeunload", (event) => {
        if (!changed || submitted) {
            return;
        }

        event.preventDefault();
        event.returnValue = "";
    });

    const firstInvalidField = form.querySelector(
        ".student-form-field--error input, " +
        ".student-form-field--error textarea, " +
        ".student-form-field--error select"
    );

    if (firstInvalidField) {
        firstInvalidField.focus({
            preventScroll: true,
        });

        firstInvalidField.scrollIntoView({
            behavior: "smooth",
            block: "center",
        });
    }
})();