/**
 * Theme toggle
 *
 * Wires up the navbar's color-theme toggle button. The initial theme itself
 * (including the first-visit prefers-color-scheme fallback) is decided by a
 * small inline script in layout/base.gohtml, which must run synchronously
 * before first paint to avoid a flash of the wrong theme -- this script only
 * needs to be ready once the button exists in the DOM.
 *
 * Copyright VirtualTam 2022, 2026
 * SPDX-License-Identifier: MIT
 */

function setTheme(theme) {
    document.documentElement.setAttribute("data-bs-theme", theme);
    document.cookie = "theme=" + theme + ";path=/;max-age=31536000;samesite=lax";
}

function initThemeToggle() {
    const toggle = document.getElementById("theme-toggle");
    if (!toggle) {
        return;
    }

    toggle.addEventListener("click", () => {
        const current = document.documentElement.getAttribute("data-bs-theme");
        setTheme(current === "dark" ? "light" : "dark");
    });
}

document.addEventListener("DOMContentLoaded", () => initThemeToggle());
