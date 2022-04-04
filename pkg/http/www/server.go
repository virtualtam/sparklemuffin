package www

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"github.com/virtualtam/yawbe/pkg/http/www/rand"
	"github.com/virtualtam/yawbe/pkg/http/www/static"
	"github.com/virtualtam/yawbe/pkg/session"
	"github.com/virtualtam/yawbe/pkg/user"
)

var _ http.Handler = &Server{}

// Server represents the Web service.
type Server struct {
	router         *mux.Router
	sessionService *session.Service
	userService    *user.Service

	accountView *view

	adminView           *view
	adminUserAddView    *view
	adminUserDeleteView *view
	adminUserEditView   *view

	homeView      *view
	userLoginView *view
}

// NewServer initializes and returns a new Server.
func NewServer(sessionService *session.Service, userService *user.Service) *Server {
	s := &Server{
		router: mux.NewRouter(),

		sessionService: sessionService,
		userService:    userService,

		accountView: newView("account/account.gohtml"),

		adminView:           newView("admin/admin.gohtml"),
		adminUserAddView:    newView("admin/user_add.gohtml"),
		adminUserDeleteView: newView("admin/user_delete.gohtml"),
		adminUserEditView:   newView("admin/user_edit.gohtml"),

		homeView:      newView("static/home.gohtml"),
		userLoginView: newView("user/login.gohtml"),
	}

	s.addRoutes()

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

// handleAccountView displays the user account management page.
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
			http.Redirect(w, r, r.URL.Path, http.StatusBadRequest)
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
			http.Redirect(w, r, r.URL.Path, http.StatusBadRequest)
			return
		}

		s.PutFlashSuccess(w, "Your account information has been successfully updated")
		http.Redirect(w, r, "/account", http.StatusFound)
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
			http.Redirect(w, r, r.URL.Path, http.StatusBadRequest)
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
			http.Redirect(w, r, r.URL.Path, http.StatusInternalServerError)
			return
		}

		s.PutFlashSuccess(w, "Your account password has been successfully updated")
		http.Redirect(w, r, r.URL.Path, http.StatusFound)
	}
}

// handleAdmin displays the main administration page.
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
			http.Redirect(w, r, r.URL.Path, http.StatusBadRequest)
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
			http.Redirect(w, r, r.URL.Path, http.StatusInternalServerError)
			return
		}

		s.PutFlashSuccess(w, fmt.Sprintf("user %q has been successfully created", newUser.Email))
		http.Redirect(w, r, "/admin", http.StatusFound)
	}
}

// handleAdminUserDeleteView displays the user deletion form.
func (s *Server) handleAdminUserDeleteView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		userUUID := vars["uuid"]

		var viewData Data

		user, err := s.userService.ByUUID(userUUID)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve user")
			s.PutFlashError(w, err.Error())
			http.Redirect(w, r, r.URL.Path, http.StatusInternalServerError)
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
			http.Redirect(w, r, r.URL.Path, http.StatusInternalServerError)
			return
		}

		if err := s.userService.DeleteByUUID(userUUID); err != nil {
			log.Error().Err(err).Msg("failed to delete user")
			s.PutFlashError(w, err.Error())
			http.Redirect(w, r, r.URL.Path, http.StatusInternalServerError)
			return
		}

		s.PutFlashSuccess(w, fmt.Sprintf("user %q has been successfully deleted", user.Email))
		http.Redirect(w, r, "/admin", http.StatusFound)
	}
}

// handleAdminUserEditView displays the user edition form.
func (s *Server) handleAdminUserEditView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		userUUID := vars["uuid"]

		user, err := s.userService.ByUUID(userUUID)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve user")
			s.PutFlashError(w, err.Error())
			http.Redirect(w, r, r.URL.Path, http.StatusInternalServerError)
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
			http.Redirect(w, r, r.URL.Path, http.StatusBadRequest)
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
			http.Redirect(w, r, r.URL.Path, http.StatusInternalServerError)
			return
		}

		s.PutFlashSuccess(w, fmt.Sprintf("user %q has been successfully updated", editedUser.Email))
		http.Redirect(w, r, r.URL.Path, http.StatusFound)
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
			http.Redirect(w, r, r.URL.Path, http.StatusBadRequest)
			return
		}

		user, err := s.userService.Authenticate(form.Email, form.Password)
		if err != nil {
			log.Error().Err(err).Msg("failed to authenticate user")
			s.PutFlashError(w, "invalid email or password")
			http.Redirect(w, r, r.URL.Path, http.StatusInternalServerError)
			return
		}

		if err := s.setUserRememberToken(w, user.UUID); err != nil {
			log.Error().Err(err).Msg("failed to set remember token")
			s.PutFlashError(w, "failed to save session cookie")
			http.Redirect(w, r, r.URL.Path, http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusFound)
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
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		token, err := rand.RandomBase64URLString(UserRememberTokenNBytes)
		if err != nil {
			log.Error().Err(err).Msg("failed to generate a remember token")
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		userSession := session.Session{
			UserUUID:      user.UUID,
			RememberToken: token,
		}

		err = s.sessionService.Add(userSession)
		if err != nil {
			log.Error().Err(err).Msg("failed to save user session")
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		http.Redirect(w, r, "/", http.StatusFound)
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
