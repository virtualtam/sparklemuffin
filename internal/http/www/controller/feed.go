// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package controller

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"github.com/virtualtam/opml-go"

	"github.com/virtualtam/sparklemuffin/internal/http/www/htmx"
	"github.com/virtualtam/sparklemuffin/internal/http/www/httpcontext"
	"github.com/virtualtam/sparklemuffin/internal/http/www/middleware"
	"github.com/virtualtam/sparklemuffin/internal/http/www/view"
	"github.com/virtualtam/sparklemuffin/internal/paginate"
	"github.com/virtualtam/sparklemuffin/pkg/feed"
	feedexporting "github.com/virtualtam/sparklemuffin/pkg/feed/exporting"
	feedimporting "github.com/virtualtam/sparklemuffin/pkg/feed/importing"
	feedquerying "github.com/virtualtam/sparklemuffin/pkg/feed/querying"
	"github.com/virtualtam/sparklemuffin/pkg/user"
)

// RegisterFeedHandlers registers HTTP handlers for syndication feed operations.
func RegisterFeedHandlers(
	r *chi.Mux,
	feedService *feed.Service,
	exportingService *feedexporting.Service,
	importingService *feedimporting.Service,
	queryingService *feedquerying.Service,
	userService *user.Service,
) {
	fc := feedController{
		feedService:      feedService,
		exportingService: exportingService,
		importingService: importingService,
		queryingService:  queryingService,
		userService:      userService,

		feedListView: view.New("feed/feed_list.gohtml"),

		feedCategoryAddView:    view.New("feed/category_add.gohtml"),
		feedCategoryDeleteView: view.New("feed/category_delete.gohtml"),
		feedCategoryEditView:   view.New("feed/category_edit.gohtml"),

		feedSubscriptionAddView:    view.New("feed/subscription_add.gohtml"),
		feedSubscriptionDeleteView: view.New("feed/subscription_delete.gohtml"),
		feedSubscriptionEditView:   view.New("feed/subscription_edit.gohtml"),
		feedSubscriptionListView:   view.New("feed/subscription_list.gohtml"),

		feedExportView: view.New("feed/feed_export.gohtml"),
		feedImportView: view.New("feed/feed_import.gohtml"),
	}

	r.Route("/feeds", func(r chi.Router) {
		r.Use(func(h http.Handler) http.Handler {
			return middleware.AuthenticatedUser(h.ServeHTTP)
		})

		r.Get("/", fc.handleFeedListAllView())

		r.Get("/export", fc.handleFeedExportView())
		r.Post("/export", fc.handleFeedExport())
		r.Get("/import", fc.handleFeedImportView())
		r.Post("/import", fc.handleFeedImport())

		r.Route("/categories", func(sr chi.Router) {
			sr.Get("/add", fc.handleFeedCategoryAddView())
			sr.Post("/add", fc.handleFeedCategoryAdd())
			sr.Get("/{uuid}/delete", fc.handleFeedCategoryDeleteView())
			sr.Post("/{uuid}/delete", fc.handleFeedCategoryDelete())
			sr.Get("/{uuid}/edit", fc.handleFeedCategoryEditView())
			sr.Post("/{uuid}/edit", fc.handleFeedCategoryEdit())

			sr.Get("/{slug}", fc.handleFeedListByCategoryView())
			sr.Post("/{slug}/entries/mark-all-read", fc.handleHxEntryMetadataMarkAllAsReadByCategory())
		})

		r.Route("/entries", func(sr chi.Router) {
			sr.Post("/mark-all-read", fc.handleHxEntryMetadataMarkAllAsRead())
			sr.Post("/{uid}/toggle-read", fc.handleHxFeedEntryToggleRead())
		})

		r.Route("/preferences", func(sr chi.Router) {
			sr.Post("/show-entries", fc.handleHxPreferencesFeedShowEntriesUpdate())
			sr.Post("/toggle-show-entry-summaries", fc.handleHxPreferencesToggleShowEntrySummaries())
		})

		r.Route("/subscriptions", func(sr chi.Router) {
			sr.Get("/", fc.handleFeedSubscriptionListView())

			sr.Get("/add", fc.handleFeedSubscriptionAddView())
			sr.Post("/add", fc.handleFeedSubscriptionAdd())

			sr.Get("/{uuid}/delete", fc.handleFeedSubscriptionDeleteView())
			sr.Post("/{uuid}/delete", fc.handleFeedSubscriptionDelete())

			sr.Get("/{uuid}/edit", fc.handleFeedSubscriptionEditView())
			sr.Post("/{uuid}/edit", fc.handleFeedSubscriptionEdit())

			sr.Get("/{slug}", fc.handleFeedListBySubscriptionView())
			sr.Post("/{slug}/entries/mark-all-read", fc.handleHxEntryMetadataMarkAllAsReadByFeed())
		})
	})
}

type feedController struct {
	feedService      *feed.Service
	exportingService *feedexporting.Service
	importingService *feedimporting.Service
	queryingService  *feedquerying.Service
	userService      *user.Service

	feedSubscriptionAddView *view.View
	feedListView            *view.View

	feedCategoryAddView    *view.View
	feedCategoryDeleteView *view.View
	feedCategoryEditView   *view.View

	feedSubscriptionDeleteView *view.View
	feedSubscriptionEditView   *view.View
	feedSubscriptionListView   *view.View

	feedExportView *view.View
	feedImportView *view.View
}

