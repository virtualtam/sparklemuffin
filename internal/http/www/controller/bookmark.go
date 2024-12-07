// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package controller

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	"github.com/virtualtam/sparklemuffin/internal/http/www/csrf"
	"github.com/virtualtam/sparklemuffin/internal/http/www/httpcontext"
	"github.com/virtualtam/sparklemuffin/internal/http/www/middleware"
	"github.com/virtualtam/sparklemuffin/internal/http/www/view"
	"github.com/virtualtam/sparklemuffin/internal/paginate"
	"github.com/virtualtam/sparklemuffin/pkg/bookmark"
	bookmarkquerying "github.com/virtualtam/sparklemuffin/pkg/bookmark/querying"
	"github.com/virtualtam/sparklemuffin/pkg/user"
)

const (
	actionBookmarkAdd    string = "bookmark-add"
	actionBookmarkEdit   string = "bookmark-edit"
	actionBookmarkDelete string = "bookmark-delete"
)

type bookmarkHandlerContext struct {
	publicURL *url.URL

	bookmarkService *bookmark.Service
	csrfService     *csrf.Service
	queryingService *bookmarkquerying.Service
	userService     *user.Service

	bookmarkAddView    *view.View
	bookmarkDeleteView *view.View
	bookmarkEditView   *view.View
	bookmarkListView   *view.View

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

func RegisterBookmarkHandlers(
	r *chi.Mux,
	publicURL *url.URL,
	bookmarkService *bookmark.Service,
	csrfService *csrf.Service,
	queryingService *bookmarkquerying.Service,
	userService *user.Service,
) {
	hc := bookmarkHandlerContext{
		publicURL: publicURL,

		bookmarkService: bookmarkService,
		csrfService:     csrfService,
		queryingService: queryingService,
		userService:     userService,

		bookmarkAddView:    view.New("bookmark/bookmark_add.gohtml"),
		bookmarkDeleteView: view.New("bookmark/bookmark_delete.gohtml"),
		bookmarkEditView:   view.New("bookmark/bookmark_edit.gohtml"),
		bookmarkListView:   view.New("bookmark/bookmark_list.gohtml"),

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

		r.Get("/", hc.handleBookmarkListView())
		r.Get("/add", hc.handleBookmarkAddView())
		r.Post("/add", hc.handleBookmarkAdd())
		r.Get("/{uid}/delete", hc.handleBookmarkDeleteView())
		r.Post("/{uid}/delete", hc.handleBookmarkDelete())
		r.Get("/{uid}/edit", hc.handleBookmarkEditView())
		r.Post("/{uid}/edit", hc.handleBookmarkEdit())

		r.Route("/tags", func(sr chi.Router) {
			sr.Get("/", hc.handleTagListView())
			sr.Get("/{name}/delete", hc.handleTagDeleteView())
			sr.Post("/{name}/delete", hc.handleTagDelete())
			sr.Get("/{name}/edit", hc.handleTagEditView())
			sr.Post("/{name}/edit", hc.handleTagEdit())
		})
	})

	// public bookmarks
	r.Route("/u/{nickname}", func(r chi.Router) {
		r.Get("/bookmarks", hc.handlePublicBookmarkListView())
		r.Get("/bookmarks/{uid}", hc.handlePublicBookmarkPermalinkView())
		r.Get("/feed/atom", hc.handlePublicBookmarkFeedAtom())
	})
}

// handleBookmarkAddView renders the bookmark addition form.
func (hc *bookmarkHandlerContext) handleBookmarkAddView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		user := httpcontext.UserValue(r.Context())
		csrfToken := hc.csrfService.Generate(user.UUID, actionBookmarkAdd)

		tags, err := hc.queryingService.TagNamesByCount(user.UUID, bookmarkquerying.VisibilityAll)
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
		hc.bookmarkAddView.Render(w, r, viewData)
	}
}

// handleBookmarkAdd processes the bookmark addition form.
func (hc *bookmarkHandlerContext) handleBookmarkAdd() func(w http.ResponseWriter, r *http.Request) {
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

		if !hc.csrfService.Validate(form.CSRFToken, user.UUID, actionBookmarkAdd) {
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

		if err := hc.bookmarkService.Add(newBookmark); err != nil {
			log.Error().Err(err).Msg("failed to add bookmark")
			view.PutFlashError(w, "failed to add bookmark")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/bookmarks", http.StatusSeeOther)
	}
}

// handleBookmarkDeleteView renders the bookmark deletion form.
func (hc *bookmarkHandlerContext) handleBookmarkDeleteView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		bookmarkUID := chi.URLParam(r, "uid")
		user := httpcontext.UserValue(r.Context())
		csrfToken := hc.csrfService.Generate(user.UUID, actionBookmarkDelete)

		bookmark, err := hc.bookmarkService.ByUID(user.UUID, bookmarkUID)
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

		hc.bookmarkDeleteView.Render(w, r, viewData)
	}
}

