package www

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"github.com/virtualtam/netscape-go/v2"

	"github.com/virtualtam/yawbe/internal/rand"
	"github.com/virtualtam/yawbe/pkg/bookmark"
	"github.com/virtualtam/yawbe/pkg/exporting"
	"github.com/virtualtam/yawbe/pkg/http/www/static"
	"github.com/virtualtam/yawbe/pkg/importing"
	"github.com/virtualtam/yawbe/pkg/querying"
	"github.com/virtualtam/yawbe/pkg/session"
	"github.com/virtualtam/yawbe/pkg/user"
)

var _ http.Handler = &Server{}

// Server represents the Web service.
type Server struct {
	router *mux.Router

	bookmarkService  *bookmark.Service
	exportingService *exporting.Service
	importingService *importing.Service
	queryingService  *querying.Service
	sessionService   *session.Service
	userService      *user.Service

	accountView *view

	adminView           *view
	adminUserAddView    *view
	adminUserDeleteView *view
	adminUserEditView   *view

	bookmarkAddView    *view
	bookmarkDeleteView *view
	bookmarkEditView   *view
	bookmarkListView   *view

	publicBookmarkListView *view

	toolsView       *view
	toolsExportView *view
	toolsImportView *view

	homeView      *view
	userLoginView *view
}

// NewServer initializes and returns a new Server.
func NewServer() *Server {
	s := &Server{
		router: mux.NewRouter(),

		accountView: newView("account/account.gohtml"),

		adminView:           newView("admin/admin.gohtml"),
		adminUserAddView:    newView("admin/user_add.gohtml"),
		adminUserDeleteView: newView("admin/user_delete.gohtml"),
		adminUserEditView:   newView("admin/user_edit.gohtml"),

		bookmarkAddView:    newView("bookmark/add.gohtml"),
		bookmarkDeleteView: newView("bookmark/delete.gohtml"),
		bookmarkEditView:   newView("bookmark/edit.gohtml"),
		bookmarkListView:   newView("bookmark/list.gohtml"),

		publicBookmarkListView: newView("public/bookmark_list.gohtml"),

		toolsView:       newView("tools/tools.gohtml"),
		toolsExportView: newView("tools/export.gohtml"),
		toolsImportView: newView("tools/import.gohtml"),

		homeView:      newView("static/home.gohtml"),
		userLoginView: newView("user/login.gohtml"),
	}

	s.addRoutes()

	return s
}

func (s *Server) WithBookmarkService(bookmarkService *bookmark.Service) *Server {
	s.bookmarkService = bookmarkService
	return s
}

func (s *Server) WithExportingService(exportingService *exporting.Service) *Server {
	s.exportingService = exportingService
	return s
}

func (s *Server) WithImportingService(importingService *importing.Service) *Server {
	s.importingService = importingService
	return s
}

func (s *Server) WithQueryingService(queryingService *querying.Service) *Server {
	s.queryingService = queryingService
	return s
}

func (s *Server) WithSessionService(sessionService *session.Service) *Server {
	s.sessionService = sessionService
	return s
}

func (s *Server) WithUserService(userService *user.Service) *Server {
	s.userService = userService
	return s
}

