package postgresql

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/virtualtam/yawbe/pkg/bookmark"
	"github.com/virtualtam/yawbe/pkg/session"
	"github.com/virtualtam/yawbe/pkg/user"
)

var _ bookmark.Repository = &Repository{}
var _ session.Repository = &Repository{}
var _ user.Repository = &Repository{}

// Repository provides a PostgreSQL persistence layer.
type Repository struct {
	db *sqlx.DB
}

// NewRepository initializes and returns a PostgreSQL Repository.
func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) BookmarkAdd(b bookmark.Bookmark) error {
	dbBookmark := Bookmark{
		UID:         b.UID,
		UserUUID:    b.UserUUID,
		URL:         b.URL,
		Title:       b.Title,
		Description: b.Description,
		CreatedAt:   b.CreatedAt,
		UpdatedAt:   b.UpdatedAt,
	}

	_, err := r.db.NamedExec(
		`
INSERT INTO bookmarks(
	uid,
	user_uuid,
	url,
	title,
	description,
	created_at,
	updated_at
)
VALUES(
	:uid,
	:user_uuid,
	:url,
	:title,
	:description,
	:created_at,
	:updated_at
)
`,
		dbBookmark,
	)

	return err
}

func (r *Repository) BookmarkDelete(userUUID, uid string) error {
	result, err := r.db.Exec("DELETE FROM bookmarks WHERE user_uuid=$1 AND uid=$2", userUUID, uid)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected != 1 {
		return user.ErrNotFound
	}

	return nil
}

func (r *Repository) BookmarkGetAll(userUUID string) ([]bookmark.Bookmark, error) {
	rows, err := r.db.Queryx(
		`
SELECT user_uuid, uid, url, title, description, created_at, updated_at
FROM bookmarks
WHERE user_uuid=$1
ORDER BY created_at DESC`,
		userUUID,
	)
	if err != nil {
		return []bookmark.Bookmark{}, err
	}

	bookmarks := []bookmark.Bookmark{}

	for rows.Next() {
		dbBookmark := Bookmark{}

		if err := rows.StructScan(&dbBookmark); err != nil {
			return []bookmark.Bookmark{}, err
		}

		bookmark := bookmark.Bookmark{
			UserUUID:    dbBookmark.UserUUID,
			UID:         dbBookmark.UID,
			URL:         dbBookmark.URL,
			Title:       dbBookmark.Title,
			Description: dbBookmark.Description,
			CreatedAt:   dbBookmark.CreatedAt,
			UpdatedAt:   dbBookmark.UpdatedAt,
		}

		bookmarks = append(bookmarks, bookmark)
	}

	return bookmarks, nil
}

func (r *Repository) BookmarkGetByURL(userUUID, url string) (bookmark.Bookmark, error) {
	dbBookmark := &Bookmark{}

	err := r.db.QueryRowx(
		`
SELECT user_uuid, uid, url, title, description, created_at, updated_at
FROM bookmarks
WHERE user_uuid=$1
AND url=$2`,
		userUUID,
		url,
	).StructScan(dbBookmark)

	if errors.Is(err, sql.ErrNoRows) {
		return bookmark.Bookmark{}, bookmark.ErrNotFound
	}
	if err != nil {
		return bookmark.Bookmark{}, err
	}

	return bookmark.Bookmark{
		UserUUID:    dbBookmark.UserUUID,
		UID:         dbBookmark.UID,
		URL:         dbBookmark.URL,
		Title:       dbBookmark.Title,
		Description: dbBookmark.Description,
		CreatedAt:   dbBookmark.CreatedAt,
		UpdatedAt:   dbBookmark.UpdatedAt,
	}, nil
}

func (r *Repository) BookmarkUpdate(b bookmark.Bookmark) error {
	dbBookmark := Bookmark{
		UserUUID:    b.UserUUID,
		UID:         b.UID,
		URL:         b.URL,
		Title:       b.Title,
		Description: b.Description,
		UpdatedAt:   b.UpdatedAt,
	}

	_, err := r.db.NamedExec(
		`
UPDATE bookmarks
SET
	url=:url,
	title=:title,
	description=:description,
	updated_at=:updated_at
WHERE user_uuid=:user_uuid
AND uid=:uid
		`,
		dbBookmark,
	)
	return err
}

