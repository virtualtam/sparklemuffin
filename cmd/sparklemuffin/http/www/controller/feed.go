// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package controller

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	"github.com/virtualtam/sparklemuffin/cmd/sparklemuffin/http/www/csrf"
	"github.com/virtualtam/sparklemuffin/cmd/sparklemuffin/http/www/httpcontext"
	"github.com/virtualtam/sparklemuffin/cmd/sparklemuffin/http/www/middleware"
	"github.com/virtualtam/sparklemuffin/cmd/sparklemuffin/http/www/view"
	"github.com/virtualtam/sparklemuffin/pkg/feed"
	fquerying "github.com/virtualtam/sparklemuffin/pkg/feed/querying"
	"github.com/virtualtam/sparklemuffin/pkg/user"
)

const (
	actionFeedAdd string = "feed-add"
)

// RegisterFeedHandlers registers HTTP handlers for syndication feed operations.
func RegisterFeedHandlers(
	r *chi.Mux,
	csrfService *csrf.Service,
	feedService *feed.Service,
	feedQueryingService *fquerying.Service,
	userService *user.Service,
) {
	fc := feedHandlerContext{
		csrfService:         csrfService,
		feedService:         feedService,
		feedQueryingService: feedQueryingService,
		userService:         userService,

		feedListView: view.New("feed/list.gohtml"),
		feedAddView:  view.New("feed/add.gohtml"),
	}

	// feeds
	r.Route("/feeds", func(r chi.Router) {
		r.Use(func(h http.Handler) http.Handler {
			return middleware.AuthenticatedUser(h.ServeHTTP)
		})

		r.Get("/", fc.handleFeedListView())
		r.Get("/add", fc.handleFeedAddView())
		r.Post("/add", fc.handleFeedAdd())
	})
}

type feedHandlerContext struct {
	csrfService         *csrf.Service
	feedService         *feed.Service
	feedQueryingService *fquerying.Service
	userService         *user.Service

	feedAddView  *view.View
	feedListView *view.View
}

type feedFormContent struct {
	CSRFToken  string
	Categories []feed.Category
}

// handleFeedAddView renders the feed subscription form.
func (fc *feedHandlerContext) handleFeedAddView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		user := httpcontext.UserValue(r.Context())
		csrfToken := fc.csrfService.Generate(user.UUID, actionFeedAdd)

		categories, err := fc.feedService.Categories(user.UUID)
		if err != nil {
			log.Error().Err(err).Str("user_uuid", user.UUID).Msg("failed to retrieve feed categories")
			view.PutFlashError(w, "failed to retrieve existing feed categories")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		viewData := view.Data{
			Content: feedFormContent{
				CSRFToken:  csrfToken,
				Categories: categories,
			},
			Title: "Add feed",
		}

		fc.feedAddView.Render(w, r, viewData)
	}
}

// handleFeedAdd processes the feed addition form.
func (fc *feedHandlerContext) handleFeedAdd() func(w http.ResponseWriter, r *http.Request) {
	type feedAddForm struct {
		CSRFToken    string `schema:"csrf_token"`
		URL          string `schema:"url"`
		CategoryUUID string `schema:"category"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctxUser := httpcontext.UserValue(r.Context())

		var form feedAddForm
		if err := decodeForm(r, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse feed subscription form")
			view.PutFlashError(w, "There was an error processing the form")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		if !fc.csrfService.Validate(form.CSRFToken, ctxUser.UUID, actionFeedAdd) {
			log.Warn().Msg("failed to validate CSRF token")
			view.PutFlashError(w, "There was an error processing the form")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		if err := fc.feedService.Subscribe(ctxUser.UUID, form.CategoryUUID, form.URL); err != nil {
			log.Error().Err(err).Msg("failed to subscribe to feed")
			view.PutFlashError(w, "failed to subscribe to feed")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/feeds", http.StatusSeeOther)
	}
}

// handleFeedListView renders the syndication feed for the current authenticated user.
func (fc *feedHandlerContext) handleFeedListView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		user := httpcontext.UserValue(r.Context())

		feedPage, err := fc.feedQueryingService.FeedsByPage(user.UUID)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve feeds")
			view.PutFlashError(w, "failed to retrieve feeds")
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		var viewData view.Data
		viewData.Title = "Feeds"
		viewData.Content = feedPage

		fc.feedListView.Render(w, r, viewData)
	}
}
