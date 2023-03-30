package www

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
	"github.com/virtualtam/yawbe/pkg/http/www/templates"
	"github.com/virtualtam/yawbe/pkg/user"
)

// Data holds the data that can be rendered by views.
type Data struct {
	AtomFeedURL string
	Content     any
	Flash       *Flash
	User        *user.User
}

// view represents a Web view that will be rendered by the server in response to
// an HTTP client request.
type view struct {
	Template *template.Template
}

// newView returns an initialized View, preconfigured with the default
// application templates and page-specific templates.
func newView(templateFiles ...string) *view {
	templateFiles = append(templateFiles, layoutTemplateFiles()...)

	t, err := template.New("base").
		Funcs(template.FuncMap{
			"Join":           strings.Join,
			"MarkdownToHTML": markdownToHTMLFunc(),
		}).
		ParseFS(templates.FS, templateFiles...)

	if err != nil {
		panic(err)
	}

	return &view{
		Template: t,
	}
}

func (v *view) handle(w http.ResponseWriter, r *http.Request) {
	v.render(w, r, nil)
}

func (v *view) render(w http.ResponseWriter, r *http.Request, data interface{}) {
	w.Header().Set("Content-Type", "text/html")

	var viewData Data

	switch d := data.(type) {
	case Data:
		viewData = d
	default:
		viewData = Data{Content: data}
	}

	viewData.popFlash(w, r)
	viewData.User = userValue(r.Context())

	var buf bytes.Buffer

	if err := v.Template.Execute(&buf, viewData); err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	// Skip error checking as the HTTP headers have already been sent
	_, _ = io.Copy(w, &buf)
}

func (d *Data) popFlash(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(flashCookieName)
	if errors.Is(err, http.ErrNoCookie) {
		return
	} else if err != nil {
		log.Error().Err(err).Msg("failed to retrieve flash cookie")
		return
	}

	flash := &Flash{}
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