// handleBookmarkDelete processes the bookmark deletion form.
func (hc *bookmarkHandlerContext) handleBookmarkDelete() func(w http.ResponseWriter, r *http.Request) {
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

		if !hc.csrfService.Validate(form.CSRFToken, user.UUID, actionBookmarkDelete) {
			log.Warn().Msg("failed to validate CSRF token")
			view.PutFlashError(w, "There was an error processing the form")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		if err := hc.bookmarkService.Delete(user.UUID, bookmarkUID); err != nil {
			log.Error().Err(err).Msg("failed to delete bookmark")
			view.PutFlashError(w, "failed to delete bookmark")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/bookmarks", http.StatusSeeOther)
	}
}

// handleBookmarkEditView renders the bookmark edition form.
func (hc *bookmarkHandlerContext) handleBookmarkEditView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		bookmarkUID := chi.URLParam(r, "uid")
		user := httpcontext.UserValue(r.Context())
		csrfToken := hc.csrfService.Generate(user.UUID, actionBookmarkEdit)

		tags, err := hc.queryingService.TagNamesByCount(user.UUID, bookmarkquerying.VisibilityAll)
		if err != nil {
			log.Error().Err(err).Str("user_uuid", user.UUID).Msg("failed to retrieve tags")
			view.PutFlashError(w, "failed to retrieve existing tags")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		bookmark, err := hc.bookmarkService.ByUID(user.UUID, bookmarkUID)
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

		hc.bookmarkEditView.Render(w, r, viewData)
	}
}

