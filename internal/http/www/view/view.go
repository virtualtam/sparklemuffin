// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package view

import (
	"bytes"
	"encoding/json"
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
	CSP         ContentSecurityPolicy
	Flash       *flash
	Title       string
	User        *user.User
}

// ContentSecurityPolicy exposes information required to enforce the Content Security Policy.
type ContentSecurityPolicy struct {
	// Nonce is a cryptographically secure nonce, unique to the request being served.
	Nonce string
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
			"dict":           dictFunc,
			"toJSON":         toJSONFunc,
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
	viewData.CSP.Nonce = httpcontext.CSPNonceValue(r.Context())

	var buf bytes.Buffer

	if err := v.Template.Execute(&buf, viewData); err != nil {
		log.Error().Err(err).Msg("failed to render view")
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	// Skip error checking as the HTTP headers have already been sent
	_, _ = io.Copy(w, &buf) // nolint:errcheck
}

// RenderTemplate renders a single named template, without the base page layout.
// It is intended for rendering HTML fragments, e.g. in response to htmx requests.
func (v *View) RenderTemplate(w http.ResponseWriter, name string, data any) error {
	var buf bytes.Buffer

	if err := v.Template.ExecuteTemplate(&buf, name, data); err != nil {
		return err
	}

	w.Header().Set("Content-Type", "text/html")

	_, err := io.Copy(w, &buf)
	return err
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

// dictFunc builds a map[string]any from an alternating list of string keys and
// values, so that a template can pass more than one value to a sub-template
// invoked via {{template "name" pipeline}}.
func dictFunc(values ...any) (map[string]any, error) {
	if len(values)%2 != 0 {
		return nil, fmt.Errorf("dict: expected an even number of arguments, got %d", len(values))
	}

	d := make(map[string]any, len(values)/2)

	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			return nil, fmt.Errorf("dict: keys must be strings, got %T", values[i])
		}

		d[key] = values[i+1]
	}

	return d, nil
}

// toJSONFunc builds a JSON object from an alternating list of string keys and
// values, for use as the value of an htmx hx-vals attribute. It returns a plain
// string (not template.JS) so that html/template still HTML-escapes it for the
// surrounding attribute, on top of the JSON-escaping encoding/json already
// applies to the values themselves.
func toJSONFunc(values ...any) (string, error) {
	d, err := dictFunc(values...)
	if err != nil {
		return "", err
	}

	b, err := json.Marshal(d)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func layoutTemplateFiles() []string {
	files, err := fs.Glob(templates.FS, "layout/*.gohtml")
	if err != nil {
		panic(err)
	}
	return files
}
