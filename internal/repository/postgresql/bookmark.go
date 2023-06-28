package postgresql

import (
	"fmt"
	"strings"
	"time"

	"github.com/virtualtam/sparklemuffin/pkg/bookmark"
)

var (
	fullTextSearchReplacer = strings.NewReplacer("/", " ", ".", " ")
)

type Bookmark struct {
	UID      string `db:"uid"`
	UserUUID string `db:"user_uuid"`

	URL         string `db:"url"`
	Title       string `db:"title"`
	Description string `db:"description"`

	Private bool     `db:"private"`
	Tags    []string `db:"tags"`

	FullTextSearchString string `db:"fulltextsearch_string"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func bookmarkToFullTextSearchString(b bookmark.Bookmark) string {
	return fmt.Sprintf(
		"%s %s %s",
		b.Title,
		fullTextSearchReplacer.Replace(b.Description),
		fullTextSearchReplacer.Replace(strings.Join(b.Tags, " ")),
	)
}
