/**
 * Copyright VirtualTam 2022, 2026
 * SPDX-License-Identifier: MIT
 */

// Initializes EasyMDE on textareas marked with `data-easymde`.
//
// Usage: <textarea data-easymde></textarea>
import EasyMDE from "easymde";

let currentEditor = null;
let currentEditorElement = null;

function initEasyMDE(root) {
    const target = root.querySelector("[data-easymde]");
    if (!target || target.dataset.easymdeInit) {
        return;
    }
    target.dataset.easymdeInit = "true";

    // cleanup() removes the previous instance's document-level keydown listener.
    if (currentEditor) {
        currentEditor.cleanup();
        currentEditor = null;
        currentEditorElement = null;
    }

    currentEditorElement = target;
    currentEditor = new EasyMDE({
        element: target,
        // Disabled: the CSP blocks the third-party CDNs these features fetch from.
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
            "guide",
        ],
        // Disabled: would require maintaining two syntax highlighting implementations (Chroma).
        shortcuts: {
            togglePreview: null,
            toggleSideBySide: null,
        },
        indentWithTabs: false,
        tabSize: 4,
    });

    const editor = currentEditor;
    const modalEl = target.closest(".modal");
    if (modalEl) {
        refreshOnceVisible(editor, modalEl);
    }
}

// A modal starts hidden, so CodeMirror first lays out against a zero-width element.
// Poll with setTimeout, since requestAnimationFrame stalls on a background tab.
function refreshOnceVisible(editor, modalEl, attemptsLeft = 30) {
    if (getComputedStyle(modalEl).display === "none") {
        if (attemptsLeft > 0) {
            setTimeout(() => refreshOnceVisible(editor, modalEl, attemptsLeft - 1), 16);
        }
        return;
    }
    editor.codemirror.refresh();
}

document.addEventListener("DOMContentLoaded", () => initEasyMDE(document));
// Use afterSettle, not afterSwap: htmx reverts the inline style CodeMirror sets if we init any earlier.
document.addEventListener("htmx:afterSettle", (event) => initEasyMDE(event.detail.target));

// CodeMirror writes back to the textarea only on submit, after htmx has already read the stale value.
// Patch the parameter here instead, since configRequest fires just before serialization.
document.addEventListener("htmx:configRequest", (event) => {
    if (currentEditor && currentEditorElement && event.detail.elt.contains(currentEditorElement)) {
        event.detail.parameters[currentEditorElement.name] = currentEditor.value();
    }
});
