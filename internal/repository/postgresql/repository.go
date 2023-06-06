package postgresql

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"github.com/virtualtam/sparklemuffin/pkg/bookmark"
	"github.com/virtualtam/sparklemuffin/pkg/exporting"
	"github.com/virtualtam/sparklemuffin/pkg/importing"
	"github.com/virtualtam/sparklemuffin/pkg/querying"
	"github.com/virtualtam/sparklemuffin/pkg/session"
	"github.com/virtualtam/sparklemuffin/pkg/user"
)

var _ bookmark.Repository = &Repository{}
var _ exporting.Repository = &Repository{}
var _ importing.Repository = &Repository{}
var _ querying.Repository = &Repository{}
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
	dbTags := tagsToTextArray(b.Tags)
	fullTextSearchString := bookmarkToFullTextSearchString(b)

	dbBookmark := Bookmark{
		UID:                  b.UID,
		UserUUID:             b.UserUUID,
		URL:                  b.URL,
		Title:                b.Title,
		Description:          b.Description,
		Private:              b.Private,
		Tags:                 dbTags,
		FullTextSearchString: fullTextSearchString,
		CreatedAt:            b.CreatedAt,
		UpdatedAt:            b.UpdatedAt,
	}

	_, err := r.db.NamedExec(
		`
INSERT INTO bookmarks(
	uid,
	user_uuid,
	url,
	title,
	description,
	private,
	tags,
	fulltextsearch_tsv,
	created_at,
	updated_at
)
VALUES(
	:uid,
	:user_uuid,
	:url,
	:title,
	:description,
	:private,
	:tags,
	to_tsvector(:fulltextsearch_string),
	:created_at,
	:updated_at
)
`,
		dbBookmark,
	)

	return err
}

func (r *Repository) BookmarkAddMany(bookmarks []bookmark.Bookmark) (int64, error) {
	return r.bookmarkUpsertMany("ON CONFLICT DO NOTHING", bookmarks)
}

func (r *Repository) BookmarkUpsertMany(bookmarks []bookmark.Bookmark) (int64, error) {
	return r.bookmarkUpsertMany(
		`
ON CONFLICT (user_uuid, url) DO UPDATE
SET
	title              = EXCLUDED.title,
	description        = EXCLUDED.description,
	private            = EXCLUDED.private,
	tags               = EXCLUDED.tags,
	fulltextsearch_tsv = EXCLUDED.fulltextsearch_tsv,
	created_at         = EXCLUDED.created_at,
	updated_at         = EXCLUDED.updated_at
`,
		bookmarks,
	)
}