// ServeHTTP satisfies the http.Handler interface,
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// addRoutes registers all HTTP handlers for the Web interface.
func (s *Server) addRoutes() {
	// static pages
	s.router.HandleFunc("/", s.homeView.handle)

	// user account
	accountRouter := s.router.PathPrefix("/account").Subrouter()

	accountRouter.HandleFunc("", s.handleAccountView()).Methods(http.MethodGet)
	accountRouter.HandleFunc("/info", s.handleAccountInfoUpdate()).Methods(http.MethodPost)
	accountRouter.HandleFunc("/password", s.handleAccountPasswordUpdate()).Methods(http.MethodPost)

	accountRouter.Use(func(h http.Handler) http.Handler {
		return s.authenticatedUser(h.ServeHTTP)
	})

	// administration
	adminRouter := s.router.PathPrefix("/admin").Subrouter()

	adminRouter.HandleFunc("", s.handleAdmin()).Methods(http.MethodGet)
	adminRouter.HandleFunc("/users/add", s.adminUserAddView.handle).Methods(http.MethodGet)
	adminRouter.HandleFunc("/users", s.handleAdminUserAdd()).Methods(http.MethodPost)
	adminRouter.HandleFunc("/users/{uuid}", s.handleAdminUserEditView()).Methods(http.MethodGet)
	adminRouter.HandleFunc("/users/{uuid}", s.handleAdminUserEdit()).Methods(http.MethodPost)
	adminRouter.HandleFunc("/users/{uuid}/delete", s.handleAdminUserDeleteView()).Methods(http.MethodGet)
	adminRouter.HandleFunc("/users/{uuid}/delete", s.handleAdminUserDelete()).Methods(http.MethodPost)

	adminRouter.Use(func(h http.Handler) http.Handler {
		return s.adminUser(h.ServeHTTP)
	})

	// authentication
	s.router.HandleFunc("/login", s.userLoginView.handle).Methods(http.MethodGet)
	s.router.HandleFunc("/login", s.handleUserLogin()).Methods(http.MethodPost)
	s.router.HandleFunc("/logout", s.handleUserLogout()).Methods(http.MethodPost)

	// bookmarks
	bookmarkRouter := s.router.PathPrefix("/bookmarks").Subrouter()

	bookmarkRouter.HandleFunc("", s.handleBookmarkListView()).Methods(http.MethodGet)
	bookmarkRouter.HandleFunc("/add", s.handleBookmarkAddView()).Methods(http.MethodGet)
	bookmarkRouter.HandleFunc("/add", s.handleBookmarkAdd()).Methods(http.MethodPost)
	bookmarkRouter.HandleFunc("/{uid}/delete", s.handleBookmarkDeleteView()).Methods(http.MethodGet)
	bookmarkRouter.HandleFunc("/{uid}/delete", s.handleBookmarkDelete()).Methods(http.MethodPost)
	bookmarkRouter.HandleFunc("/{uid}/edit", s.handleBookmarkEditView()).Methods(http.MethodGet)
	bookmarkRouter.HandleFunc("/{uid}/edit", s.handleBookmarkEdit()).Methods(http.MethodPost)

	bookmarkRouter.Use(func(h http.Handler) http.Handler {
		return s.authenticatedUser(h.ServeHTTP)
	})

	// bookmark management tools
	toolsRouter := s.router.PathPrefix("/tools").Subrouter()

	toolsRouter.HandleFunc("", s.toolsView.handle).Methods(http.MethodGet)
	toolsRouter.HandleFunc("/export", s.toolsExportView.handle).Methods(http.MethodGet)
	toolsRouter.HandleFunc("/export", s.handleToolsExport()).Methods(http.MethodPost)
	toolsRouter.HandleFunc("/import", s.toolsImportView.handle).Methods(http.MethodGet)
	toolsRouter.HandleFunc("/import", s.handleToolsImport()).Methods(http.MethodPost)

	toolsRouter.Use(func(h http.Handler) http.Handler {
		return s.authenticatedUser(h.ServeHTTP)
	})

	// public bookmarks
	publicBookmarkRouter := s.router.PathPrefix("/u/{nickname}").Subrouter()
	publicBookmarkRouter.HandleFunc("/bookmarks", s.handlePublicBookmarkListView()).Methods(http.MethodGet)

	// static assets
	s.router.HandleFunc("/static/", http.NotFound)
	s.router.PathPrefix("/static/").Handler(http.StripPrefix(
		"/static/",
		s.staticCacheControl(http.FileServer(http.FS(static.FS)))))

	// global middleware
	s.router.Use(func(h http.Handler) http.Handler {
		return s.rememberUser(h.ServeHTTP)
	})
}

