// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package controller

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	"github.com/virtualtam/sparklemuffin/cmd/sparklemuffin/http/www/csrf"
	"github.com/virtualtam/sparklemuffin/cmd/sparklemuffin/http/www/httpcontext"
	"github.com/virtualtam/sparklemuffin/cmd/sparklemuffin/http/www/middleware"
	"github.com/virtualtam/sparklemuffin/cmd/sparklemuffin/http/www/view"
	"github.com/virtualtam/sparklemuffin/internal/paginate"
	"github.com/virtualtam/sparklemuffin/pkg/feed"
	feedquerying "github.com/virtualtam/sparklemuffin/pkg/feed/querying"
	"github.com/virtualtam/sparklemuffin/pkg/user"
)

const (
	actionFeedAdd            string = "feed-add"
	actionFeedCategoryAdd    string = "feed-category-add"
	actionFeedCategoryDelete string = "feed-category-delete"
	actionFeedCategoryEdit   string = "feed-category-edit"
)

// RegisterFeedHandlers registers HTTP handlers for syndication feed operations.
func RegisterFeedHandlers(
	r *chi.Mux,
	csrfService *csrf.Service,
	feedService *feed.Service,
	feedQueryingService *feedquerying.Service,
	userService *user.Service,
) {
	fc := feedHandlerContext{
		csrfService:         csrfService,
		feedService:         feedService,
		feedQueryingService: feedQueryingService,
		userService:         userService,

		feedListView: view.New("feed/list.gohtml"),
		feedAddView:  view.New("feed/feed_add.gohtml"),

		feedCategoryAddView:      view.New("feed/category_add.gohtml"),
		feedCategoryDeleteView:   view.New("feed/category_delete.gohtml"),
		feedCategoryEditView:     view.New("feed/category_edit.gohtml"),
		feedSubscriptionListView: view.New("feed/subscription_list.gohtml"),
	}

	// feeds
	r.Route("/feeds", func(r chi.Router) {
		r.Use(func(h http.Handler) http.Handler {
			return middleware.AuthenticatedUser(h.ServeHTTP)
		})

		r.Get("/", fc.handleFeedListView())
		r.Get("/categories/{slug}/entries", fc.handleFeedListByCategoryView())
		r.Get("/{slug}/entries", fc.handleFeedListBySubscriptionView())

		r.Get("/add", fc.handleFeedAddView())
		r.Post("/add", fc.handleFeedAdd())

		r.Route("/categories", func(sr chi.Router) {
			sr.Get("/add", fc.handleFeedCategoryAddView())
			sr.Post("/add", fc.handleFeedCategoryAdd())
			sr.Get("/{uuid}/delete", fc.handleFeedCategoryDeleteView())
			sr.Post("/{uuid}/delete", fc.handleFeedCategoryDelete())
			sr.Get("/{uuid}/edit", fc.handleFeedCategoryEditView())
			sr.Post("/{uuid}/edit", fc.handleFeedCategoryEdit())
		})

		r.Get("/subscriptions", fc.handleFeedSubscriptionListView())
	})
}

type feedHandlerContext struct {
	csrfService         *csrf.Service
	feedService         *feed.Service
	feedQueryingService *feedquerying.Service
	userService         *user.Service

	feedAddView  *view.View
	feedListView *view.View

	feedCategoryAddView      *view.View
	feedCategoryDeleteView   *view.View
	feedCategoryEditView     *view.View
	feedSubscriptionListView *view.View
}

type feedFormContent struct {
	CSRFToken  string
	Categories []feed.Category
}

type feedCategoryFormContent struct {
	CSRFToken string
	Category  feed.Category
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

		pageNumber, pageNumberStr, err := paginate.GetPageNumber(r.URL.Query())
		if err != nil {
			log.Error().Err(err).Str("page_number", pageNumberStr).Msg("invalid page number")
			view.PutFlashError(w, fmt.Sprintf("invalid page number: %q", pageNumberStr))
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		feedPage, err := fc.feedQueryingService.FeedsByPage(user.UUID, pageNumber)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve feeds")
			view.PutFlashError(w, "failed to retrieve feeds")
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		viewData := view.Data{
			Title:   "Feeds",
			Content: feedPage,
		}

		fc.feedListView.Render(w, r, viewData)
	}
}

