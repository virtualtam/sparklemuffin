// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package controller

import (
	"net/http"

	"github.com/gorilla/schema"
)

var (
	schemaDecoder = schema.NewDecoder()
)

// decodeForm parses and decodes values from a submitted HTML form.
func decodeForm(r *http.Request, dst interface{}) error {
	if err := r.ParseForm(); err != nil {
		return err
	}

	if err := schemaDecoder.Decode(dst, r.PostForm); err != nil {
		return err
	}

	return nil
}
