package www

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	"github.com/virtualtam/sparklemuffin/pkg/bookmark"
	"github.com/virtualtam/sparklemuffin/pkg/http/www/csrf"
	"github.com/virtualtam/sparklemuffin/pkg/querying"
	"github.com/virtualtam/sparklemuffin/pkg/user"
)

const (
	actionBookmarkAdd    string = "bookmark-add"
	actionBookmarkEdit   string = "bookmark-edit"
	actionBookmarkDelete string = "bookmark-delete"
)

type bookmarkHandlerContext struct {
	publicURL *url.URL

	bookmarkService *bookmark.Service
	csrfService     *csrf.Service
	queryingService *querying.Service
	userService     *user.Service

	bookmarkAddView    *view
	bookmarkDeleteView *view
	bookmarkEditView   *view
	bookmarkListView   *view

	publicBookmarkListView *view
}

type bookmarkFormContent struct {
	CSRFToken string
	Bookmark  *bookmark.Bookmark
	Tags      []string
}

func registerBookmarkHandlers(
	r *chi.Mux,
	publicURL *url.URL,
	bookmarkService *bookmark.Service,
	csrfService *csrf.Service,
	queryingService *querying.Service,
	userService *user.Service,
) {
	hc := bookmarkHandlerContext{
		publicURL: publicURL,

		bookmarkService: bookmarkService,
		csrfService:     csrfService,
		queryingService: queryingService,
		userService:     userService,

		bookmarkAddView:    newView("bookmark/add.gohtml"),
		bookmarkDeleteView: newView("bookmark/delete.gohtml"),
		bookmarkEditView:   newView("bookmark/edit.gohtml"),
		bookmarkListView:   newView("bookmark/list.gohtml"),

		publicBookmarkListView: newView("public/bookmark_list.gohtml"),
	}

	// bookmarks
	r.Route("/bookmarks", func(r chi.Router) {
		r.Use(func(h http.Handler) http.Handler {
			return authenticatedUser(h.ServeHTTP)
		})

		r.Get("/", hc.handleBookmarkListView())
		r.Get("/add", hc.handleBookmarkAddView())
		r.Post("/add", hc.handleBookmarkAdd())
		r.Get("/{uid}/delete", hc.handleBookmarkDeleteView())
		r.Post("/{uid}/delete", hc.handleBookmarkDelete())
		r.Get("/{uid}/edit", hc.handleBookmarkEditView())
		r.Post("/{uid}/edit", hc.handleBookmarkEdit())
	})

	// public bookmarks
	r.Route("/u/{nickname}", func(r chi.Router) {
		r.Get("/bookmarks", hc.handlePublicBookmarkListView())
		r.Get("/bookmarks/{uid}", hc.handlePublicBookmarkPermalinkView())
		r.Get("/feed/atom", hc.handlePublicBookmarkFeedAtom())
	})
}

// handleBookmarkAddView renders the bookmark addition form.
func (hc *bookmarkHandlerContext) handleBookmarkAddView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		user := userValue(r.Context())
		csrfToken := hc.csrfService.Generate(user.UUID, actionBookmarkAdd)

		tags, err := hc.queryingService.TagNamesByCount(user.UUID, querying.VisibilityAll)
		if err != nil {
			log.Error().Err(err).Str("user_uuid", user.UUID).Msg("failed to retrieve tags")
			PutFlashError(w, "failed to retrieve existing tags")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		viewData := Data{
			Content: bookmarkFormContent{
				CSRFToken: csrfToken,
				Tags:      tags,
			},
			Title: "Add bookmark",
		}
		hc.bookmarkAddView.render(w, r, viewData)
	}
}

