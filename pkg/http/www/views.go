package www

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"path/filepath"
)

const (
	// LayoutDir points to the location of the templates that will be used when
	// rendering views.
	LayoutDir = "pkg/http/www/templates/layout/"

	// TemplateExt defines which file types will be looked for when auto-adding
	// layout templates.
	TemplateExt = ".gohtml"
)

var (
	HomeView = NewView("pkg/http/www/templates/static/home.gohtml")
)

// Data holds the data that can be rendered by views.
type Data struct {
	Content any
}

// View represents a Web view that will be rendered by the server in response to
// an HTTP client request.
type View struct {
	Template *template.Template
}

// View represents a Web view that will be rendered by the server in response to
// an HTTP client request.
func (v *View) Render(w http.ResponseWriter, r *http.Request, data interface{}) {
	w.Header().Set("Content-Type", "text/html")

	var vd Data
	switch d := data.(type) {
	case Data:
		vd = d
	default:
		vd = Data{Content: data}
	}

	var buf bytes.Buffer

	if err := v.Template.Execute(&buf, vd); err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	// Skip error checking as the HTTP headers have already been sent
	_, _ = io.Copy(w, &buf)
}

func (v *View) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	v.Render(w, r, nil)
}

func layoutTemplateFiles() []string {
	files, err := filepath.Glob(LayoutDir + "*" + TemplateExt)
	if err != nil {
		panic(err)
	}
	return files
}

// NewView returns an initialized View, preconfigured with the default
// application templates and page-specific templates.
func NewView(templateFiles ...string) *View {
	templateFiles = append(templateFiles, layoutTemplateFiles()...)

	t, err := template.New("base").ParseFiles(templateFiles...)
	if err != nil {
		panic(err)
	}

	return &View{
		Template: t,
	}
}
