/**
 * EasyMDE Initialization Script
 *
 * Initializes EasyMDE markdown editors on textareas marked with the
 * data-easymde attribute. Runs on page load, and again after any htmx swap
 * (e.g. a form loaded into a modal), scoped to the swapped subtree so
 * unrelated swaps elsewhere on the page are ignored.
 *
 * Listens for htmx:afterSettle, not htmx:afterSwap: htmx settles a fixed
 * set of attributes (class, style, width, height) ~20ms after a swap,
 * reconciling them back toward their value at swap time. Initializing on
 * afterSwap meant CodeMirror's textarea.style.display = "none" (set
 * synchronously during construction) got silently reverted moments later by
 * that settle step, since the server-rendered textarea has no style
 * attribute at swap time. Waiting for afterSettle avoids the reconciliation
 * entirely. The source textarea itself is hidden unconditionally via a CSS
 * rule on [data-easymde] (see easymde.css), not JS, so there's no flash of
 * raw textarea while waiting for afterSettle to fire.
 *
 * A previously-created instance's cleanup() is called first, to remove the
 * document-level keydown listener EasyMDE adds on top of CodeMirror (its
 * own DOM removal, via htmx replacing the old content, doesn't clean that
 * up).
 *
 * CodeMirror's fromTextArea only writes its content back to the source
 * textarea on the form's native "submit" event (see codemirror.js, on(
 * textarea.form, "submit", save)), listening for that event was attached
 * here, during htmx:afterSettle -- after htmx has already attached its own
 * submit listener to the form while processing the swapped-in content.
 * Listeners for the same event fire in attachment order, so htmx read the
 * textarea's stale value before CodeMirror's sync ran, and the edited
 * description was silently discarded. Hooking htmx:configRequest instead --
 * fired synchronously right before parameters are serialized -- sidesteps
 * that ordering entirely, by overwriting the stale parameter with the
 * editor's current value directly.
 *
 * autoDownloadFontAwesome and spellChecker are both disabled: by default
 * EasyMDE fetches a Font Awesome stylesheet and spell-check dictionaries
 * from third-party CDNs, which this application's CSP blocks (we already
 * bundle Font Awesome ourselves, and have no self-hosted dictionaries).
 *
 * Usage:
 *   <textarea id="description" data-easymde></textarea>
 *
 * Copyright VirtualTam 2022, 2026
 * SPDX-License-Identifier: MIT
 */
import EasyMDE from "easymde";

let currentEditor = null;
let currentEditorElement = null;

function initEasyMDE(root) {
    const target = root.querySelector("[data-easymde]");
    if (!target || target.dataset.easymdeInit) {
        return;
    }
    target.dataset.easymdeInit = "true";

    if (currentEditor) {
        currentEditor.cleanup();
        currentEditor = null;
        currentEditorElement = null;
    }

    currentEditorElement = target;
    currentEditor = new EasyMDE({
        element: target,
        autoDownloadFontAwesome: false,
        spellChecker: false,
        status: ["lines", "words", "cursor"],
        toolbar: [
            "bold",
            "italic",
            "heading",
            "|",
            "quote",
            "unordered-list",
            "ordered-list",
            "|",
            "link",
            "image",
            "|",
            "preview",
            "|",
            "guide",
        ],
        indentWithTabs: false,
        tabSize: 4,
    });
}

document.addEventListener("DOMContentLoaded", () => initEasyMDE(document));
document.addEventListener("htmx:afterSettle", (event) => initEasyMDE(event.detail.target));

document.addEventListener("htmx:configRequest", (event) => {
    if (currentEditor && currentEditorElement && event.detail.elt.contains(currentEditorElement)) {
        event.detail.parameters[currentEditorElement.name] = currentEditor.value();
    }
});
