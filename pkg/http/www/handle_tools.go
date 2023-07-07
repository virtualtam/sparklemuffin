package www

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
	"github.com/virtualtam/netscape-go/v2"

	"github.com/virtualtam/sparklemuffin/pkg/exporting"
	"github.com/virtualtam/sparklemuffin/pkg/importing"
)

type toolsHandlerContext struct {
	exportingService *exporting.Service
	importingService *importing.Service

	toolsView       *view
	toolsExportView *view
	toolsImportView *view
}

func registerToolsHandlers(
	r *chi.Mux,
	exportingService *exporting.Service,
	importingService *importing.Service,
) {
	hc := toolsHandlerContext{
		exportingService: exportingService,
		importingService: importingService,

		toolsView:       newView("tools/tools.gohtml"),
		toolsExportView: newView("tools/export.gohtml"),
		toolsImportView: newView("tools/import.gohtml"),
	}

	// bookmark management tools
	r.Route("/tools", func(r chi.Router) {
		r.Use(func(h http.Handler) http.Handler {
			return authenticatedUser(h.ServeHTTP)
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
		user := userValue(r.Context())
		viewData := Data{
			Content: user,
			Title:   "Tools",
		}

		hc.toolsView.render(w, r, viewData)
	}
}

// handleToolsExportView renders the bookmark export page.
func (hc *toolsHandlerContext) handleToolsExportView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		user := userValue(r.Context())
		viewData := Data{
			Content: user,
			Title:   "Export bookmarks",
		}

		hc.toolsExportView.render(w, r, viewData)
	}
}

// handleToolsExport processes the bookmarks export form and sends the
// corresponding file to the client.
func (hc *toolsHandlerContext) handleToolsExport() func(w http.ResponseWriter, r *http.Request) {
	type exportForm struct {
		Visibility exporting.Visibility `form:"visibility"`
	}

	var form exportForm

	return func(w http.ResponseWriter, r *http.Request) {
		if err := render.DecodeForm(r.Body, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse bookmark export form")
			PutFlashError(w, "failed to process form")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		user := userValue(r.Context())

		document, err := hc.exportingService.ExportAsNetscapeDocument(user.UUID, form.Visibility)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve bookmarks")
			PutFlashError(w, "failed to export bookmarks")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		marshaled, err := netscape.Marshal(document)
		if err != nil {
			log.Error().Err(err).Msg("failed to marshal Netscape bookmarks")
			PutFlashError(w, "failed to export bookmarks")
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
		user := userValue(r.Context())
		viewData := Data{
			Content: user,
			Title:   "Import bookmarks",
		}

		hc.toolsImportView.render(w, r, viewData)
	}
}

// handleToolsImport processes data submitted through the bookmark import form.
func (hc *toolsHandlerContext) handleToolsImport() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		multipartReader, err := r.MultipartReader()
		if err != nil {
			log.Error().Err(err).Msg("failed to access multipart reader")
			PutFlashError(w, "failed to process import form")
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
				PutFlashError(w, "failed to process import form")
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
				PutFlashError(w, "failed to process import form")
				http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
				return
			}
		}

		document, err := netscape.Unmarshal(importFileBuffer.Bytes())
		if err != nil {
			log.Error().Err(err).Msg("failed to process Netscape bookmark file")
			PutFlashError(w, "failed to import bookmarks from the uploaded file")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		user := userValue(r.Context())
		overwrite := importing.OnConflictStrategy(onConflictStrategyBuffer.String())
		visibility := importing.Visibility(visibilityBuffer.String())

		importStatus, err := hc.importingService.ImportFromNetscapeDocument(user.UUID, document, visibility, overwrite)
		if err != nil {
			log.Error().Err(err).Msg("failed to save imported bookmarks")
			PutFlashError(w, "failed to save imported bookmarks")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		PutFlashSuccess(w, fmt.Sprintf("Import status: %s", importStatus.Summary()))
		http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
	}
}
