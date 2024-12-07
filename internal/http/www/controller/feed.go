// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package controller

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	"github.com/virtualtam/sparklemuffin/internal/http/www/csrf"
	"github.com/virtualtam/sparklemuffin/internal/http/www/httpcontext"
	"github.com/virtualtam/sparklemuffin/internal/http/www/middleware"
	"github.com/virtualtam/sparklemuffin/internal/http/www/view"
	"github.com/virtualtam/sparklemuffin/internal/paginate"
	"github.com/virtualtam/sparklemuffin/pkg/feed"
	feedquerying "github.com/virtualtam/sparklemuffin/pkg/feed/querying"
	"github.com/virtualtam/sparklemuffin/pkg/user"
)

const (
	actionFeedCategoryAdd        string = "feed-category-add"
	actionFeedCategoryDelete     string = "feed-category-delete"
	actionFeedCategoryEdit       string = "feed-category-edit"
	actionFeedEntryMetadataEdit  string = "feed-entry-metadata-edit"
	actionFeedSubscriptionAdd    string = "feed-subscription-add"
	actionFeedSubscriptionDelete string = "feed-subscription-delete"
	actionFeedSubscriptionEdit   string = "feed-subscription-edit"
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

		feedListView: view.New("feed/feed_list.gohtml"),
		feedAddView:  view.New("feed/feed_add.gohtml"),

		feedCategoryAddView:    view.New("feed/category_add.gohtml"),
		feedCategoryDeleteView: view.New("feed/category_delete.gohtml"),
		feedCategoryEditView:   view.New("feed/category_edit.gohtml"),

		feedSubscriptionDeleteView: view.New("feed/subscription_delete.gohtml"),
		feedSubscriptionEditView:   view.New("feed/subscription_edit.gohtml"),
		feedSubscriptionListView:   view.New("feed/subscription_list.gohtml"),
	}

	r.Route("/feeds", func(r chi.Router) {
		r.Use(func(h http.Handler) http.Handler {
			return middleware.AuthenticatedUser(h.ServeHTTP)
		})

		r.Get("/", fc.handleFeedListAllView())
		r.Get("/add", fc.handleFeedAddView())
		r.Post("/add", fc.handleFeedAdd())

		r.Route("/categories", func(sr chi.Router) {
			sr.Get("/add", fc.handleFeedCategoryAddView())
			sr.Post("/add", fc.handleFeedCategoryAdd())
			sr.Get("/{uuid}/delete", fc.handleFeedCategoryDeleteView())
			sr.Post("/{uuid}/delete", fc.handleFeedCategoryDelete())
			sr.Get("/{uuid}/edit", fc.handleFeedCategoryEditView())
			sr.Post("/{uuid}/edit", fc.handleFeedCategoryEdit())

			sr.Get("/{slug}", fc.handleFeedListByCategoryView())
			sr.Post("/{slug}/entries/mark-all-read", fc.handleEntryMetadataMarkAllAsReadByCategory())
		})

		r.Route("/entries", func(sr chi.Router) {
			sr.Post("/mark-all-read", fc.handleEntryMetadataMarkAllAsRead())
			sr.Post("/{uid}/toggle-read", fc.handleFeedEntryToggleRead())
		})

		r.Route("/subscriptions", func(sr chi.Router) {
			sr.Get("/", fc.handleFeedSubscriptionListView())
			sr.Get("/{uuid}/delete", fc.handleFeedSubscriptionDeleteView())
			sr.Post("/{uuid}/delete", fc.handleFeedSubscriptionDelete())
			sr.Get("/{uuid}/edit", fc.handleFeedSubscriptionEditView())
			sr.Post("/{uuid}/edit", fc.handleFeedSubscriptionEdit())

			sr.Get("/{slug}", fc.handleFeedListBySubscriptionView())
			sr.Post("/{slug}/entries/mark-all-read", fc.handleEntryMetadataMarkAllAsReadByFeed())
		})
	})
}

