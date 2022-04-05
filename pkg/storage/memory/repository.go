package memory

import (
	"github.com/virtualtam/yawbe/pkg/bookmark"
	"github.com/virtualtam/yawbe/pkg/session"
	"github.com/virtualtam/yawbe/pkg/user"
)

var _ bookmark.Repository = &Repository{}
var _ session.Repository = &Repository{}
var _ user.Repository = &Repository{}

type Repository struct {
	bookmarks []bookmark.Bookmark
	sessions  []session.Session
	users     []user.User
}

func (r *Repository) BookmarkAdd(b bookmark.Bookmark) error {
	r.bookmarks = append(r.bookmarks, b)
	return nil
}

func (r *Repository) BookmarkDelete(userUUID, uid string) error {
	for index, b := range r.bookmarks {
		if b.UserUUID == userUUID && b.UID == uid {
			r.bookmarks = append(r.bookmarks[:index], r.bookmarks[index+1:]...)
			return nil
		}
	}

	return bookmark.ErrNotFound
}

func (r *Repository) BookmarkGetAll(userUUID string) ([]bookmark.Bookmark, error) {
	bookmarks := []bookmark.Bookmark{}

	for _, b := range r.bookmarks {
		if b.UserUUID == userUUID {
			bookmarks = append(bookmarks, b)
		}
	}

	return bookmarks, nil
}

func (r *Repository) BookmarkGetByUID(userUUID, uid string) (bookmark.Bookmark, error) {
	for _, b := range r.bookmarks {
		if b.UserUUID == userUUID && b.UID == uid {
			return b, nil
		}
	}

	return bookmark.Bookmark{}, bookmark.ErrNotFound
}

func (r *Repository) BookmarkGetByURL(userUUID, url string) (bookmark.Bookmark, error) {
	for _, b := range r.bookmarks {
		if b.UserUUID == userUUID && b.URL == url {
			return b, nil
		}
	}

	return bookmark.Bookmark{}, bookmark.ErrNotFound
}

func (r *Repository) BookmarkIsURLRegistered(userUUID, url string) (bool, error) {
	for _, b := range r.bookmarks {
		if b.UserUUID == userUUID && b.URL == url {
			return true, nil
		}
	}
	return false, nil
}

func (r *Repository) BookmarkUpdate(updated bookmark.Bookmark) error {
	for index, b := range r.bookmarks {
		if b.UserUUID == updated.UserUUID && b.UID == updated.UID {
			r.bookmarks[index] = updated
			return nil
		}
	}

	return bookmark.ErrNotFound
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