type (
	feedsByPageCallback         func(ctx context.Context, r *http.Request, user *user.User, preferences feed.Preferences, pageNumber uint) (feedquerying.FeedPage, error)
	feedsByQueryAndPageCallback func(ctx context.Context, r *http.Request, user *user.User, preferences feed.Preferences, query string, pageNumber uint) (feedquerying.FeedPage, error)
)

func (fc *feedController) handleFeedListView(
	feedsByPage feedsByPageCallback,
	feedsByQueryAndPage feedsByQueryAndPageCallback,
) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctxUser := httpcontext.UserValue(ctx)

		// TODO: cache user preferences to avoid unnecessary SQL queries
		preferences, err := fc.feedService.PreferencesByUserUUID(ctx, ctxUser.UUID)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve account preferences")
			view.PutFlashError(w, "There was an error retrieving your preferences")
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		var viewData view.Data
		feedQueryingPage := feedQueryingPage{
			URLPath:     r.URL.Path,
			Preferences: preferences,
		}

		pageNumber, pageNumberStr, err := paginate.GetPageNumber(r.URL.Query())
		if err != nil {
			log.Warn().Err(err).Str("page_number", pageNumberStr).Msg("invalid page number")
			view.PutFlashError(w, fmt.Sprintf("invalid page number: %q", pageNumberStr))
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		searchQuery := r.URL.Query().Get("search")
		if searchQuery == "" {
			feedPage, err := feedsByPage(ctx, r, ctxUser, preferences, pageNumber)
			if errors.Is(err, paginate.ErrPageNumberOutOfBounds) {
				msg := fmt.Sprintf("invalid page number: %d", pageNumber)
				log.Warn().Err(err).Msg(msg)
				view.PutFlashError(w, msg)
				http.Redirect(w, r, "/feeds", http.StatusSeeOther)
				return
			} else if err != nil {
				log.Error().Err(err).Msg("failed to retrieve feeds")
				view.PutFlashError(w, "failed to retrieve feeds")
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}

			viewData.Title = fmt.Sprintf("Feeds: %s", feedPage.PageTitle)
			feedQueryingPage.FeedPage = feedPage
		} else {
			feedPage, err := feedsByQueryAndPage(ctx, r, ctxUser, preferences, searchQuery, pageNumber)
			if errors.Is(err, paginate.ErrPageNumberOutOfBounds) {
				msg := fmt.Sprintf("invalid page number: %d", pageNumber)
				log.Warn().Err(err).Msg(msg)
				view.PutFlashError(w, msg)
				http.Redirect(w, r, "/feeds", http.StatusSeeOther)
				return
			} else if err != nil {
				log.Error().Err(err).Msg("failed to retrieve feeds")
				view.PutFlashError(w, "failed to retrieve feeds")
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}

			viewData.Title = fmt.Sprintf("Feed search in %s: %q", feedPage.PageTitle, searchQuery)
			feedQueryingPage.FeedPage = feedPage
		}

		viewData.Content = feedQueryingPage

		fc.feedListView.Render(w, r, viewData)
	}
}

// handleFeedListAllView renders the syndication feed for the current authenticated user.
func (fc *feedController) handleFeedListAllView() func(w http.ResponseWriter, r *http.Request) {
	feedsByPage := func(ctx context.Context, _ *http.Request, user *user.User, preferences feed.Preferences, pageNumber uint) (feedquerying.FeedPage, error) {
		return fc.queryingService.FeedsByPage(ctx, user.UUID, preferences, pageNumber)
	}

	feedsByQueryAndPage := func(ctx context.Context, _ *http.Request, user *user.User, preferences feed.Preferences, query string, pageNumber uint) (feedquerying.FeedPage, error) {
		return fc.queryingService.FeedsByQueryAndPage(ctx, user.UUID, preferences, query, pageNumber)
	}

	return fc.handleFeedListView(feedsByPage, feedsByQueryAndPage)
}

// handleFeedListByCategoryView renders the syndication feed for the current authenticated user.
func (fc *feedController) handleFeedListByCategoryView() func(w http.ResponseWriter, r *http.Request) {
	feedsByPage := func(ctx context.Context, r *http.Request, user *user.User, preferences feed.Preferences, pageNumber uint) (feedquerying.FeedPage, error) {
		categorySlug := chi.URLParam(r, "slug")

		category, err := fc.feedService.CategoryBySlug(ctx, user.UUID, categorySlug)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve feed category")
			return feedquerying.FeedPage{}, err
		}

		return fc.queryingService.FeedsByCategoryAndPage(ctx, user.UUID, preferences, category, pageNumber)
	}

	feedsByQueryAndPage := func(ctx context.Context, r *http.Request, user *user.User, preferences feed.Preferences, query string, pageNumber uint) (feedquerying.FeedPage, error) {
		categorySlug := chi.URLParam(r, "slug")

		category, err := fc.feedService.CategoryBySlug(ctx, user.UUID, categorySlug)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve feed category")
			return feedquerying.FeedPage{}, err
		}

		return fc.queryingService.FeedsByCategoryAndQueryAndPage(ctx, user.UUID, preferences, category, query, pageNumber)
	}

	return fc.handleFeedListView(feedsByPage, feedsByQueryAndPage)
}