// handleAccountView renders the user account management page.
func (s *Server) handleAccountView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		user := userValue(r.Context())
		viewData := Data{
			Content: user,
		}

		s.accountView.render(w, r, viewData)
	}
}

// handleAccountInfoUpdate processes the account information update form.
func (s *Server) handleAccountInfoUpdate() func(w http.ResponseWriter, r *http.Request) {
	type infoUpdateForm struct {
		Email       string `schema:"email"`
		NickName    string `schema:"nick_name"`
		DisplayName string `schema:"display_name"`
	}

	var form infoUpdateForm

	return func(w http.ResponseWriter, r *http.Request) {
		ctxUser := userValue(r.Context())

		if err := parseForm(r, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse account information update form")
			s.PutFlashError(w, "There was an error processing the form")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		userInfo := user.InfoUpdate{
			UUID:        ctxUser.UUID,
			Email:       form.Email,
			NickName:    form.NickName,
			DisplayName: form.DisplayName,
		}

		if err := s.userService.UpdateInfo(userInfo); err != nil {
			log.Error().Err(err).Msg("failed to update account information")
			s.PutFlashError(w, "There was an error updating your information")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		s.PutFlashSuccess(w, "Your account information has been successfully updated")
		http.Redirect(w, r, "/account", http.StatusSeeOther)
	}
}

// handleAccountPasswordUpdate processes the user account password update form.
func (s *Server) handleAccountPasswordUpdate() func(w http.ResponseWriter, r *http.Request) {
	type passwordUpdateForm struct {
		CurrentPassword         string `schema:"current_password"`
		NewPassword             string `schema:"new_password"`
		NewPasswordConfirmation string `schema:"new_password_confirmation"`
	}

	var form passwordUpdateForm

	return func(w http.ResponseWriter, r *http.Request) {
		ctxUser := userValue(r.Context())

		if err := parseForm(r, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse account password update form")
			s.PutFlashError(w, "There was an error processing the form")
			http.Redirect(w, r, "/account", http.StatusSeeOther)
			return
		}

		userPassword := user.PasswordUpdate{
			UUID:                    ctxUser.UUID,
			CurrentPassword:         form.CurrentPassword,
			NewPassword:             form.NewPassword,
			NewPasswordConfirmation: form.NewPasswordConfirmation,
		}

		if err := s.userService.UpdatePassword(userPassword); err != nil {
			log.Error().Err(err).Msg("failed to update account password")
			s.PutFlashError(w, fmt.Sprintf("There was an error updating your password: %s", err))
			http.Redirect(w, r, "/account", http.StatusSeeOther)
			return
		}

		s.PutFlashSuccess(w, "Your account password has been successfully updated")
		http.Redirect(w, r, "/account", http.StatusSeeOther)
	}
}

// handleAdmin renders the main administration page.
func (s *Server) handleAdmin() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var viewData Data

		users, err := s.userService.All()
		if err != nil {
			s.PutFlashError(w, err.Error())
		} else {
			viewData.Content = users
		}

		s.adminView.render(w, r, viewData)
	}
}

// handleAdminUserAdd processes data submitted through the user creation form.
func (s *Server) handleAdminUserAdd() func(w http.ResponseWriter, r *http.Request) {
	type userAddForm struct {
		Email       string `schema:"email"`
		NickName    string `schema:"nick_name"`
		DisplayName string `schema:"display_name"`
		Password    string `schema:"password"`
		IsAdmin     bool   `schema:"is_admin"`
	}

	var form userAddForm

	return func(w http.ResponseWriter, r *http.Request) {
		if err := parseForm(r, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse user creation form")
			s.PutFlashError(w, err.Error())
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		newUser := user.User{
			Email:       form.Email,
			NickName:    form.NickName,
			DisplayName: form.DisplayName,
			Password:    form.Password,
			IsAdmin:     form.IsAdmin,
		}

		if err := s.userService.Add(newUser); err != nil {
			log.Error().Err(err).Msg("failed to persist user")
			s.PutFlashError(w, err.Error())
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		s.PutFlashSuccess(w, fmt.Sprintf("user %q has been successfully created", newUser.Email))
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	}
}

// handleAdminUserDeleteView renders the user deletion form.
func (s *Server) handleAdminUserDeleteView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		userUUID := vars["uuid"]

		var viewData Data

		user, err := s.userService.ByUUID(userUUID)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve user")
			s.PutFlashError(w, err.Error())
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		viewData.Content = user

		s.adminUserDeleteView.render(w, r, viewData)
	}
}

