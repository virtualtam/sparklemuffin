// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package controller

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	"github.com/virtualtam/sparklemuffin/internal/http/www/middleware"
	"github.com/virtualtam/sparklemuffin/internal/http/www/view"
	"github.com/virtualtam/sparklemuffin/pkg/user"
)

// RegisterAdminHandlers registers handlers for administration operations.
func RegisterAdminHandlers(
	r *chi.Mux,
	userService *user.Service,
) {
	ac := adminController{
		userService: userService,

		adminUserAddView:    view.New("admin/user_add.gohtml"),
		adminUserDeleteView: view.New("admin/user_delete.gohtml"),
		adminUserEditView:   view.New("admin/user_edit.gohtml"),
		adminUserListView:   view.New("admin/user_list.gohtml"),
	}

	// administration
	r.Route("/admin", func(r chi.Router) {
		r.Use(func(h http.Handler) http.Handler {
			return middleware.AdminUser(h.ServeHTTP)
		})

		r.Get("/users", ac.handleUserListView())
		r.Get("/users/add", ac.handleUserAddView())
		r.Post("/users", ac.handleUserAdd())
		r.Get("/users/{uuid}", ac.handleUserEditView())
		r.Post("/users/{uuid}", ac.handleUserEdit())
		r.Get("/users/{uuid}/delete", ac.handleUserDeleteView())
		r.Post("/users/{uuid}/delete", ac.handleUserDelete())
	})
}

type adminController struct {
	userService *user.Service

	adminUserAddView    *view.View
	adminUserDeleteView *view.View
	adminUserEditView   *view.View
	adminUserListView   *view.View
}

// handleUserListView renders the users list view.
func (ac *adminController) handleUserListView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		viewData := view.Data{Title: "Administration"}

		users, err := ac.userService.All(ctx)
		if err != nil {
			view.PutFlashError(w, err.Error())
		} else {
			viewData.Content = users
		}

		ac.adminUserListView.Render(w, r, viewData)
	}
}

// handleUserAddView renders the user creation form.
func (ac *adminController) handleUserAddView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		viewData := view.Data{
			Title: "Add user",
		}

		ac.adminUserAddView.Render(w, r, viewData)
	}
}

// handleUserAdd processes view.data submitted through the user creation form.
func (ac *adminController) handleUserAdd() func(w http.ResponseWriter, r *http.Request) {
	type userAddForm struct {
		Email       string `schema:"email"`
		NickName    string `schema:"nick_name"`
		DisplayName string `schema:"display_name"`
		Password    string `schema:"password"`
		IsAdmin     bool   `schema:"is_admin"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var form userAddForm
		if err := decodeForm(r, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse user creation form")
			view.PutFlashError(w, err.Error())
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

		if err := ac.userService.Add(ctx, newUser); err != nil {
			log.Error().Err(err).Msg("failed to persist user")
			view.PutFlashError(w, err.Error())
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		view.PutFlashSuccess(w, fmt.Sprintf("user %q has been successfully created", newUser.Email))
		http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
	}
}

// handleUserDeleteView renders the user deletion form.
func (ac *adminController) handleUserDeleteView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userUUID := chi.URLParam(r, "uuid")

		userToDelete, err := ac.userService.ByUUID(ctx, userUUID)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve user")
			view.PutFlashError(w, err.Error())
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		viewData := view.Data{
			Title:   fmt.Sprintf("Delete user: %s", userToDelete.NickName),
			Content: userToDelete,
		}

		ac.adminUserDeleteView.Render(w, r, viewData)
	}
}

// handleUserDelete processes the user deletion form.
func (ac *adminController) handleUserDelete() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userUUID := chi.URLParam(r, "uuid")

		userToDelete, err := ac.userService.ByUUID(ctx, userUUID)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve user")
			view.PutFlashError(w, err.Error())
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		if err := ac.userService.DeleteByUUID(ctx, userUUID); err != nil {
			log.Error().Err(err).Msg("failed to delete user")
			view.PutFlashError(w, err.Error())
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		view.PutFlashSuccess(w, fmt.Sprintf("user %q has been successfully deleted", userToDelete.Email))
		http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
	}
}

// handleUserEditView renders the user edition form.
func (ac *adminController) handleUserEditView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userUUID := chi.URLParam(r, "uuid")

		userToEdit, err := ac.userService.ByUUID(ctx, userUUID)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve user")
			view.PutFlashError(w, err.Error())
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		viewData := view.Data{
			Title:   fmt.Sprintf("Edit user: %s", userToEdit.NickName),
			Content: userToEdit,
		}
		ac.adminUserEditView.Render(w, r, viewData)
	}
}

// handleUserEdit processes the user edition form.
func (ac *adminController) handleUserEdit() func(w http.ResponseWriter, r *http.Request) {
	type userEditForm struct {
		Email       string `schema:"email"`
		NickName    string `schema:"nick_name"`
		DisplayName string `schema:"display_name"`
		Password    string `schema:"password"`
		IsAdmin     bool   `schema:"is_admin"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userUUID := chi.URLParam(r, "uuid")

		var form userEditForm
		if err := decodeForm(r, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse user edition form")
			view.PutFlashError(w, err.Error())
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

		if err := ac.userService.Update(ctx, editedUser); err != nil {
			log.Error().Err(err).Msg("failed to update user")
			view.PutFlashError(w, err.Error())
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		view.PutFlashSuccess(w, fmt.Sprintf("user %q has been successfully updated", editedUser.Email))
		http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
	}
}
