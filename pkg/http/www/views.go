package www

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"

	"github.com/virtualtam/yawbe/pkg/http/www/templates"
)

var (
	HomeView  = NewView("static/home.gohtml")
	loginView = NewView("user/login.gohtml")
)

// Data holds the data that can be rendered by views.
type Data struct {
	Alert   *Alert
	Content any
}

func (d *Data) alert(level alertLevel, message string) {
	d.Alert = &Alert{
		Level:   level,
		Message: message,
	}
}

// AlertError sets an Alert that will be rendered as an error message.
func (d *Data) AlertError(err error) {
	d.alert(alertLevelError, fmt.Sprintf("Error: %s", err))
}

// AlertInfo sets an Alert that will be rendered as an information message.
func (d *Data) AlertInfo(message string) {
	d.alert(alertLevelInfo, message)
}

// AlertSuccess sets an Alert that will be rendered as a success message.
func (d *Data) AlertSuccess(message string) {
	d.alert(alertLevelSuccess, message)
}

// AlertWarning sets an Alert that will be rendered as a warning message.
func (d *Data) AlertWarning(message string) {
	d.alert(alertLevelWarning, fmt.Sprintf("Warning: %s", message))
}

// alertLevel represents the severity level of an Alert that will be displayed
// to the user.
type alertLevel string

const (
	alertLevelError   alertLevel = "error"
	alertLevelInfo    alertLevel = "info"
	alertLevelSuccess alertLevel = "success"
	alertLevelWarning alertLevel = "warning"
)

// Alert represents an alert message that will be displayed to the user when
// rendering a View.
type Alert struct {
	Level   alertLevel
	Message string
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
	files, err := fs.Glob(templates.FS, "layout/*.gohtml")
	if err != nil {
		panic(err)
	}
	return files
}

// NewView returns an initialized View, preconfigured with the default
// application templates and page-specific templates.
func NewView(templateFiles ...string) *View {
	templateFiles = append(templateFiles, layoutTemplateFiles()...)

	t, err := template.New("base").ParseFS(templates.FS, templateFiles...)
	if err != nil {
		panic(err)
	}

	return &View{
		Template: t,
	}
}