type feedHandlerContext struct {
	csrfService         *csrf.Service
	feedService         *feed.Service
	feedQueryingService *feedquerying.Service
	userService         *user.Service

	feedAddView  *view.View
	feedListView *view.View

	feedCategoryAddView    *view.View
	feedCategoryDeleteView *view.View
	feedCategoryEditView   *view.View

	feedSubscriptionDeleteView *view.View
	feedSubscriptionEditView   *view.View
	feedSubscriptionListView   *view.View
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
		csrfToken := fc.csrfService.Generate(user.UUID, actionFeedSubscriptionAdd)

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

		if !fc.csrfService.Validate(form.CSRFToken, ctxUser.UUID, actionFeedSubscriptionAdd) {
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

type (
	feedsByPageCallback         func(r *http.Request, user *user.User, pageNumber uint) (feedquerying.FeedPage, error)
	feedsByQueryAndPageCallback func(r *http.Request, user *user.User, query string, pageNumber uint) (feedquerying.FeedPage, error)
)

func (fc *feedHandlerContext) handleFeedListView(
	feedsByPage feedsByPageCallback,
	feedsByQueryAndPage feedsByQueryAndPageCallback,
) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		user := httpcontext.UserValue(r.Context())

		csrfToken := fc.csrfService.Generate(user.UUID, actionFeedEntryMetadataEdit)

		var viewData view.Data
		feedQueryingPage := feedQueryingPage{
			CSRFToken: csrfToken,
			URLPath:   r.URL.Path,
		}

		pageNumber, pageNumberStr, err := paginate.GetPageNumber(r.URL.Query())
		if err != nil {
			log.Error().Err(err).Str("page_number", pageNumberStr).Msg("invalid page number")
			view.PutFlashError(w, fmt.Sprintf("invalid page number: %q", pageNumberStr))
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		searchQuery := r.URL.Query().Get("search")
		if searchQuery == "" {
			feedPage, err := feedsByPage(r, user, pageNumber)
			if err != nil {
				log.Error().Err(err).Msg("failed to retrieve feeds")
				view.PutFlashError(w, "failed to retrieve feeds")
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}

			viewData.Title = fmt.Sprintf("Feeds: %s", feedPage.Header)
			feedQueryingPage.FeedPage = feedPage
		} else {
			feedPage, err := feedsByQueryAndPage(r, user, searchQuery, pageNumber)
			if err != nil {
				log.Error().Err(err).Msg("failed to retrieve feeds")
				view.PutFlashError(w, "failed to retrieve feeds")
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}

			viewData.Title = fmt.Sprintf("Feed search in %s: %q", feedPage.Header, searchQuery)
			feedQueryingPage.FeedPage = feedPage
		}

		viewData.Content = feedQueryingPage

		fc.feedListView.Render(w, r, viewData)
	}
}

// handleFeedListAllView renders the syndication feed for the current authenticated user.
func (fc *feedHandlerContext) handleFeedListAllView() func(w http.ResponseWriter, r *http.Request) {
	feedsByPage := func(_ *http.Request, user *user.User, pageNumber uint) (feedquerying.FeedPage, error) {
		return fc.feedQueryingService.FeedsByPage(user.UUID, pageNumber)
	}

	feedsByQueryAndPage := func(_ *http.Request, user *user.User, query string, pageNumber uint) (feedquerying.FeedPage, error) {
		return fc.feedQueryingService.FeedsByQueryAndPage(user.UUID, query, pageNumber)
	}

	return fc.handleFeedListView(feedsByPage, feedsByQueryAndPage)
}

// handleFeedListByCategoryView renders the syndication feed for the current authenticated user.
func (fc *feedHandlerContext) handleFeedListByCategoryView() func(w http.ResponseWriter, r *http.Request) {
	feedsByPage := func(r *http.Request, user *user.User, pageNumber uint) (feedquerying.FeedPage, error) {
		categorySlug := chi.URLParam(r, "slug")

		category, err := fc.feedService.CategoryBySlug(user.UUID, categorySlug)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve feed category")
			return feedquerying.FeedPage{}, err
		}

		return fc.feedQueryingService.FeedsByCategoryAndPage(user.UUID, category, pageNumber)
	}

	feedsByQueryAndPage := func(r *http.Request, user *user.User, query string, pageNumber uint) (feedquerying.FeedPage, error) {
		categorySlug := chi.URLParam(r, "slug")

		category, err := fc.feedService.CategoryBySlug(user.UUID, categorySlug)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve feed category")
			return feedquerying.FeedPage{}, err
		}

		return fc.feedQueryingService.FeedsByCategoryAndQueryAndPage(user.UUID, category, query, pageNumber)
	}

	return fc.handleFeedListView(feedsByPage, feedsByQueryAndPage)
}

