/**
 * Awesomplete Tag Completion
 *
 * Initializes Awesomplete autocomplete on input elements marked with the
 * tags ID. The suggestions are read from the data-list attribute. Runs on
 * page load, and again after any htmx swap (e.g. a form loaded into a
 * modal), scoped to the swapped subtree so unrelated swaps elsewhere on the
 * page are ignored.
 *
 * Listens for htmx:afterSettle, not htmx:afterSwap: htmx settles a fixed
 * set of attributes (class, style, width, height) ~20ms after a swap,
 * reconciling them back toward their value at swap time, which can clobber
 * DOM changes made by JS running during afterSwap (see easymde-init.js for
 * the concrete case that surfaced this). Using afterSettle here too, even
 * though Awesomplete doesn't hide anything itself, keeps both scripts
 * consistent.
 *
 * Usage:
 *   <input type="text" id="tags" data-list="tag1,tag2,tag3">
 *
 * Copyright VirtualTam 2022, 2026
 * SPDX-License-Identifier: MIT
 */
import Awesomplete from "awesomplete";

let currentAutocomplete = null;

function initCompleteTags(root) {
    const tags = root.querySelector("#tags");
    if (!tags || tags.dataset.completeTagsInit) {
        return;
    }
    tags.dataset.completeTagsInit = "true";

    if (currentAutocomplete) {
        currentAutocomplete.destroy();
        currentAutocomplete = null;
    }

    currentAutocomplete = new Awesomplete(tags, {
        autoFirst: true,
        sort: false,

        filter: function (text, input) {
            return Awesomplete.FILTER_CONTAINS(text, input.match(/[^ ]*$/)[0]);
        },

        item: function (text, input) {
            return Awesomplete.ITEM(text, input.match(/[^ ]*$/)[0]);
        },

        replace: function (text) {
            const before = this.input.value.match(/^.+\s+|/)[0];
            this.input.value = before + text + " ";
        },
    });
}

document.addEventListener("DOMContentLoaded", () => initCompleteTags(document));
document.addEventListener("htmx:afterSettle", (event) => initCompleteTags(event.detail.target));
