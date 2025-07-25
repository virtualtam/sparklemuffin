// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package view

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/virtualtam/sparklemuffin/internal/http/www/httpcontext"
	"github.com/virtualtam/sparklemuffin/internal/http/www/templates"
	"github.com/virtualtam/sparklemuffin/pkg/user"
)

const (
	appTitle string = "SparkleMuffin"
)

// Data holds the data that can be rendered by views.
type Data struct {
	AtomFeedURL string
	Content     any
	Flash       *flash
	Title       string
	User        *user.User
}

// FormContent holds the data that can be rendered by a form, protected with a CSRF token.
type FormContent struct {
	CSRFToken string
	Content   any
}

// View represents a Web View that will be rendered by the server in response to
// an HTTP client request.
type View struct {
	Template *template.Template
}

// New returns an initialized View, preconfigured with the default
// application templates and page-specific templates.
func New(templateFiles ...string) *View {
	templateFiles = append(templateFiles, layoutTemplateFiles()...)

	t, err := template.New("base").
		Funcs(template.FuncMap{
			"Join":           strings.Join,
			"MarkdownToHTML": MarkdownToHTMLFunc(),
			"mod":            func(i, j int) int { return i % j },
		}).
		ParseFS(templates.FS, templateFiles...)

	if err != nil {
		panic(err)
	}

	return &View{
		Template: t,
	}
}

// Handle renders the view with no data.
func (v *View) Handle(w http.ResponseWriter, r *http.Request) {
	v.Render(w, r, nil)
}

// Render renders the view with the given data.
func (v *View) Render(w http.ResponseWriter, r *http.Request, data any) {
	w.Header().Set("Content-Type", "text/html")

	var viewData Data

	switch d := data.(type) {
	case Data:
		viewData = d
	default:
		viewData = Data{Content: data}
	}

	if viewData.Title == "" {
		viewData.Title = appTitle
	} else {
		viewData.Title = fmt.Sprintf("%s | %s", viewData.Title, appTitle)
	}

	viewData.popFlash(w, r)
	viewData.User = httpcontext.UserValue(r.Context())

	var buf bytes.Buffer

	if err := v.Template.Execute(&buf, viewData); err != nil {
		log.Error().Err(err).Msg("failed to render view")
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	// Skip error checking as the HTTP headers have already been sent
	_, _ = io.Copy(w, &buf) // nolint:errcheck
}

func (d *Data) popFlash(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(flashCookieName)
	if errors.Is(err, http.ErrNoCookie) {
		return
	} else if err != nil {
		log.Error().Err(err).Msg("failed to retrieve flash cookie")
		return
	}

	flash := &flash{}
	if err := flash.base64URLDecode(cookie.Value); err != nil {
		log.Error().Err(err).Msg("failed to decode flash cookie")
		return
	}

	d.Flash = flash

	cookie = &http.Cookie{
		Name:     flashCookieName,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
		Expires:  time.Unix(1, 0),
	}
	http.SetCookie(w, cookie)
}

func layoutTemplateFiles() []string {
	files, err := fs.Glob(templates.FS, "layout/*.gohtml")
	if err != nil {
		panic(err)
	}
	return files
}