// handleAdminUserDelete processes the user deletion form.
func (s *Server) handleAdminUserDelete() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		userUUID := vars["uuid"]

		user, err := s.userService.ByUUID(userUUID)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve user")
			s.PutFlashError(w, err.Error())
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		if err := s.userService.DeleteByUUID(userUUID); err != nil {
			log.Error().Err(err).Msg("failed to delete user")
			s.PutFlashError(w, err.Error())
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		s.PutFlashSuccess(w, fmt.Sprintf("user %q has been successfully deleted", user.Email))
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	}
}

// handleAdminUserEditView renders the user edition form.
func (s *Server) handleAdminUserEditView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		userUUID := vars["uuid"]

		user, err := s.userService.ByUUID(userUUID)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve user")
			s.PutFlashError(w, err.Error())
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		viewData := Data{
			Content: user,
		}
		s.adminUserEditView.render(w, r, viewData)
	}
}

// handleAdminUserEdit processes the user edition form.
func (s *Server) handleAdminUserEdit() func(w http.ResponseWriter, r *http.Request) {
	type userEditForm struct {
		Email       string `schema:"email"`
		NickName    string `schema:"nick_name"`
		DisplayName string `schema:"display_name"`
		Password    string `schema:"password"`
		IsAdmin     bool   `schema:"is_admin"`
	}

	var form userEditForm

	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		userUUID := vars["uuid"]

		if err := parseForm(r, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse user edition form")
			s.PutFlashError(w, err.Error())
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		editedUser := user.User{
			UUID:        userUUID,
			Email:       form.Email,
			NickName:    form.NickName,
			DisplayName: form.DisplayName,
			Password:    form.Password,
			IsAdmin:     form.IsAdmin,
		}

		if err := s.userService.Update(editedUser); err != nil {
			log.Error().Err(err).Msg("failed to update user")
			s.PutFlashError(w, err.Error())
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		s.PutFlashSuccess(w, fmt.Sprintf("user %q has been successfully updated", editedUser.Email))
		http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
	}
}

// handleBookmarkAddView renders the bookmark addition form.
func (s *Server) handleBookmarkAddView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		s.bookmarkAddView.render(w, r, nil)
	}
}

// handleBookmarkAdd processes the bookmark addition form.
func (s *Server) handleBookmarkAdd() func(w http.ResponseWriter, r *http.Request) {
	type bookmarkAddForm struct {
		URL         string `schema:"url"`
		Title       string `schema:"title"`
		Description string `schema:"description"`
		Private     bool   `schema:"private"`
		Tags        string `schema:"tags"`
	}

	var form bookmarkAddForm

	return func(w http.ResponseWriter, r *http.Request) {
		if err := parseForm(r, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse bookmark creation form")
			s.PutFlashError(w, "failed to process form")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		user := userValue(r.Context())

		newBookmark := bookmark.Bookmark{
			UserUUID:    user.UUID,
			URL:         form.URL,
			Title:       form.Title,
			Description: form.Description,
			Private:     form.Private,
			Tags:        strings.Split(form.Tags, " "),
		}

		if err := s.bookmarkService.Add(newBookmark); err != nil {
			log.Error().Err(err).Msg("failed to add bookmark")
			s.PutFlashError(w, "failed to add bookmark")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/bookmarks", http.StatusSeeOther)
	}
}

// handleBookmarkDeleteView renders the bookmark deletion form.
func (s *Server) handleBookmarkDeleteView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		uid := vars["uid"]

		user := userValue(r.Context())

		bookmark, err := s.bookmarkService.ByUID(user.UUID, uid)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve bookmark")
			s.PutFlashError(w, "failed to retrieve bookmark")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		viewData := Data{
			Content: bookmark,
		}

		s.bookmarkDeleteView.render(w, r, viewData)
	}
}

