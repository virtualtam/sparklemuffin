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
 * Opens the bookmark edit modal only once its form has loaded: unlike the
 * other modals on this page (opened immediately via data-bs-toggle, then
 * filled in by htmx), the bookmark edit modal is opened after its form has
 * swapped in, so it never shows empty and then grows once EasyMDE's editor
 * arrives. Waits for htmx:afterSettle, not htmx:afterSwap, for the same
 * reconciliation reasons as easymde-init.js and complete-tags.js.
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
    if (event.detail.target.id !== "bookmark-edit-modal-body") {
        return;
    }

    const modalEl = document.getElementById("bookmarkEditModal");
    window.bootstrap.Modal.getOrCreateInstance(modalEl).show();
});