// handleFeedListBySubscriptionView renders the syndication feed for the current authenticated user.
func (fc *feedHandlerContext) handleFeedListBySubscriptionView() func(w http.ResponseWriter, r *http.Request) {
	feedsByPage := func(r *http.Request, user *user.User, pageNumber uint) (feedquerying.FeedPage, error) {
		feedSlug := chi.URLParam(r, "slug")

		feed, err := fc.feedService.FeedBySlug(feedSlug)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve feed")
			return feedquerying.FeedPage{}, err
		}

		subscription, err := fc.feedService.SubscriptionByFeed(user.UUID, feed.UUID)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve feed subscription")
			return feedquerying.FeedPage{}, err
		}

		return fc.feedQueryingService.FeedsBySubscriptionAndPage(user.UUID, subscription, pageNumber)
	}

	feedsByQueryAndPage := func(r *http.Request, user *user.User, query string, pageNumber uint) (feedquerying.FeedPage, error) {
		feedSlug := chi.URLParam(r, "slug")

		feed, err := fc.feedService.FeedBySlug(feedSlug)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve feed")
			return feedquerying.FeedPage{}, err
		}

		subscription, err := fc.feedService.SubscriptionByFeed(user.UUID, feed.UUID)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve feed subscription")
			return feedquerying.FeedPage{}, err
		}

		return fc.feedQueryingService.FeedsBySubscriptionAndQueryAndPage(user.UUID, subscription, query, pageNumber)
	}

	return fc.handleFeedListView(feedsByPage, feedsByQueryAndPage)
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

		if _, err := fc.feedService.CreateCategory(ctxUser.UUID, form.Name); err != nil {
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

		http.Redirect(w, r, "/feeds/subscriptions", http.StatusSeeOther)
	}
}

