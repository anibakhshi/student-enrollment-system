"use strict";

document.addEventListener("DOMContentLoaded", () => {
    const form = document.querySelector(
        "[data-enrollment-form]"
    );

    if (!form) {
        return;
    }

    const submitButton = form.querySelector(
        "[data-submit-button]"
    );

    const submitLabel = form.querySelector(
        "[data-submit-label]"
    );

    const submitLoader = form.querySelector(
        "[data-submit-loader]"
    );

    let formChanged = false;
    let formSubmitted = false;

    form.addEventListener("input", () => {
        formChanged = true;
    });

    form.addEventListener("change", () => {
        formChanged = true;
    });

    form.addEventListener("submit", () => {
        formSubmitted = true;

        if (!submitButton) {
            return;
        }

        submitButton.disabled = true;
        submitButton.classList.add("is-loading");

        if (submitLabel) {
            submitLabel.hidden = true;
        }

        if (submitLoader) {
            submitLoader.hidden = false;
        }
    });

    window.addEventListener("beforeunload", (event) => {
        if (!formChanged || formSubmitted) {
            return;
        }

        event.preventDefault();
        event.returnValue = "";
    });
});