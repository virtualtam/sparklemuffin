// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package controller

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	"github.com/virtualtam/netscape-go/v2"
	"github.com/virtualtam/sparklemuffin/internal/http/www/csrf"
	"github.com/virtualtam/sparklemuffin/internal/http/www/httpcontext"
	"github.com/virtualtam/sparklemuffin/internal/http/www/middleware"
	"github.com/virtualtam/sparklemuffin/internal/http/www/view"
	"github.com/virtualtam/sparklemuffin/internal/paginate"
	"github.com/virtualtam/sparklemuffin/pkg/bookmark"
	bookmarkexporting "github.com/virtualtam/sparklemuffin/pkg/bookmark/exporting"
	bookmarkimporting "github.com/virtualtam/sparklemuffin/pkg/bookmark/importing"
	bookmarkquerying "github.com/virtualtam/sparklemuffin/pkg/bookmark/querying"
	"github.com/virtualtam/sparklemuffin/pkg/user"
)

const (
	actionBookmarkAdd    string = "bookmark-add"
	actionBookmarkDelete string = "bookmark-delete"
	actionBookmarkEdit   string = "bookmark-edit"
	actionBookmarkExport string = "bookmark-export"
	actionBookmarkImport string = "bookmark-import"
)

// RegisterBookmarkHandlers registers handlers to manage and display bookmarks.
func RegisterBookmarkHandlers(
	r *chi.Mux,
	publicURL *url.URL,
	bookmarkService *bookmark.Service,
	csrfService *csrf.Service,
	exportingService *bookmarkexporting.Service,
	importingService *bookmarkimporting.Service,
	queryingService *bookmarkquerying.Service,
	userService *user.Service,
) {
	bc := bookmarkController{
		publicURL: publicURL,

		bookmarkService:  bookmarkService,
		csrfService:      csrfService,
		exportingService: exportingService,
		importingService: importingService,
		queryingService:  queryingService,
		userService:      userService,

		bookmarkAddView:    view.New("bookmark/bookmark_add.gohtml"),
		bookmarkDeleteView: view.New("bookmark/bookmark_delete.gohtml"),
		bookmarkEditView:   view.New("bookmark/bookmark_edit.gohtml"),
		bookmarkListView:   view.New("bookmark/bookmark_list.gohtml"),

		bookmarkExportView: view.New("bookmark/bookmark_export.gohtml"),
		bookmarkImportView: view.New("bookmark/bookmark_import.gohtml"),

		publicBookmarkListView: view.New("public/bookmark_list.gohtml"),

		tagDeleteView: view.New("bookmark/tag_delete.gohtml"),
		tagEditView:   view.New("bookmark/tag_edit.gohtml"),
		tagListView:   view.New("bookmark/tag_list.gohtml"),
	}

	// bookmarks
	r.Route("/bookmarks", func(r chi.Router) {
		r.Use(func(h http.Handler) http.Handler {
			return middleware.AuthenticatedUser(h.ServeHTTP)
		})

		r.Get("/", bc.handleBookmarkListView())
		r.Get("/add", bc.handleBookmarkAddView())
		r.Post("/add", bc.handleBookmarkAdd())
		r.Get("/{uid}/delete", bc.handleBookmarkDeleteView())
		r.Post("/{uid}/delete", bc.handleBookmarkDelete())
		r.Get("/{uid}/edit", bc.handleBookmarkEditView())
		r.Post("/{uid}/edit", bc.handleBookmarkEdit())

		r.Get("/export", bc.handleBookmarkExportView())
		r.Post("/export", bc.handleBookmarkExport())
		r.Get("/import", bc.handleBookmarkImportView())
		r.Post("/import", bc.handleBookmarkImport())

		r.Route("/tags", func(sr chi.Router) {
			sr.Get("/", bc.handleTagListView())
			sr.Get("/{name}/delete", bc.handleTagDeleteView())
			sr.Post("/{name}/delete", bc.handleTagDelete())
			sr.Get("/{name}/edit", bc.handleTagEditView())
			sr.Post("/{name}/edit", bc.handleTagEdit())
		})
	})

	// public bookmarks
	r.Route("/u/{nickname}", func(r chi.Router) {
		r.Get("/bookmarks", bc.handlePublicBookmarkListView())
		r.Get("/bookmarks/{uid}", bc.handlePublicBookmarkPermalinkView())
		r.Get("/feed/atom", bc.handlePublicBookmarkFeedAtom())
	})
}

