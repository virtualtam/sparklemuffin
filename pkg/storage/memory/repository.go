package memory

import (
	"github.com/virtualtam/yawbe/pkg/session"
	"github.com/virtualtam/yawbe/pkg/user"
)

var _ session.Repository = &Repository{}
var _ user.Repository = &Repository{}

type Repository struct {
	sessions []session.Session
	users    []user.User
}

func (r *Repository) SessionAdd(session session.Session) error {
	r.sessions = append(r.sessions, session)
	return nil
}

func (r *Repository) SessionGetByRememberTokenHash(hash string) (session.Session, error) {
	for _, s := range r.sessions {
		if s.RememberTokenHash == hash {
			return s, nil
		}
	}

	return session.Session{}, session.ErrNotFound
}

func (r *Repository) UserAdd(u user.User) error {
	r.users = append(r.users, u)
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
	return r.users, nil
}

func (r *Repository) UserGetByNickName(nick string) (user.User, error) {
	for _, existingUser := range r.users {
		if existingUser.NickName == nick {
			return existingUser, nil
		}
	}

	return user.User{}, user.ErrNotFound
}

func (r *Repository) UserGetByEmail(email string) (user.User, error) {
	for _, existingUser := range r.users {
		if existingUser.Email == email {
			return existingUser, nil
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

func (r *Repository) UserIsNickNameRegistered(nick string) (bool, error) {
	registered := false

	for _, existingUser := range r.users {
		if existingUser.NickName == nick {
			registered = true
			break
		}
	}

	return registered, nil
}

func (r *Repository) UserGetByUUID(userUUID string) (user.User, error) {
	for _, existingUser := range r.users {
		if existingUser.UUID == userUUID {
			return existingUser, nil
		}
	}

	return user.User{}, user.ErrNotFound
}

func (r *Repository) UserUpdate(u user.User) error {
	for index, existingUser := range r.users {
		if existingUser.UUID == u.UUID {
			r.users[index] = u
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
