// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package view

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

// ErrorView represents a Web View that will be rendered by the server in response to
// an HTTP client request resulting in an HTTP 4xx error status.
type ErrorView struct {
	*View
}

// NewError returns an initialized ErrorView, preconfigured with the default
// application templates and page-specific templates.
func NewError() *ErrorView {
	return &ErrorView{
		View: New("page/error.gohtml"),
	}
}

func (v *ErrorView) Render(w http.ResponseWriter, r *http.Request, statusCode int) {
	var viewData Data
	viewData.popFlash(w, r)

	switch statusCode {
	case http.StatusUnauthorized:
		viewData.Title = "401 unauthorized"
		viewData.Content = "401 unauthorized"
	case http.StatusNotFound:
		viewData.Title = "404 page not found"
		viewData.Content = "404 page not found"
	default:
		viewData.Title = "Error"
		viewData.Content = fmt.Sprintf("%d", statusCode)
	}

	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "text/html")

	var buf bytes.Buffer

	if err := v.Template.Execute(&buf, viewData); err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	// Skip error checking as the HTTP headers have already been sent
	_, _ = io.Copy(w, &buf) // nolint:errcheck
}
