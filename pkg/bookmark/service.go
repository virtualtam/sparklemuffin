package bookmark

import (
	"time"
)

// Service handles operations for the bookmark domain.
type Service struct {
	r Repository
}

// NewService initializes and returns a Bookmark Service.
func NewService(r Repository) *Service {
	return &Service{
		r: r,
	}
}

func (s *Service) Add(bookmark Bookmark) error {
	now := time.Now().UTC()

	bookmark.generateUID()
	bookmark.CreatedAt = now
	bookmark.UpdatedAt = now

	bookmark.Normalize()

	if err := bookmark.ValidateForAddition(s.r); err != nil {
		return err
	}

	return s.r.BookmarkAdd(bookmark)
}

func (s *Service) All(userUUID string) ([]Bookmark, error) {
	return s.r.BookmarkGetAll(userUUID)
}

func (s *Service) ByUID(userUUID string, uid string) (Bookmark, error) {
	b := &Bookmark{
		UserUUID: userUUID,
		UID:      uid,
	}

	fns := []func() error{
		b.requireUID,
		b.validateUID,
		b.requireUserUUID,
	}

	for _, fn := range fns {
		if err := fn(); err != nil {
			return Bookmark{}, err
		}
	}

	return s.r.BookmarkGetByUID(userUUID, uid)
}

func (s *Service) Delete(userUUID, uid string) error {
	b := Bookmark{
		UserUUID: userUUID,
		UID:      uid,
	}

	fns := []func() error{
		b.requireUID,
		b.validateUID,
		b.requireUserUUID,
	}

	for _, fn := range fns {
		if err := fn(); err != nil {
			return err
		}
	}

	return s.r.BookmarkDelete(userUUID, uid)
}

func (s *Service) Update(bookmark Bookmark) error {
	now := time.Now().UTC()
	bookmark.UpdatedAt = now

	bookmark.Normalize()

	if err := bookmark.ValidateForUpdate(s.r); err != nil {
		return err
	}

	return s.r.BookmarkUpdate(bookmark)
}

func (s *Service) UpdateTag(tagNameUpdate TagNameUpdate) (int64, error) {
	now := time.Now().UTC()

	// todo normalize tag name
	// todo check non-empty
	// todo ensure single string
	// todo check not equal to current name

	bookmarks, err := s.r.BookmarkGetByTag(tagNameUpdate.UserUUID, tagNameUpdate.CurrentName)
	if err != nil {
		return 0, err
	}

	for i, bookmark := range bookmarks {
		for j, name := range bookmark.Tags {
			if name == tagNameUpdate.CurrentName {
				bookmark.Tags[j] = tagNameUpdate.NewName
			}
		}

		bookmark.normalizeTags()
		bookmark.UpdatedAt = now

		bookmarks[i] = bookmark
	}

	return s.r.BookmarkTagUpdateMany(bookmarks)
}