// handleFeedListBySubscriptionView renders the syndication feed for the current authenticated user.
func (fc *feedController) handleFeedListBySubscriptionView() func(w http.ResponseWriter, r *http.Request) {
	feedsByPage := func(ctx context.Context, r *http.Request, user *user.User, preferences feed.Preferences, pageNumber uint) (feedquerying.FeedPage, error) {
		feedSlug := chi.URLParam(r, "slug")

		f, err := fc.feedService.FeedBySlug(ctx, feedSlug)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve feed")
			return feedquerying.FeedPage{}, err
		}

		subscription, err := fc.feedService.SubscriptionByFeed(ctx, user.UUID, f.UUID)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve feed subscription")
			return feedquerying.FeedPage{}, err
		}

		return fc.queryingService.FeedsBySubscriptionAndPage(ctx, user.UUID, preferences, subscription, pageNumber)
	}

	feedsByQueryAndPage := func(ctx context.Context, r *http.Request, user *user.User, preferences feed.Preferences, query string, pageNumber uint) (feedquerying.FeedPage, error) {
		feedSlug := chi.URLParam(r, "slug")

		f, err := fc.feedService.FeedBySlug(ctx, feedSlug)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve feed")
			return feedquerying.FeedPage{}, err
		}

		subscription, err := fc.feedService.SubscriptionByFeed(ctx, user.UUID, f.UUID)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve feed subscription")
			return feedquerying.FeedPage{}, err
		}

		return fc.queryingService.FeedsBySubscriptionAndQueryAndPage(ctx, user.UUID, preferences, subscription, query, pageNumber)
	}

	return fc.handleFeedListView(feedsByPage, feedsByQueryAndPage)
}

// handleFeedCategoryAddView renders the feed category addition form.
func (fc *feedController) handleFeedCategoryAddView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		viewData := view.Data{
			Title: "Add feed category",
		}

		fc.feedCategoryAddView.Render(w, r, viewData)
	}
}