func (r *Repository) SessionAdd(sess session.Session) error {
	dbSession := Session{
		UserUUID:               sess.UserUUID,
		RememberTokenHash:      sess.RememberTokenHash,
		RememberTokenExpiresAt: sess.RememberTokenExpiresAt,
	}

	_, err := r.db.NamedExec(
		`
INSERT INTO sessions(
	user_uuid,
	remember_token_hash,
	remember_token_expires_at
)
VALUES(
	:user_uuid,
	:remember_token_hash,
	:remember_token_expires_at
)`,
		dbSession,
	)

	return err
}

func (r *Repository) SessionGetByRememberTokenHash(hash string) (session.Session, error) {
	dbSession := &Session{}

	err := r.db.QueryRowx(
		`SELECT user_uuid, remember_token_hash
FROM sessions
WHERE remember_token_hash=$1`,
		hash,
	).StructScan(dbSession)

	if errors.Is(err, sql.ErrNoRows) {
		return session.Session{}, session.ErrNotFound
	}
	if err != nil {
		return session.Session{}, err
	}

	return session.Session{
		UserUUID:          dbSession.UserUUID,
		RememberTokenHash: dbSession.RememberTokenHash,
	}, nil
}

func (r *Repository) UserAdd(u user.User) error {
	dbUser := User{
		UUID:         u.UUID,
		Email:        u.Email,
		NickName:     u.NickName,
		DisplayName:  u.DisplayName,
		PasswordHash: u.PasswordHash,
		IsAdmin:      u.IsAdmin,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}

	_, err := r.db.NamedExec(
		`
INSERT INTO users(
	uuid,
	email,
	nick_name,
	display_name,
	password_hash,
	is_admin,
	created_at,
	updated_at
)
VALUES(
	:uuid,
	:email,
	:nick_name,
	:display_name,
	:password_hash,
	:is_admin,
	:created_at,
	:updated_at
)`,
		dbUser,
	)

	return err
}