// handleBookmarkAdd processes the bookmark addition form.
func (hc *bookmarkHandlerContext) handleBookmarkAdd() func(w http.ResponseWriter, r *http.Request) {
	type bookmarkAddForm struct {
		CSRFToken   string `schema:"csrf_token"`
		URL         string `schema:"url"`
		Title       string `schema:"title"`
		Description string `schema:"description"`
		Private     bool   `schema:"private"`
		Tags        string `schema:"tags"`
	}

	var form bookmarkAddForm

	return func(w http.ResponseWriter, r *http.Request) {
		ctxUser := userValue(r.Context())

		if err := decodeForm(r, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse bookmark creation form")
			PutFlashError(w, "There was an error processing the form")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		if !hc.csrfService.Validate(form.CSRFToken, ctxUser.UUID, actionBookmarkAdd) {
			log.Warn().Msg("failed to validate CSRF token")
			PutFlashError(w, "There was an error processing the form")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		newBookmark := bookmark.Bookmark{
			UserUUID:    ctxUser.UUID,
			URL:         form.URL,
			Title:       form.Title,
			Description: form.Description,
			Private:     form.Private,
			Tags:        strings.Split(form.Tags, " "),
		}

		if err := hc.bookmarkService.Add(newBookmark); err != nil {
			log.Error().Err(err).Msg("failed to add bookmark")
			PutFlashError(w, "failed to add bookmark")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/bookmarks", http.StatusSeeOther)
	}
}

// handleBookmarkDeleteView renders the bookmark deletion form.
func (hc *bookmarkHandlerContext) handleBookmarkDeleteView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		uid := chi.URLParam(r, "uid")
		user := userValue(r.Context())
		csrfToken := hc.csrfService.Generate(user.UUID, actionBookmarkDelete)

		bookmark, err := hc.bookmarkService.ByUID(user.UUID, uid)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve bookmark")
			PutFlashError(w, "failed to retrieve bookmark")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		viewData := Data{
			Content: FormContent{
				CSRFToken: csrfToken,
				Content:   bookmark,
			},
			Title: fmt.Sprintf("Delete bookmark: %s", bookmark.Title),
		}

		hc.bookmarkDeleteView.render(w, r, viewData)
	}
}

// handleBookmarkDelete processes the bookmark deletion form.
func (hc *bookmarkHandlerContext) handleBookmarkDelete() func(w http.ResponseWriter, r *http.Request) {
	type bookmarkDeleteForm struct {
		CSRFToken string `schema:"csrf_token"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		uid := chi.URLParam(r, "uid")
		ctxUser := userValue(r.Context())

		var form bookmarkDeleteForm
		if err := decodeForm(r, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse bookmark deletion form")
			PutFlashError(w, "There was an error processing the form")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		if !hc.csrfService.Validate(form.CSRFToken, ctxUser.UUID, actionBookmarkDelete) {
			log.Warn().Msg("failed to validate CSRF token")
			PutFlashError(w, "There was an error processing the form")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		if err := hc.bookmarkService.Delete(ctxUser.UUID, uid); err != nil {
			log.Error().Err(err).Msg("failed to delete bookmark")
			PutFlashError(w, "failed to delete bookmark")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/bookmarks", http.StatusSeeOther)
	}
}

// handleBookmarkEditView renders the bookmark edition form.
func (hc *bookmarkHandlerContext) handleBookmarkEditView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		uid := chi.URLParam(r, "uid")
		user := userValue(r.Context())
		csrfToken := hc.csrfService.Generate(user.UUID, actionBookmarkEdit)

		tags, err := hc.queryingService.TagNamesByCount(user.UUID, querying.VisibilityAll)
		if err != nil {
			log.Error().Err(err).Str("user_uuid", user.UUID).Msg("failed to retrieve tags")
			PutFlashError(w, "failed to retrieve existing tags")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		bookmark, err := hc.bookmarkService.ByUID(user.UUID, uid)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve bookmark")
			PutFlashError(w, "failed to retrieve bookmark")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		viewData := Data{
			Content: bookmarkFormContent{
				CSRFToken: csrfToken,
				Bookmark:  &bookmark,
				Tags:      tags,
			},
			Title: fmt.Sprintf("Edit bookmark: %s", bookmark.Title),
		}

		hc.bookmarkEditView.render(w, r, viewData)
	}
}

// handleBookmarkEdit processes the bookmark edition form.
func (hc *bookmarkHandlerContext) handleBookmarkEdit() func(w http.ResponseWriter, r *http.Request) {
	type bookmarkEditForm struct {
		CSRFToken   string `schema:"csrf_token"`
		URL         string `schema:"url"`
		Title       string `schema:"title"`
		Description string `schema:"description"`
		Private     bool   `schema:"private"`
		Tags        string `schema:"tags"`
	}

	var form bookmarkEditForm

	return func(w http.ResponseWriter, r *http.Request) {
		uid := chi.URLParam(r, "uid")
		ctxUser := userValue(r.Context())

		if err := decodeForm(r, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse bookmark edition form")
			PutFlashError(w, "failed to process form")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		if !hc.csrfService.Validate(form.CSRFToken, ctxUser.UUID, actionBookmarkEdit) {
			log.Warn().Msg("failed to validate CSRF token")
			PutFlashError(w, "There was an error processing the form")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		editedBookmark := bookmark.Bookmark{
			UserUUID:    ctxUser.UUID,
			UID:         uid,
			URL:         form.URL,
			Title:       form.Title,
			Description: form.Description,
			Private:     form.Private,
			Tags:        strings.Split(form.Tags, " "),
		}

		if err := hc.bookmarkService.Update(editedBookmark); err != nil {
			log.Error().Err(err).Msg("failed to edit bookmark")
			PutFlashError(w, "failed to edit bookmark")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/bookmarks", http.StatusSeeOther)
	}
}

// handleBookmarkListView renders the bookmark list for the current authenticated user.
func (hc *bookmarkHandlerContext) handleBookmarkListView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var viewData Data
		user := userValue(r.Context())

		pageNumberParam := r.URL.Query().Get("page")
		pageNumber, err := getPageNumber(pageNumberParam)
		if err != nil {
			log.Error().Err(err).Str("page_number", pageNumberParam).Msg("invalid page number")
			PutFlashError(w, fmt.Sprintf("invalid page number: %q", pageNumberParam))
			http.Redirect(w, r, "/bookmarks", http.StatusSeeOther)
			return
		}

		searchTermsParam := r.URL.Query().Get("search")
		if searchTermsParam != "" {
			bookmarksSearchPage, err := hc.queryingService.BookmarksBySearchQueryAndPage(
				user.UUID,
				querying.VisibilityAll,
				searchTermsParam,
				pageNumber,
			)
			if errors.Is(err, querying.ErrPageNumberOutOfBounds) {
				msg := fmt.Sprintf("invalid page number: %d", pageNumber)
				log.Error().Err(err).Msg(msg)
				PutFlashError(w, msg)
				http.Redirect(w, r, "/bookmarks", http.StatusSeeOther)
				return
			} else if err != nil {
				log.Error().Err(err).Msg("failed to retrieve bookmarks")
				PutFlashError(w, "failed to retrieve bookmarks")
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}

			viewData.Title = fmt.Sprintf("Bookmark search: %s", searchTermsParam)
			viewData.Content = bookmarksSearchPage

		} else {
			bookmarksPage, err := hc.queryingService.BookmarksByPage(
				user.UUID,
				querying.VisibilityAll,
				pageNumber,
			)
			if errors.Is(err, querying.ErrPageNumberOutOfBounds) {
				msg := fmt.Sprintf("invalid page number: %d", pageNumber)
				log.Error().Err(err).Msg(msg)
				PutFlashError(w, msg)
				http.Redirect(w, r, "/bookmarks", http.StatusSeeOther)
				return
			} else if err != nil {
				log.Error().Err(err).Msg("failed to retrieve bookmarks")
				PutFlashError(w, "failed to retrieve bookmarks")
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}

			viewData.Title = "Bookmarks"
			viewData.Content = bookmarksPage
		}

		hc.bookmarkListView.render(w, r, viewData)
	}
}

// handlePublicBookmarkListView renders the public bookmark list for a registered user.
func (hc *bookmarkHandlerContext) handlePublicBookmarkListView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var viewData Data

		nickName := chi.URLParam(r, "nickname")

		// Retrieve the owner UUID via user.Service to avoid duplicating the normalization/validation layer
		// in querying.Service.
		// In practice, this requires performing an extra database query.
		owner, err := hc.userService.ByNickName(nickName)
		if err != nil {
			log.Error().Err(err).Str("nickname", nickName).Msg("failed to retrieve user")
			PutFlashError(w, "unknown user")
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		pageNumberParam := r.URL.Query().Get("page")
		pageNumber, err := getPageNumber(pageNumberParam)
		if err != nil {
			log.Error().Err(err).Str("page_number", pageNumberParam).Msg("invalid page number")
			PutFlashError(w, fmt.Sprintf("invalid page number: %q", pageNumberParam))
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		var bookmarkPage querying.BookmarkPage

		searchTermsParam := r.URL.Query().Get("search")
		if searchTermsParam != "" {
			bookmarksSearchPage, err := hc.queryingService.PublicBookmarksBySearchQueryAndPage(
				owner.UUID,
				searchTermsParam,
				pageNumber,
			)
			if errors.Is(err, querying.ErrPageNumberOutOfBounds) {
				msg := fmt.Sprintf("invalid page number: %d", pageNumber)
				log.Error().Err(err).Msg(msg)
				PutFlashError(w, msg)
				http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
				return
			} else if err != nil {
				log.Error().Err(err).Msg("failed to retrieve bookmarks")
				PutFlashError(w, "failed to retrieve bookmarks")
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}

			bookmarkPage = bookmarksSearchPage
			viewData.Title = fmt.Sprintf("%s's bookmarks: %s", owner.DisplayName, searchTermsParam)

		} else {
			bookmarksPage, err := hc.queryingService.PublicBookmarksByPage(
				owner.UUID,
				pageNumber,
			)
			if errors.Is(err, querying.ErrPageNumberOutOfBounds) {
				msg := fmt.Sprintf("invalid page number: %d", pageNumber)
				log.Error().Err(err).Msg(msg)
				PutFlashError(w, msg)
				http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
				return
			} else if err != nil {
				log.Error().Err(err).Msg("failed to retrieve bookmarks")
				PutFlashError(w, "failed to retrieve bookmarks")
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}

			bookmarkPage = bookmarksPage
			viewData.Title = fmt.Sprintf("%s's bookmarks", owner.DisplayName)
		}

		viewData.AtomFeedURL = fmt.Sprintf("/u/%s/feed/atom", bookmarkPage.Owner.NickName)
		viewData.Content = bookmarkPage

		hc.publicBookmarkListView.render(w, r, viewData)
	}
}

// handlePublicBookmarkPermalinkView renders a given public bookmark for a registered user.
func (hc *bookmarkHandlerContext) handlePublicBookmarkPermalinkView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var viewData Data

		nickName := chi.URLParam(r, "nickname")
		uid := chi.URLParam(r, "uid")

		// Retrieve the owner UUID via user.Service to avoid duplicating the normalization/validation layer
		// in querying.Service.
		// In practice, this requires performing an extra database query.
		owner, err := hc.userService.ByNickName(nickName)
		if err != nil {
			log.Error().Err(err).Str("nickname", nickName).Msg("failed to retrieve user")
			PutFlashError(w, "unknown user")
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		bookmarkPage, err := hc.queryingService.PublicBookmarkByUID(owner.UUID, uid)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve bookmarks")
			PutFlashError(w, "failed to retrieve bookmarks")
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		viewData.AtomFeedURL = fmt.Sprintf("/u/%s/feed/atom", bookmarkPage.Owner.NickName)
		viewData.Title = fmt.Sprintf("%s's bookmarks", owner.DisplayName)
		viewData.Content = bookmarkPage

		hc.publicBookmarkListView.render(w, r, viewData)
	}
}

// handlePublicBookmarkFeedAtom renders the public Atom feed for a registered user.
func (hc *bookmarkHandlerContext) handlePublicBookmarkFeedAtom() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		nickName := chi.URLParam(r, "nickname")

		// Retrieve the owner UUID via user.Service to avoid duplicating the normalization/validation layer
		// in querying.Service.
		// In practice, this requires performing an extra database query.
		owner, err := hc.userService.ByNickName(nickName)
		if err != nil {
			log.Error().Err(err).Str("nickname", nickName).Msg("failed to retrieve user")
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		bookmarksPage, err := hc.queryingService.PublicBookmarksByPage(
			owner.UUID,
			1,
		)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve bookmarks")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		feed, err := bookmarksToFeed(hc.publicURL, bookmarksPage.Owner, bookmarksPage.Bookmarks)
		if err != nil {
			log.Error().Err(err).Msg("failed to create feed")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.atom", bookmarksPage.Owner.NickName))
		w.Header().Add("Content-Type", "application/atom+xml")

		if err := feed.WriteAtom(w); err != nil {
			log.Error().Err(err).Msg("failed to marshal Atom feed")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

func getPageNumber(pageNumberParam string) (uint, error) {
	if pageNumberParam == "" {
		return 1, nil
	}

	pageNumber64, err := strconv.ParseUint(pageNumberParam, 10, 64)
	if err != nil {
		return 0, err
	}

	return uint(pageNumber64), nil
}
