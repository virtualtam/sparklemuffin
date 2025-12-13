/**
 * EasyMDE Initialization Script
 *
 * Initializes EasyMDE markdown editors on textareas marked with the
 * data-easymde attribute.
 *
 * Usage:
 *   <textarea id="description" data-easymde></textarea>
 *
 * Copyright (c) VirtualTam
 * SPDX-License-Identifier: MIT
 */
import EasyMDE from "easymde";

document.addEventListener("DOMContentLoaded", function () {
    const target = document.querySelector("[data-easymde]");
    if (!target) {
        return;
    }

    new EasyMDE({
        element: target,
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
});
