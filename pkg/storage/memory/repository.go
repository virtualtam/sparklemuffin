package memory

import (
	"github.com/virtualtam/yawbe/pkg/user"
)

var _ user.Repository = &Repository{}

type Repository struct {
	users []User
}

func (r *Repository) AddUser(u user.User) error {
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

func (r *Repository) GetAllUsers() ([]user.User, error) {
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

func (r *Repository) GetUserByEmail(email string) (user.User, error) {
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

func (r *Repository) GetUserByRememberTokenHash(rememberTokenHash string) (user.User, error) {
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

func (r *Repository) UpdateUser(u user.User) error {
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

func (r *Repository) UpdateUserRememberTokenHash(u user.User) error {
	for index, existingUser := range r.users {
		if existingUser.UUID == u.UUID {
			r.users[index].RememberTokenHash = u.RememberTokenHash
			return nil
		}
	}

	return user.ErrNotFound
}
