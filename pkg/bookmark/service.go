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

func (s *Service) DeleteTag(dq TagDeleteQuery) (int64, error) {
	now := time.Now().UTC()

	dq.normalize()

	fns := []func() error{
		dq.requireUserUUID,
		dq.requireName,
		dq.ensureNameHasNoWhitespace,
	}

	for _, fn := range fns {
		if err := fn(); err != nil {
			return 0, err
		}
	}

	bookmarks, err := s.r.BookmarkGetByTag(dq.UserUUID, dq.Name)
	if err != nil {
		return 0, err
	}

	for i, bookmark := range bookmarks {
		for j, bookmarkTag := range bookmark.Tags {
			if bookmarkTag == dq.Name {
				bookmark.Tags = append(bookmark.Tags[:j], bookmark.Tags[j+1:]...)
				break
			}
		}

		bookmark.UpdatedAt = now

		bookmarks[i] = bookmark
	}

	return s.r.BookmarkTagUpdateMany(bookmarks)
}

func (s *Service) UpdateTag(uq TagUpdateQuery) (int64, error) {
	now := time.Now().UTC()

	uq.normalize()

	fns := []func() error{
		uq.requireUserUUID,
		uq.requireCurrentName,
		uq.ensureCurrentNameHasNoWhitespace,
		uq.requireNewName,
		uq.ensureNewNameHasNoWhitespace,
		uq.ensureNewNameIsNotEqualToCurrentName,
	}

	for _, fn := range fns {
		if err := fn(); err != nil {
			return 0, err
		}
	}

	bookmarks, err := s.r.BookmarkGetByTag(uq.UserUUID, uq.CurrentName)
	if err != nil {
		return 0, err
	}

	for i, bookmark := range bookmarks {
		for j, bookmarkTag := range bookmark.Tags {
			if bookmarkTag == uq.CurrentName {
				bookmark.Tags[j] = uq.NewName
			}
		}

		bookmark.deduplicateTags()
		bookmark.sortTags()
		bookmark.UpdatedAt = now

		bookmarks[i] = bookmark
	}

	return s.r.BookmarkTagUpdateMany(bookmarks)
}
