package www

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"github.com/virtualtam/yawbe/pkg/http/www/rand"
	"github.com/virtualtam/yawbe/pkg/http/www/static"
	"github.com/virtualtam/yawbe/pkg/user"
)

var _ http.Handler = &Server{}

// Server represents the Web service.
type Server struct {
	router      *mux.Router
	userService *user.Service

	accountView *view

	adminView           *view
	adminUserAddView    *view
	adminUserDeleteView *view
	adminUserEditView   *view

	homeView      *view
	userLoginView *view
}

// NewServer initializes and returns a new Server.
func NewServer(userService *user.Service) *Server {
	s := &Server{
		router:      mux.NewRouter(),
		userService: userService,

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
	s.router.HandleFunc("/", s.rememberUser(s.homeView.handle))

	// user account
	s.router.HandleFunc("/account", s.rememberUser(s.authenticatedUser(s.handleAccountView()))).Methods("GET")
	s.router.HandleFunc("/account/info", s.rememberUser(s.authenticatedUser(s.handleAccountInfoUpdate()))).Methods("POST")
	s.router.HandleFunc("/account/password", s.rememberUser(s.authenticatedUser(s.handleAccountPasswordUpdate()))).Methods("POST")

	// administration
	s.router.HandleFunc("/admin", s.rememberUser(s.adminUser(s.handleAdmin()))).Methods("GET")
	s.router.HandleFunc("/admin/users/add", s.rememberUser(s.adminUser(s.adminUserAddView.handle))).Methods("GET")
	s.router.HandleFunc("/admin/users", s.rememberUser(s.adminUser(s.handleAdminUserAdd()))).Methods("POST")
	s.router.HandleFunc("/admin/users/{uuid}", s.rememberUser(s.adminUser(s.handleAdminUserEditView()))).Methods("GET")
	s.router.HandleFunc("/admin/users/{uuid}", s.rememberUser(s.adminUser(s.handleAdminUserEdit()))).Methods("POST")
	s.router.HandleFunc("/admin/users/{uuid}/delete", s.rememberUser(s.adminUser(s.handleAdminUserDeleteView()))).Methods("GET")
	s.router.HandleFunc("/admin/users/{uuid}/delete", s.rememberUser(s.adminUser(s.handleAdminUserDelete()))).Methods("POST")

	// authentication
	s.router.HandleFunc("/login", s.rememberUser(s.userLoginView.handle)).Methods("GET")
	s.router.HandleFunc("/login", s.rememberUser(s.handleUserLogin())).Methods("POST")
	s.router.HandleFunc("/logout", s.rememberUser(s.handleUserLogout())).Methods("POST")

	// static assets
	s.router.HandleFunc("/static/", http.NotFound)
	s.router.PathPrefix("/static/").Handler(http.StripPrefix(
		"/static/",
		s.staticCacheControl(http.FileServer(http.FS(static.FS)))))
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
		Email string `schema:"email"`
	}

	var form infoUpdateForm
	var viewData Data

	return func(w http.ResponseWriter, r *http.Request) {
		ctxUser := userValue(r.Context())

		if err := parseForm(r, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse account information update form")
			viewData.AlertErrorStr("There was an error processing the form")
			viewData.User = ctxUser
			s.accountView.render(w, r, viewData)
			return
		}

		userInfo := user.InfoUpdate{
			UUID:  ctxUser.UUID,
			Email: form.Email,
		}

		if err := s.userService.UpdateInfo(userInfo); err != nil {
			log.Error().Err(err).Msg("failed to update account information")
			viewData.AlertErrorStr("There was an error updating your information")
			viewData.User = ctxUser
			s.accountView.render(w, r, viewData)
			return
		}

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
	var viewData Data

	return func(w http.ResponseWriter, r *http.Request) {
		ctxUser := userValue(r.Context())

		if err := parseForm(r, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse account password update form")
			viewData.AlertErrorStr("There was an error processing the form")
			viewData.User = ctxUser
			s.accountView.render(w, r, viewData)
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
			viewData.AlertErrorStr("There was an error updating your password")
			viewData.User = ctxUser
			s.accountView.render(w, r, viewData)
			return
		}

		http.Redirect(w, r, "/account", http.StatusFound)
	}
}

// handleAdmin displays the main administration page.
func (s *Server) handleAdmin() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var viewData Data

		users, err := s.userService.All()
		if err != nil {
			viewData.AlertError(err)
		} else {
			viewData.Content = users
		}

		s.adminView.render(w, r, viewData)
	}
}

