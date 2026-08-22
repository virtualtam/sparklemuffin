// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package controller

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	"github.com/virtualtam/sparklemuffin/internal/http/www/htmx"
	"github.com/virtualtam/sparklemuffin/internal/http/www/httpcontext"
	"github.com/virtualtam/sparklemuffin/internal/http/www/middleware"
	"github.com/virtualtam/sparklemuffin/internal/http/www/view"
	"github.com/virtualtam/sparklemuffin/internal/paginate"
	bookmarkquerying "github.com/virtualtam/sparklemuffin/pkg/bookmark/querying"
	feedquerying "github.com/virtualtam/sparklemuffin/pkg/feed/querying"
	"github.com/virtualtam/sparklemuffin/pkg/taxonomy"
)

const tagsPerPage uint = 60

// RegisterTagsHandlers registers HTTP handlers for the unified tag
// management page, spanning both the bookmark and feed domains.
//
// This bridge calls into taxonomyService, bookmarkQueryingService and
// feedQueryingService directly; it never calls the bookmark or feed
// controllers' own HTTP handlers, so pkg/bookmark and pkg/feed stay
// decoupled from each other.
func RegisterTagsHandlers(
	r *chi.Mux,
	taxonomyService *taxonomy.Service,
	bookmarkQueryingService *bookmarkquerying.Service,
	feedQueryingService *feedquerying.Service,
) {
	tc := tagsController{
		taxonomyService:         taxonomyService,
		bookmarkQueryingService: bookmarkQueryingService,
		feedQueryingService:     feedQueryingService,

		tagListView:   view.New("taxonomy/tag_list.gohtml"),
		tagEditView:   view.New("taxonomy/tag_edit.gohtml"),
		tagDeleteView: view.New("taxonomy/tag_delete.gohtml"),
	}

	r.Route("/tags", func(r chi.Router) {
		r.Use(func(h http.Handler) http.Handler {
			return middleware.AuthenticatedUser(h.ServeHTTP)
		})

		r.Get("/", tc.handleTagListView())

		r.Get("/{uuid}/edit", tc.handleTagEditView())
		r.Post("/{uuid}/edit", tc.handleTagEdit())

		r.Get("/{uuid}/delete", tc.handleTagDeleteView())
		r.Post("/{uuid}/delete", tc.handleTagDelete())
	})
}

type tagsController struct {
	taxonomyService         *taxonomy.Service
	bookmarkQueryingService *bookmarkquerying.Service
	feedQueryingService     *feedquerying.Service

	tagListView   *view.View
	tagEditView   *view.View
	tagDeleteView *view.View
}

// tagRow composes a taxonomy Tag with its per-domain usage counts, for
// display on the unified tag list page.
type tagRow struct {
	UUID string
	Name string

	BookmarkCount uint
	FeedCount     uint
}

// tagListPage holds a paginated set of tagRows.
type tagListPage struct {
	paginate.Page

	Tags []tagRow
}

// tagCounts fetches a user's bookmark and feed usage counts, grouped by tag
// name, in exactly two queries regardless of how many tags the user has.
func (tc *tagsController) tagCounts(ctx context.Context, userUUID string) (bookmarkCounts, feedCounts map[string]uint, err error) {
	bookmarkCounts, err = tc.bookmarkQueryingService.BookmarkCountsByTag(ctx, userUUID)
	if err != nil {
		return nil, nil, err
	}

	feedCounts, err = tc.feedQueryingService.SubscriptionCountsByTag(ctx, userUUID)
	if err != nil {
		return nil, nil, err
	}

	return bookmarkCounts, feedCounts, nil
}

// tagRows composes taxonomy Tags with their per-domain usage counts (looked
// up from maps built by tagCounts), and sorts the result by descending
// total usage (bookmark count + feed count), tags with equal counts ordered
// by name.
func tagRows(bookmarkCounts, feedCounts map[string]uint, tags []taxonomy.Tag) []tagRow {
	rows := make([]tagRow, len(tags))

	for i, tag := range tags {
		rows[i] = tagRow{
			UUID:          tag.UUID,
			Name:          tag.Name,
			BookmarkCount: bookmarkCounts[tag.Name],
			FeedCount:     feedCounts[tag.Name],
		}
	}

	sort.Slice(rows, func(i, j int) bool {
		iCount := rows[i].BookmarkCount + rows[i].FeedCount
		jCount := rows[j].BookmarkCount + rows[j].FeedCount
		if iCount != jCount {
			return iCount > jCount
		}
		return rows[i].Name < rows[j].Name
	})

	return rows
}

// filterTagsByName returns the tags whose name contains searchTerms,
// case-insensitively.
func filterTagsByName(tags []taxonomy.Tag, searchTerms string) []taxonomy.Tag {
	term := strings.ToLower(searchTerms)

	filtered := make([]taxonomy.Tag, 0, len(tags))
	for _, tag := range tags {
		if strings.Contains(strings.ToLower(tag.Name), term) {
			filtered = append(filtered, tag)
		}
	}

	return filtered
}