func (r *Repository) UserDeleteByUUID(userUUID string) error {
	result, err := r.db.Exec("DELETE FROM users WHERE uuid=$1", userUUID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected != 1 {
		return user.ErrNotFound
	}

	return nil
}

func (r *Repository) UserGetAll() ([]user.User, error) {
	rows, err := r.db.Queryx("SELECT uuid, email, nick_name, display_name, is_admin, created_at, updated_at FROM users")
	if err != nil {
		return []user.User{}, err
	}

	users := []user.User{}

	for rows.Next() {
		dbUser := User{}

		if err := rows.StructScan(&dbUser); err != nil {
			return []user.User{}, err
		}

		user := user.User{
			UUID:        dbUser.UUID,
			Email:       dbUser.Email,
			NickName:    dbUser.NickName,
			DisplayName: dbUser.DisplayName,
			IsAdmin:     dbUser.IsAdmin,
			CreatedAt:   dbUser.CreatedAt,
			UpdatedAt:   dbUser.UpdatedAt,
		}

		users = append(users, user)
	}

	return users, nil
}

func (r *Repository) UserGetByEmail(email string) (user.User, error) {
	dbUser := &User{}

	err := r.db.QueryRowx(
		`SELECT uuid, email, nick_name, display_name, password_hash, is_admin, created_at, updated_at
FROM users
WHERE email=$1`,
		email,
	).StructScan(dbUser)

	if errors.Is(err, sql.ErrNoRows) {
		return user.User{}, user.ErrNotFound
	}
	if err != nil {
		return user.User{}, err
	}

	return user.User{
		UUID:         dbUser.UUID,
		Email:        dbUser.Email,
		NickName:     dbUser.NickName,
		DisplayName:  dbUser.DisplayName,
		PasswordHash: dbUser.PasswordHash,
		IsAdmin:      dbUser.IsAdmin,
		CreatedAt:    dbUser.CreatedAt,
		UpdatedAt:    dbUser.UpdatedAt,
	}, nil
}

func (r *Repository) UserGetByNickName(nick string) (user.User, error) {
	dbUser := &User{}

	err := r.db.QueryRowx(
		`SELECT uuid, email, nick_name, display_name, password_hash, is_admin, created_at, updated_at
FROM users
WHERE nick_name=$1`,
		nick,
	).StructScan(dbUser)

	if errors.Is(err, sql.ErrNoRows) {
		return user.User{}, user.ErrNotFound
	}
	if err != nil {
		return user.User{}, err
	}

	return user.User{
		UUID:         dbUser.UUID,
		Email:        dbUser.Email,
		NickName:     dbUser.NickName,
		DisplayName:  dbUser.DisplayName,
		PasswordHash: dbUser.PasswordHash,
		IsAdmin:      dbUser.IsAdmin,
		CreatedAt:    dbUser.CreatedAt,
		UpdatedAt:    dbUser.UpdatedAt,
	}, nil
}

func (r *Repository) UserGetByUUID(userUUID string) (user.User, error) {
	dbUser := &User{}

	err := r.db.QueryRowx(
		`SELECT uuid, email, nick_name, display_name, password_hash, is_admin, created_at, updated_at
FROM users
WHERE uuid=$1`,
		userUUID,
	).StructScan(dbUser)

	if errors.Is(err, sql.ErrNoRows) {
		return user.User{}, user.ErrNotFound
	}
	if err != nil {
		return user.User{}, err
	}

	return user.User{
		UUID:         dbUser.UUID,
		Email:        dbUser.Email,
		NickName:     dbUser.NickName,
		DisplayName:  dbUser.DisplayName,
		PasswordHash: dbUser.PasswordHash,
		IsAdmin:      dbUser.IsAdmin,
		CreatedAt:    dbUser.CreatedAt,
		UpdatedAt:    dbUser.UpdatedAt,
	}, nil
}

func (r *Repository) UserIsEmailRegistered(email string) (bool, error) {
	dbUser := &User{}

	err := r.db.QueryRowx("SELECT email FROM users WHERE email=$1", email).StructScan(dbUser)

	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

func (r *Repository) UserIsNickNameRegistered(nick string) (bool, error) {
	dbUser := &User{}

	err := r.db.QueryRowx("SELECT email FROM users WHERE nick_name=$1", nick).StructScan(dbUser)

	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

func (r *Repository) UserUpdate(u user.User) error {
	dbUser := User{
		UUID:         u.UUID,
		Email:        u.Email,
		NickName:     u.NickName,
		DisplayName:  u.DisplayName,
		PasswordHash: u.PasswordHash,
		IsAdmin:      u.IsAdmin,
		UpdatedAt:    u.UpdatedAt,
	}

	_, err := r.db.NamedExec(`UPDATE users
SET
	email=:email,
	nick_name=:nick_name,
	display_name=:display_name,
	password_hash=:password_hash,
	is_admin=:is_admin,
	updated_at=:updated_at
WHERE uuid=:uuid`,
		dbUser,
	)

	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) UserUpdateInfo(info user.InfoUpdate) error {
	dbUser := User{
		UUID:        info.UUID,
		Email:       info.Email,
		NickName:    info.NickName,
		DisplayName: info.DisplayName,
		UpdatedAt:   info.UpdatedAt,
	}

	_, err := r.db.NamedExec(`UPDATE users
SET
	email=:email,
	nick_name=:nick_name,
	display_name=:display_name,
	updated_at=:updated_at
WHERE uuid=:uuid`,
		dbUser,
	)

	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) UserUpdatePasswordHash(passwordHash user.PasswordHashUpdate) error {
	dbUser := User{
		UUID:         passwordHash.UUID,
		PasswordHash: passwordHash.PasswordHash,
		UpdatedAt:    passwordHash.UpdatedAt,
	}

	_, err := r.db.NamedExec(`UPDATE users
SET
	password_hash=:password_hash,
	updated_at=:updated_at
WHERE uuid=:uuid`,
		dbUser,
	)

	if err != nil {
		return err
	}

	return nil
}