// handleFeedCategoryAdd processes the feed category addition form.
func (fc *feedController) handleFeedCategoryAdd() func(w http.ResponseWriter, r *http.Request) {
	type feedAddForm struct {
		Name string `schema:"name"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctxUser := httpcontext.UserValue(ctx)

		var form feedAddForm
		if err := decodeForm(r, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse feed category addition form")
			view.PutFlashError(w, "There was an error processing the form")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		if _, err := fc.feedService.CreateCategory(ctx, ctxUser.UUID, form.Name); err != nil {
			log.Error().Err(err).Msg("failed to add feed category")
			view.PutFlashError(w, "failed to add feed category")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/feeds", http.StatusSeeOther)
	}
}

// handleFeedCategoryDeleteView renders the feed category deletion form.
func (fc *feedController) handleFeedCategoryDeleteView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		categoryUUID := chi.URLParam(r, "uuid")
		ctx := r.Context()
		ctxUser := httpcontext.UserValue(ctx)

		category, err := fc.feedService.CategoryByUUID(ctx, ctxUser.UUID, categoryUUID)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve feed category")
			view.PutFlashError(w, "failed to retrieve feed category")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		viewData := view.Data{
			Content: category,
			Title:   fmt.Sprintf("Delete category: %s", category.Name),
		}

		fc.feedCategoryDeleteView.Render(w, r, viewData)
	}
}

func (fc *feedController) handleFeedCategoryDelete() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		categoryUUID := chi.URLParam(r, "uuid")
		ctx := r.Context()
		ctxUser := httpcontext.UserValue(ctx)

		if err := fc.feedService.DeleteCategory(ctx, ctxUser.UUID, categoryUUID); err != nil {
			log.Error().Err(err).Msg("failed to delete feed category")
			view.PutFlashError(w, "failed to delete feed category")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/feeds", http.StatusSeeOther)
	}
}

// handleFeedCategoryEditView renders the feed category edition form.
func (fc *feedController) handleFeedCategoryEditView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctxUser := httpcontext.UserValue(ctx)

		categoryUUID := chi.URLParam(r, "uuid")

		category, err := fc.feedService.CategoryByUUID(ctx, ctxUser.UUID, categoryUUID)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve feed category")
			view.PutFlashError(w, "failed to retrieve feed category")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		viewData := view.Data{
			Content: category,
			Title:   "Edit feed category",
		}

		fc.feedCategoryEditView.Render(w, r, viewData)
	}
}

// handleFeedCategoryEdit processes the feed category edition form.
func (fc *feedController) handleFeedCategoryEdit() func(w http.ResponseWriter, r *http.Request) {
	type feedCategoryEditForm struct {
		Name string `schema:"name"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctxUser := httpcontext.UserValue(ctx)
		categoryUUID := chi.URLParam(r, "uuid")

		var form feedCategoryEditForm
		if err := decodeForm(r, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse feed category edition form")
			view.PutFlashError(w, "There was an error processing the form")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		updatedCategory := feed.Category{
			UserUUID: ctxUser.UUID,
			UUID:     categoryUUID,
			Name:     form.Name,
		}

		if err := fc.feedService.UpdateCategory(ctx, updatedCategory); err != nil {
			log.Error().Err(err).Msg("failed to edit feed category")
			view.PutFlashError(w, "failed to edit feed category")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/feeds/subscriptions", http.StatusSeeOther)
	}
}

// requireHxRequest rejects a request that carries no proof of having been
// issued by htmx (the HX-Request header), returning true if the caller
// should continue handling it.
//
// The handlers guarded by this only ever render an HTML fragment: they have
// no full-page layout to fall back to, so a request that isn't htmx-issued
// (e.g. a direct hit on the route, or a non-JS client) has no well-defined
// response to send.
func (fc *feedController) requireHxRequest(w http.ResponseWriter, r *http.Request) bool {
	if r.Header.Get(htmx.HeaderRequest) == "true" {
		return true
	}

	log.Error().Err(htmx.ErrMissingRequestHeader).Msg("rejected non-htmx request")
	http.Error(w, htmx.ErrMissingRequestHeader.Error(), http.StatusBadRequest)
	return false
}

// feedPageForContext returns the FeedPage matching the view the user was on (All,
// a category, or a subscription, with an optional search query), so that counts
// derived from it (unread badges, entry count) stay consistent with that view.
func (fc *feedController) feedPageForContext(
	ctx context.Context,
	userUUID string,
	preferences feed.Preferences,
	urlPath string,
	searchTerms string,
	pageNumber uint,
) (feedquerying.FeedPage, error) {
	switch {
	case strings.HasPrefix(urlPath, "/feeds/categories/"):
		category, err := fc.feedService.CategoryBySlug(ctx, userUUID, strings.TrimPrefix(urlPath, "/feeds/categories/"))
		if err != nil {
			return feedquerying.FeedPage{}, err
		}

		if searchTerms == "" {
			return fc.queryingService.FeedsByCategoryAndPage(ctx, userUUID, preferences, category, pageNumber)
		}
		return fc.queryingService.FeedsByCategoryAndQueryAndPage(ctx, userUUID, preferences, category, searchTerms, pageNumber)

	case strings.HasPrefix(urlPath, "/feeds/subscriptions/"):
		f, err := fc.feedService.FeedBySlug(ctx, strings.TrimPrefix(urlPath, "/feeds/subscriptions/"))
		if err != nil {
			return feedquerying.FeedPage{}, err
		}

		subscription, err := fc.feedService.SubscriptionByFeed(ctx, userUUID, f.UUID)
		if err != nil {
			return feedquerying.FeedPage{}, err
		}

		if searchTerms == "" {
			return fc.queryingService.FeedsBySubscriptionAndPage(ctx, userUUID, preferences, subscription, pageNumber)
		}
		return fc.queryingService.FeedsBySubscriptionAndQueryAndPage(ctx, userUUID, preferences, subscription, searchTerms, pageNumber)

	default:
		if searchTerms == "" {
			return fc.queryingService.FeedsByPage(ctx, userUUID, preferences, pageNumber)
		}
		return fc.queryingService.FeedsByQueryAndPage(ctx, userUUID, preferences, searchTerms, pageNumber)
	}
}

// renderFeedListUpdate re-renders the entry list and every fragment whose
// content depends on it (unread badges, entry count, filter button state,
// both pagination widgets), as an htmx response covering all of it in one
// go. It is shared by the feed list's preference/bulk-action endpoints,
// which all end up changing which entries are visible and/or how many
// pages there are, unlike the single-entry toggle-read swap.
//
// On any failure it falls back to the same flash+redirect behavior used
// throughout this file, which htmx follows as a full page reload.
func (fc *feedController) renderFeedListUpdate(
	w http.ResponseWriter,
	r *http.Request,
	userUUID string,
	preferences feed.Preferences,
	urlPath string,
	searchTerms string,
	pageNumber uint,
) {
	ctx := r.Context()

	ctxPage, err := fc.feedPageForContext(ctx, userUUID, preferences, urlPath, searchTerms, pageNumber)
	if err != nil {
		log.Error().Err(err).Msg("failed to retrieve feeds")
		view.RedirectWithFlashError(w, "/feeds", "failed to retrieve feeds")
		return
	}

	var buf bytes.Buffer

	renderFragment := func(name string, data any) bool {
		if err := fc.feedListView.Template.ExecuteTemplate(&buf, name, data); err != nil {
			log.Error().Err(err).Msg("failed to render feed fragment")
			view.RedirectWithFlashError(w, r.Referer(), "failed to render feed fragment")
			return false
		}
		return true
	}

	entryListData := map[string]any{
		"Entries":            ctxPage.Entries,
		"ItemOffset":         ctxPage.ItemOffset,
		"ShowEntrySummaries": preferences.ShowEntrySummaries,
		"URLPath":            urlPath,
		"SearchTerms":        searchTerms,
		"PageNumber":         ctxPage.PageNumber,
	}
	if !renderFragment("entryList", entryListData) {
		return
	}

	if !renderFragment("unreadCountAll", ctxPage.Unread) {
		return
	}

	for _, category := range ctxPage.Categories {
		if !renderFragment("unreadCountCategory", category) {
			return
		}

		for _, subscribedFeed := range category.SubscribedFeeds {
			if !renderFragment("unreadCountFeed", subscribedFeed) {
				return
			}
		}
	}

	if !renderFragment("entryCount", ctxPage.Page) {
		return
	}

	showEntriesButtonsData := map[string]any{
		"ShowEntries": preferences.ShowEntries,
		"URLPath":     urlPath,
		"SearchTerms": searchTerms,
	}
	if !renderFragment("showEntriesButtons", showEntriesButtonsData) {
		return
	}

	compactButtonData := map[string]any{
		"ShowEntrySummaries": preferences.ShowEntrySummaries,
		"URLPath":            urlPath,
		"SearchTerms":        searchTerms,
		"PageNumber":         ctxPage.PageNumber,
	}
	if !renderFragment("compactButton", compactButtonData) {
		return
	}

	if !renderFragment("paginationTop", ctxPage.Page) {
		return
	}

	if !renderFragment("paginationBottom", ctxPage.Page) {
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if _, err := buf.WriteTo(w); err != nil {
		log.Error().Err(err).Msg("failed to write response")
	}
}

// handleHxFeedEntryToggleRead handles a request to toggle the read status of a feed entry.
//
// On success, it responds with an HTML fragment: the re-rendered entry (or nothing,
// if the entry no longer matches the current read/unread filter, so htmx removes it),
// plus out-of-band fragments refreshing the unread badges and entry count that the
// toggle affects. On error, it falls back to the same flash+redirect behavior used
// throughout this file, which htmx follows as a full page reload.
func (fc *feedController) handleHxFeedEntryToggleRead() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if !fc.requireHxRequest(w, r) {
			return
		}

		ctx := r.Context()
		ctxUser := httpcontext.UserValue(ctx)
		entryUID := chi.URLParam(r, "uid")

		if err := fc.feedService.ToggleEntryRead(ctx, ctxUser.UUID, entryUID); err != nil {
			log.Error().Err(err).Msg("failed to set entry metadata")
			view.RedirectWithFlashError(w, r.Referer(), "failed to set entry metadata")
			return
		}

		preferences, err := fc.feedService.PreferencesByUserUUID(ctx, ctxUser.UUID)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve account preferences")
			view.RedirectWithFlashError(w, r.Referer(), "There was an error retrieving your preferences")
			return
		}

		entry, err := fc.queryingService.SubscribedFeedEntryByUID(ctx, ctxUser.UUID, entryUID)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve feed entry")
			view.RedirectWithFlashError(w, r.Referer(), "failed to retrieve feed entry")
			return
		}

		if err := r.ParseForm(); err != nil {
			log.Error().Err(err).Msg("failed to parse request form")
			view.RedirectWithFlashError(w, r.Referer(), "There was an error processing the request")
			return
		}

		urlPath := r.PostForm.Get("urlPath")
		searchTerms := r.PostForm.Get("search")

		pageNumber, _, err := paginate.GetPageNumber(r.PostForm)
		if err != nil {
			pageNumber = 1
		}

		ctxPage, err := fc.feedPageForContext(ctx, ctxUser.UUID, preferences, urlPath, searchTerms, pageNumber)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve feeds")
			view.RedirectWithFlashError(w, "/feeds", "failed to retrieve feeds")
			return
		}

		var buf bytes.Buffer

		renderFragment := func(name string, data any) bool {
			if err := fc.feedListView.Template.ExecuteTemplate(&buf, name, data); err != nil {
				log.Error().Err(err).Msg("failed to render feed fragment")
				view.RedirectWithFlashError(w, r.Referer(), "failed to render feed fragment")
				return false
			}
			return true
		}

		stillVisible := true
		switch preferences.ShowEntries {
		case feed.EntryVisibilityRead:
			stillVisible = entry.Read
		case feed.EntryVisibilityUnread:
			stillVisible = !entry.Read
		}

		if stillVisible {
			entryData := map[string]any{
				"Entry":              entry,
				"ShowEntrySummaries": preferences.ShowEntrySummaries,
				"URLPath":            urlPath,
				"SearchTerms":        searchTerms,
				"PageNumber":         pageNumber,
			}

			if !renderFragment("feedEntry", entryData) {
				return
			}
		}

		if !renderFragment("unreadCountAll", ctxPage.Unread) {
			return
		}

		for _, category := range ctxPage.Categories {
			if !renderFragment("unreadCountCategory", category) {
				return
			}

			for _, subscribedFeed := range category.SubscribedFeeds {
				if !renderFragment("unreadCountFeed", subscribedFeed) {
					return
				}
			}
		}

		if !renderFragment("entryCount", ctxPage.Page) {
			return
		}

		w.Header().Set("Content-Type", "text/html")
		if _, err := buf.WriteTo(w); err != nil {
			log.Error().Err(err).Msg("failed to write response")
		}
	}
}