// handleTagListView renders the tag list for the current authenticated user,
// sorted by descending total usage (bookmark count + feed count).
//
// Sorting by usage requires knowing every tag's counts up front, so this
// composes and sorts the full set of the user's tags before paginating in
// memory, rather than paginating at the taxonomy layer (which only knows
// tag names, not cross-domain usage) and sorting each page in isolation.
//
// On an htmx request, it responds with only the list content fragment
// (search form, tags, pagination), so that searching or paginating swaps the
// list in place instead of reloading the full page. On a plain request, it
// renders the full page as usual.
func (tc *tagsController) handleTagListView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var viewData view.Data

		ctx := r.Context()
		ctxUser := httpcontext.UserValue(ctx)

		pageNumber, pageNumberStr, err := paginate.GetPageNumber(r.URL.Query())
		if err != nil {
			log.Warn().Err(err).Str("page_number", pageNumberStr).Msg("invalid page number")
			view.RedirectOnError(w, r, "/tags", fmt.Sprintf("invalid page number: %q", pageNumberStr))
			return
		}

		searchTermsParam := r.URL.Query().Get("search")

		tags, err := tc.taxonomyService.AllTags(ctx, ctxUser.UUID)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve tags")
			view.RedirectOnError(w, r, "/tags", "failed to retrieve tags")
			return
		}

		if searchTermsParam != "" {
			tags = filterTagsByName(tags, searchTermsParam)
		}

		bookmarkCounts, feedCounts, err := tc.tagCounts(ctx, ctxUser.UUID)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve tag usage counts")
			view.RedirectOnError(w, r, "/tags", "failed to retrieve tag usage counts")
			return
		}

		rows := tagRows(bookmarkCounts, feedCounts, tags)

		itemCount := uint(len(rows))
		totalPages := paginate.PageCount(itemCount, tagsPerPage)

		if pageNumber < 1 || pageNumber > totalPages {
			msg := fmt.Sprintf("invalid page number: %d", pageNumber)
			log.Error().Err(paginate.ErrPageNumberOutOfBounds).Msg(msg)
			view.RedirectOnError(w, r, "/tags", msg)
			return
		}

		offset := (pageNumber - 1) * tagsPerPage
		pageRows := []tagRow{}
		if offset < itemCount {
			end := min(offset+tagsPerPage, itemCount)
			pageRows = rows[offset:end]
		}

		page := paginate.NewPage(pageNumber, totalPages, tagsPerPage, itemCount)
		page.SearchTerms = searchTermsParam

		if searchTermsParam != "" {
			viewData.Title = fmt.Sprintf("Tag search: %s", searchTermsParam)
		} else {
			viewData.Title = "Tags"
		}
		viewData.Content = tagListPage{Page: page, Tags: pageRows}

		if r.Header.Get(htmx.HeaderRequest) == "true" {
			if err := tc.tagListView.RenderTemplate(w, "content", viewData.Content); err != nil {
				log.Error().Err(err).Msg("failed to render tag list fragment")
				http.Error(w, "Something went wrong", http.StatusInternalServerError)
			}
			return
		}

		tc.tagListView.Render(w, r, viewData)
	}
}

// handleTagEditView renders the tag edition form.
//
// On an htmx request, it responds with only the form fragment, meant to be
// loaded into the tag list page's edit modal. On a plain request, it renders
// the full page as usual, so the URL stays independently navigable.
func (tc *tagsController) handleTagEditView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		tagUUID := chi.URLParam(r, "uuid")

		ctx := r.Context()
		ctxUser := httpcontext.UserValue(ctx)

		tag, err := tc.taxonomyService.TagByUUID(ctx, ctxUser.UUID, tagUUID)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve tag")
			view.RedirectOnError(w, r, r.URL.Path, "failed to retrieve tag")
			return
		}

		if r.Header.Get(htmx.HeaderRequest) == "true" {
			formData := map[string]any{"Tag": tag, "InModal": true}
			if err := tc.tagEditView.RenderTemplate(w, "tagEditForm", formData); err != nil {
				log.Error().Err(err).Msg("failed to render tag edit form fragment")
				http.Error(w, "Something went wrong", http.StatusInternalServerError)
			}
			return
		}

		viewData := view.Data{
			Content: tag,
			Title:   fmt.Sprintf("Edit tag: %s", tag.Name),
		}

		tc.tagEditView.Render(w, r, viewData)
	}
}

