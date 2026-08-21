// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package controller

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	"github.com/virtualtam/sparklemuffin/internal/http/www/htmx"
	"github.com/virtualtam/sparklemuffin/internal/http/www/httpcontext"
	"github.com/virtualtam/sparklemuffin/internal/http/www/middleware"
	"github.com/virtualtam/sparklemuffin/internal/http/www/view"
	"github.com/virtualtam/sparklemuffin/pkg/bookmark"
	bookmarkquerying "github.com/virtualtam/sparklemuffin/pkg/bookmark/querying"
	feedquerying "github.com/virtualtam/sparklemuffin/pkg/feed/querying"
)

// RegisterFeedBookmarkHandlers registers HTTP handlers bridging the feed and
// bookmark domains, letting a user bookmark a feed entry.
//
// This bridge calls into feedquerying.Service and bookmark.Service directly;
// it never calls the bookmark controller's own HTTP handlers, so pkg/feed
// and pkg/bookmark stay decoupled from each other.
func RegisterFeedBookmarkHandlers(
	r *chi.Mux,
	feedQueryingService *feedquerying.Service,
	bookmarkService *bookmark.Service,
	bookmarkQueryingService *bookmarkquerying.Service,
) {
	fbc := feedBookmarkController{
		feedQueryingService:     feedQueryingService,
		bookmarkService:         bookmarkService,
		bookmarkQueryingService: bookmarkQueryingService,

		entryBookmarkView: view.New("feed/entry_bookmark.gohtml"),
	}

	r.Get("/feeds/entries/{uid}/bookmark", middleware.AuthenticatedUser(fbc.handleFeedEntryBookmarkView()))
	r.Post("/feeds/entries/{uid}/bookmark", middleware.AuthenticatedUser(fbc.handleFeedEntryBookmark()))
}

type feedBookmarkController struct {
	feedQueryingService     *feedquerying.Service
	bookmarkService         *bookmark.Service
	bookmarkQueryingService *bookmarkquerying.Service

	entryBookmarkView *view.View
}

type entryBookmarkFormContent struct {
	EntryUID string
	Bookmark *bookmark.Bookmark
	Tags     []string
	Existing bool
}

// handleFeedEntryBookmarkView renders a bookmark creation form pre-filled
// from a feed entry's URL, title and subscription tags.
//
// On an htmx request, it responds with only the form fragment, meant to be
// loaded into the feed entry list's bookmark modal. On a plain request, it
// renders the full page as usual, so the URL stays independently navigable.
func (fbc *feedBookmarkController) handleFeedEntryBookmarkView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		entryUID := chi.URLParam(r, "uid")
		ctx := r.Context()
		ctxUser := httpcontext.UserValue(ctx)

		entry, err := fbc.feedQueryingService.SubscribedFeedEntryByUID(ctx, ctxUser.UUID, entryUID)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve feed entry")
			view.RedirectOnError(w, r, r.URL.Path, "failed to retrieve feed entry")
			return
		}

		tags, err := fbc.bookmarkQueryingService.TagNamesByCount(ctx, ctxUser.UUID, bookmarkquerying.VisibilityAll)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve tags")
			view.RedirectOnError(w, r, r.URL.Path, "failed to retrieve existing tags")
			return
		}

		draft := &bookmark.Bookmark{
			URL:   entry.URL,
			Title: entry.Title,
			Tags:  entry.SubscriptionTags,
		}
		existing := false

		if existingBookmark, err := fbc.bookmarkService.ByURL(ctx, ctxUser.UUID, entry.URL); err == nil {
			draft = &existingBookmark
			existing = true
		} else if !errors.Is(err, bookmark.ErrNotFound) {
			log.Error().Err(err).Msg("failed to check for an existing bookmark")
			view.RedirectOnError(w, r, r.URL.Path, "failed to check for an existing bookmark")
			return
		}

		if r.Header.Get(htmx.HeaderRequest) == "true" {
			formData := map[string]any{"EntryUID": entryUID, "Bookmark": draft, "Tags": tags, "InModal": true, "Existing": existing}
			if err := fbc.entryBookmarkView.RenderTemplate(w, "entryBookmarkForm", formData); err != nil {
				log.Error().Err(err).Msg("failed to render entry bookmark form fragment")
				http.Error(w, "Something went wrong", http.StatusInternalServerError)
			}
			return
		}

		viewData := view.Data{
			Content: entryBookmarkFormContent{EntryUID: entryUID, Bookmark: draft, Tags: tags, Existing: existing},
			Title:   fmt.Sprintf("Bookmark: %s", entry.Title),
		}
		fbc.entryBookmarkView.Render(w, r, viewData)
	}
}

// handleFeedEntryBookmark processes the pre-filled bookmark form submitted
// from a feed entry.
func (fbc *feedBookmarkController) handleFeedEntryBookmark() func(w http.ResponseWriter, r *http.Request) {
	type entryBookmarkForm struct {
		URL         string `schema:"url"`
		Title       string `schema:"title"`
		Description string `schema:"description"`
		Private     bool   `schema:"private"`
		Tags        string `schema:"tags"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctxUser := httpcontext.UserValue(ctx)

		var form entryBookmarkForm
		if err := decodeForm(r, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse feed entry bookmark form")
			view.RedirectOnError(w, r, r.URL.Path, "failed to process form")
			return
		}

		submittedBookmark := bookmark.Bookmark{
			UserUUID:    ctxUser.UUID,
			URL:         form.URL,
			Title:       form.Title,
			Description: form.Description,
			Private:     form.Private,
			Tags:        strings.Split(form.Tags, " "),
		}

		if existing, err := fbc.bookmarkService.ByURL(ctx, ctxUser.UUID, form.URL); err == nil {
			submittedBookmark.UID = existing.UID
			if err := fbc.bookmarkService.Update(ctx, submittedBookmark); err != nil {
				log.Error().Err(err).Msg("failed to update bookmark")
				view.RedirectOnError(w, r, r.URL.Path, "failed to update bookmark")
				return
			}
		} else if errors.Is(err, bookmark.ErrNotFound) {
			if err := fbc.bookmarkService.Add(ctx, submittedBookmark); err != nil {
				log.Error().Err(err).Msg("failed to add bookmark")
				view.RedirectOnError(w, r, r.URL.Path, "failed to add bookmark")
				return
			}
		} else {
			log.Error().Err(err).Msg("failed to check for an existing bookmark")
			view.RedirectOnError(w, r, r.URL.Path, "failed to check for an existing bookmark")
			return
		}

		if r.Header.Get(htmx.HeaderRequest) != "true" {
			http.Redirect(w, r, "/bookmarks", http.StatusSeeOther)
			return
		}

		saved, err := fbc.bookmarkService.ByURL(ctx, ctxUser.UUID, form.URL)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve saved bookmark")
			view.RedirectOnError(w, r, "/feeds", "failed to retrieve saved bookmark")
			return
		}

		if err := fbc.entryBookmarkView.RenderTemplate(w, "entryBookmarkConfirmation", map[string]any{"Bookmark": saved}); err != nil {
			log.Error().Err(err).Msg("failed to render entry bookmark confirmation fragment")
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
	}
}