// handleBookmarkEdit processes the bookmark edition form.
func (hc *bookmarkHandlerContext) handleBookmarkEdit() func(w http.ResponseWriter, r *http.Request) {
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

		if !hc.csrfService.Validate(form.CSRFToken, user.UUID, actionBookmarkEdit) {
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

		if err := hc.bookmarkService.Update(editedBookmark); err != nil {
			log.Error().Err(err).Msg("failed to edit bookmark")
			view.PutFlashError(w, "failed to edit bookmark")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/bookmarks", http.StatusSeeOther)
	}
}

// handleBookmarkListView renders the bookmark list for the current authenticated user.
func (hc *bookmarkHandlerContext) handleBookmarkListView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var viewData view.Data
		user := httpcontext.UserValue(r.Context())

		pageNumber, pageNumberStr, err := paginate.GetPageNumber(r.URL.Query())
		if err != nil {
			log.Error().Err(err).Str("page_number", pageNumberStr).Msg("invalid page number")
			view.PutFlashError(w, fmt.Sprintf("invalid page number: %q", pageNumberStr))
			http.Redirect(w, r, "/bookmarks", http.StatusSeeOther)
			return
		}

		searchTermsParam := r.URL.Query().Get("search")
		if searchTermsParam != "" {
			bookmarksSearchPage, err := hc.queryingService.BookmarksBySearchQueryAndPage(
				user.UUID,
				bookmarkquerying.VisibilityAll,
				searchTermsParam,
				pageNumber,
			)
			if errors.Is(err, bookmarkquerying.ErrPageNumberOutOfBounds) {
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
			bookmarksPage, err := hc.queryingService.BookmarksByPage(
				user.UUID,
				bookmarkquerying.VisibilityAll,
				pageNumber,
			)
			if errors.Is(err, bookmarkquerying.ErrPageNumberOutOfBounds) {
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

		hc.bookmarkListView.Render(w, r, viewData)
	}
}

// handlePublicBookmarkListView renders the public bookmark list for a registered user.
func (hc *bookmarkHandlerContext) handlePublicBookmarkListView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var viewData view.Data

		nickName := chi.URLParam(r, "nickname")

		// Retrieve the owner UUID via user.Service to avoid duplicating the normalization/validation layer
		// in bookmarkquerying.Service.
		// In practice, this requires performing an extra database query.
		owner, err := hc.userService.ByNickName(nickName)
		if err != nil {
			log.Error().Err(err).Str("nickname", nickName).Msg("failed to retrieve user")
			view.PutFlashError(w, "unknown user")
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		pageNumber, pageNumberStr, err := paginate.GetPageNumber(r.URL.Query())
		if err != nil {
			log.Error().Err(err).Str("page_number", pageNumberStr).Msg("invalid page number")
			view.PutFlashError(w, fmt.Sprintf("invalid page number: %q", pageNumberStr))
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		var bookmarkPage bookmarkquerying.BookmarkPage

		searchTermsParam := r.URL.Query().Get("search")
		if searchTermsParam != "" {
			bookmarksSearchPage, err := hc.queryingService.PublicBookmarksBySearchQueryAndPage(
				owner.UUID,
				searchTermsParam,
				pageNumber,
			)
			if errors.Is(err, bookmarkquerying.ErrPageNumberOutOfBounds) {
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
			bookmarksPage, err := hc.queryingService.PublicBookmarksByPage(
				owner.UUID,
				pageNumber,
			)
			if errors.Is(err, bookmarkquerying.ErrPageNumberOutOfBounds) {
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

		hc.publicBookmarkListView.Render(w, r, viewData)
	}
}

// handlePublicBookmarkPermalinkView renders a given public bookmark for a registered user.
func (hc *bookmarkHandlerContext) handlePublicBookmarkPermalinkView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var viewData view.Data

		nickName := chi.URLParam(r, "nickname")
		bookmarkUID := chi.URLParam(r, "uid")

		// Retrieve the owner UUID via user.Service to avoid duplicating the normalization/validation layer
		// in bookmarkquerying.Service.
		// In practice, this requires performing an extra database query.
		owner, err := hc.userService.ByNickName(nickName)
		if err != nil {
			log.Error().Err(err).Str("nickname", nickName).Msg("failed to retrieve user")
			view.PutFlashError(w, "unknown user")
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		bookmarkPage, err := hc.queryingService.PublicBookmarkByUID(owner.UUID, bookmarkUID)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve bookmarks")
			view.PutFlashError(w, "failed to retrieve bookmarks")
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		viewData.AtomFeedURL = fmt.Sprintf("/u/%s/feed/atom", bookmarkPage.Owner.NickName)
		viewData.Title = fmt.Sprintf("%s's bookmarks", owner.DisplayName)
		viewData.Content = bookmarkPage

		hc.publicBookmarkListView.Render(w, r, viewData)
	}
}

// handlePublicBookmarkFeedAtom renders the public Atom feed for a registered user.
func (hc *bookmarkHandlerContext) handlePublicBookmarkFeedAtom() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		nickName := chi.URLParam(r, "nickname")

		// Retrieve the owner UUID via user.Service to avoid duplicating the normalization/validation layer
		// in bookmarkquerying.Service.
		// In practice, this requires performing an extra database query.
		owner, err := hc.userService.ByNickName(nickName)
		if err != nil {
			log.Error().Err(err).Str("nickname", nickName).Msg("failed to retrieve user")
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		bookmarksPage, err := hc.queryingService.PublicBookmarksByPage(
			owner.UUID,
			1,
		)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve bookmarks")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		feed, err := bookmarksToFeed(hc.publicURL, bookmarksPage.Owner, bookmarksPage.Bookmarks)
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
func (hc *bookmarkHandlerContext) handleTagDeleteView() func(w http.ResponseWriter, r *http.Request) {
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

		hc.tagDeleteView.Render(w, r, viewData)
	}
}

// handleTagDelete processes the tag deletion form.
func (hc *bookmarkHandlerContext) handleTagDelete() func(w http.ResponseWriter, r *http.Request) {
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

		updated, err := hc.bookmarkService.DeleteTag(tagDelete)
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
func (hc *bookmarkHandlerContext) handleTagEditView() func(w http.ResponseWriter, r *http.Request) {
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

		hc.tagEditView.Render(w, r, viewData)
	}
}

// handleTagEdit processes the tag edition form.
func (hc *bookmarkHandlerContext) handleTagEdit() func(w http.ResponseWriter, r *http.Request) {
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

		updated, err := hc.bookmarkService.UpdateTag(tagNameUpdate)
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
func (hc *bookmarkHandlerContext) handleTagListView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var viewData view.Data
		user := httpcontext.UserValue(r.Context())

		pageNumber, pageNumberStr, err := paginate.GetPageNumber(r.URL.Query())
		if err != nil {
			log.Error().Err(err).Str("page_number", pageNumberStr).Msg("invalid page number")
			view.PutFlashError(w, fmt.Sprintf("invalid page number: %q", pageNumberStr))
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		filterTermParam := r.URL.Query().Get("filter")

		if filterTermParam != "" {
			tagSearchPage, err := hc.queryingService.TagsByFilterQueryAndPage(
				user.UUID,
				bookmarkquerying.VisibilityAll,
				filterTermParam,
				pageNumber,
			)

			if errors.Is(err, bookmarkquerying.ErrPageNumberOutOfBounds) {
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
			tagPage, err := hc.queryingService.TagsByPage(
				user.UUID,
				bookmarkquerying.VisibilityAll,
				pageNumber,
			)

			if errors.Is(err, bookmarkquerying.ErrPageNumberOutOfBounds) {
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

		hc.tagListView.Render(w, r, viewData)
	}
}
