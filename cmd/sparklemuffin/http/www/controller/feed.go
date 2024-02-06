// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package controller

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	"github.com/virtualtam/sparklemuffin/cmd/sparklemuffin/http/www/httpcontext"
	"github.com/virtualtam/sparklemuffin/cmd/sparklemuffin/http/www/middleware"
	"github.com/virtualtam/sparklemuffin/cmd/sparklemuffin/http/www/view"
	"github.com/virtualtam/sparklemuffin/pkg/feed"
	fquerying "github.com/virtualtam/sparklemuffin/pkg/feed/querying"
	"github.com/virtualtam/sparklemuffin/pkg/user"
)

// RegisterFeedHandlers registers HTTP handlers for syndication feed operations.
func RegisterFeedHandlers(
	r *chi.Mux,
	feedService *feed.Service,
	feedQueryingService *fquerying.Service,
	userService *user.Service,
) {
	fc := feedHandlerContext{
		feedService:         feedService,
		feedQueryingService: feedQueryingService,
		userService:         userService,

		feedListView: view.New("feed/list.gohtml"),
	}

	// feeds
	r.Route("/feeds", func(r chi.Router) {
		r.Use(func(h http.Handler) http.Handler {
			return middleware.AuthenticatedUser(h.ServeHTTP)
		})

		r.Get("/", fc.handleFeedListView())
	})
}

type feedHandlerContext struct {
	feedService         *feed.Service
	feedQueryingService *fquerying.Service
	userService         *user.Service

	feedListView *view.View
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
