/**
 * Bootstrap Modal Bridge
 *
 * Bridges htmx to Bootstrap's own Modal JS component, which this app uses
 * instead of hand-driving modal visibility with Alpine.js.
 *
 * Closes modals on the server's close signal: htmx dispatches a
 * "modal:close" event on the document when a response carries the
 * HX-Trigger: modal:close header (set by the server after a successful edit
 * or delete). Any currently-open Bootstrap modal is hidden in response.
 *
 * Opens modals only once their content has loaded: trigger elements carry
 * hx-get/hx-target but no data-bs-toggle, so a modal never becomes visible
 * with the previous row's content (or nothing at all, on the first open)
 * still in it. Waits for htmx:afterSettle, not htmx:afterSwap, for the same
 * reconciliation reasons as easymde-init.js and complete-tags.js.
 *
 * The modal id is derived from its body's id by convention:
 * "foo-bar-modal-body" -> "fooBarModal".
 *
 * Copyright VirtualTam 2022, 2026
 * SPDX-License-Identifier: MIT
 */
document.addEventListener("modal:close", () => {
    document.querySelectorAll(".modal.show").forEach((modalEl) => {
        window.bootstrap.Modal.getInstance(modalEl)?.hide();
    });
});

document.addEventListener("htmx:afterSettle", (event) => {
    const bodyId = event.detail.target.id;
    if (!bodyId.endsWith("-modal-body")) {
        return;
    }

    const modalId = bodyId
        .slice(0, -"-modal-body".length)
        .split("-")
        .map((word, i) => (i === 0 ? word : word[0].toUpperCase() + word.slice(1)))
        .join("") + "Modal";

    const modalEl = document.getElementById(modalId);
    window.bootstrap.Modal.getOrCreateInstance(modalEl).show();
});