type bookmarkController struct {
	publicURL *url.URL

	bookmarkService  *bookmark.Service
	csrfService      *csrf.Service
	exportingService *bookmarkexporting.Service
	importingService *bookmarkimporting.Service
	queryingService  *bookmarkquerying.Service
	userService      *user.Service

	bookmarkAddView    *view.View
	bookmarkDeleteView *view.View
	bookmarkEditView   *view.View
	bookmarkListView   *view.View

	bookmarkExportView *view.View
	bookmarkImportView *view.View

	publicBookmarkListView *view.View

	tagDeleteView *view.View
	tagEditView   *view.View
	tagListView   *view.View
}

type bookmarkFormContent struct {
	CSRFToken string
	Bookmark  *bookmark.Bookmark
	Tags      []string
}

// handleBookmarkAddView renders the bookmark addition form.
func (bc *bookmarkController) handleBookmarkAddView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		user := httpcontext.UserValue(r.Context())
		csrfToken := bc.csrfService.Generate(user.UUID, actionBookmarkAdd)

		tags, err := bc.queryingService.TagNamesByCount(user.UUID, bookmarkquerying.VisibilityAll)
		if err != nil {
			log.Error().Err(err).Str("user_uuid", user.UUID).Msg("failed to retrieve tags")
			view.PutFlashError(w, "failed to retrieve existing tags")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		viewData := view.Data{
			Content: bookmarkFormContent{
				CSRFToken: csrfToken,
				Tags:      tags,
			},
			Title: "Add bookmark",
		}
		bc.bookmarkAddView.Render(w, r, viewData)
	}
}

