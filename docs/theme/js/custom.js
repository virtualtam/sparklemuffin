/**
 * Copyright VirtualTam 2022, 2026
 * SPDX-License-Identifier: MIT
 */

// mdBook only auto-generates fold toggles for h3+; add the same toggle at the h2 level, changelog only.
document.addEventListener('DOMContentLoaded', () => {
    if (!document.location.pathname.endsWith('/changelog.html')) {
        return;
    }

    const onThisPage = document.querySelector('.on-this-page');
    if (!onThisPage) {
        return;
    }

    const topLevelItems = onThisPage.querySelectorAll(':scope > ol > li.header-item');
    const collapsible = [];

    topLevelItems.forEach(li => {
        const nestedOl = li.querySelector(':scope > ol.section');
        if (!nestedOl) {
            return;
        }

        collapsible.push(li);

        const span = li.querySelector(':scope > span.chapter-link-wrapper');
        const toggle = document.createElement('a');
        toggle.classList.add('chapter-fold-toggle', 'header-toggle');
        toggle.addEventListener('click', () => {
            li.classList.toggle('expanded');
        });
        const toggleDiv = document.createElement('div');
        toggleDiv.textContent = '❱';
        toggle.appendChild(toggleDiv);
        span.appendChild(toggle);
    });

    if (collapsible.length === 0) {
        return;
    }

    // Keep only the section containing mdBook's own .current-header expanded.
    function collapseInactiveSections() {
        const current = onThisPage.querySelector('a.current-header');
        const activeSection = current ? collapsible.find(li => li.contains(current)) : null;

        collapsible.forEach(li => {
            if (li !== activeSection) {
                li.classList.remove('expanded');
            }
        });
    }

    collapseInactiveSections();
    // Registered after mdBook's own scroll listener (this script loads last), so
    // .current-header and its expanded ancestor chain are already up to date here.
    document.addEventListener('scroll', collapseInactiveSections, { passive: true });
});
