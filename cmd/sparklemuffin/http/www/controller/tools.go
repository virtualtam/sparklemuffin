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
	"github.com/virtualtam/opml-go"

	"github.com/virtualtam/sparklemuffin/cmd/sparklemuffin/http/www/csrf"
	"github.com/virtualtam/sparklemuffin/cmd/sparklemuffin/http/www/httpcontext"
	"github.com/virtualtam/sparklemuffin/cmd/sparklemuffin/http/www/middleware"
	"github.com/virtualtam/sparklemuffin/cmd/sparklemuffin/http/www/view"
	bookmarkexporting "github.com/virtualtam/sparklemuffin/pkg/bookmark/exporting"
	bookmarkimporting "github.com/virtualtam/sparklemuffin/pkg/bookmark/importing"
	feedexporting "github.com/virtualtam/sparklemuffin/pkg/feed/exporting"
)

const (
	actionToolsBookmarkExport string = "tools-bookmark-export"
	actionToolsBookmarkImport string = "tools-bookmark-import"
	actionToolsFeedExport     string = "tools-feed-export"
)

type toolsHandlerContext struct {
	bookmarkExportingService *bookmarkexporting.Service
	bookmarkImportingService *bookmarkimporting.Service
	csrfService              *csrf.Service
	feedExportingService     *feedexporting.Service

	toolsView          *view.View
	bookmarkExportView *view.View
	bookmarkImportView *view.View
	feedExportView     *view.View
}

func RegisterToolsHandlers(
	r *chi.Mux,
	bookmarkExportingService *bookmarkexporting.Service,
	bookmarkImportingService *bookmarkimporting.Service,
	csrfService *csrf.Service,
	feedExportingService *feedexporting.Service,
) {
	hc := toolsHandlerContext{
		bookmarkExportingService: bookmarkExportingService,
		bookmarkImportingService: bookmarkImportingService,
		csrfService:              csrfService,
		feedExportingService:     feedExportingService,

		toolsView: view.New("tools/tools.gohtml"),

		bookmarkExportView: view.New("tools/bookmark_export.gohtml"),
		bookmarkImportView: view.New("tools/bookmark_import.gohtml"),
		feedExportView:     view.New("tools/feed_export.gohtml"),
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

		// feed management tools
		r.Route("/feeds", func(sr chi.Router) {
			sr.Get("/export", hc.handleFeedExportView())
			sr.Post("/export", hc.handleFeedExport())
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
		ctxUser := httpcontext.UserValue(r.Context())
		csrfToken := hc.csrfService.Generate(ctxUser.UUID, actionToolsBookmarkExport)

		viewData := view.Data{
			Content: csrf.Data{
				CSRFToken: csrfToken,
			},
			Title: "Export bookmarks",
		}

		hc.bookmarkExportView.Render(w, r, viewData)
	}
}

// handleBookmarkExport processes the bookmarks export form and sends the
// corresponding file to the client.
func (hc *toolsHandlerContext) handleBookmarkExport() func(w http.ResponseWriter, r *http.Request) {
	type exportForm struct {
		CSRFToken  string                       `schema:"csrf_token"`
		Visibility bookmarkexporting.Visibility `schema:"visibility"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var form exportForm
		if err := decodeForm(r, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse bookmark export form")
			view.PutFlashError(w, "failed to process form")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		ctxUser := httpcontext.UserValue(r.Context())

		if !hc.csrfService.Validate(form.CSRFToken, ctxUser.UUID, actionToolsBookmarkExport) {
			log.Warn().Msg("failed to validate CSRF token")
			view.PutFlashError(w, "There was an error processing the form")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		document, err := hc.bookmarkExportingService.ExportAsNetscapeDocument(ctxUser.UUID, form.Visibility)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve bookmarks")
			view.PutFlashError(w, "failed to export bookmarks")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		marshaled, err := netscape.Marshal(document)
		if err != nil {
			log.Error().Err(err).Msg("failed to marshal Netscape bookmarks")
			view.PutFlashError(w, "failed to export bookmarks")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
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
		ctxUser := httpcontext.UserValue(r.Context())
		csrfToken := hc.csrfService.Generate(ctxUser.UUID, actionToolsBookmarkImport)

		viewData := view.Data{
			Content: csrf.Data{
				CSRFToken: csrfToken,
			},
			Title: "Import bookmarks",
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
			csrfTokenBuffer          bytes.Buffer
			importFileBuffer         bytes.Buffer
			onConflictStrategyBuffer bytes.Buffer
			visibilityBuffer         bytes.Buffer
		)
		csrfTokenWriter := bufio.NewWriter(&csrfTokenBuffer)
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
			case "csrf_token":
				_, err = io.Copy(csrfTokenWriter, part)
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

		ctxUser := httpcontext.UserValue(r.Context())

		if !hc.csrfService.Validate(csrfTokenBuffer.String(), ctxUser.UUID, actionToolsBookmarkImport) {
			log.Warn().Msg("failed to validate CSRF token")
			view.PutFlashError(w, "There was an error processing the form")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
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