// handleBookmarkAdd processes the bookmark addition form.
func (bc *bookmarkController) handleBookmarkAdd() func(w http.ResponseWriter, r *http.Request) {
	type bookmarkAddForm struct {
		CSRFToken   string `schema:"csrf_token"`
		URL         string `schema:"url"`
		Title       string `schema:"title"`
		Description string `schema:"description"`
		Private     bool   `schema:"private"`
		Tags        string `schema:"tags"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		user := httpcontext.UserValue(r.Context())

		var form bookmarkAddForm
		if err := decodeForm(r, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse bookmark creation form")
			view.PutFlashError(w, "There was an error processing the form")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		if !bc.csrfService.Validate(form.CSRFToken, user.UUID, actionBookmarkAdd) {
			log.Warn().Msg("failed to validate CSRF token")
			view.PutFlashError(w, "There was an error processing the form")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		newBookmark := bookmark.Bookmark{
			UserUUID:    user.UUID,
			URL:         form.URL,
			Title:       form.Title,
			Description: form.Description,
			Private:     form.Private,
			Tags:        strings.Split(form.Tags, " "),
		}

		if err := bc.bookmarkService.Add(newBookmark); err != nil {
			log.Error().Err(err).Msg("failed to add bookmark")
			view.PutFlashError(w, "failed to add bookmark")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/bookmarks", http.StatusSeeOther)
	}
}

// handleBookmarkDeleteView renders the bookmark deletion form.
func (bc *bookmarkController) handleBookmarkDeleteView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		bookmarkUID := chi.URLParam(r, "uid")
		user := httpcontext.UserValue(r.Context())
		csrfToken := bc.csrfService.Generate(user.UUID, actionBookmarkDelete)

		bookmark, err := bc.bookmarkService.ByUID(user.UUID, bookmarkUID)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve bookmark")
			view.PutFlashError(w, "failed to retrieve bookmark")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		viewData := view.Data{
			Content: view.FormContent{
				CSRFToken: csrfToken,
				Content:   bookmark,
			},
			Title: fmt.Sprintf("Delete bookmark: %s", bookmark.Title),
		}

		bc.bookmarkDeleteView.Render(w, r, viewData)
	}
}

// handleBookmarkDelete processes the bookmark deletion form.
func (bc *bookmarkController) handleBookmarkDelete() func(w http.ResponseWriter, r *http.Request) {
	type bookmarkDeleteForm struct {
		CSRFToken string `schema:"csrf_token"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		bookmarkUID := chi.URLParam(r, "uid")
		user := httpcontext.UserValue(r.Context())

		var form bookmarkDeleteForm
		if err := decodeForm(r, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse bookmark deletion form")
			view.PutFlashError(w, "There was an error processing the form")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		if !bc.csrfService.Validate(form.CSRFToken, user.UUID, actionBookmarkDelete) {
			log.Warn().Msg("failed to validate CSRF token")
			view.PutFlashError(w, "There was an error processing the form")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		if err := bc.bookmarkService.Delete(user.UUID, bookmarkUID); err != nil {
			log.Error().Err(err).Msg("failed to delete bookmark")
			view.PutFlashError(w, "failed to delete bookmark")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/bookmarks", http.StatusSeeOther)
	}
}

// handleBookmarkEditView renders the bookmark edition form.
func (bc *bookmarkController) handleBookmarkEditView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		bookmarkUID := chi.URLParam(r, "uid")
		user := httpcontext.UserValue(r.Context())
		csrfToken := bc.csrfService.Generate(user.UUID, actionBookmarkEdit)

		tags, err := bc.queryingService.TagNamesByCount(user.UUID, bookmarkquerying.VisibilityAll)
		if err != nil {
			log.Error().Err(err).Str("user_uuid", user.UUID).Msg("failed to retrieve tags")
			view.PutFlashError(w, "failed to retrieve existing tags")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		bookmark, err := bc.bookmarkService.ByUID(user.UUID, bookmarkUID)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve bookmark")
			view.PutFlashError(w, "failed to retrieve bookmark")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		viewData := view.Data{
			Content: bookmarkFormContent{
				CSRFToken: csrfToken,
				Bookmark:  &bookmark,
				Tags:      tags,
			},
			Title: fmt.Sprintf("Edit bookmark: %s", bookmark.Title),
		}

		bc.bookmarkEditView.Render(w, r, viewData)
	}
}

// handleBookmarkEdit processes the bookmark edition form.
func (bc *bookmarkController) handleBookmarkEdit() func(w http.ResponseWriter, r *http.Request) {
	type bookmarkEditForm struct {
		CSRFToken   string `schema:"csrf_token"`
		URL         string `schema:"url"`
		Title       string `schema:"title"`
		Description string `schema:"description"`
		Private     bool   `schema:"private"`
		Tags        string `schema:"tags"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		bookmarkUID := chi.URLParam(r, "uid")
		user := httpcontext.UserValue(r.Context())

		var form bookmarkEditForm
		if err := decodeForm(r, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse bookmark edition form")
			view.PutFlashError(w, "failed to process form")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		if !bc.csrfService.Validate(form.CSRFToken, user.UUID, actionBookmarkEdit) {
			log.Warn().Msg("failed to validate CSRF token")
			view.PutFlashError(w, "There was an error processing the form")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		editedBookmark := bookmark.Bookmark{
			UserUUID:    user.UUID,
			UID:         bookmarkUID,
			URL:         form.URL,
			Title:       form.Title,
			Description: form.Description,
			Private:     form.Private,
			Tags:        strings.Split(form.Tags, " "),
		}

		if err := bc.bookmarkService.Update(editedBookmark); err != nil {
			log.Error().Err(err).Msg("failed to edit bookmark")
			view.PutFlashError(w, "failed to edit bookmark")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/bookmarks", http.StatusSeeOther)
	}
}

// handleBookmarkListView renders the bookmark list for the current authenticated user.
func (bc *bookmarkController) handleBookmarkListView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var viewData view.Data
		user := httpcontext.UserValue(r.Context())

		pageNumber, pageNumberStr, err := paginate.GetPageNumber(r.URL.Query())
		if err != nil {
			log.Warn().Err(err).Str("page_number", pageNumberStr).Msg("invalid page number")
			view.PutFlashError(w, fmt.Sprintf("invalid page number: %q", pageNumberStr))
			http.Redirect(w, r, "/bookmarks", http.StatusSeeOther)
			return
		}

		searchTermsParam := r.URL.Query().Get("search")
		if searchTermsParam != "" {
			bookmarksSearchPage, err := bc.queryingService.BookmarksBySearchQueryAndPage(
				user.UUID,
				bookmarkquerying.VisibilityAll,
				searchTermsParam,
				pageNumber,
			)
			if errors.Is(err, paginate.ErrPageNumberOutOfBounds) {
				msg := fmt.Sprintf("invalid page number: %d", pageNumber)
				log.Error().Err(err).Msg(msg)
				view.PutFlashError(w, msg)
				http.Redirect(w, r, "/bookmarks", http.StatusSeeOther)
				return
			} else if err != nil {
				log.Error().Err(err).Msg("failed to retrieve bookmarks")
				view.PutFlashError(w, "failed to retrieve bookmarks")
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}

			viewData.Title = fmt.Sprintf("Bookmark search: %s", searchTermsParam)
			viewData.Content = bookmarksSearchPage

		} else {
			bookmarksPage, err := bc.queryingService.BookmarksByPage(
				user.UUID,
				bookmarkquerying.VisibilityAll,
				pageNumber,
			)
			if errors.Is(err, paginate.ErrPageNumberOutOfBounds) {
				msg := fmt.Sprintf("invalid page number: %d", pageNumber)
				log.Error().Err(err).Msg(msg)
				view.PutFlashError(w, msg)
				http.Redirect(w, r, "/bookmarks", http.StatusSeeOther)
				return
			} else if err != nil {
				log.Error().Err(err).Msg("failed to retrieve bookmarks")
				view.PutFlashError(w, "failed to retrieve bookmarks")
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}

			viewData.Title = "Bookmarks"
			viewData.Content = bookmarksPage
		}

		bc.bookmarkListView.Render(w, r, viewData)
	}
}

// handleBookmarkExportView renders the bookmark export page.
func (bc *bookmarkController) handleBookmarkExportView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctxUser := httpcontext.UserValue(r.Context())
		csrfToken := bc.csrfService.Generate(ctxUser.UUID, actionBookmarkExport)

		viewData := view.Data{
			Content: csrf.Data{
				CSRFToken: csrfToken,
			},
			Title: "Export bookmarks",
		}

		bc.bookmarkExportView.Render(w, r, viewData)
	}
}

// handleBookmarkExport processes the bookmarks export form and sends the
// corresponding file to the user.
func (bc *bookmarkController) handleBookmarkExport() func(w http.ResponseWriter, r *http.Request) {
	type exportForm struct {
		CSRFToken  string                       `schema:"csrf_token"`
		Format     bookmarkexporting.Format     `schema:"format"`
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

		if !bc.csrfService.Validate(form.CSRFToken, ctxUser.UUID, actionBookmarkExport) {
			log.Warn().Msg("failed to validate CSRF token")
			view.PutFlashError(w, "There was an error processing the form")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		var marshaled []byte
		var fileExtension string

		switch form.Format {
		case bookmarkexporting.FormatJSON:
			fileExtension = "json"

			jsonDocument, err := bc.exportingService.ExportAsJSONDocument(ctxUser.UUID, form.Visibility)
			if err != nil {
				log.Error().Err(err).Msg("bookmark: failed to retrieve bookmarks")
				view.PutFlashError(w, "failed to export bookmarks")
				http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
				return
			}

			marshaled, err = json.MarshalIndent(jsonDocument, "", "  ")
			if err != nil {
				log.Error().Err(err).Msg("bookmark: failed to marshal JSON document")
				view.PutFlashError(w, "failed to export bookmarks")
				http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
				return
			}

		case bookmarkexporting.FormatNetscape:
			fileExtension = "htm"

			netscapeDocument, err := bc.exportingService.ExportAsNetscapeDocument(ctxUser.UUID, form.Visibility)
			if err != nil {
				log.Error().Err(err).Msg("bookmark: failed to retrieve bookmarks")
				view.PutFlashError(w, "failed to export bookmarks")
				http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
				return
			}

			marshaled, err = netscape.Marshal(netscapeDocument)
			if err != nil {
				log.Error().Err(err).Msg("bookmark: failed to marshal Netscape document")
				view.PutFlashError(w, "failed to export bookmarks")
				http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
				return
			}

		default:
			log.Error().Str("format", string(form.Format)).Msg("bookmark: invalid export format")
			view.PutFlashError(w, "failed to export bookmarks")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		filename := fmt.Sprintf("bookmarks-%s.%s", form.Visibility, fileExtension)

		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
		w.Header().Set("Content-Type", "application/octet-stream")

		if _, err := w.Write(marshaled); err != nil {
			log.Error().Err(err).Str("format", string(form.Format)).Msg("bookmark: failed to send marshaled export")
		}
	}
}

// handleBookmarkImportView renders the bookmark import page.
func (bc *bookmarkController) handleBookmarkImportView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctxUser := httpcontext.UserValue(r.Context())
		csrfToken := bc.csrfService.Generate(ctxUser.UUID, actionBookmarkImport)

		viewData := view.Data{
			Content: csrf.Data{
				CSRFToken: csrfToken,
			},
			Title: "Import bookmarks",
		}

		bc.bookmarkImportView.Render(w, r, viewData)
	}
}

// handleBookmarkImport processes data submitted through the bookmark import form.
func (bc *bookmarkController) handleBookmarkImport() func(w http.ResponseWriter, r *http.Request) {
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

		if !bc.csrfService.Validate(csrfTokenBuffer.String(), ctxUser.UUID, actionBookmarkImport) {
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

		importStatus, err := bc.importingService.ImportFromNetscapeDocument(user.UUID, document, visibility, overwrite)
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

// handlePublicBookmarkListView renders the public bookmark list for a registered user.
func (bc *bookmarkController) handlePublicBookmarkListView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var viewData view.Data

		nickName := chi.URLParam(r, "nickname")

		// Retrieve the owner UUID via user.Service to avoid duplicating the normalization/validation layer
		// in bookmarkquerying.Service.
		// In practice, this requires performing an extra database query.
		owner, err := bc.userService.ByNickName(nickName)
		if err != nil {
			log.Error().Err(err).Str("nickname", nickName).Msg("failed to retrieve user")
			view.PutFlashError(w, "unknown user")
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		pageNumber, pageNumberStr, err := paginate.GetPageNumber(r.URL.Query())
		if err != nil {
			log.Warn().Err(err).Str("page_number", pageNumberStr).Msg("invalid page number")
			view.PutFlashError(w, fmt.Sprintf("invalid page number: %q", pageNumberStr))
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		var bookmarkPage bookmarkquerying.BookmarkPage

		searchTermsParam := r.URL.Query().Get("search")
		if searchTermsParam != "" {
			bookmarksSearchPage, err := bc.queryingService.PublicBookmarksBySearchQueryAndPage(
				owner.UUID,
				searchTermsParam,
				pageNumber,
			)
			if errors.Is(err, paginate.ErrPageNumberOutOfBounds) {
				msg := fmt.Sprintf("invalid page number: %d", pageNumber)
				log.Error().Err(err).Msg(msg)
				view.PutFlashError(w, msg)
				http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
				return
			} else if err != nil {
				log.Error().Err(err).Msg("failed to retrieve bookmarks")
				view.PutFlashError(w, "failed to retrieve bookmarks")
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}

			bookmarkPage = bookmarksSearchPage
			viewData.Title = fmt.Sprintf("%s's bookmarks: %s", owner.DisplayName, searchTermsParam)

		} else {
			bookmarksPage, err := bc.queryingService.PublicBookmarksByPage(
				owner.UUID,
				pageNumber,
			)
			if errors.Is(err, paginate.ErrPageNumberOutOfBounds) {
				msg := fmt.Sprintf("invalid page number: %d", pageNumber)
				log.Error().Err(err).Msg(msg)
				view.PutFlashError(w, msg)
				http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
				return
			} else if err != nil {
				log.Error().Err(err).Msg("failed to retrieve bookmarks")
				view.PutFlashError(w, "failed to retrieve bookmarks")
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}

			bookmarkPage = bookmarksPage
			viewData.Title = fmt.Sprintf("%s's bookmarks", owner.DisplayName)
		}

		viewData.AtomFeedURL = fmt.Sprintf("/u/%s/feed/atom", bookmarkPage.Owner.NickName)
		viewData.Content = bookmarkPage

		bc.publicBookmarkListView.Render(w, r, viewData)
	}
}

// handlePublicBookmarkPermalinkView renders a given public bookmark for a registered user.
func (bc *bookmarkController) handlePublicBookmarkPermalinkView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var viewData view.Data

		nickName := chi.URLParam(r, "nickname")
		bookmarkUID := chi.URLParam(r, "uid")

		// Retrieve the owner UUID via user.Service to avoid duplicating the normalization/validation layer
		// in bookmarkquerying.Service.
		// In practice, this requires performing an extra database query.
		owner, err := bc.userService.ByNickName(nickName)
		if err != nil {
			log.Error().Err(err).Str("nickname", nickName).Msg("failed to retrieve user")
			view.PutFlashError(w, "unknown user")
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		bookmarkPage, err := bc.queryingService.PublicBookmarkByUID(owner.UUID, bookmarkUID)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve bookmarks")
			view.PutFlashError(w, "failed to retrieve bookmarks")
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		viewData.AtomFeedURL = fmt.Sprintf("/u/%s/feed/atom", bookmarkPage.Owner.NickName)
		viewData.Title = fmt.Sprintf("%s's bookmarks", owner.DisplayName)
		viewData.Content = bookmarkPage

		bc.publicBookmarkListView.Render(w, r, viewData)
	}
}

// handlePublicBookmarkFeedAtom renders the public Atom feed for a registered user.
func (bc *bookmarkController) handlePublicBookmarkFeedAtom() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		nickName := chi.URLParam(r, "nickname")

		// Retrieve the owner UUID via user.Service to avoid duplicating the normalization/validation layer
		// in bookmarkquerying.Service.
		// In practice, this requires performing an extra database query.
		owner, err := bc.userService.ByNickName(nickName)
		if err != nil {
			log.Error().Err(err).Str("nickname", nickName).Msg("failed to retrieve user")
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		bookmarksPage, err := bc.queryingService.PublicBookmarksByPage(
			owner.UUID,
			1,
		)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve bookmarks")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		feed, err := bookmarksToFeed(bc.publicURL, bookmarksPage.Owner, bookmarksPage.Bookmarks)
		if err != nil {
			log.Error().Err(err).Msg("failed to create feed")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.atom", bookmarksPage.Owner.NickName))
		w.Header().Add("Content-Type", "application/atom+xml")

		if err := feed.WriteAtom(w); err != nil {
			log.Error().Err(err).Msg("failed to marshal Atom feed")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

// handleTagDeleteView renders the tag deletion form.
func (bc *bookmarkController) handleTagDeleteView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		nameBase64 := chi.URLParam(r, "name")

		nameBytes, err := base64.URLEncoding.DecodeString(nameBase64)
		if err != nil {
			log.Error().Err(err).Msg("invalid tag")
			view.PutFlashError(w, "invalid tag")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		name := string(nameBytes)
		tag := bookmarkquerying.NewTag(name, 0)

		viewData := view.Data{
			Content: tag,
			Title:   fmt.Sprintf("Delete tag: %s", name),
		}

		bc.tagDeleteView.Render(w, r, viewData)
	}
}

// handleTagDelete processes the tag deletion form.
func (bc *bookmarkController) handleTagDelete() func(w http.ResponseWriter, r *http.Request) {
	type tagDeleteForm struct {
		Name string `schema:"name"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var form tagDeleteForm
		if err := decodeForm(r, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse tag deletion form")
			view.PutFlashError(w, "failed to process form")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		nameBase64 := chi.URLParam(r, "name")

		nameBytes, err := base64.URLEncoding.DecodeString(nameBase64)
		if err != nil {
			log.Error().Err(err).Msg("invalid tag")
			view.PutFlashError(w, "invalid tag")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		name := string(nameBytes)

		user := httpcontext.UserValue(r.Context())

		tagDelete := bookmark.TagDeleteQuery{
			UserUUID: user.UUID,
			Name:     name,
		}

		updated, err := bc.bookmarkService.DeleteTag(tagDelete)
		if err != nil {
			log.Error().Err(err).Msg("failed to delete tag")
			view.PutFlashError(w, "failed to delete tag")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		view.PutFlashSuccess(w, fmt.Sprintf("Tag deleted from %d bookmarks", updated))
		http.Redirect(w, r, "/bookmarks/tags", http.StatusSeeOther)
	}
}

// handleTagEditView renders the tag edition form.
func (bc *bookmarkController) handleTagEditView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		nameBase64 := chi.URLParam(r, "name")

		nameBytes, err := base64.URLEncoding.DecodeString(nameBase64)
		if err != nil {
			log.Error().Err(err).Msg("invalid tag")
			view.PutFlashError(w, "invalid tag")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		name := string(nameBytes)
		tag := bookmarkquerying.NewTag(name, 0)

		viewData := view.Data{
			Content: tag,
			Title:   fmt.Sprintf("Edit tag: %s", name),
		}

		bc.tagEditView.Render(w, r, viewData)
	}
}

// handleTagEdit processes the tag edition form.
func (bc *bookmarkController) handleTagEdit() func(w http.ResponseWriter, r *http.Request) {
	type tagEditForm struct {
		Name string `schema:"name"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var form tagEditForm
		if err := decodeForm(r, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse tag edition form")
			view.PutFlashError(w, "failed to process form")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		nameBase64 := chi.URLParam(r, "name")

		nameBytes, err := base64.URLEncoding.DecodeString(nameBase64)
		if err != nil {
			log.Error().Err(err).Msg("invalid tag")
			view.PutFlashError(w, "invalid tag")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		name := string(nameBytes)

		user := httpcontext.UserValue(r.Context())

		tagNameUpdate := bookmark.TagUpdateQuery{
			UserUUID:    user.UUID,
			CurrentName: name,
			NewName:     form.Name,
		}

		updated, err := bc.bookmarkService.UpdateTag(tagNameUpdate)
		if err != nil {
			log.Error().Err(err).Msg("failed to rename tag")
			view.PutFlashError(w, "failed to rename tag")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		view.PutFlashSuccess(w, fmt.Sprintf("Tag updated for %d bookmarks", updated))
		http.Redirect(w, r, "/bookmarks/tags", http.StatusSeeOther)
	}
}

// handleTagListView renders the tag list view for the current authenticated user.
func (bc *bookmarkController) handleTagListView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var viewData view.Data
		user := httpcontext.UserValue(r.Context())

		pageNumber, pageNumberStr, err := paginate.GetPageNumber(r.URL.Query())
		if err != nil {
			log.Warn().Err(err).Str("page_number", pageNumberStr).Msg("invalid page number")
			view.PutFlashError(w, fmt.Sprintf("invalid page number: %q", pageNumberStr))
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		filterTermParam := r.URL.Query().Get("filter")

		if filterTermParam != "" {
			tagSearchPage, err := bc.queryingService.TagsByFilterQueryAndPage(
				user.UUID,
				bookmarkquerying.VisibilityAll,
				filterTermParam,
				pageNumber,
			)

			if errors.Is(err, paginate.ErrPageNumberOutOfBounds) {
				msg := fmt.Sprintf("invalid page number: %d", pageNumber)
				log.Error().Err(err).Msg(msg)
				view.PutFlashError(w, msg)
				http.Redirect(w, r, "/bookmarks/tags", http.StatusSeeOther)
				return
			} else if err != nil {
				log.Error().Err(err).Msg("failed to retrieve tags")
				view.PutFlashError(w, "failed to retrieve tags")
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}

			viewData.Title = fmt.Sprintf("Tag search: %s", filterTermParam)
			viewData.Content = tagSearchPage

		} else {
			tagPage, err := bc.queryingService.TagsByPage(
				user.UUID,
				bookmarkquerying.VisibilityAll,
				pageNumber,
			)

			if errors.Is(err, paginate.ErrPageNumberOutOfBounds) {
				msg := fmt.Sprintf("invalid page number: %d", pageNumber)
				log.Error().Err(err).Msg(msg)
				view.PutFlashError(w, msg)
				http.Redirect(w, r, "/bookmarks/tags", http.StatusSeeOther)
				return
			} else if err != nil {
				log.Error().Err(err).Msg("failed to retrieve tags")
				view.PutFlashError(w, "failed to retrieve tags")
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}

			viewData.Title = "Tags"
			viewData.Content = tagPage
		}

		bc.tagListView.Render(w, r, viewData)
	}
}
