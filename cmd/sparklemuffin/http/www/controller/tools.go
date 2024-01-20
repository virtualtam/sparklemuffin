// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package controller

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"github.com/virtualtam/netscape-go/v2"

	"github.com/virtualtam/sparklemuffin/cmd/sparklemuffin/http/www/httpcontext"
	"github.com/virtualtam/sparklemuffin/cmd/sparklemuffin/http/www/middleware"
	"github.com/virtualtam/sparklemuffin/cmd/sparklemuffin/http/www/view"
	"github.com/virtualtam/sparklemuffin/pkg/bookmark/exporting"
	"github.com/virtualtam/sparklemuffin/pkg/bookmark/importing"
)

type toolsHandlerContext struct {
	exportingService *exporting.Service
	importingService *importing.Service

	toolsView       *view.View
	toolsExportView *view.View
	toolsImportView *view.View
}

func RegisterToolsHandlers(
	r *chi.Mux,
	exportingService *exporting.Service,
	importingService *importing.Service,
) {
	hc := toolsHandlerContext{
		exportingService: exportingService,
		importingService: importingService,

		toolsView:       view.New("tools/tools.gohtml"),
		toolsExportView: view.New("tools/export.gohtml"),
		toolsImportView: view.New("tools/import.gohtml"),
	}

	// bookmark management tools
	r.Route("/tools", func(r chi.Router) {
		r.Use(func(h http.Handler) http.Handler {
			return middleware.AuthenticatedUser(h.ServeHTTP)
		})

		r.Get("/", hc.handleToolsView())
		r.Get("/export", hc.handleToolsExportView())
		r.Post("/export", hc.handleToolsExport())
		r.Get("/import", hc.handleToolsImportView())
		r.Post("/import", hc.handleToolsImport())
	})
}

// handleToolsView renders the tools page.
func (hc *toolsHandlerContext) handleToolsView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		user := httpcontext.UserValue(r.Context())
		viewData := view.Data{
			Content: user,
			Title:   "Tools",
		}

		hc.toolsView.Render(w, r, viewData)
	}
}

// handleToolsExportView renders the bookmark export page.
func (hc *toolsHandlerContext) handleToolsExportView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		user := httpcontext.UserValue(r.Context())
		viewData := view.Data{
			Content: user,
			Title:   "Export bookmarks",
		}

		hc.toolsExportView.Render(w, r, viewData)
	}
}

// handleToolsExport processes the bookmarks export form and sends the
// corresponding file to the client.
func (hc *toolsHandlerContext) handleToolsExport() func(w http.ResponseWriter, r *http.Request) {
	type exportForm struct {
		Visibility exporting.Visibility `schema:"visibility"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var form exportForm
		if err := decodeForm(r, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse bookmark export form")
			view.PutFlashError(w, "failed to process form")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		user := httpcontext.UserValue(r.Context())

		document, err := hc.exportingService.ExportAsNetscapeDocument(user.UUID, form.Visibility)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve bookmarks")
			view.PutFlashError(w, "failed to export bookmarks")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		marshaled, err := netscape.Marshal(document)
		if err != nil {
			log.Error().Err(err).Msg("failed to marshal Netscape bookmarks")
			view.PutFlashError(w, "failed to export bookmarks")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		filename := fmt.Sprintf("bookmarks-%s.htm", form.Visibility)

		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
		w.Header().Set("Content-Type", "application/octet-stream")
		_, err = w.Write(marshaled)

		if err != nil {
			log.Error().Err(err).Msg("failed to send marshaled Netscape bookmark export")
		}
	}
}

// handleToolsExportView renders the bookmark import page.
func (hc *toolsHandlerContext) handleToolsImportView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		user := httpcontext.UserValue(r.Context())
		viewData := view.Data{
			Content: user,
			Title:   "Import bookmarks",
		}

		hc.toolsImportView.Render(w, r, viewData)
	}
}

// handleToolsImport processes data submitted through the bookmark import form.
func (hc *toolsHandlerContext) handleToolsImport() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		multipartReader, err := r.MultipartReader()
		if err != nil {
			log.Error().Err(err).Msg("failed to access multipart reader")
			view.PutFlashError(w, "failed to process import form")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		var (
			importFileBuffer         bytes.Buffer
			onConflictStrategyBuffer bytes.Buffer
			visibilityBuffer         bytes.Buffer
		)
		importFileWriter := bufio.NewWriter(&importFileBuffer)
		onConflictStrategyWriter := bufio.NewWriter(&onConflictStrategyBuffer)
		visibilityWriter := bufio.NewWriter(&visibilityBuffer)

		for {
			part, err := multipartReader.NextPart()

			if errors.Is(err, io.EOF) {
				// no more parts to process
				break
			}

			if err != nil {
				log.Error().Err(err).Msg("failed to access multipart form data")
				view.PutFlashError(w, "failed to process import form")
				http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
				return
			}

			switch part.FormName() {
			case "importfile":
				_, err = io.Copy(importFileWriter, part)
			case "on-conflict":
				_, err = io.Copy(onConflictStrategyWriter, part)
			case "visibility":
				_, err = io.Copy(visibilityWriter, part)
			default:
				err = fmt.Errorf("unexpected multipart form field: %q", part.FormName())
			}

			if err != nil {
				log.Error().Err(err).Msg(fmt.Sprintf("failed to process multipart form part %q", part.FormName()))
				view.PutFlashError(w, "failed to process import form")
				http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
				return
			}
		}

		document, err := netscape.Unmarshal(importFileBuffer.Bytes())
		if err != nil {
			log.Error().Err(err).Msg("failed to process Netscape bookmark file")
			view.PutFlashError(w, "failed to import bookmarks from the uploaded file")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		user := httpcontext.UserValue(r.Context())
		overwrite := importing.OnConflictStrategy(onConflictStrategyBuffer.String())
		visibility := importing.Visibility(visibilityBuffer.String())

		importStatus, err := hc.importingService.ImportFromNetscapeDocument(user.UUID, document, visibility, overwrite)
		if err != nil {
			log.Error().Err(err).Msg("failed to save imported bookmarks")
			view.PutFlashError(w, "failed to save imported bookmarks")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		view.PutFlashSuccess(w, fmt.Sprintf("Import status: %s", importStatus.Summary()))
		http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
	}
}