func (r *Repository) bookmarkUpsertMany(onConflictStmt string, bookmarks []bookmark.Bookmark) (int64, error) {
	dbBookmarks := make([]Bookmark, len(bookmarks))

	for index, b := range bookmarks {
		dbTags := tagsToTextArray(b.Tags)
		fullTextSearchString := bookmarkToFullTextSearchString(b)

		dbBookmark := Bookmark{
			UID:                  b.UID,
			UserUUID:             b.UserUUID,
			URL:                  b.URL,
			Title:                b.Title,
			Description:          b.Description,
			Private:              b.Private,
			Tags:                 dbTags,
			FullTextSearchString: fullTextSearchString,
			CreatedAt:            b.CreatedAt,
			UpdatedAt:            b.UpdatedAt,
		}
		dbBookmarks[index] = dbBookmark
	}

	insertQuery := `
INSERT INTO bookmarks(
	uid,
	user_uuid,
	url,
	title,
	description,
	private,
	tags,
	fulltextsearch_tsv,
	created_at,
	updated_at
)
VALUES(
	:uid,
	:user_uuid,
	:url,
	:title,
	:description,
	:private,
	:tags,
	to_tsvector(:fulltextsearch_string),
	:created_at,
	:updated_at
)`

	query := insertQuery + onConflictStmt
	res, err := r.db.NamedExec(query, dbBookmarks)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	return rowsAffected, err
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

func (r *Repository) bookmarkGetQuery(query string, queryParams ...any) ([]bookmark.Bookmark, error) {
	rows, err := r.db.Queryx(query, queryParams...)
	if err != nil {
		return []bookmark.Bookmark{}, err
	}

	bookmarks := []bookmark.Bookmark{}

	for rows.Next() {
		dbBookmark := Bookmark{}

		if err := rows.StructScan(&dbBookmark); err != nil {
			return []bookmark.Bookmark{}, err
		}

		tags := textArrayToTags(dbBookmark.Tags)

		bookmark := bookmark.Bookmark{
			UserUUID:    dbBookmark.UserUUID,
			UID:         dbBookmark.UID,
			URL:         dbBookmark.URL,
			Title:       dbBookmark.Title,
			Description: dbBookmark.Description,
			Private:     dbBookmark.Private,
			Tags:        tags,
			CreatedAt:   dbBookmark.CreatedAt,
			UpdatedAt:   dbBookmark.UpdatedAt,
		}

		bookmarks = append(bookmarks, bookmark)
	}

	return bookmarks, nil
}

func (r *Repository) BookmarkGetAll(userUUID string) ([]bookmark.Bookmark, error) {
	return r.bookmarkGetQuery(
		`
SELECT user_uuid, uid, url, title, description, private, tags, created_at, updated_at
FROM bookmarks
WHERE user_uuid=$1
ORDER BY created_at DESC`,
		userUUID,
	)
}

func (r *Repository) BookmarkGetAllPrivate(userUUID string) ([]bookmark.Bookmark, error) {
	return r.bookmarkGetQuery(
		`
SELECT user_uuid, uid, url, title, description, private, tags, created_at, updated_at
FROM bookmarks
WHERE user_uuid=$1
AND   private=TRUE
ORDER BY created_at DESC`,
		userUUID,
	)
}

func (r *Repository) BookmarkGetAllPublic(userUUID string) ([]bookmark.Bookmark, error) {
	return r.bookmarkGetQuery(
		`
SELECT user_uuid, uid, url, title, description, private, tags, created_at, updated_at
FROM bookmarks
WHERE user_uuid=$1
AND   private=FALSE
ORDER BY created_at DESC`,
		userUUID,
	)
}

func (r *Repository) BookmarkGetByTag(userUUID string, tag string) ([]bookmark.Bookmark, error) {
	return r.bookmarkGetQuery(
		`
SELECT user_uuid, uid, url, title, description, private, tags, created_at, updated_at
FROM bookmarks
WHERE user_uuid=$1
AND   $2=ANY(tags)
		`,
		userUUID,
		tag,
	)
}

func (r *Repository) BookmarkGetByUID(userUUID, uid string) (bookmark.Bookmark, error) {
	dbBookmark := &Bookmark{}

	err := r.db.QueryRowx(
		`
SELECT user_uuid, uid, url, title, description, private, tags, created_at, updated_at
FROM bookmarks
WHERE user_uuid=$1
AND uid=$2`,
		userUUID,
		uid,
	).StructScan(dbBookmark)

	if errors.Is(err, sql.ErrNoRows) {
		return bookmark.Bookmark{}, bookmark.ErrNotFound
	}
	if err != nil {
		return bookmark.Bookmark{}, err
	}

	tags := textArrayToTags(dbBookmark.Tags)

	return bookmark.Bookmark{
		UserUUID:    dbBookmark.UserUUID,
		UID:         dbBookmark.UID,
		URL:         dbBookmark.URL,
		Title:       dbBookmark.Title,
		Description: dbBookmark.Description,
		Private:     dbBookmark.Private,
		Tags:        tags,
		CreatedAt:   dbBookmark.CreatedAt,
		UpdatedAt:   dbBookmark.UpdatedAt,
	}, nil
}

func (r *Repository) BookmarkGetCount(userUUID string, visibility querying.Visibility) (uint, error) {
	var query string

	switch visibility {
	case querying.VisibilityPrivate:
		query = `
		SELECT COUNT(*)
		FROM  bookmarks
		WHERE user_uuid=$1
		AND   private=TRUE`

	case querying.VisibilityPublic:
		query = `
		SELECT COUNT(*)
		FROM  bookmarks
		WHERE user_uuid=$1
		AND   private=FALSE`

	default:
		query = `
		SELECT COUNT(*)
		FROM bookmarks
		WHERE user_uuid=$1`
	}

	var count uint

	err := r.db.Get(
		&count,
		query,
		userUUID,
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *Repository) BookmarkGetN(userUUID string, visibility querying.Visibility, n uint, offset uint) ([]bookmark.Bookmark, error) {
	var query string

	switch visibility {
	case querying.VisibilityPrivate:
		query = `
		SELECT user_uuid, uid, url, title, description, private, tags, created_at, updated_at
		FROM  bookmarks
		WHERE user_uuid=$1
		AND   private=TRUE
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	case querying.VisibilityPublic:
		query = `
		SELECT user_uuid, uid, url, title, description, private, tags, created_at, updated_at
		FROM  bookmarks
		WHERE user_uuid=$1
		AND   private=FALSE
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	default:
		query = `
		SELECT user_uuid, uid, url, title, description, private, tags, created_at, updated_at
		FROM  bookmarks
		WHERE user_uuid=$1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`
	}

	return r.bookmarkGetQuery(
		query,
		userUUID,
		n,
		offset,
	)
}

func (r *Repository) BookmarkGetPublicByUID(userUUID, uid string) (bookmark.Bookmark, error) {
	dbBookmark := &Bookmark{}

	err := r.db.QueryRowx(
		`
SELECT user_uuid, uid, url, title, description, private, tags, created_at, updated_at
FROM bookmarks
WHERE user_uuid=$1
AND uid=$2
AND private=false`,
		userUUID,
		uid,
	).StructScan(dbBookmark)

	if errors.Is(err, sql.ErrNoRows) {
		return bookmark.Bookmark{}, bookmark.ErrNotFound
	}
	if err != nil {
		return bookmark.Bookmark{}, err
	}

	tags := textArrayToTags(dbBookmark.Tags)

	return bookmark.Bookmark{
		UserUUID:    dbBookmark.UserUUID,
		UID:         dbBookmark.UID,
		URL:         dbBookmark.URL,
		Title:       dbBookmark.Title,
		Description: dbBookmark.Description,
		Private:     dbBookmark.Private,
		Tags:        tags,
		CreatedAt:   dbBookmark.CreatedAt,
		UpdatedAt:   dbBookmark.UpdatedAt,
	}, nil
}

func (r *Repository) BookmarkSearchCount(userUUID string, visibility querying.Visibility, searchTerms string) (uint, error) {
	var query string

	switch visibility {
	case querying.VisibilityPrivate:
		query = `
		SELECT COUNT(*)
		FROM bookmarks
		WHERE user_uuid=$1
		AND PRIVATE=TRUE
		AND fulltextsearch_tsv @@ websearch_to_tsquery($2)`

	case querying.VisibilityPublic:
		query = `
		SELECT COUNT(*)
		FROM bookmarks
		WHERE user_uuid=$1
		AND PRIVATE=FALSE
		AND fulltextsearch_tsv @@ websearch_to_tsquery($2)`

	default:
		query = `
		SELECT COUNT(*)
		FROM bookmarks
		WHERE user_uuid=$1
		AND fulltextsearch_tsv @@ websearch_to_tsquery($2)`
	}

	var count uint
	fullTextSearchTerms := fullTextSearchReplacer.Replace(searchTerms)

	err := r.db.Get(
		&count,
		query,
		userUUID,
		fullTextSearchTerms,
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *Repository) BookmarkSearchN(userUUID string, visibility querying.Visibility, searchTerms string, n uint, offset uint) ([]bookmark.Bookmark, error) {
	var query string

	switch visibility {
	case querying.VisibilityPrivate:
		query = `
		SELECT user_uuid, uid, url, title, description, private, tags, created_at, updated_at
		FROM bookmarks
		WHERE user_uuid=$1
		AND private=TRUE
		AND fulltextsearch_tsv @@ websearch_to_tsquery($2)
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`

	case querying.VisibilityPublic:
		query = `
		SELECT user_uuid, uid, url, title, description, private, tags, created_at, updated_at
		FROM bookmarks
		WHERE user_uuid=$1
		AND private=FALSE
		AND fulltextsearch_tsv @@ websearch_to_tsquery($2)
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`

	default:
		query = `
		SELECT user_uuid, uid, url, title, description, private, tags, created_at, updated_at
		FROM bookmarks
		WHERE user_uuid=$1
		AND fulltextsearch_tsv @@ websearch_to_tsquery($2)
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`
	}

	fullTextSearchTerms := fullTextSearchReplacer.Replace(searchTerms)

	return r.bookmarkGetQuery(
		query,
		userUUID,
		fullTextSearchTerms,
		n,
		offset,
	)
}

func (r *Repository) BookmarkIsURLRegistered(userUUID, url string) (bool, error) {
	dbBookmark := &Bookmark{}

	err := r.db.QueryRowx(
		"SELECT url FROM bookmarks WHERE user_uuid=$1 AND url=$2",
		userUUID,
		url,
	).StructScan(dbBookmark)

	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

func (r *Repository) BookmarkIsURLRegisteredToAnotherUID(userUUID, url, uid string) (bool, error) {
	dbBookmark := &Bookmark{}

	err := r.db.QueryRowx(
		"SELECT url FROM bookmarks WHERE user_uuid=$1 AND url=$2 AND uid!=$3",
		userUUID,
		url,
		uid,
	).StructScan(dbBookmark)

	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

func (r *Repository) BookmarkTagUpdateMany(bookmarks []bookmark.Bookmark) (int64, error) {
	// sqlx does not support PostgreSQL's bulk update syntax.
	// As a workaround, we perform a bulk upsert to update existing bookmarks.
	// https://github.com/jmoiron/sqlx/issues/796
	return r.BookmarkUpsertMany(bookmarks)
}

func (r *Repository) BookmarkUpdate(b bookmark.Bookmark) error {
	dbTags := tagsToTextArray(b.Tags)
	fullTextSearchString := bookmarkToFullTextSearchString(b)

	dbBookmark := Bookmark{
		UserUUID:             b.UserUUID,
		UID:                  b.UID,
		URL:                  b.URL,
		Title:                b.Title,
		Description:          b.Description,
		Private:              b.Private,
		Tags:                 dbTags,
		FullTextSearchString: fullTextSearchString,
		UpdatedAt:            b.UpdatedAt,
	}

	_, err := r.db.NamedExec(
		`
UPDATE bookmarks
SET
	url=:url,
	title=:title,
	description=:description,
	private=:private,
	tags=:tags,
	fulltextsearch_tsv=to_tsvector(:fulltextsearch_string),
	updated_at=:updated_at
WHERE user_uuid=:user_uuid
AND uid=:uid
		`,
		dbBookmark,
	)
	return err
}

func (r *Repository) OwnerGetByUUID(userUUID string) (querying.Owner, error) {
	dbUser := &User{}

	err := r.db.QueryRowx(
		`SELECT uuid, nick_name, display_name
FROM users
WHERE uuid=$1`,
		userUUID,
	).StructScan(dbUser)

	if errors.Is(err, sql.ErrNoRows) {
		return querying.Owner{}, querying.ErrOwnerNotFound
	}
	if err != nil {
		return querying.Owner{}, err
	}

	return querying.Owner{
		UUID:        dbUser.UUID,
		NickName:    dbUser.NickName,
		DisplayName: dbUser.DisplayName,
	}, nil
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

func (r *Repository) tagGetQuery(query string, queryParams ...any) ([]querying.Tag, error) {
	rows, err := r.db.Queryx(query, queryParams...)
	if err != nil {
		return []querying.Tag{}, err
	}

	tags := []querying.Tag{}

	for rows.Next() {
		dbTag := Tag{}

		if err := rows.StructScan(&dbTag); err != nil {
			return []querying.Tag{}, err
		}

		tag := querying.NewTag(dbTag.Name, dbTag.Count)

		tags = append(tags, tag)
	}

	return tags, nil
}

func (r *Repository) TagGetCount(userUUID string, visibility querying.Visibility) (uint, error) {
	var query string

	switch visibility {
	case querying.VisibilityPrivate:
		query = `
		SELECT COUNT(DISTINCT name)
		FROM (
			SELECT unnest(tags) as name
			FROM bookmarks
			WHERE user_uuid=$1
			AND   private=TRUE
		) s`

	case querying.VisibilityPublic:
		query = `
		SELECT COUNT(DISTINCT name)
		FROM (
			SELECT unnest(tags) as name
			FROM bookmarks
			WHERE user_uuid=$1
			AND   private=FALSE
		) s`

	default:
		query = `
		SELECT COUNT(DISTINCT name)
		FROM (
			SELECT unnest(tags) as name
			FROM bookmarks
			WHERE user_uuid=$1
		) s`
	}

	var count uint

	err := r.db.Get(
		&count,
		query,
		userUUID,
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *Repository) TagGetAll(userUUID string, visibility querying.Visibility) ([]querying.Tag, error) {
	var query string

	switch visibility {
	case querying.VisibilityPrivate:
		query = `
		SELECT name, COUNT(name) as count
		FROM (
			SELECT unnest(tags) as name
			FROM  bookmarks
			WHERE user_uuid=$1
			AND   private=TRUE
		) s
		GROUP BY name
		ORDER BY count DESC`

	case querying.VisibilityPublic:
		query = `
		SELECT name, COUNT(name) as count
		FROM (
			SELECT unnest(tags) as name
			FROM  bookmarks
			WHERE user_uuid=$1
			AND   private=FALSE
		) s
		GROUP BY name
		ORDER BY count DESC`

	default:
		query = `
		SELECT name, COUNT(name) as count
		FROM (
			SELECT unnest(tags) as name
			FROM  bookmarks
			WHERE user_uuid=$1
		) s
		GROUP BY name
		ORDER BY count DESC`
	}

	return r.tagGetQuery(query, userUUID)
}

func (r *Repository) TagGetN(userUUID string, visibility querying.Visibility, n uint, offset uint) ([]querying.Tag, error) {
	var query string

	switch visibility {
	case querying.VisibilityPrivate:
		query = `
		SELECT name, COUNT(name) as count
		FROM (
			SELECT unnest(tags) as name
			FROM  bookmarks
			WHERE user_uuid=$1
			AND   private=TRUE
		) s
		GROUP BY name
		ORDER BY count DESC
		LIMIT $2 OFFSET $3`

	case querying.VisibilityPublic:
		query = `
		SELECT name, COUNT(name) as count
		FROM (
			SELECT unnest(tags) as name
			FROM  bookmarks
			WHERE user_uuid=$1
			AND   private=FALSE
		) s
		GROUP BY name
		ORDER BY count DESC
		LIMIT $2 OFFSET $3`

	default:
		query = `
		SELECT name, COUNT(name) as count
		FROM (
			SELECT unnest(tags) as name
			FROM  bookmarks
			WHERE user_uuid=$1
		) s
		GROUP BY name
		ORDER BY count DESC
		LIMIT $2 OFFSET $3`
	}

	return r.tagGetQuery(query, userUUID, n, offset)
}

func (r *Repository) TagFilterCount(userUUID string, visibility querying.Visibility, filterTerm string) (uint, error) {
	var query string

	switch visibility {
	case querying.VisibilityPrivate:
		query = `
		SELECT COUNT(DISTINCT name)
		FROM (
			SELECT unnest(tags) as name
			FROM bookmarks
			WHERE user_uuid=$1
			AND   private=TRUE
		) s
		WHERE name ILIKE $2`

	case querying.VisibilityPublic:
		query = `
		SELECT COUNT(DISTINCT name)
		FROM (
			SELECT unnest(tags) as name
			FROM bookmarks
			WHERE user_uuid=$1
			AND   private=FALSE
		) s
		WHERE name ILIKE $2`

	default:
		query = `
		SELECT COUNT(DISTINCT name)
		FROM (
			SELECT unnest(tags) as name
			FROM bookmarks
			WHERE user_uuid=$1
		) s
		WHERE name ILIKE $2`
	}

	var count uint

	err := r.db.Get(
		&count,
		query,
		userUUID,
		"%"+filterTerm+"%",
	)
	if err != nil {
		log.Warn().Msg("plop")
		return 0, err
	}

	return count, nil
}

func (r *Repository) TagFilterN(userUUID string, visibility querying.Visibility, filterTerm string, n uint, offset uint) ([]querying.Tag, error) {
	var query string

	switch visibility {
	case querying.VisibilityPrivate:
		query = `
		SELECT name, COUNT(name) as count
		FROM (
			SELECT unnest(tags) as name
			FROM  bookmarks
			WHERE user_uuid=$1
			AND   private=TRUE
		) s
		WHERE name ILIKE $2
		GROUP BY name
		ORDER BY count DESC
		LIMIT $3 OFFSET $4`

	case querying.VisibilityPublic:
		query = `
		SELECT name, COUNT(name) as count
		FROM (
			SELECT unnest(tags) as name
			FROM  bookmarks
			WHERE user_uuid=$1
			AND   private=FALSE
		) s
		WHERE name ILIKE $2
		GROUP BY name
		ORDER BY count DESC
		LIMIT $3 OFFSET $4`

	default:
		query = `
		SELECT name, COUNT(name) as count
		FROM (
			SELECT unnest(tags) as name
			FROM  bookmarks
			WHERE user_uuid=$1
		) s
		WHERE name ILIKE $2
		GROUP BY name
		ORDER BY count DESC
		LIMIT $3 OFFSET $4`
	}

	return r.tagGetQuery(query, userUUID, "%"+filterTerm+"%", n, offset)
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