// handleFeedListByCategoryView renders the syndication feed for the current authenticated user.
func (fc *feedHandlerContext) handleFeedListByCategoryView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		user := httpcontext.UserValue(r.Context())
		categorySlug := chi.URLParam(r, "slug")

		pageNumber, pageNumberStr, err := paginate.GetPageNumber(r.URL.Query())
		if err != nil {
			log.Error().Err(err).Str("page_number", pageNumberStr).Msg("invalid page number")
			view.PutFlashError(w, fmt.Sprintf("invalid page number: %q", pageNumberStr))
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		category, err := fc.feedService.CategoryBySlug(user.UUID, categorySlug)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve feed category")
			view.PutFlashError(w, "failed to retrieve feed category")
			http.Redirect(w, r, "/feeds", http.StatusSeeOther)
			return
		}

		feedPage, err := fc.feedQueryingService.FeedsByCategoryAndPage(user.UUID, category, pageNumber)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve feeds")
			view.PutFlashError(w, "failed to retrieve feeds")
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		viewData := view.Data{
			Title:   fmt.Sprintf("Feeds: %s", category.Name),
			Content: feedPage,
		}

		fc.feedListView.Render(w, r, viewData)
	}
}

// handleFeedListBySubscriptionView renders the syndication feed for the current authenticated user.
func (fc *feedHandlerContext) handleFeedListBySubscriptionView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		user := httpcontext.UserValue(r.Context())
		feedSlug := chi.URLParam(r, "slug")

		pageNumber, pageNumberStr, err := paginate.GetPageNumber(r.URL.Query())
		if err != nil {
			log.Error().Err(err).Str("page_number", pageNumberStr).Msg("invalid page number")
			view.PutFlashError(w, fmt.Sprintf("invalid page number: %q", pageNumberStr))
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		feed, err := fc.feedService.FeedBySlug(user.UUID, feedSlug)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve feed")
			view.PutFlashError(w, "failed to retrieve feed")
			http.Redirect(w, r, "/feeds", http.StatusSeeOther)
			return
		}

		subscription, err := fc.feedService.SubscriptionByFeed(user.UUID, feed.UUID)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve feed subscription")
			view.PutFlashError(w, "failed to retrieve feed subscription")
			http.Redirect(w, r, "/feeds", http.StatusSeeOther)
			return
		}

		feedPage, err := fc.feedQueryingService.FeedsBySubscriptionAndPage(user.UUID, subscription, pageNumber)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve feeds")
			view.PutFlashError(w, "failed to retrieve feeds")
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		viewData := view.Data{
			Title:   fmt.Sprintf("Feeds: %s", feed.Title),
			Content: feedPage,
		}

		fc.feedListView.Render(w, r, viewData)
	}
}

// handleFeedCategoryAddView renders the feed category addition form.
func (fc *feedHandlerContext) handleFeedCategoryAddView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		user := httpcontext.UserValue(r.Context())
		csrfToken := fc.csrfService.Generate(user.UUID, actionFeedCategoryAdd)

		viewData := view.Data{
			Content: feedFormContent{
				CSRFToken: csrfToken,
			},
			Title: "Add feed category",
		}

		fc.feedCategoryAddView.Render(w, r, viewData)
	}
}

