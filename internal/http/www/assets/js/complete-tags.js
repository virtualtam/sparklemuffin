/**
 * Awesomplete Tag Completion
 *
 * Initializes Awesomplete autocomplete on input elements marked with the
 * tags ID. The suggestions are read from the data-list attribute.
 *
 * Usage:
 *   <input type="text" id="tags" data-list="tag1,tag2,tag3">
 *
 * Copyright (c) VirtualTam
 * SPDX-License-Identifier: MIT
 */
import Awesomplete from "awesomplete";

document.addEventListener("DOMContentLoaded", function () {
    const tags = document.getElementById("tags");
    if (!tags) {
        return;
    }

    new Awesomplete(tags, {
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
});
