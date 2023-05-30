package www

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"github.com/virtualtam/sparklemuffin/pkg/querying"
)

type tagHandlerContext struct {
	queryingService *querying.Service

	tagListView *view
}

func registerTagHandlers(
	r *mux.Router,
	queryingService *querying.Service,
) {
	tc := tagHandlerContext{
		queryingService: queryingService,
		tagListView:     newView("tag/list.gohtml"),
	}

	// bookmark tags
	tagRouter := r.PathPrefix("/tags").Subrouter()
	tagRouter.HandleFunc("", tc.handleTagListView()).Methods(http.MethodGet)

	tagRouter.Use(func(h http.Handler) http.Handler {
		return authenticatedUser(h.ServeHTTP)
	})
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

			viewData.Content = tagPage
		}

		tc.tagListView.render(w, r, viewData)
	}
}