// handleFeedCategoryAdd processes the feed category addition form.
func (fc *feedHandlerContext) handleFeedCategoryAdd() func(w http.ResponseWriter, r *http.Request) {
	type feedAddForm struct {
		CSRFToken string `schema:"csrf_token"`
		Name      string `schema:"name"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctxUser := httpcontext.UserValue(r.Context())

		var form feedAddForm
		if err := decodeForm(r, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse feed category addition form")
			view.PutFlashError(w, "There was an error processing the form")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		if !fc.csrfService.Validate(form.CSRFToken, ctxUser.UUID, actionFeedCategoryAdd) {
			log.Warn().Msg("failed to validate CSRF token")
			view.PutFlashError(w, "There was an error processing the form")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		if _, err := fc.feedService.AddCategory(ctxUser.UUID, form.Name); err != nil {
			log.Error().Err(err).Msg("failed to add feed category")
			view.PutFlashError(w, "failed to add feed category")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/feeds", http.StatusSeeOther)
	}
}

// handleFeedCategoryDeleteView renders the feed category deletion form.
func (fc *feedHandlerContext) handleFeedCategoryDeleteView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		categoryUUID := chi.URLParam(r, "uuid")
		ctxUser := httpcontext.UserValue(r.Context())
		csrfToken := fc.csrfService.Generate(ctxUser.UUID, actionFeedCategoryDelete)

		category, err := fc.feedService.CategoryByUUID(ctxUser.UUID, categoryUUID)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve feed category")
			view.PutFlashError(w, "failed to retrieve feed category")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		viewData := view.Data{
			Content: feedCategoryFormContent{
				CSRFToken: csrfToken,
				Category:  category,
			},
			Title: fmt.Sprintf("Delete category: %s", category.Name),
		}

		fc.feedCategoryDeleteView.Render(w, r, viewData)
	}
}

func (fc *feedHandlerContext) handleFeedCategoryDelete() func(w http.ResponseWriter, r *http.Request) {
	type feedCategoryDeleteForm struct {
		CSRFToken string `schema:"csrf_token"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		categoryUUID := chi.URLParam(r, "uuid")
		ctxUser := httpcontext.UserValue(r.Context())

		var form feedCategoryDeleteForm
		if err := decodeForm(r, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse feed category deletion form")
			view.PutFlashError(w, "There was an error processing the form")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		if !fc.csrfService.Validate(form.CSRFToken, ctxUser.UUID, actionFeedCategoryDelete) {
			log.Warn().Msg("failed to validate CSRF token")
			view.PutFlashError(w, "There was an error processing the form")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		if err := fc.feedService.DeleteCategory(ctxUser.UUID, categoryUUID); err != nil {
			log.Error().Err(err).Msg("failed to delete feed category")
			view.PutFlashError(w, "failed to delete feed category")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/feeds", http.StatusSeeOther)
	}
}

// handleFeedCategoryEditView renders the feed category edition form.
func (fc *feedHandlerContext) handleFeedCategoryEditView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctxUser := httpcontext.UserValue(r.Context())
		csrfToken := fc.csrfService.Generate(ctxUser.UUID, actionFeedCategoryEdit)

		categoryUUID := chi.URLParam(r, "uuid")

		category, err := fc.feedService.CategoryByUUID(ctxUser.UUID, categoryUUID)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve feed category")
			view.PutFlashError(w, "failed to retrieve feed category")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		viewData := view.Data{
			Content: feedCategoryFormContent{
				CSRFToken: csrfToken,
				Category:  category,
			},
			Title: "Edit feed category",
		}

		fc.feedCategoryEditView.Render(w, r, viewData)
	}
}

// handleFeedCategoryEdit processes the feed category edition form.
func (fc *feedHandlerContext) handleFeedCategoryEdit() func(w http.ResponseWriter, r *http.Request) {
	type feedCategoryEditForm struct {
		CSRFToken string `schema:"csrf_token"`
		Name      string `schema:"name"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctxUser := httpcontext.UserValue(r.Context())
		categoryUUID := chi.URLParam(r, "uuid")

		var form feedCategoryEditForm
		if err := decodeForm(r, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse feed category edition form")
			view.PutFlashError(w, "There was an error processing the form")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		if !fc.csrfService.Validate(form.CSRFToken, ctxUser.UUID, actionFeedCategoryEdit) {
			log.Warn().Msg("failed to validate CSRF token")
			view.PutFlashError(w, "There was an error processing the form")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		updatedCategory := feed.Category{
			UserUUID: ctxUser.UUID,
			UUID:     categoryUUID,
			Name:     form.Name,
		}

		if err := fc.feedService.UpdateCategory(updatedCategory); err != nil {
			log.Error().Err(err).Msg("failed to edit feed category")
			view.PutFlashError(w, "failed to edit feed category")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/feeds/categories", http.StatusSeeOther)
	}
}

// handleFeedSubscriptionListView renders the feed category list view.
func (fc *feedHandlerContext) handleFeedSubscriptionListView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		user := httpcontext.UserValue(r.Context())

		subscriptionsByCategory, err := fc.feedQueryingService.SubscriptionTitlesByCategory(user.UUID)
		if err != nil {
			log.Error().Err(err).Str("user_uuid", user.UUID).Msg("failed to retrieve feed subscriptions")
			view.PutFlashError(w, "failed to retrieve feed subscriptions")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		viewData := view.Data{
			Content: subscriptionsByCategory,
			Title:   "Feed Subscriptions",
		}

		fc.feedSubscriptionListView.Render(w, r, viewData)
	}
}
