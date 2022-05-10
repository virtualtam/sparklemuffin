package bookmark

import "time"

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