// handleFeedSubscriptionListView renders the feed category list view.
func (fc *feedController) handleFeedSubscriptionListView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctxUser := httpcontext.UserValue(ctx)

		subscriptionsByCategory, err := fc.queryingService.SubscriptionsByCategory(ctx, ctxUser.UUID)
		if err != nil {
			log.Error().Err(err).Str("user_uuid", ctxUser.UUID).Msg("failed to retrieve feed subscriptions")
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

// handleFeedSubscriptionAddView renders the feed subscription addition form.
func (fc *feedController) handleFeedSubscriptionAddView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctxUser := httpcontext.UserValue(ctx)

		categories, err := fc.feedService.Categories(ctx, ctxUser.UUID)
		if err != nil {
			log.Error().Err(err).Str("user_uuid", ctxUser.UUID).Msg("failed to retrieve feed categories")
			view.PutFlashError(w, "failed to retrieve existing feed categories")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		viewData := view.Data{
			Content: categories,
			Title:   "Add feed",
		}

		fc.feedSubscriptionAddView.Render(w, r, viewData)
	}
}

// handleFeedSubscriptionAdd processes the feed subscription addition form.
func (fc *feedController) handleFeedSubscriptionAdd() func(w http.ResponseWriter, r *http.Request) {
	type feedAddForm struct {
		URL          string `schema:"url"`
		CategoryUUID string `schema:"category"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctxUser := httpcontext.UserValue(ctx)

		var form feedAddForm
		if err := decodeForm(r, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse feed subscription form")
			view.PutFlashError(w, "There was an error processing the form")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		if err := fc.feedService.Subscribe(ctx, ctxUser.UUID, form.CategoryUUID, form.URL); err != nil {
			log.Error().Err(err).Msg("failed to subscribe to feed")
			view.PutFlashError(w, "failed to subscribe to feed")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/feeds", http.StatusSeeOther)
	}
}

// handleFeedSubscriptionDeleteView renders the feed subscription deletion form.
func (fc *feedController) handleFeedSubscriptionDeleteView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		subscriptionUUID := chi.URLParam(r, "uuid")
		ctx := r.Context()
		ctxUser := httpcontext.UserValue(ctx)

		subscription, err := fc.queryingService.SubscriptionByUUID(ctx, ctxUser.UUID, subscriptionUUID)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve feed subscription")
			view.PutFlashError(w, "failed to retrieve feed subscription")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		viewData := view.Data{
			Content: subscription,
			Title:   fmt.Sprintf("Delete subscription: %s", subscription.FeedTitle),
		}

		fc.feedSubscriptionDeleteView.Render(w, r, viewData)
	}
}

// handleFeedSubscriptionDelete processes the feed subscription deletion form.
func (fc *feedController) handleFeedSubscriptionDelete() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		subscriptionUUID := chi.URLParam(r, "uuid")
		ctx := r.Context()
		ctxUser := httpcontext.UserValue(ctx)

		if err := fc.feedService.DeleteSubscription(ctx, ctxUser.UUID, subscriptionUUID); err != nil {
			log.Error().Err(err).Msg("failed to delete feed subscription")
			view.PutFlashError(w, "failed to delete feed subscription")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/feeds", http.StatusSeeOther)
	}
}

// handleFeedSubscriptionEditView renders the feed subscription edition form.
func (fc *feedController) handleFeedSubscriptionEditView() func(w http.ResponseWriter, r *http.Request) {
	type feedSubscriptionEditFormContent struct {
		Subscription feedquerying.Subscription
		Categories   []feed.Category
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctxUser := httpcontext.UserValue(ctx)
		subscriptionUUID := chi.URLParam(r, "uuid")

		subscription, err := fc.queryingService.SubscriptionByUUID(ctx, ctxUser.UUID, subscriptionUUID)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve feed subscription")
			view.PutFlashError(w, "failed to retrieve feed subscription")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		categories, err := fc.feedService.Categories(ctx, ctxUser.UUID)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve feed categories")
			view.PutFlashError(w, "failed to retrieve feed categories")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		viewData := view.Data{
			Content: feedSubscriptionEditFormContent{
				Subscription: subscription,
				Categories:   categories,
			},
			Title: "Edit feed subscription",
		}

		fc.feedSubscriptionEditView.Render(w, r, viewData)
	}
}

// handleFeedSubscriptionEdit processes the feed subscription edition form.
func (fc *feedController) handleFeedSubscriptionEdit() func(w http.ResponseWriter, r *http.Request) {
	type feedSubscriptionEditForm struct {
		Alias        string `schema:"alias"`
		CategoryUUID string `schema:"category"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctxUser := httpcontext.UserValue(ctx)
		subscriptionUUID := chi.URLParam(r, "uuid")

		var form feedSubscriptionEditForm
		if err := decodeForm(r, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse feed subscription edition form")
			view.PutFlashError(w, "There was an error processing the form")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		updatedSubscription := feed.Subscription{
			UserUUID:     ctxUser.UUID,
			UUID:         subscriptionUUID,
			Alias:        form.Alias,
			CategoryUUID: form.CategoryUUID,
		}

		if err := fc.feedService.UpdateSubscription(ctx, updatedSubscription); err != nil {
			log.Error().Err(err).Msg("failed to edit feed subscription")
			view.PutFlashError(w, "failed to edit feed subscription")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/feeds/subscriptions", http.StatusSeeOther)
	}
}

// handleHxEntryMetadataMarkAllAsRead handles a request to mark all feed entries as read.
//
// On success, it responds with the re-rendered entry list (reset to page 1,
// since marking everything read can shrink or empty an Unread-only view)
// plus every fragment that depends on it. On error, it falls back to the
// same flash+redirect behavior used throughout this file.
func (fc *feedController) handleHxEntryMetadataMarkAllAsRead() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if !fc.requireHxRequest(w, r) {
			return
		}

		ctx := r.Context()
		ctxUser := httpcontext.UserValue(ctx)

		if err := fc.feedService.MarkAllEntriesAsRead(ctx, ctxUser.UUID); err != nil {
			log.Error().Err(err).Msg("failed to mark feed entries as read")
			view.RedirectWithFlashError(w, r.Referer(), "failed to mark feed entries as read")
			return
		}

		preferences, err := fc.feedService.PreferencesByUserUUID(ctx, ctxUser.UUID)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve account preferences")
			view.RedirectWithFlashError(w, r.Referer(), "There was an error retrieving your preferences")
			return
		}

		if err := r.ParseForm(); err != nil {
			log.Error().Err(err).Msg("failed to parse request form")
			view.RedirectWithFlashError(w, r.Referer(), "There was an error processing the request")
			return
		}

		urlPath := r.PostForm.Get("urlPath")
		searchTerms := r.PostForm.Get("search")

		fc.renderFeedListUpdate(w, r, ctxUser.UUID, preferences, urlPath, searchTerms, 1)
	}
}