// handleTagEdit processes the tag edition form.
//
// On success:
//   - htmx request: re-renders the tag's row and retargets/reswaps the
//     response into it (outerHTML), and fires a "modal:close" client-side
//     event so the tag list page's edit modal closes.
//   - plain request: flash + redirect to the tag list, as before.
//
// On error, it falls back to the same flash+redirect (or HX-Redirect, for
// htmx requests) behavior used throughout this file.
func (tc *tagsController) handleTagEdit() func(w http.ResponseWriter, r *http.Request) {
	type tagEditForm struct {
		Name string `schema:"name"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var form tagEditForm
		if err := decodeForm(r, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse tag edition form")
			view.RedirectOnError(w, r, r.URL.Path, "failed to process form")
			return
		}

		tagUUID := chi.URLParam(r, "uuid")

		ctx := r.Context()
		ctxUser := httpcontext.UserValue(ctx)

		tag, err := tc.taxonomyService.TagByUUID(ctx, ctxUser.UUID, tagUUID)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve tag")
			view.RedirectOnError(w, r, r.URL.Path, "failed to retrieve tag")
			return
		}

		tagNameUpdate := taxonomy.TagUpdateQuery{
			UserUUID:    ctxUser.UUID,
			CurrentName: tag.Name,
			NewName:     form.Name,
		}

		renamedTag, err := tc.taxonomyService.RenameTag(ctx, tagNameUpdate)
		if err != nil {
			log.Error().Err(err).Msg("failed to rename tag")
			view.RedirectOnError(w, r, r.URL.Path, "failed to rename tag")
			return
		}

		if r.Header.Get(htmx.HeaderRequest) == "true" {
			bookmarkCounts, feedCounts, err := tc.tagCounts(ctx, ctxUser.UUID)
			if err != nil {
				log.Error().Err(err).Msg("failed to retrieve tag usage counts")
				view.RedirectOnError(w, r, r.URL.Path, "failed to retrieve tag usage counts")
				return
			}

			rows := tagRows(bookmarkCounts, feedCounts, []taxonomy.Tag{renamedTag})

			w.Header().Set(htmx.HeaderRetarget, fmt.Sprintf("[id='tag-row-%s']", tagUUID))
			w.Header().Set(htmx.HeaderReswap, "outerHTML")
			w.Header().Set(htmx.HeaderTrigger, "modal:close")

			if err := tc.tagListView.RenderTemplate(w, "tagRow", rows[0]); err != nil {
				log.Error().Err(err).Msg("failed to render tag row fragment")
				http.Error(w, "Something went wrong", http.StatusInternalServerError)
			}
			return
		}

		view.PutFlashSuccess(w, "Tag renamed")
		http.Redirect(w, r, "/tags", http.StatusSeeOther)
	}
}

// handleTagDeleteView renders the tag deletion form.
//
// On an htmx request, it responds with only the form fragment, meant to be
// loaded into the tag list page's delete modal. On a plain request, it
// renders the full page as usual, so the URL stays independently navigable.
func (tc *tagsController) handleTagDeleteView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		tagUUID := chi.URLParam(r, "uuid")

		ctx := r.Context()
		ctxUser := httpcontext.UserValue(ctx)

		tag, err := tc.taxonomyService.TagByUUID(ctx, ctxUser.UUID, tagUUID)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve tag")
			view.RedirectOnError(w, r, r.URL.Path, "failed to retrieve tag")
			return
		}

		if r.Header.Get(htmx.HeaderRequest) == "true" {
			formData := map[string]any{"Tag": tag, "InModal": true}
			if err := tc.tagDeleteView.RenderTemplate(w, "tagDeleteForm", formData); err != nil {
				log.Error().Err(err).Msg("failed to render tag delete form fragment")
				http.Error(w, "Something went wrong", http.StatusInternalServerError)
			}
			return
		}

		viewData := view.Data{
			Content: tag,
			Title:   fmt.Sprintf("Delete tag: %s", tag.Name),
		}

		tc.tagDeleteView.Render(w, r, viewData)
	}
}

// handleTagDelete processes the tag deletion form.
//
// On success:
//   - htmx request: retargets/reswaps an empty response into the tag's row
//     (outerHTML), removing it, and fires a "modal:close" client-side event
//     so the tag list page's delete modal closes.
//   - plain request: flash + redirect to the tag list, as before.
//
// On error, it falls back to the same flash+redirect (or HX-Redirect, for
// htmx requests) behavior used throughout this file.
func (tc *tagsController) handleTagDelete() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		tagUUID := chi.URLParam(r, "uuid")

		ctx := r.Context()
		ctxUser := httpcontext.UserValue(ctx)

		tag, err := tc.taxonomyService.TagByUUID(ctx, ctxUser.UUID, tagUUID)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve tag")
			view.RedirectOnError(w, r, r.URL.Path, "failed to retrieve tag")
			return
		}

		tagDelete := taxonomy.TagDeleteQuery{
			UserUUID: ctxUser.UUID,
			Name:     tag.Name,
		}

		if err := tc.taxonomyService.DeleteTag(ctx, tagDelete); err != nil {
			log.Error().Err(err).Msg("failed to delete tag")
			view.RedirectOnError(w, r, r.URL.Path, "failed to delete tag")
			return
		}

		if r.Header.Get(htmx.HeaderRequest) == "true" {
			w.Header().Set(htmx.HeaderRetarget, fmt.Sprintf("[id='tag-row-%s']", tagUUID))
			w.Header().Set(htmx.HeaderReswap, "outerHTML")
			w.Header().Set(htmx.HeaderTrigger, "modal:close")
			return
		}

		view.PutFlashSuccess(w, "Tag deleted")
		http.Redirect(w, r, "/tags", http.StatusSeeOther)
	}
}
