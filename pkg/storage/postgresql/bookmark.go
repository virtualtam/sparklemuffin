package postgresql

import (
	"time"

	"github.com/jackc/pgtype"
)

type Bookmark struct {
	UID      string `db:"uid"`
	UserUUID string `db:"user_uuid"`

	URL         string `db:"url"`
	Title       string `db:"title"`
	Description string `db:"description"`

	Tags pgtype.TextArray

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func tagsToTextArray(tags []string) pgtype.TextArray {
	dbTags := pgtype.TextArray{
		Dimensions: []pgtype.ArrayDimension{
			{
				Length:     int32(len(tags)),
				LowerBound: 1,
			},
		},
		Status: pgtype.Present,
	}

	for _, tag := range tags {
		dbText := pgtype.Text{String: tag, Status: pgtype.Present}
		dbTags.Elements = append(dbTags.Elements, dbText)
	}

	return dbTags
}

func textArrayToTags(dbTags pgtype.TextArray) []string {
	tags := []string{}

	for _, dbText := range dbTags.Elements {
		tags = append(tags, dbText.String)
	}

	return tags
}
