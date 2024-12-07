// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package controller

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	"github.com/virtualtam/sparklemuffin/internal/http/www/csrf"
	"github.com/virtualtam/sparklemuffin/internal/http/www/httpcontext"
	"github.com/virtualtam/sparklemuffin/internal/http/www/middleware"
	"github.com/virtualtam/sparklemuffin/internal/http/www/view"
	"github.com/virtualtam/sparklemuffin/pkg/user"
)

const (
	actionAdminUserAdd    string = "admin-user-add"
	actionAdminUserDelete string = "admin-user-delete"
	actionAdminUserEdit   string = "admin-user-edit"
)

type adminHandlerContext struct {
	csrfService *csrf.Service
	userService *user.Service

	adminView           *view.View
	adminUserAddView    *view.View
	adminUserDeleteView *view.View
	adminUserEditView   *view.View
}

func RegisterAdminHandlers(
	r *chi.Mux,
	csrfService *csrf.Service,
	userService *user.Service,
) {
	hc := adminHandlerContext{
		csrfService: csrfService,
		userService: userService,

		adminView:           view.New("admin/admin.gohtml"),
		adminUserAddView:    view.New("admin/user_add.gohtml"),
		adminUserDeleteView: view.New("admin/user_delete.gohtml"),
		adminUserEditView:   view.New("admin/user_edit.gohtml"),
	}

	// administration
	r.Route("/admin", func(r chi.Router) {
		r.Use(func(h http.Handler) http.Handler {
			return middleware.AdminUser(h.ServeHTTP)
		})

		r.Get("/", hc.handleAdmin())
		r.Get("/users/add", hc.handleAdminUserAddView())
		r.Post("/users", hc.handleAdminUserAdd())
		r.Get("/users/{uuid}", hc.handleAdminUserEditView())
		r.Post("/users/{uuid}", hc.handleAdminUserEdit())
		r.Get("/users/{uuid}/delete", hc.handleAdminUserDeleteView())
		r.Post("/users/{uuid}/delete", hc.handleAdminUserDelete())
	})
}

// handleAdmin renders the main administration page.
func (hc *adminHandlerContext) handleAdmin() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		viewData := view.Data{Title: "Administration"}

		users, err := hc.userService.All()
		if err != nil {
			view.PutFlashError(w, err.Error())
		} else {
			viewData.Content = users
		}

		hc.adminView.Render(w, r, viewData)
	}
}

// handleAdminUserAddView renders the user creation form.
func (hc *adminHandlerContext) handleAdminUserAddView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		adminUser := httpcontext.UserValue(r.Context())
		csrfToken := hc.csrfService.Generate(adminUser.UUID, actionAdminUserAdd)

		viewData := view.Data{
			Title: "Add user",
			Content: view.FormContent{
				CSRFToken: csrfToken,
			},
		}

		hc.adminUserAddView.Render(w, r, viewData)
	}
}

// handleAdminUserAdd processes view.data submitted through the user creation form.
func (hc *adminHandlerContext) handleAdminUserAdd() func(w http.ResponseWriter, r *http.Request) {
	type userAddForm struct {
		CSRFToken   string `schema:"csrf_token"`
		Email       string `schema:"email"`
		NickName    string `schema:"nick_name"`
		DisplayName string `schema:"display_name"`
		Password    string `schema:"password"`
		IsAdmin     bool   `schema:"is_admin"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		adminUser := httpcontext.UserValue(r.Context())

		var form userAddForm
		if err := decodeForm(r, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse user creation form")
			view.PutFlashError(w, err.Error())
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		if !hc.csrfService.Validate(form.CSRFToken, adminUser.UUID, actionAdminUserAdd) {
			log.Warn().Msg("failed to validate CSRF token")
			view.PutFlashError(w, "There was an error processing the form")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
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
			view.PutFlashError(w, err.Error())
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		view.PutFlashSuccess(w, fmt.Sprintf("user %q has been successfully created", newUser.Email))
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	}
}

// handleAdminUserDeleteView renders the user deletion form.
func (hc *adminHandlerContext) handleAdminUserDeleteView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		adminUser := httpcontext.UserValue(r.Context())
		userUUID := chi.URLParam(r, "uuid")

		csrfToken := hc.csrfService.Generate(adminUser.UUID, actionAdminUserDelete)

		user, err := hc.userService.ByUUID(userUUID)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve user")
			view.PutFlashError(w, err.Error())
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		viewData := view.Data{
			Title: fmt.Sprintf("Delete user: %s", user.NickName),
			Content: view.FormContent{
				CSRFToken: csrfToken,
				Content:   user,
			},
		}

		hc.adminUserDeleteView.Render(w, r, viewData)
	}
}

// handleAdminUserDelete processes the user deletion form.
func (hc *adminHandlerContext) handleAdminUserDelete() func(w http.ResponseWriter, r *http.Request) {
	type userDeleteForm struct {
		CSRFToken string `schema:"csrf_token"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		adminUser := httpcontext.UserValue(r.Context())
		userUUID := chi.URLParam(r, "uuid")

		var form userDeleteForm
		if err := decodeForm(r, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse user deletion form")
			view.PutFlashError(w, "There was an error processing the form")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		if !hc.csrfService.Validate(form.CSRFToken, adminUser.UUID, actionAdminUserDelete) {
			log.Warn().Msg("failed to validate CSRF token")
			view.PutFlashError(w, "There was an error processing the form")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		user, err := hc.userService.ByUUID(userUUID)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve user")
			view.PutFlashError(w, err.Error())
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		if err := hc.userService.DeleteByUUID(userUUID); err != nil {
			log.Error().Err(err).Msg("failed to delete user")
			view.PutFlashError(w, err.Error())
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		view.PutFlashSuccess(w, fmt.Sprintf("user %q has been successfully deleted", user.Email))
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	}
}

// handleAdminUserEditView renders the user edition form.
func (hc *adminHandlerContext) handleAdminUserEditView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		adminUser := httpcontext.UserValue(r.Context())
		userUUID := chi.URLParam(r, "uuid")

		csrfToken := hc.csrfService.Generate(adminUser.UUID, actionAdminUserEdit)

		user, err := hc.userService.ByUUID(userUUID)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve user")
			view.PutFlashError(w, err.Error())
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		viewData := view.Data{
			Title: fmt.Sprintf("Edit user: %s", user.NickName),
			Content: view.FormContent{
				CSRFToken: csrfToken,
				Content:   user,
			},
		}
		hc.adminUserEditView.Render(w, r, viewData)
	}
}

// handleAdminUserEdit processes the user edition form.
func (hc *adminHandlerContext) handleAdminUserEdit() func(w http.ResponseWriter, r *http.Request) {
	type userEditForm struct {
		CSRFToken   string `schema:"csrf_token"`
		Email       string `schema:"email"`
		NickName    string `schema:"nick_name"`
		DisplayName string `schema:"display_name"`
		Password    string `schema:"password"`
		IsAdmin     bool   `schema:"is_admin"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		adminUser := httpcontext.UserValue(r.Context())
		userUUID := chi.URLParam(r, "uuid")

		var form userEditForm
		if err := decodeForm(r, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse user edition form")
			view.PutFlashError(w, err.Error())
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		if !hc.csrfService.Validate(form.CSRFToken, adminUser.UUID, actionAdminUserEdit) {
			log.Warn().Msg("failed to validate CSRF token")
			view.PutFlashError(w, "There was an error processing the form")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
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
			view.PutFlashError(w, err.Error())
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		view.PutFlashSuccess(w, fmt.Sprintf("user %q has been successfully updated", editedUser.Email))
		http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
	}
}
