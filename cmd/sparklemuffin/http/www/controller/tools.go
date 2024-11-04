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
	bookmarkexporting "github.com/virtualtam/sparklemuffin/pkg/bookmark/exporting"
	bookmarkimporting "github.com/virtualtam/sparklemuffin/pkg/bookmark/importing"
)

type toolsHandlerContext struct {
	bookmarkExportingService *bookmarkexporting.Service
	bookmarkImportingService *bookmarkimporting.Service

	toolsView          *view.View
	bookmarkExportView *view.View
	bookmarkImportView *view.View
}

func RegisterToolsHandlers(
	r *chi.Mux,
	bookmarkExportingService *bookmarkexporting.Service,
	bookmarkImportingService *bookmarkimporting.Service,
) {
	hc := toolsHandlerContext{
		bookmarkExportingService: bookmarkExportingService,
		bookmarkImportingService: bookmarkImportingService,

		toolsView: view.New("tools/tools.gohtml"),

		bookmarkExportView: view.New("tools/bookmark_export.gohtml"),
		bookmarkImportView: view.New("tools/bookmark_import.gohtml"),
	}

	r.Route("/tools", func(r chi.Router) {
		r.Use(func(h http.Handler) http.Handler {
			return middleware.AuthenticatedUser(h.ServeHTTP)
		})

		r.Get("/", hc.handleToolsView())

		// bookmark management tools
		r.Route("/bookmarks", func(sr chi.Router) {
			sr.Get("/export", hc.handleBookmarkExportView())
			sr.Post("/export", hc.handleBookmarkExport())
			sr.Get("/import", hc.handleBookmarkImportView())
			sr.Post("/import", hc.handleBookmarkImport())
		})
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

// handleBookmarkExportView renders the bookmark export page.
func (hc *toolsHandlerContext) handleBookmarkExportView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		user := httpcontext.UserValue(r.Context())
		viewData := view.Data{
			Content: user,
			Title:   "Export bookmarks",
		}

		hc.bookmarkExportView.Render(w, r, viewData)
	}
}

// handleBookmarkExport processes the bookmarks export form and sends the
// corresponding file to the client.
func (hc *toolsHandlerContext) handleBookmarkExport() func(w http.ResponseWriter, r *http.Request) {
	type exportForm struct {
		Visibility bookmarkexporting.Visibility `schema:"visibility"`
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

		document, err := hc.bookmarkExportingService.ExportAsNetscapeDocument(user.UUID, form.Visibility)
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
func (hc *toolsHandlerContext) handleBookmarkImportView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		user := httpcontext.UserValue(r.Context())
		viewData := view.Data{
			Content: user,
			Title:   "Import bookmarks",
		}

		hc.bookmarkImportView.Render(w, r, viewData)
	}
}

// handleBookmarkImport processes data submitted through the bookmark import form.
func (hc *toolsHandlerContext) handleBookmarkImport() func(w http.ResponseWriter, r *http.Request) {
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
		overwrite := bookmarkimporting.OnConflictStrategy(onConflictStrategyBuffer.String())
		visibility := bookmarkimporting.Visibility(visibilityBuffer.String())

		importStatus, err := hc.bookmarkImportingService.ImportFromNetscapeDocument(user.UUID, document, visibility, overwrite)
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