// handleAdminUserAdd processes data submitted through the user creation form.
func (s *Server) handleAdminUserAdd() func(w http.ResponseWriter, r *http.Request) {
	var viewData Data

	type userAddForm struct {
		Email    string `schema:"email"`
		Password string `schema:"password"`
		IsAdmin  bool   `schema:"is_admin"`
	}

	var form userAddForm

	return func(w http.ResponseWriter, r *http.Request) {
		if err := parseForm(r, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse user creation form")
			viewData.AlertError(err)
			s.adminUserAddView.render(w, r, viewData)
			return
		}

		newUser := user.User{
			Email:    form.Email,
			Password: form.Password,
			IsAdmin:  form.IsAdmin,
		}

		if err := s.userService.Add(newUser); err != nil {
			log.Error().Err(err).Msg("failed to persist user")
			viewData.AlertError(err)
			s.adminUserAddView.render(w, r, viewData)
			return
		}

		viewData.AlertSuccess(fmt.Sprintf("user %q has been successfully created", newUser.Email))
		s.adminUserAddView.render(w, r, viewData)
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
			viewData.AlertError(err)
			s.adminUserDeleteView.render(w, r, viewData)
			return
		}

		viewData.Content = user

		s.adminUserDeleteView.render(w, r, viewData)
	}
}

// handleAdminUserDelete processes the user deletion form.
func (s *Server) handleAdminUserDelete() func(w http.ResponseWriter, r *http.Request) {
	var viewData Data

	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		userUUID := vars["uuid"]

		if err := s.userService.DeleteByUUID(userUUID); err != nil {
			log.Error().Err(err).Msg("failed to delete user")
			viewData.AlertError(err)
			s.adminUserEditView.render(w, r, viewData)
			return
		}

		viewData.AlertSuccess(fmt.Sprintf("user %q has been successfully deleted", userUUID))
		s.adminView.render(w, r, viewData)
	}
}

// handleAdminUserEditView displays the user edition form.
func (s *Server) handleAdminUserEditView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		userUUID := vars["uuid"]

		var viewData Data

		user, err := s.userService.ByUUID(userUUID)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve user")
			viewData.AlertError(err)
			s.adminUserEditView.render(w, r, viewData)
			return
		}

		viewData.Content = user

		s.adminUserEditView.render(w, r, viewData)
	}
}

// handleAdminUserEdit processes the user edition form.
func (s *Server) handleAdminUserEdit() func(w http.ResponseWriter, r *http.Request) {
	var viewData Data

	type userEditForm struct {
		Email    string `schema:"email"`
		Password string `schema:"password"`
		IsAdmin  bool   `schema:"is_admin"`
	}

	var form userEditForm

	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		userUUID := vars["uuid"]

		if err := parseForm(r, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse user edition form")
			viewData.AlertError(err)
			s.adminUserEditView.render(w, r, viewData)
			return
		}

		editedUser := user.User{
			UUID:     userUUID,
			Email:    form.Email,
			Password: form.Password,
			IsAdmin:  form.IsAdmin,
		}

		if err := s.userService.Update(editedUser); err != nil {
			log.Error().Err(err).Msg("failed to update user")
			viewData.AlertError(err)
			s.adminUserEditView.render(w, r, viewData)
			return
		}

		viewData.AlertSuccess(fmt.Sprintf("user %q has been successfully updated", editedUser.Email))
		s.adminUserEditView.render(w, r, viewData)
	}
}

// handleUserLogin processes data submitted through the user login form.
func (s *Server) handleUserLogin() func(w http.ResponseWriter, r *http.Request) {
	var viewData Data

	type loginForm struct {
		Email    string `schema:"email"`
		Password string `schema:"password"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var form loginForm

		if err := parseForm(r, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse login form")
			viewData.AlertError(err)
			s.userLoginView.render(w, r, viewData)
			return
		}

		user, err := s.userService.Authenticate(form.Email, form.Password)
		if err != nil {
			log.Error().Err(err).Msg("failed to authenticate user")
			viewData.AlertError(err)
			s.userLoginView.render(w, r, viewData)
			return
		}

		if err := s.setUserRememberToken(w, &user); err != nil {
			log.Error().Err(err).Msg("failed to set remember token")
			viewData.AlertError(err)
			s.userLoginView.render(w, r, viewData)
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
			Expires:  time.Now(),
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

		user.RememberToken = token
		err = s.userService.UpdateRememberToken(*user)
		if err != nil {
			log.Error().Err(err).Msg("failed to update user")
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
func (s *Server) setUserRememberToken(w http.ResponseWriter, user *user.User) error {
	if user.RememberToken == "" {
		token, err := rand.RandomBase64URLString(UserRememberTokenNBytes)
		if err != nil {
			log.Error().Err(err).Msg("failed to generate a remember token")
			return err
		}

		user.RememberToken = token
		err = s.userService.UpdateRememberToken(*user)
		if err != nil {
			log.Error().Err(err).Msg("failed to update user")
			return err
		}
	}

	cookie := http.Cookie{
		Name:     UserRememberTokenCookieName,
		Value:    user.RememberToken,
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

		user, err := s.userService.ByRememberToken(cookie.Value)
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
