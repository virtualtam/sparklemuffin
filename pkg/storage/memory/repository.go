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
		Email:             u.Email,
		PasswordHash:      u.PasswordHash,
		RememberTokenHash: u.RememberTokenHash,
		IsAdmin:           u.IsAdmin,
	}

	r.users = append(r.users, newUser)

	return nil
}

func (r *Repository) GetUserByEmail(email string) (user.User, error) {
	for _, existingUser := range r.users {
		if existingUser.Email == email {
			return user.User{
				Email:             existingUser.Email,
				PasswordHash:      existingUser.PasswordHash,
				RememberTokenHash: existingUser.RememberTokenHash,
				IsAdmin:           existingUser.IsAdmin,
			}, nil
		}
	}

	return user.User{}, user.ErrNotFound
}

func (r *Repository) GetUserByRememberTokenHash(rememberTokenHash string) (user.User, error) {
	for _, existingUser := range r.users {
		if existingUser.RememberTokenHash == rememberTokenHash {
			return user.User{
				Email:             existingUser.Email,
				PasswordHash:      existingUser.PasswordHash,
				RememberTokenHash: existingUser.RememberTokenHash,
				IsAdmin:           existingUser.IsAdmin,
			}, nil
		}
	}

	return user.User{}, user.ErrNotFound
}

func (r *Repository) UpdateUser(u user.User) error {
	for index, existingUser := range r.users {
		if existingUser.Email == u.Email {
			r.users[index] = User{
				Email:             u.Email,
				PasswordHash:      u.PasswordHash,
				RememberTokenHash: u.RememberTokenHash,
				IsAdmin:           u.IsAdmin,
			}
			return nil
		}
	}

	return user.ErrNotFound
}
