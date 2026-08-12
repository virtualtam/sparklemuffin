// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package view

import "net/http"

const (
	themeCookieName string = "theme"

	themeDark  string = "dark"
	themeLight string = "light"
)

// themeFromCookie returns the color theme ("light" or "dark") selected by
// the visitor, read from the cookie set by the client-side toggle (see
// js/theme-toggle.js). It returns "" if the cookie is unset or holds an
// unrecognized value, leaving the choice to the client-side
// prefers-color-scheme fallback rather than forcing a default here.
func themeFromCookie(r *http.Request) string {
	cookie, err := r.Cookie(themeCookieName)
	if err != nil {
		return ""
	}

	switch cookie.Value {
	case themeDark, themeLight:
		return cookie.Value
	default:
		return ""
	}
}