func (fc *feedHandlerContext) handleFeedEntryToggleRead() func(w http.ResponseWriter, r *http.Request) {
	type feedEntryReadForm struct {
		CSRFToken string `schema:"csrf_token"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctxUser := httpcontext.UserValue(r.Context())
		entryUID := chi.URLParam(r, "uid")

		var form feedEntryReadForm
		if err := decodeForm(r, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse feed entry read toggle form")
			view.PutFlashError(w, "There was an error processing the form")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		if !fc.csrfService.Validate(form.CSRFToken, ctxUser.UUID, actionFeedEntryMetadataEdit) {
			log.Warn().Msg("failed to validate CSRF token")
			view.PutFlashError(w, "There was an error processing the form")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		if err := fc.feedService.ToggleEntryRead(ctxUser.UUID, entryUID); err != nil {
			log.Error().Err(err).Msg("failed to set entry metadata")
			view.PutFlashError(w, "failed to set entry metadata")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
	}
}

// handleFeedSubscriptionListView renders the feed category list view.
func (fc *feedHandlerContext) handleFeedSubscriptionListView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		user := httpcontext.UserValue(r.Context())

		subscriptionsByCategory, err := fc.feedQueryingService.SubscriptionsByCategory(user.UUID)
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

// handleFeedSubscriptionDeleteView renders the feed subscription deletion form.
func (fc *feedHandlerContext) handleFeedSubscriptionDeleteView() func(w http.ResponseWriter, r *http.Request) {
	type feedSubscriptionTitleFormContent struct {
		CSRFToken         string
		SubscriptionTitle feedquerying.Subscription
	}

	return func(w http.ResponseWriter, r *http.Request) {
		subscriptionUUID := chi.URLParam(r, "uuid")
		user := httpcontext.UserValue(r.Context())
		csrfToken := fc.csrfService.Generate(user.UUID, actionFeedSubscriptionDelete)

		subscriptionTitle, err := fc.feedQueryingService.SubscriptionByUUID(user.UUID, subscriptionUUID)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve feed subscription")
			view.PutFlashError(w, "failed to retrieve feed subscription")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		viewData := view.Data{
			Content: feedSubscriptionTitleFormContent{
				CSRFToken:         csrfToken,
				SubscriptionTitle: subscriptionTitle,
			},
			Title: fmt.Sprintf("Delete subscription: %s", subscriptionTitle.FeedTitle),
		}

		fc.feedSubscriptionDeleteView.Render(w, r, viewData)
	}
}

func (fc *feedHandlerContext) handleFeedSubscriptionDelete() func(w http.ResponseWriter, r *http.Request) {
	type feedSubscriptionDeleteForm struct {
		CSRFToken string `schema:"csrf_token"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		subscriptionUUID := chi.URLParam(r, "uuid")
		user := httpcontext.UserValue(r.Context())

		var form feedSubscriptionDeleteForm
		if err := decodeForm(r, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse feed subscription deletion form")
			view.PutFlashError(w, "There was an error processing the form")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		if !fc.csrfService.Validate(form.CSRFToken, user.UUID, actionFeedSubscriptionDelete) {
			log.Warn().Msg("failed to validate CSRF token")
			view.PutFlashError(w, "There was an error processing the form")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		if err := fc.feedService.DeleteSubscription(user.UUID, subscriptionUUID); err != nil {
			log.Error().Err(err).Msg("failed to delete feed subscription")
			view.PutFlashError(w, "failed to delete feed subscription")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/feeds", http.StatusSeeOther)
	}
}

// handleFeedSubscriptionEditView renders the feed subscription edition form.
func (fc *feedHandlerContext) handleFeedSubscriptionEditView() func(w http.ResponseWriter, r *http.Request) {
	type feedSubscriptionEditFormContent struct {
		CSRFToken    string
		Subscription feedquerying.Subscription
		Categories   []feed.Category
	}

	return func(w http.ResponseWriter, r *http.Request) {
		user := httpcontext.UserValue(r.Context())
		subscriptionUUID := chi.URLParam(r, "uuid")

		csrfToken := fc.csrfService.Generate(user.UUID, actionFeedSubscriptionEdit)

		subscription, err := fc.feedQueryingService.SubscriptionByUUID(user.UUID, subscriptionUUID)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve feed subscription")
			view.PutFlashError(w, "failed to retrieve feed subscription")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		categories, err := fc.feedService.Categories(user.UUID)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve feed categories")
			view.PutFlashError(w, "failed to retrieve feed categories")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		viewData := view.Data{
			Content: feedSubscriptionEditFormContent{
				CSRFToken:    csrfToken,
				Subscription: subscription,
				Categories:   categories,
			},
			Title: "Edit feed subscription",
		}

		fc.feedSubscriptionEditView.Render(w, r, viewData)
	}
}

// handleFeedSubscriptionEdit processes the feed subscription edition form.
func (fc *feedHandlerContext) handleFeedSubscriptionEdit() func(w http.ResponseWriter, r *http.Request) {
	type feedSubscriptionEditForm struct {
		CSRFToken    string `schema:"csrf_token"`
		Alias        string `schema:"alias"`
		CategoryUUID string `schema:"category"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		user := httpcontext.UserValue(r.Context())
		subscriptionUUID := chi.URLParam(r, "uuid")

		var form feedSubscriptionEditForm
		if err := decodeForm(r, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse feed subscription edition form")
			view.PutFlashError(w, "There was an error processing the form")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		if !fc.csrfService.Validate(form.CSRFToken, user.UUID, actionFeedSubscriptionEdit) {
			log.Warn().Msg("failed to validate CSRF token")
			view.PutFlashError(w, "There was an error processing the form")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		updatedSubscription := feed.Subscription{
			UserUUID:     user.UUID,
			UUID:         subscriptionUUID,
			Alias:        form.Alias,
			CategoryUUID: form.CategoryUUID,
		}

		if err := fc.feedService.UpdateSubscription(updatedSubscription); err != nil {
			log.Error().Err(err).Msg("failed to edit feed subscription")
			view.PutFlashError(w, "failed to edit feed subscription")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/feeds/subscriptions", http.StatusSeeOther)
	}
}

func (fc *feedHandlerContext) handleEntryMetadataMarkAllAsRead() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		user := httpcontext.UserValue(r.Context())

		var form feedEntryMetadataMarkReadForm
		if err := decodeForm(r, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse feed entry metadata edition form")
			view.PutFlashError(w, "There was an error processing the form")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		if !fc.csrfService.Validate(form.CSRFToken, user.UUID, actionFeedEntryMetadataEdit) {
			log.Warn().Msg("failed to validate CSRF token")
			view.PutFlashError(w, "There was an error processing the form")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		if err := fc.feedService.MarkAllEntriesAsRead(user.UUID); err != nil {
			log.Error().Err(err).Msg("failed to mark feed entries as read")
			view.PutFlashError(w, "failed to mark feed entries as read")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
	}
}

func (fc *feedHandlerContext) handleEntryMetadataMarkAllAsReadByCategory() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		user := httpcontext.UserValue(r.Context())
		categorySlug := chi.URLParam(r, "slug")

		var form feedEntryMetadataMarkReadForm
		if err := decodeForm(r, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse feed entry metadata edition form")
			view.PutFlashError(w, "There was an error processing the form")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		if !fc.csrfService.Validate(form.CSRFToken, user.UUID, actionFeedEntryMetadataEdit) {
			log.Warn().Msg("failed to validate CSRF token")
			view.PutFlashError(w, "There was an error processing the form")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		category, err := fc.feedService.CategoryBySlug(user.UUID, categorySlug)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve feed category")
			view.PutFlashError(w, "failed to retrieve feed category")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		if err := fc.feedService.MarkAllEntriesAsReadByCategory(user.UUID, category.UUID); err != nil {
			log.Error().Err(err).Msg("failed to mark feed entries as read")
			view.PutFlashError(w, "failed to mark feed entries as read")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
	}
}

func (fc *feedHandlerContext) handleEntryMetadataMarkAllAsReadByFeed() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		user := httpcontext.UserValue(r.Context())
		feedSlug := chi.URLParam(r, "slug")

		var form feedEntryMetadataMarkReadForm
		if err := decodeForm(r, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse feed entry metadata edition form")
			view.PutFlashError(w, "There was an error processing the form")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		if !fc.csrfService.Validate(form.CSRFToken, user.UUID, actionFeedEntryMetadataEdit) {
			log.Warn().Msg("failed to validate CSRF token")
			view.PutFlashError(w, "There was an error processing the form")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		feed, err := fc.feedService.FeedBySlug(feedSlug)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve feed")
			view.PutFlashError(w, "failed to retrieve feed")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		subscription, err := fc.feedService.SubscriptionByFeed(user.UUID, feed.UUID)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve feed subscription")
			view.PutFlashError(w, "failed to retrieve feed subscription")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		if err := fc.feedService.MarkAllEntriesAsReadBySubscription(user.UUID, subscription.UUID); err != nil {
			log.Error().Err(err).Msg("failed to mark feed entries as read")
			view.PutFlashError(w, "failed to mark feed entries as read")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
	}
}