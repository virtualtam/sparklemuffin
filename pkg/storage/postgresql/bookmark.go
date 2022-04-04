package postgresql

import "time"

type Bookmark struct {
	UID      string `db:"uid"`
	UserUUID string `db:"user_uuid"`

	URL         string `db:"url"`
	Title       string `db:"title"`
	Description string `db:"description"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
