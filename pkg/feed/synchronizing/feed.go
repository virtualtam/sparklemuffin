package synchronizing

import "time"

type FeedFetchMetadata struct {
	UUID string

	ETag string

	UpdatedAt time.Time
	FetchedAt time.Time
}