// handleBookmarkDelete processes the bookmark deletion form.
func (s *Server) handleBookmarkDelete() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		uid := vars["uid"]

		user := userValue(r.Context())

		if err := s.bookmarkService.Delete(user.UUID, uid); err != nil {
			log.Error().Err(err).Msg("failed to delete bookmark")
			s.PutFlashError(w, "failed to delete bookmark")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/bookmarks", http.StatusSeeOther)
	}
}

// handleBookmarkEditView renders the bookmark edition form.
func (s *Server) handleBookmarkEditView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		uid := vars["uid"]

		user := userValue(r.Context())

		bookmark, err := s.bookmarkService.ByUID(user.UUID, uid)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve bookmark")
			s.PutFlashError(w, "failed to retrieve bookmark")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		viewData := Data{
			Content: bookmark,
		}

		s.bookmarkEditView.render(w, r, viewData)
	}
}

// handleBookmarkEdit processes the bookmark edition form.
func (s *Server) handleBookmarkEdit() func(w http.ResponseWriter, r *http.Request) {
	type bookmarkEditForm struct {
		URL         string `schema:"url"`
		Title       string `schema:"title"`
		Description string `schema:"description"`
		Private     bool   `schema:"private"`
		Tags        string `schema:"tags"`
	}

	var form bookmarkEditForm

	return func(w http.ResponseWriter, r *http.Request) {
		if err := parseForm(r, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse bookmark edition form")
			s.PutFlashError(w, "failed to process form")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		vars := mux.Vars(r)
		uid := vars["uid"]

		user := userValue(r.Context())

		editedBookmark := bookmark.Bookmark{
			UserUUID:    user.UUID,
			UID:         uid,
			URL:         form.URL,
			Title:       form.Title,
			Description: form.Description,
			Private:     form.Private,
			Tags:        strings.Split(form.Tags, " "),
		}

		if err := s.bookmarkService.Update(editedBookmark); err != nil {
			log.Error().Err(err).Msg("failed to edit bookmark")
			s.PutFlashError(w, "failed to edit bookmark")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/bookmarks", http.StatusSeeOther)
	}
}

// handleBookmarkListView renders the bookmark list for the current authenticated user.
func (s *Server) handleBookmarkListView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var viewData Data
		user := userValue(r.Context())

		pageNumberParam := r.URL.Query().Get("page")
		pageNumber, err := strconv.Atoi(pageNumberParam)
		if err != nil {
			pageNumber = 1
		}

		searchTermsParam := r.URL.Query().Get("search")
		if searchTermsParam != "" {
			bookmarksSearchPage, err := s.queryingService.BySearchQueryAndPage(user.UUID, searchTermsParam, pageNumber)
			if errors.Is(err, querying.ErrPageNumberOutOfBounds) {
				msg := fmt.Sprintf("invalid page number: %d", pageNumber)
				log.Error().Err(err).Msg(msg)
				s.PutFlashError(w, msg)
				http.Redirect(w, r, "/bookmarks", http.StatusSeeOther)
				return
			} else if err != nil {
				log.Error().Err(err).Msg("failed to retrieve bookmarks")
				s.PutFlashError(w, "failed to retrieve bookmarks")
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}

			viewData.Content = bookmarksSearchPage

		} else {
			bookmarksPage, err := s.queryingService.ByPage(user.UUID, pageNumber)
			if errors.Is(err, querying.ErrPageNumberOutOfBounds) {
				msg := fmt.Sprintf("invalid page number: %d", pageNumber)
				log.Error().Err(err).Msg(msg)
				s.PutFlashError(w, msg)
				http.Redirect(w, r, "/bookmarks", http.StatusSeeOther)
				return
			} else if err != nil {
				log.Error().Err(err).Msg("failed to retrieve bookmarks")
				s.PutFlashError(w, "failed to retrieve bookmarks")
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}

			viewData.Content = bookmarksPage
		}

		s.bookmarkListView.render(w, r, viewData)
	}
}

