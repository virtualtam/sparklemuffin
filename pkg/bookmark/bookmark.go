package bookmark

import (
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/segmentio/ksuid"
)

// Bookmark represents a Web bookmark.
type Bookmark struct {
	UID      string
	UserUUID string

	URL         string
	Title       string
	Description string

	Private bool
	Tags    []string

	CreatedAt time.Time
	UpdatedAt time.Time
}

// Normalize sanitizes and normalizes all fields.
func (b *Bookmark) Normalize() {
	b.normalizeURL()
	b.normalizeTitle()
	b.normalizeDescription()
	b.normalizeTags()
	b.deduplicateTags()
	b.sortTags()
}

// ValidateForAddition ensures mandatory fields are properly set when adding an
// new Bookmark.
func (b *Bookmark) ValidateForAddition(r ValidationRepository) error {
	fns := []func() error{
		b.requireUserUUID,
		b.requireURL,
		b.ensureURLIsParseable,
		b.ensureURLIsNotRegistered(r),
		b.requireTitle,
		b.validateUID,
	}

	for _, fn := range fns {
		if err := fn(); err != nil {
			return err
		}
	}

	return nil
}

// ValidateForUpdate ensures mandatory fields are properly set when updating an
// existing Bookmark.
func (b *Bookmark) ValidateForUpdate(r Repository) error {
	fns := []func() error{
		b.requireUID,
		b.validateUID,
		b.requireUserUUID,
		b.requireURL,
		b.ensureURLIsParseable,
		b.ensureURLIsNotRegisteredToAnotherBookmark(r),
		b.requireTitle,
	}

	for _, fn := range fns {
		if err := fn(); err != nil {
			return err
		}
	}

	return nil
}

func (b *Bookmark) normalizeDescription() {
	b.Description = strings.TrimSpace(b.Description)
}

func (b *Bookmark) normalizeTags() {
	tags := []string{}

	for _, tag := range b.Tags {
		tag := strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		tags = append(tags, tag)
	}

	b.Tags = tags
}

func (b *Bookmark) deduplicateTags() {
	tagNames := map[string]bool{}
	tags := []string{}

	for _, tag := range b.Tags {
		_, exists := tagNames[tag]

		if exists {
			continue
		}

		tagNames[tag] = true
		tags = append(tags, tag)
	}

	b.Tags = tags
}

func (b *Bookmark) sortTags() {
	sort.Strings(b.Tags)
}

func (b *Bookmark) normalizeTitle() {
	b.Title = strings.TrimSpace(b.Title)
}

func (b *Bookmark) normalizeURL() {
	b.URL = strings.TrimSpace(b.URL)
}

func (b *Bookmark) generateUID() {
	b.UID = ksuid.New().String()
}

func (b *Bookmark) ensureURLIsParseable() error {
	_, err := url.Parse(b.URL)
	if err != nil {
		return ErrURLInvalid
	}

	return nil
}

func (b *Bookmark) ensureURLIsNotRegistered(r ValidationRepository) func() error {
	return func() error {
		registered, err := r.BookmarkIsURLRegistered(b.UserUUID, b.URL)
		if err != nil {
			return err
		}
		if registered {
			return ErrURLAlreadyRegistered
		}
		return nil
	}
}

func (b *Bookmark) ensureURLIsNotRegisteredToAnotherBookmark(r ValidationRepository) func() error {
	return func() error {
		registered, err := r.BookmarkIsURLRegisteredToAnotherUID(b.UserUUID, b.URL, b.UID)

		if err != nil {
			return err
		}
		if registered {
			return ErrURLAlreadyRegistered
		}
		return nil
	}
}

func (b *Bookmark) requireUID() error {
	if b.UID == "" {
		return ErrUIDRequired
	}
	return nil
}

func (b *Bookmark) requireTitle() error {
	if b.Title == "" {
		return ErrTitleRequired
	}
	return nil
}

func (b *Bookmark) requireURL() error {
	if b.URL == "" {
		return ErrURLRequired
	}
	return nil
}

func (b *Bookmark) requireUserUUID() error {
	if b.UserUUID == "" {
		return ErrUserUUIDRequired
	}
	return nil
}

func (b *Bookmark) validateUID() error {
	_, err := ksuid.Parse(b.UID)
	if err != nil {
		return ErrUIDInvalid
	}

	return nil
}
