// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package view

import (
	"html/template"
	"net/http/httptest"
	"testing"
)

func TestViewRenderTemplate(t *testing.T) {
	tmpl := template.Must(template.New("base").Parse(`{{define "greeting"}}Hello, {{.}}!{{end}}`))
	v := &View{Template: tmpl}

	cases := []struct {
		tname   string
		name    string
		data    any
		want    string
		wantErr bool
	}{
		{
			tname: "renders a named template fragment",
			name:  "greeting",
			data:  "World",
			want:  "Hello, World!",
		},
		{
			tname:   "unknown template name",
			name:    "does-not-exist",
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			w := httptest.NewRecorder()

			err := v.RenderTemplate(w, tc.name, tc.data)

			if tc.wantErr {
				if err == nil {
					t.Fatal("want error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("want no error, got %q", err)
			}

			if got := w.Body.String(); got != tc.want {
				t.Errorf("want body %q, got %q", tc.want, got)
			}

			if got := w.Header().Get("Content-Type"); got != "text/html" {
				t.Errorf("want Content-Type %q, got %q", "text/html", got)
			}
		})
	}
}
