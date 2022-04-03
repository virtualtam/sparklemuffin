package www

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"

	"github.com/virtualtam/yawbe/pkg/http/www/templates"
	"github.com/virtualtam/yawbe/pkg/user"
)

// Data holds the data that can be rendered by views.
type Data struct {
	Content any
	Flash   *Flash
	User    *user.User
}

func (d *Data) putFlash(level flashLevel, message string) {
	d.Flash = &Flash{
		Level:   level,
		Message: message,
	}
}

// PutFlashError sets a Flash that will be rendered as an error message.
func (d *Data) PutFlashError(message string) {
	d.putFlash(flashLevelError, fmt.Sprintf("Error: %s", message))
}

// PutFlashInfo sets a Flash that will be rendered as an information message.
func (d *Data) PutFlashInfo(message string) {
	d.putFlash(flashLevelInfo, message)
}

// PutFlashSuccess sets a Flash that will be rendered as a success message.
func (d *Data) PutFlashSuccess(message string) {
	d.putFlash(flashLevelSuccess, message)
}

// PutFlashWarning sets a Flash that will be rendered as a warning message.
func (d *Data) PutFlashWarning(message string) {
	d.putFlash(flashLevelWarning, fmt.Sprintf("Warning: %s", message))
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

	t, err := template.New("base").ParseFS(templates.FS, templateFiles...)
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

	var vd Data
	switch d := data.(type) {
	case Data:
		vd = d
	default:
		vd = Data{Content: data}
	}

	vd.User = userValue(r.Context())

	var buf bytes.Buffer

	if err := v.Template.Execute(&buf, vd); err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	// Skip error checking as the HTTP headers have already been sent
	_, _ = io.Copy(w, &buf)
}

func layoutTemplateFiles() []string {
	files, err := fs.Glob(templates.FS, "layout/*.gohtml")
	if err != nil {
		panic(err)
	}
	return files
}
