// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package controller

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/virtualtam/sparklemuffin/cmd/sparklemuffin/http/www/middleware"
	"github.com/virtualtam/sparklemuffin/cmd/sparklemuffin/http/www/view"
	"github.com/virtualtam/sparklemuffin/pkg/user"
)

// RegisterFeedHandlers registers HTTP handlers for syndication feed operations.
func RegisterFeedHandlers(
	r *chi.Mux,
	userService *user.Service,
) {
	fc := feedHandlerContext{
		userService: userService,

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
	userService *user.Service

	feedListView *view.View
}

// handleFeedListView renders the syndication feed for the current authenticated user.
func (fc *feedHandlerContext) handleFeedListView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var viewData view.Data

		viewData.Title = "Feeds"

		fc.feedListView.Render(w, r, viewData)
	}
}