// handleHxEntryMetadataMarkAllAsReadByCategory handles a request to mark all feed entries as read for a given category.
//
// See handleHxEntryMetadataMarkAllAsRead for the response behavior.
func (fc *feedController) handleHxEntryMetadataMarkAllAsReadByCategory() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if !fc.requireHxRequest(w, r) {
			return
		}

		ctx := r.Context()
		ctxUser := httpcontext.UserValue(ctx)
		categorySlug := chi.URLParam(r, "slug")

		category, err := fc.feedService.CategoryBySlug(ctx, ctxUser.UUID, categorySlug)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve feed category")
			view.RedirectWithFlashError(w, "/feeds", "failed to retrieve feed category")
			return
		}

		if err := fc.feedService.MarkAllEntriesAsReadByCategory(ctx, ctxUser.UUID, category.UUID); err != nil {
			log.Error().Err(err).Msg("failed to mark feed entries as read")
			view.RedirectWithFlashError(w, r.Referer(), "failed to mark feed entries as read")
			return
		}

		preferences, err := fc.feedService.PreferencesByUserUUID(ctx, ctxUser.UUID)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve account preferences")
			view.RedirectWithFlashError(w, r.Referer(), "There was an error retrieving your preferences")
			return
		}

		if err := r.ParseForm(); err != nil {
			log.Error().Err(err).Msg("failed to parse request form")
			view.RedirectWithFlashError(w, r.Referer(), "There was an error processing the request")
			return
		}

		urlPath := r.PostForm.Get("urlPath")
		searchTerms := r.PostForm.Get("search")

		fc.renderFeedListUpdate(w, r, ctxUser.UUID, preferences, urlPath, searchTerms, 1)
	}
}