// handlePublicBookmarkListView renders the public bookmark list for a registered user.
func (s *Server) handlePublicBookmarkListView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var viewData Data

		vars := mux.Vars(r)
		nickName := vars["nickname"]

		owner, err := s.userService.ByNickName(nickName)
		if err != nil {
			log.Error().Err(err).Str("nickname", nickName).Msg("failed to retrieve user")
			s.PutFlashError(w, "unknown user")
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		// TODO public bookmarks only!

		pageNumberParam := r.URL.Query().Get("page")
		pageNumber, err := strconv.Atoi(pageNumberParam)
		if err != nil {
			pageNumber = 1
		}

		var bookmarkPage querying.Page

		searchTermsParam := r.URL.Query().Get("search")
		if searchTermsParam != "" {
			bookmarksSearchPage, err := s.queryingService.BySearchQueryAndPage(owner.UUID, searchTermsParam, pageNumber)
			if errors.Is(err, querying.ErrPageNumberOutOfBounds) {
				msg := fmt.Sprintf("invalid page number: %d", pageNumber)
				log.Error().Err(err).Msg(msg)
				s.PutFlashError(w, msg)
				http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
				return
			} else if err != nil {
				log.Error().Err(err).Msg("failed to retrieve bookmarks")
				s.PutFlashError(w, "failed to retrieve bookmarks")
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}

			bookmarkPage = bookmarksSearchPage

		} else {
			bookmarksPage, err := s.queryingService.ByPage(owner.UUID, pageNumber)
			if errors.Is(err, querying.ErrPageNumberOutOfBounds) {
				msg := fmt.Sprintf("invalid page number: %d", pageNumber)
				log.Error().Err(err).Msg(msg)
				s.PutFlashError(w, msg)
				http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
				return
			} else if err != nil {
				log.Error().Err(err).Msg("failed to retrieve bookmarks")
				s.PutFlashError(w, "failed to retrieve bookmarks")
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}

			bookmarkPage = bookmarksPage
		}

		// TODO: rework with a proper struct
		viewData.Content = struct {
			Owner user.User
			Page  querying.Page
		}{
			Owner: owner,
			Page:  bookmarkPage,
		}

		s.publicBookmarkListView.render(w, r, viewData)
	}
}

