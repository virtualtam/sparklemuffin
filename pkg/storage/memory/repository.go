package memory

import (
	"github.com/virtualtam/yawbe/pkg/user"
)

var _ user.Repository = &Repository{}

type Repository struct {
	users []User
}

func (r *Repository) UserAdd(u user.User) error {
	newUser := User{
		UUID:              u.UUID,
		Email:             u.Email,
		PasswordHash:      u.PasswordHash,
		RememberTokenHash: u.RememberTokenHash,
		IsAdmin:           u.IsAdmin,
		CreatedAt:         u.CreatedAt,
		UpdatedAt:         u.UpdatedAt,
	}

	r.users = append(r.users, newUser)

	return nil
}

func (r *Repository) UserDeleteByUUID(userUUID string) error {
	for index, user := range r.users {
		if user.UUID == userUUID {
			r.users = append(r.users[:index], r.users[index+1:]...)
		}
	}

	return user.ErrNotFound
}

func (r *Repository) UserGetAll() ([]user.User, error) {
	users := make([]user.User, len(r.users))

	for index, u := range r.users {
		user := user.User{
			UUID:              u.UUID,
			Email:             u.Email,
			PasswordHash:      u.PasswordHash,
			RememberTokenHash: u.RememberTokenHash,
			IsAdmin:           u.IsAdmin,
			CreatedAt:         u.CreatedAt,
			UpdatedAt:         u.UpdatedAt,
		}

		users[index] = user
	}

	return users, nil
}

func (r *Repository) UserGetByEmail(email string) (user.User, error) {
	for _, existingUser := range r.users {
		if existingUser.Email == email {
			return user.User{
				UUID:              existingUser.UUID,
				Email:             existingUser.Email,
				PasswordHash:      existingUser.PasswordHash,
				RememberTokenHash: existingUser.RememberTokenHash,
				IsAdmin:           existingUser.IsAdmin,
				CreatedAt:         existingUser.CreatedAt,
				UpdatedAt:         existingUser.UpdatedAt,
			}, nil
		}
	}

	return user.User{}, user.ErrNotFound
}

func (r *Repository) UserIsEmailRegistered(email string) (bool, error) {
	registered := false

	for _, existingUser := range r.users {
		if existingUser.Email == email {
			registered = true
			break
		}
	}

	return registered, nil
}

func (r *Repository) UserGetByRememberTokenHash(rememberTokenHash string) (user.User, error) {
	for _, existingUser := range r.users {
		if existingUser.RememberTokenHash == rememberTokenHash {
			return user.User{
				UUID:              existingUser.UUID,
				Email:             existingUser.Email,
				PasswordHash:      existingUser.PasswordHash,
				RememberTokenHash: existingUser.RememberTokenHash,
				IsAdmin:           existingUser.IsAdmin,
				CreatedAt:         existingUser.CreatedAt,
				UpdatedAt:         existingUser.UpdatedAt,
			}, nil
		}
	}

	return user.User{}, user.ErrNotFound
}

func (r *Repository) UserGetByUUID(userUUID string) (user.User, error) {
	for _, existingUser := range r.users {
		if existingUser.UUID == userUUID {
			return user.User{
				UUID:              existingUser.UUID,
				Email:             existingUser.Email,
				PasswordHash:      existingUser.PasswordHash,
				RememberTokenHash: existingUser.RememberTokenHash,
				IsAdmin:           existingUser.IsAdmin,
				CreatedAt:         existingUser.CreatedAt,
				UpdatedAt:         existingUser.UpdatedAt,
			}, nil
		}
	}

	return user.User{}, user.ErrNotFound
}

func (r *Repository) UserUpdate(u user.User) error {
	for index, existingUser := range r.users {
		if existingUser.UUID == u.UUID {
			r.users[index] = User{
				UUID:              u.UUID,
				Email:             u.Email,
				PasswordHash:      u.PasswordHash,
				RememberTokenHash: u.RememberTokenHash,
				IsAdmin:           u.IsAdmin,
				CreatedAt:         u.CreatedAt,
				UpdatedAt:         u.UpdatedAt,
			}
			return nil
		}
	}

	return user.ErrNotFound
}

func (r *Repository) UserUpdateInfo(info user.InfoUpdate) error {
	for index, existingUser := range r.users {
		if existingUser.UUID == info.UUID {
			r.users[index].Email = info.Email
			r.users[index].UpdatedAt = info.UpdatedAt
			return nil
		}
	}

	return user.ErrNotFound
}

func (r *Repository) UserUpdatePasswordHash(passwordHash user.PasswordHashUpdate) error {
	for index, existingUser := range r.users {
		if existingUser.UUID == passwordHash.UUID {
			r.users[index].PasswordHash = passwordHash.PasswordHash
			r.users[index].UpdatedAt = passwordHash.UpdatedAt
			return nil
		}
	}

	return user.ErrNotFound
}

func (r *Repository) UserUpdateRememberTokenHash(u user.User) error {
	for index, existingUser := range r.users {
		if existingUser.UUID == u.UUID {
			r.users[index].RememberTokenHash = u.RememberTokenHash
			return nil
		}
	}

	return user.ErrNotFound
}