// handleHxEntryMetadataMarkAllAsReadByFeed handles a request to mark all feed entries as read for a given feed.
//
// See handleHxEntryMetadataMarkAllAsRead for the response behavior.
func (fc *feedController) handleHxEntryMetadataMarkAllAsReadByFeed() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if !fc.requireHxRequest(w, r) {
			return
		}

		ctx := r.Context()
		ctxUser := httpcontext.UserValue(ctx)
		feedSlug := chi.URLParam(r, "slug")

		f, err := fc.feedService.FeedBySlug(ctx, feedSlug)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve feed")
			view.RedirectWithFlashError(w, "/feeds", "failed to retrieve feed")
			return
		}

		subscription, err := fc.feedService.SubscriptionByFeed(ctx, ctxUser.UUID, f.UUID)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve feed subscription")
			view.RedirectWithFlashError(w, "/feeds", "failed to retrieve feed subscription")
			return
		}

		if err := fc.feedService.MarkAllEntriesAsReadBySubscription(ctx, ctxUser.UUID, subscription.UUID); err != nil {
			log.Error().Err(err).Msg("failed to mark feed entries as read")
			view.RedirectWithFlashError(w, r.Referer(), "failed to mark feed entries as read")
			return
		}

		preferences, err := fc.feedService.PreferencesByUserUUID(ctx, ctxUser.UUID)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve account preferences")
			view.RedirectWithFlashError(w, r.Referer(), "There was an error retrieving your preferences")
			return
		}

		if err := r.ParseForm(); err != nil {
			log.Error().Err(err).Msg("failed to parse request form")
			view.RedirectWithFlashError(w, r.Referer(), "There was an error processing the request")
			return
		}

		urlPath := r.PostForm.Get("urlPath")
		searchTerms := r.PostForm.Get("search")

		fc.renderFeedListUpdate(w, r, ctxUser.UUID, preferences, urlPath, searchTerms, 1)
	}
}

// handleFeedExportView renders the feed export page.
func (fc *feedController) handleFeedExportView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		viewData := view.Data{
			Title: "Export feed subscriptions",
		}

		fc.feedExportView.Render(w, r, viewData)
	}
}

// handleFeedExport processes the feed subscription export form and sends the
// corresponding file to the client.
func (fc *feedController) handleFeedExport() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctxUser := httpcontext.UserValue(ctx)

		opmlDocument, err := fc.exportingService.ExportAsOPMLDocument(ctx, *ctxUser)
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

