package controller

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	"github.com/virtualtam/sparklemuffin/cmd/sparklemuffin/http/www/httpcontext"
	"github.com/virtualtam/sparklemuffin/cmd/sparklemuffin/http/www/middleware"
	"github.com/virtualtam/sparklemuffin/cmd/sparklemuffin/http/www/view"
	"github.com/virtualtam/sparklemuffin/pkg/bookmark"
	"github.com/virtualtam/sparklemuffin/pkg/bookmark/querying"
)

type tagHandlerContext struct {
	bookmarkService *bookmark.Service
	queryingService *querying.Service

	tagDeleteView *view.View
	tagEditView   *view.View
	tagListView   *view.View
}

func RegisterBookmarkTagHandlers(
	r *chi.Mux,
	bookmarkService *bookmark.Service,
	queryingService *querying.Service,
) {
	tc := tagHandlerContext{
		bookmarkService: bookmarkService,
		queryingService: queryingService,

		tagDeleteView: view.New("tag/delete.gohtml"),
		tagEditView:   view.New("tag/edit.gohtml"),
		tagListView:   view.New("tag/list.gohtml"),
	}

	// bookmark tags
	r.Route("/tags", func(r chi.Router) {
		r.Use(func(h http.Handler) http.Handler {
			return middleware.AuthenticatedUser(h.ServeHTTP)
		})

		r.Get("/", tc.handleTagListView())
		r.Get("/{name}/delete", tc.handleTagDeleteView())
		r.Post("/{name}/delete", tc.handleTagDelete())
		r.Get("/{name}/edit", tc.handleTagEditView())
		r.Post("/{name}/edit", tc.handleTagEdit())
	})
}

// handleTagDeleteView renders the tag deletion form.
func (tc *tagHandlerContext) handleTagDeleteView() func(w http.ResponseWriter, r *http.Request) {
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
		tag := querying.NewTag(name, 0)

		viewData := view.Data{
			Content: tag,
			Title:   fmt.Sprintf("Delete tag: %s", name),
		}

		tc.tagDeleteView.Render(w, r, viewData)
	}
}

// handleTagDelete processes the tag deletion form.
func (tc *tagHandlerContext) handleTagDelete() func(w http.ResponseWriter, r *http.Request) {
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

		updated, err := tc.bookmarkService.DeleteTag(tagDelete)
		if err != nil {
			log.Error().Err(err).Msg("failed to delete tag")
			view.PutFlashError(w, "failed to delete tag")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		view.PutFlashSuccess(w, fmt.Sprintf("Tag deleted from %d bookmarks", updated))
		http.Redirect(w, r, "/tags", http.StatusSeeOther)
	}
}

// handleTagEditView renders the tag edition form.
func (tc *tagHandlerContext) handleTagEditView() func(w http.ResponseWriter, r *http.Request) {
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
		tag := querying.NewTag(name, 0)

		viewData := view.Data{
			Content: tag,
			Title:   fmt.Sprintf("Edit tag: %s", name),
		}

		tc.tagEditView.Render(w, r, viewData)
	}
}

// handleTagEdit processes the tag edition form.
func (tc *tagHandlerContext) handleTagEdit() func(w http.ResponseWriter, r *http.Request) {
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

		updated, err := tc.bookmarkService.UpdateTag(tagNameUpdate)
		if err != nil {
			log.Error().Err(err).Msg("failed to rename tag")
			view.PutFlashError(w, "failed to rename tag")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		view.PutFlashSuccess(w, fmt.Sprintf("Tag updated for %d bookmarks", updated))
		http.Redirect(w, r, "/tags", http.StatusSeeOther)
	}
}

// handleTagListView renders the tag list view for the current authenticated user.
func (tc *tagHandlerContext) handleTagListView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var viewData view.Data
		user := httpcontext.UserValue(r.Context())

		pageNumberParam := r.URL.Query().Get("page")
		pageNumber, err := getPageNumber(pageNumberParam)
		if err != nil {
			log.Error().Err(err).Str("page_number", pageNumberParam).Msg("invalid page number")
			view.PutFlashError(w, fmt.Sprintf("invalid page number: %q", pageNumberParam))
			http.Redirect(w, r, "/tags", http.StatusSeeOther)
			return
		}

		filterTermParam := r.URL.Query().Get("filter")

		if filterTermParam != "" {
			tagSearchPage, err := tc.queryingService.TagsByFilterQueryAndPage(
				user.UUID,
				querying.VisibilityAll,
				filterTermParam,
				pageNumber,
			)

			if errors.Is(err, querying.ErrPageNumberOutOfBounds) {
				msg := fmt.Sprintf("invalid page number: %d", pageNumber)
				log.Error().Err(err).Msg(msg)
				view.PutFlashError(w, msg)
				http.Redirect(w, r, "/tags", http.StatusSeeOther)
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
			tagPage, err := tc.queryingService.TagsByPage(
				user.UUID,
				querying.VisibilityAll,
				pageNumber,
			)

			if errors.Is(err, querying.ErrPageNumberOutOfBounds) {
				msg := fmt.Sprintf("invalid page number: %d", pageNumber)
				log.Error().Err(err).Msg(msg)
				view.PutFlashError(w, msg)
				http.Redirect(w, r, "/tags", http.StatusSeeOther)
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

		tc.tagListView.Render(w, r, viewData)
	}
}
