package www

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"github.com/virtualtam/sparklemuffin/pkg/bookmark"
	"github.com/virtualtam/sparklemuffin/pkg/querying"
)

type tagHandlerContext struct {
	bookmarkService *bookmark.Service
	queryingService *querying.Service

	tagDeleteView *view
	tagEditView   *view
	tagListView   *view
}

func registerTagHandlers(
	r *mux.Router,
	bookmarkService *bookmark.Service,
	queryingService *querying.Service,
) {
	tc := tagHandlerContext{
		bookmarkService: bookmarkService,
		queryingService: queryingService,

		tagDeleteView: newView("tag/delete.gohtml"),
		tagEditView:   newView("tag/edit.gohtml"),
		tagListView:   newView("tag/list.gohtml"),
	}

	// bookmark tags
	tagRouter := r.PathPrefix("/tags").Subrouter()
	tagRouter.HandleFunc("", tc.handleTagListView()).Methods(http.MethodGet)
	tagRouter.HandleFunc("/{name}/delete", tc.handleTagDeleteView()).Methods(http.MethodGet)
	tagRouter.HandleFunc("/{name}/delete", tc.handleTagDelete()).Methods(http.MethodPost)
	tagRouter.HandleFunc("/{name}/edit", tc.handleTagEditView()).Methods(http.MethodGet)
	tagRouter.HandleFunc("/{name}/edit", tc.handleTagEdit()).Methods(http.MethodPost)

	tagRouter.Use(func(h http.Handler) http.Handler {
		return authenticatedUser(h.ServeHTTP)
	})
}

// handleTagDeleteView renders the tag deletion form.
func (tc *tagHandlerContext) handleTagDeleteView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		nameBase64 := vars["name"]

		nameBytes, err := base64.URLEncoding.DecodeString(nameBase64)
		if err != nil {
			log.Error().Err(err).Msg("invalid tag")
			PutFlashError(w, "invalid tag")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		name := string(nameBytes)
		tag := querying.NewTag(name, 0)

		viewData := Data{
			Content: tag,
			Title:   fmt.Sprintf("Delete tag: %s", name),
		}

		tc.tagDeleteView.render(w, r, viewData)
	}
}

// handleTagDelete processes the tag deletion form.
func (tc *tagHandlerContext) handleTagDelete() func(w http.ResponseWriter, r *http.Request) {
	type tagDeleteForm struct {
		Name string `schema:"name"`
	}

	var form tagDeleteForm

	return func(w http.ResponseWriter, r *http.Request) {
		if err := parseForm(r, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse tag deletion form")
			PutFlashError(w, "failed to process form")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		vars := mux.Vars(r)
		nameBase64 := vars["name"]

		nameBytes, err := base64.URLEncoding.DecodeString(nameBase64)
		if err != nil {
			log.Error().Err(err).Msg("invalid tag")
			PutFlashError(w, "invalid tag")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		name := string(nameBytes)

		user := userValue(r.Context())

		tagDelete := bookmark.TagDeleteQuery{
			UserUUID: user.UUID,
			Name:     name,
		}

		updated, err := tc.bookmarkService.DeleteTag(tagDelete)
		if err != nil {
			log.Error().Err(err).Msg("failed to delete tag")
			PutFlashError(w, "failed to delete tag")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		PutFlashSuccess(w, fmt.Sprintf("Tag deleted from %d bookmarks", updated))
		http.Redirect(w, r, "/tags", http.StatusSeeOther)
	}
}

// handleTagEditView renders the tag edition form.
func (tc *tagHandlerContext) handleTagEditView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		nameBase64 := vars["name"]

		nameBytes, err := base64.URLEncoding.DecodeString(nameBase64)
		if err != nil {
			log.Error().Err(err).Msg("invalid tag")
			PutFlashError(w, "invalid tag")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		name := string(nameBytes)
		tag := querying.NewTag(name, 0)

		viewData := Data{
			Content: tag,
			Title:   fmt.Sprintf("Edit tag: %s", name),
		}

		tc.tagEditView.render(w, r, viewData)
	}
}

// handleTagEdit processes the tag edition form.
func (tc *tagHandlerContext) handleTagEdit() func(w http.ResponseWriter, r *http.Request) {
	type tagEditForm struct {
		Name string `schema:"name"`
	}

	var form tagEditForm

	return func(w http.ResponseWriter, r *http.Request) {
		if err := parseForm(r, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse tag edition form")
			PutFlashError(w, "failed to process form")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		vars := mux.Vars(r)
		nameBase64 := vars["name"]

		nameBytes, err := base64.URLEncoding.DecodeString(nameBase64)
		if err != nil {
			log.Error().Err(err).Msg("invalid tag")
			PutFlashError(w, "invalid tag")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		name := string(nameBytes)

		user := userValue(r.Context())

		tagNameUpdate := bookmark.TagUpdateQuery{
			UserUUID:    user.UUID,
			CurrentName: name,
			NewName:     form.Name,
		}

		updated, err := tc.bookmarkService.UpdateTag(tagNameUpdate)
		if err != nil {
			log.Error().Err(err).Msg("failed to rename tag")
			PutFlashError(w, "failed to rename tag")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		PutFlashSuccess(w, fmt.Sprintf("Tag updated for %d bookmarks", updated))
		http.Redirect(w, r, "/tags", http.StatusSeeOther)
	}
}

// handleTagListView renders the tag list view for the current authenticated user.
func (tc *tagHandlerContext) handleTagListView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var viewData Data
		user := userValue(r.Context())

		pageNumberParam := r.URL.Query().Get("page")
		pageNumber, err := getPageNumber(pageNumberParam)
		if err != nil {
			log.Error().Err(err).Str("page_number", pageNumberParam).Msg("invalid page number")
			PutFlashError(w, fmt.Sprintf("invalid page number: %q", pageNumberParam))
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
				PutFlashError(w, msg)
				http.Redirect(w, r, "/tags", http.StatusSeeOther)
				return
			} else if err != nil {
				log.Error().Err(err).Msg("failed to retrieve tags")
				PutFlashError(w, "failed to retrieve tags")
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
				PutFlashError(w, msg)
				http.Redirect(w, r, "/tags", http.StatusSeeOther)
				return
			} else if err != nil {
				log.Error().Err(err).Msg("failed to retrieve tags")
				PutFlashError(w, "failed to retrieve tags")
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}

			viewData.Title = "Tags"
			viewData.Content = tagPage
		}

		tc.tagListView.render(w, r, viewData)
	}
}
