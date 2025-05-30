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
	"github.com/virtualtam/opml-go"

	"github.com/virtualtam/sparklemuffin/internal/http/www/csrf"
	"github.com/virtualtam/sparklemuffin/internal/http/www/httpcontext"
	"github.com/virtualtam/sparklemuffin/internal/http/www/middleware"
	"github.com/virtualtam/sparklemuffin/internal/http/www/view"

	feedexporting "github.com/virtualtam/sparklemuffin/pkg/feed/exporting"
	feedimporting "github.com/virtualtam/sparklemuffin/pkg/feed/importing"
)

const (
	actionToolsFeedExport string = "tools-feed-export"
	actionToolsFeedImport string = "tools-feed-import"
)

type toolsHandlerContext struct {
	csrfService          *csrf.Service
	feedExportingService *feedexporting.Service
	feedImportingService *feedimporting.Service

	toolsView *view.View

	feedExportView *view.View
	feedImportView *view.View
}

func RegisterToolsHandlers(
	r *chi.Mux,
	csrfService *csrf.Service,
	feedExportingService *feedexporting.Service,
	feedImportingService *feedimporting.Service,
) {
	hc := toolsHandlerContext{
		csrfService:          csrfService,
		feedExportingService: feedExportingService,
		feedImportingService: feedImportingService,

		toolsView: view.New("tools/tools.gohtml"),

		feedExportView: view.New("tools/feed_export.gohtml"),
		feedImportView: view.New("tools/feed_import.gohtml"),
	}

	r.Route("/tools", func(r chi.Router) {
		r.Use(func(h http.Handler) http.Handler {
			return middleware.AuthenticatedUser(h.ServeHTTP)
		})

		r.Get("/", hc.handleToolsView())

		// feed management tools
		r.Route("/feeds", func(sr chi.Router) {
			sr.Get("/export", hc.handleFeedExportView())
			sr.Post("/export", hc.handleFeedExport())
			sr.Get("/import", hc.handleFeedImportView())
			sr.Post("/import", hc.handleFeedImport())
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

// handleFeedExportView renders the feed export page.
func (hc *toolsHandlerContext) handleFeedExportView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctxUser := httpcontext.UserValue(r.Context())
		csrfToken := hc.csrfService.Generate(ctxUser.UUID, actionToolsFeedExport)

		viewData := view.Data{
			Content: csrf.Data{
				CSRFToken: csrfToken,
			},
			Title: "Export feed subscriptions",
		}

		hc.feedExportView.Render(w, r, viewData)
	}
}

// handleFeedExport processes the feed subscription export form and sends the
// corresponding file to the client.
func (hc *toolsHandlerContext) handleFeedExport() func(w http.ResponseWriter, r *http.Request) {
	type exportForm struct {
		CSRFToken string `schema:"csrf_token"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctxUser := httpcontext.UserValue(r.Context())

		var form exportForm
		if err := decodeForm(r, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse feed export form")
			view.PutFlashError(w, "failed to process form")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		if !hc.csrfService.Validate(form.CSRFToken, ctxUser.UUID, actionToolsFeedExport) {
			log.Warn().Msg("failed to validate CSRF token")
			view.PutFlashError(w, "There was an error processing the form")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		opmlDocument, err := hc.feedExportingService.ExportAsOPMLDocument(*ctxUser)
		if err != nil {
			log.Error().Err(err).Msg("failed to encode feeds as OPML")
			view.PutFlashError(w, "failed to process form")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		marshaled, err := opml.Marshal(opmlDocument)
		if err != nil {
			log.Error().Err(err).Msg("failed to marshal OPML feed subscriptions")
			view.PutFlashError(w, "failed to export feed subscriptions")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		w.Header().Set("Content-Disposition", "attachment; filename=feeds.opml")
		w.Header().Set("Content-Type", "application/xml")

		if _, err := w.Write(marshaled); err != nil {
			log.Error().Err(err).Msg("failed to send OPML export")
		}
	}
}

// handleFeedExportView renders the feed subscription import page.
func (hc *toolsHandlerContext) handleFeedImportView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctxUser := httpcontext.UserValue(r.Context())
		csrfToken := hc.csrfService.Generate(ctxUser.UUID, actionToolsFeedImport)

		viewData := view.Data{
			Content: csrf.Data{
				CSRFToken: csrfToken,
			},
			Title: "Import feed subscriptions",
		}

		hc.feedImportView.Render(w, r, viewData)
	}
}

// handleFeedImport processes data submitted through the feed subscription import form.
func (hc *toolsHandlerContext) handleFeedImport() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		multipartReader, err := r.MultipartReader()
		if err != nil {
			log.Error().Err(err).Msg("failed to access multipart reader")
			view.PutFlashError(w, "failed to process import form")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		var (
			csrfTokenBuffer  bytes.Buffer
			importFileBuffer bytes.Buffer
		)
		csrfTokenWriter := bufio.NewWriter(&csrfTokenBuffer)
		importFileWriter := bufio.NewWriter(&importFileBuffer)

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
			case "csrf_token":
				_, err = io.Copy(csrfTokenWriter, part)
			case "importfile":
				_, err = io.Copy(importFileWriter, part)
			default:
				err = fmt.Errorf("unexpected multipart form field: %q", part.FormName())
			}

			if err != nil {
				log.Error().Err(err).Msg(fmt.Sprintf("failed to process multipart form part %q", part.FormName()))
				view.PutFlashError(w, "failed to process import form")
				http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
				return
			}
		}

		ctxUser := httpcontext.UserValue(r.Context())

		if !hc.csrfService.Validate(csrfTokenBuffer.String(), ctxUser.UUID, actionToolsFeedImport) {
			log.Warn().Msg("failed to validate CSRF token")
			view.PutFlashError(w, "There was an error processing the form")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		document, err := opml.Unmarshal(importFileBuffer.Bytes())
		if err != nil {
			log.Error().Err(err).Msg("failed to process OPML feed subscription file")
			view.PutFlashError(w, "failed to import feed subscriptions from the uploaded file")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		importStatus, err := hc.feedImportingService.ImportFromOPMLDocument(ctxUser.UUID, document)
		if err != nil {
			log.Error().Err(err).Msg("failed to save imported feed subscriptions")
			view.PutFlashError(w, "failed to save imported feed subscriptions")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		view.PutFlashSuccess(w, fmt.Sprintf("Import status: %s", importStatus.UserSummary()))
		http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
	}
}