// handleFeedExportView renders the feed subscription import page.
func (fc *feedController) handleFeedImportView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		viewData := view.Data{
			Title: "Import feed subscriptions",
		}

		fc.feedImportView.Render(w, r, viewData)
	}
}

// handleFeedImport processes data submitted through the feed subscription import form.
func (fc *feedController) handleFeedImport() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		multipartReader, err := r.MultipartReader()
		if err != nil {
			log.Error().Err(err).Msg("failed to access multipart reader")
			view.PutFlashError(w, "failed to process import form")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		var importFileBuffer bytes.Buffer
		importFileWriter := bufio.NewWriter(&importFileBuffer)

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
			case "importfile":
				_, err = io.Copy(importFileWriter, part)
			default:
				err = fmt.Errorf("unexpected multipart form field: %q", part.FormName())
			}

			if err != nil {
				log.Error().Err(err).Msg(fmt.Sprintf("failed to process multipart form part %q", part.FormName()))
				view.PutFlashError(w, "failed to process import form")
				http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
				return
			}
		}

		ctx := r.Context()
		ctxUser := httpcontext.UserValue(ctx)

		document, err := opml.Unmarshal(importFileBuffer.Bytes())
		if err != nil {
			log.Error().Err(err).Msg("failed to process OPML feed subscription file")
			view.PutFlashError(w, "failed to import feed subscriptions from the uploaded file")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		importStatus, err := fc.importingService.ImportFromOPMLDocument(ctx, ctxUser.UUID, document)
		if err != nil {
			log.Error().Err(err).Msg("failed to save imported feed subscriptions")
			view.PutFlashError(w, "failed to save imported feed subscriptions")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		view.PutFlashSuccess(w, fmt.Sprintf("Import status: %s", importStatus.UserSummary()))
		http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
	}
}

// handlePreferencesToggleShowEntrySummaries handles a request to toggle
// whether entry summaries are shown ("Compact" mode).
//
// On success, it responds with the re-rendered entry list, preserving the
// current page (unlike the filter change, toggling summaries doesn't affect
// which entries match or how many pages there are), plus every fragment that
// depends on it. On error, it falls back to the same flash+redirect behavior
// used throughout this file.
func (fc *feedController) handleHxPreferencesToggleShowEntrySummaries() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if !fc.requireHxRequest(w, r) {
			return
		}

		ctx := r.Context()
		ctxUser := httpcontext.UserValue(ctx)

		if err := r.ParseForm(); err != nil {
			log.Error().Err(err).Msg("failed to parse request form")
			view.RedirectWithFlashError(w, r.Referer(), "There was an error processing the request")
			return
		}

		preferences, err := fc.feedService.PreferencesByUserUUID(ctx, ctxUser.UUID)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve account preferences")
			view.RedirectWithFlashError(w, r.Referer(), fmt.Sprintf("There was an error updating your preferences: %s", err))
			return
		}

		preferences.ShowEntrySummaries = !preferences.ShowEntrySummaries

		if err := fc.feedService.UpdatePreferences(ctx, preferences); err != nil {
			log.Error().Err(err).Msg("failed to update account preferences")
			view.RedirectWithFlashError(w, r.Referer(), fmt.Sprintf("There was an error updating your preferences: %s", err))
			return
		}

		urlPath := r.PostForm.Get("urlPath")
		searchTerms := r.PostForm.Get("search")

		pageNumber, _, err := paginate.GetPageNumber(r.PostForm)
		if err != nil {
			pageNumber = 1
		}

		fc.renderFeedListUpdate(w, r, ctxUser.UUID, preferences, urlPath, searchTerms, pageNumber)
	}
}

// handlePreferencesFeedShowEntriesUpdate handles a request to change the
// All/Read/Unread entry filter.
//
// On success, it responds with the re-rendered entry list (reset to page 1,
// since the current page may no longer make sense for the new filter) plus
// every fragment that depends on it. On error, it falls back to the same
// flash+redirect behavior used throughout this file.
func (fc *feedController) handleHxPreferencesFeedShowEntriesUpdate() func(w http.ResponseWriter, r *http.Request) {
	type feedShowEntriesForm struct {
		ShowEntries string `schema:"show"`
		URLPath     string `schema:"urlPath"`
		SearchTerms string `schema:"search"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if !fc.requireHxRequest(w, r) {
			return
		}

		ctx := r.Context()
		ctxUser := httpcontext.UserValue(ctx)

		var form feedShowEntriesForm
		if err := decodeForm(r, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse feed preferences update form")
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		preferences, err := fc.feedService.PreferencesByUserUUID(ctx, ctxUser.UUID)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve account preferences")
			view.RedirectWithFlashError(w, r.Referer(), fmt.Sprintf("There was an error updating your preferences: %s", err))
			return
		}

		preferences.ShowEntries = feed.EntryVisibility(form.ShowEntries)

		if err := fc.feedService.UpdatePreferences(ctx, preferences); err != nil {
			log.Error().Err(err).Msg("failed to update account preferences")
			view.RedirectWithFlashError(w, r.Referer(), fmt.Sprintf("There was an error updating your preferences: %s", err))
			return
		}

		fc.renderFeedListUpdate(w, r, ctxUser.UUID, preferences, form.URLPath, form.SearchTerms, 1)
	}
}