// handleToolsExport processes the bookmarks export form and sends the
// corresponding file to the client.
func (s *Server) handleToolsExport() func(w http.ResponseWriter, r *http.Request) {
	type exportForm struct {
		Visibility exporting.Visibility `schema:"visibility"`
	}

	var form exportForm

	return func(w http.ResponseWriter, r *http.Request) {
		if err := parseForm(r, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse bookmark export form")
			s.PutFlashError(w, "failed to process form")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		user := userValue(r.Context())

		document, err := s.exportingService.ExportAsNetscapeDocument(user.UUID, form.Visibility)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve bookmarks")
			s.PutFlashError(w, "failed to export bookmarks")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		marshaled, err := netscape.Marshal(document)
		if err != nil {
			log.Error().Err(err).Msg("failed to marshal Netscape bookmarks")
			s.PutFlashError(w, "failed to export bookmarks")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		filename := fmt.Sprintf("bookmarks-%s.htm", form.Visibility)

		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
		w.Header().Set("Content-Type", "application/octet-stream")
		_, err = w.Write(marshaled)

		if err != nil {
			log.Error().Err(err).Msg("failed to send marshaled Netscape bookmark export")
		}
	}
}

// handleToolsImport processes data submitted through the bookmark import form.
func (s *Server) handleToolsImport() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		multipartReader, err := r.MultipartReader()
		if err != nil {
			log.Error().Err(err).Msg("failed to access multipart reader")
			s.PutFlashError(w, "failed to process import form")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		var (
			importFileBuffer         bytes.Buffer
			onConflictStrategyBuffer bytes.Buffer
			visibilityBuffer         bytes.Buffer
		)
		importFileWriter := bufio.NewWriter(&importFileBuffer)
		onConflictStrategyWriter := bufio.NewWriter(&onConflictStrategyBuffer)
		visibilityWriter := bufio.NewWriter(&visibilityBuffer)

		for {
			part, err := multipartReader.NextPart()

			if errors.Is(err, io.EOF) {
				// no more parts to process
				break
			}

			if err != nil {
				log.Error().Err(err).Msg("failed to access multipart form data")
				s.PutFlashError(w, "failed to process import form")
				http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
				return
			}

			switch part.FormName() {
			case "importfile":
				_, err = io.Copy(importFileWriter, part)
			case "on-conflict":
				_, err = io.Copy(onConflictStrategyWriter, part)
			case "visibility":
				_, err = io.Copy(visibilityWriter, part)
			default:
				err = fmt.Errorf("unexpected multipart form field: %q", part.FormName())
			}

			if err != nil {
				log.Error().Err(err).Msg(fmt.Sprintf("failed to process multipart form part %q", part.FormName()))
				s.PutFlashError(w, "failed to process import form")
				http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
				return
			}
		}

		document, err := netscape.Unmarshal(importFileBuffer.Bytes())
		if err != nil {
			log.Error().Err(err).Msg("failed to process Netscape bookmark file")
			s.PutFlashError(w, "failed to import bookmarks from the uploaded file")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		user := userValue(r.Context())
		overwrite := importing.OnConflictStrategy(onConflictStrategyBuffer.String())
		visibility := importing.Visibility(visibilityBuffer.String())

		importStatus, err := s.importingService.ImportFromNetscapeDocument(user.UUID, document, visibility, overwrite)
		if err != nil {
			log.Error().Err(err).Msg("failed to save imported bookmarks")
			s.PutFlashError(w, "failed to save imported bookmarks")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		s.PutFlashSuccess(w, fmt.Sprintf("Import status: %s", importStatus.Summary()))
		http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
	}
}

// handleUserLogin processes data submitted through the user login form.
func (s *Server) handleUserLogin() func(w http.ResponseWriter, r *http.Request) {
	type loginForm struct {
		Email    string `schema:"email"`
		Password string `schema:"password"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var form loginForm

		if err := parseForm(r, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse login form")
			s.PutFlashError(w, err.Error())
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		user, err := s.userService.Authenticate(form.Email, form.Password)
		if err != nil {
			log.Error().Err(err).Msg("failed to authenticate user")
			s.PutFlashError(w, "invalid email or password")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		if err := s.setUserRememberToken(w, user.UUID); err != nil {
			log.Error().Err(err).Msg("failed to set remember token")
			s.PutFlashError(w, "failed to save session cookie")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/bookmarks", http.StatusSeeOther)
	}
}

// handleUserLogout logs a user out and clears their session data.
func (s *Server) handleUserLogout() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie := http.Cookie{
			Name:     UserRememberTokenCookieName,
			Value:    "",
			Path:     "/",
			Expires:  time.Unix(0, 1),
			HttpOnly: true,
		}
		http.SetCookie(w, &cookie)

		user := userValue(r.Context())
		if user == nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		token, err := rand.RandomBase64URLString(UserRememberTokenNBytes)
		if err != nil {
			log.Error().Err(err).Msg("failed to generate a remember token")
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		userSession := session.Session{
			UserUUID:      user.UUID,
			RememberToken: token,
		}

		err = s.sessionService.Add(userSession)
		if err != nil {
			log.Error().Err(err).Msg("failed to save user session")
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

const (
	UserRememberTokenNBytes     int    = 32
	UserRememberTokenCookieName string = "remember_me"
)

// setUserRememberToken creates and persists a new RememberToken if needed, and
// sets it as a session cookie.
func (s *Server) setUserRememberToken(w http.ResponseWriter, userUUID string) error {
	token, err := rand.RandomBase64URLString(UserRememberTokenNBytes)
	if err != nil {
		log.Error().Err(err).Msg("failed to generate a remember token")
		return err
	}

	// expires after one month
	expiresAt := time.Now().UTC().AddDate(0, 1, 0)

	userSession := session.Session{
		UserUUID:               userUUID,
		RememberToken:          token,
		RememberTokenExpiresAt: expiresAt,
	}

	err = s.sessionService.Add(userSession)
	if err != nil {
		log.Error().Err(err).Msg("failed to update user")
		return err
	}

	cookie := http.Cookie{
		Name:     UserRememberTokenCookieName,
		Value:    userSession.RememberToken,
		Expires:  expiresAt,
		Path:     "/",
		HttpOnly: true,
	}

	http.SetCookie(w, &cookie)

	return nil
}

// staticCacheControl sets the Cache-Control header for static assets.
func (s *Server) staticCacheControl(h http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "max-age=2592000") // 30 days
		h.ServeHTTP(w, r)
	})
}

// rememberUser enriches the request context with a user.User if a valid
// remember token cookie is set.
func (s *Server) rememberUser(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(UserRememberTokenCookieName)
		if err != nil {
			h(w, r)
			return
		}

		session, err := s.sessionService.ByRememberToken(cookie.Value)
		if err != nil {
			h(w, r)
			return
		}

		user, err := s.userService.ByUUID(session.UserUUID)
		if err != nil {
			h(w, r)
			return
		}

		ctx := r.Context()
		ctx = withUser(ctx, user)
		r = r.WithContext(ctx)

		h(w, r)
	})
}

// authenticatedUser requires the user to be authenticated.
func (s *Server) authenticatedUser(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := userValue(r.Context())

		if user == nil {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		h(w, r)
	})
}

// adminUser requires the user to have administration privileges to
// access content.
func (s *Server) adminUser(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := userValue(r.Context())

		if user == nil {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		if !user.IsAdmin {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		h(w, r)
	})
}

// putFlash sets a session flash message to be displayed by the next rendered
// view.
func (s *Server) putFlash(w http.ResponseWriter, level flashLevel, message string) {
	flash := Flash{
		Level:   level,
		Message: message,
	}

	encoded, err := flash.base64URLEncode()
	if err != nil {
		log.Error().Err(err).Msg("failed to put flash cookie")
		return
	}

	cookie := &http.Cookie{
		Name:     flashCookieName,
		Path:     "/",
		Value:    encoded,
		HttpOnly: true,
	}

	http.SetCookie(w, cookie)
}

// PutFlashError sets a Flash that will be rendered as an error message.
func (s *Server) PutFlashError(w http.ResponseWriter, message string) {
	s.putFlash(w, flashLevelError, fmt.Sprintf("Error: %s", message))
}

// PutFlashInfo sets a Flash that will be rendered as an information message.
func (s *Server) PutFlashInfo(w http.ResponseWriter, message string) {
	s.putFlash(w, flashLevelInfo, message)
}

// PutFlashSuccess sets a Flash that will be rendered as a success message.
func (s *Server) PutFlashSuccess(w http.ResponseWriter, message string) {
	s.putFlash(w, flashLevelSuccess, message)
}

// PutFlashWarning sets a Flash that will be rendered as a warning message.
func (s *Server) PutFlashWarning(w http.ResponseWriter, message string) {
	s.putFlash(w, flashLevelWarning, fmt.Sprintf("Warning: %s", message))
}
