package www

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	"github.com/virtualtam/sparklemuffin/pkg/user"
)

type adminHandlerContext struct {
	userService *user.Service

	adminView           *view
	adminUserAddView    *view
	adminUserDeleteView *view
	adminUserEditView   *view
}

func registerAdminHandlers(
	r *chi.Mux,
	userService *user.Service,
) {
	hc := adminHandlerContext{
		userService: userService,

		adminView:           newView("admin/admin.gohtml"),
		adminUserAddView:    newView("admin/user_add.gohtml"),
		adminUserDeleteView: newView("admin/user_delete.gohtml"),
		adminUserEditView:   newView("admin/user_edit.gohtml"),
	}

	// administration
	r.Route("/admin", func(r chi.Router) {
		r.Use(func(h http.Handler) http.Handler {
			return adminUser(h.ServeHTTP)
		})

		r.Get("/", hc.handleAdmin())
		r.Get("/users/add", hc.handleAdminUserAddView())
		r.Post("/users", hc.handleAdminUserAdd())
		r.Get("/users/{uuid}", hc.handleAdminUserEditView())
		r.Post("/users/{uuid}", hc.handleAdminUserEdit())
		r.Get("/users/{uuid}/delete", hc.handleAdminUserDeleteView())
		r.Get("/users/{uuid}/delete", hc.handleAdminUserDelete())
	})
}

// handleAdmin renders the main administration page.
func (hc *adminHandlerContext) handleAdmin() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		viewData := Data{Title: "Administration"}

		users, err := hc.userService.All()
		if err != nil {
			PutFlashError(w, err.Error())
		} else {
			viewData.Content = users
		}

		hc.adminView.render(w, r, viewData)
	}
}

// handleAdminUserAddView renders the user creation form.
func (hc *adminHandlerContext) handleAdminUserAddView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		viewData := Data{Title: "Add user"}

		hc.adminUserAddView.render(w, r, viewData)
	}
}

// handleAdminUserAdd processes data submitted through the user creation form.
func (hc *adminHandlerContext) handleAdminUserAdd() func(w http.ResponseWriter, r *http.Request) {
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
			PutFlashError(w, err.Error())
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

		if err := hc.userService.Add(newUser); err != nil {
			log.Error().Err(err).Msg("failed to persist user")
			PutFlashError(w, err.Error())
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		PutFlashSuccess(w, fmt.Sprintf("user %q has been successfully created", newUser.Email))
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	}
}

// handleAdminUserDeleteView renders the user deletion form.
func (hc *adminHandlerContext) handleAdminUserDeleteView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		userUUID := chi.URLParam(r, "uuid")

		var viewData Data

		user, err := hc.userService.ByUUID(userUUID)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve user")
			PutFlashError(w, err.Error())
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		viewData.Content = user
		viewData.Title = fmt.Sprintf("Delete user: %s", user.NickName)

		hc.adminUserDeleteView.render(w, r, viewData)
	}
}

// handleAdminUserDelete processes the user deletion form.
func (hc *adminHandlerContext) handleAdminUserDelete() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		userUUID := chi.URLParam(r, "uuid")

		user, err := hc.userService.ByUUID(userUUID)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve user")
			PutFlashError(w, err.Error())
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		if err := hc.userService.DeleteByUUID(userUUID); err != nil {
			log.Error().Err(err).Msg("failed to delete user")
			PutFlashError(w, err.Error())
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		PutFlashSuccess(w, fmt.Sprintf("user %q has been successfully deleted", user.Email))
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	}
}

// handleAdminUserEditView renders the user edition form.
func (hc *adminHandlerContext) handleAdminUserEditView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		userUUID := chi.URLParam(r, "uuid")

		user, err := hc.userService.ByUUID(userUUID)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve user")
			PutFlashError(w, err.Error())
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		viewData := Data{
			Content: user,
			Title:   fmt.Sprintf("Edit user: %s", user.NickName),
		}
		hc.adminUserEditView.render(w, r, viewData)
	}
}

// handleAdminUserEdit processes the user edition form.
func (hc *adminHandlerContext) handleAdminUserEdit() func(w http.ResponseWriter, r *http.Request) {
	type userEditForm struct {
		Email       string `schema:"email"`
		NickName    string `schema:"nick_name"`
		DisplayName string `schema:"display_name"`
		Password    string `schema:"password"`
		IsAdmin     bool   `schema:"is_admin"`
	}

	var form userEditForm

	return func(w http.ResponseWriter, r *http.Request) {
		userUUID := chi.URLParam(r, "uuid")

		if err := parseForm(r, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse user edition form")
			PutFlashError(w, err.Error())
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

		if err := hc.userService.Update(editedUser); err != nil {
			log.Error().Err(err).Msg("failed to update user")
			PutFlashError(w, err.Error())
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		PutFlashSuccess(w, fmt.Sprintf("user %q has been successfully updated", editedUser.Email))
		http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
	}
}
