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
		Email: u.Email,
	}

	r.users = append(r.users, newUser)

	return nil
}
