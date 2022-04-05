package bookmark

import (
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/segmentio/ksuid"
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
	err := s.runValidationFuncs(
		&bookmark,
		s.requireUserUUID,
		s.normalizeURL,
		s.requireURL,
		s.ensureURLIsParseable,
		s.ensureURLIsNotRegistered,
		s.normalizeTitle,
		s.requireTitle,
		s.normalizeDescription,
		s.generateUID,
		s.validateUID,
		s.setCreatedUpdatedAt,
	)
	if err != nil {
		return err
	}

	return s.r.BookmarkAdd(bookmark)
}

func (s *Service) All(userUUID string) ([]Bookmark, error) {
	return s.r.BookmarkGetAll(userUUID)
}

func (s *Service) ByUID(userUUID string, uid string) (Bookmark, error) {
	bookmark := Bookmark{
		UserUUID: userUUID,
		UID:      uid,
	}

	err := s.runValidationFuncs(
		&bookmark,
		s.requireUID,
		s.validateUID,
		s.requireUserUUID,
	)
	if err != nil {
		return Bookmark{}, err
	}

	return s.r.BookmarkGetByUID(userUUID, uid)
}

func (s *Service) Delete(userUUID, uid string) error {
	bookmark := Bookmark{
		UserUUID: userUUID,
		UID:      uid,
	}

	err := s.runValidationFuncs(
		&bookmark,
		s.requireUID,
		s.validateUID,
		s.requireUserUUID,
	)
	if err != nil {
		return err
	}

	return s.r.BookmarkDelete(userUUID, uid)
}

func (s *Service) Update(bookmark Bookmark) error {
	err := s.runValidationFuncs(
		&bookmark,
		s.requireUID,
		s.validateUID,
		s.requireUserUUID,
		s.normalizeURL,
		s.requireURL,
		s.ensureURLIsParseable,
		s.ensureURLIsNotRegisteredToAnotherBookmark,
		s.normalizeTitle,
		s.requireTitle,
		s.normalizeDescription,
		s.refreshUpdatedAt,
	)
	if err != nil {
		return err
	}

	return s.r.BookmarkUpdate(bookmark)
}

func (s *Service) ensureURLIsParseable(bookmark *Bookmark) error {
	_, err := url.Parse(bookmark.URL)
	if err != nil {
		return ErrURLInvalid
	}

	return nil
}

func (s *Service) ensureURLIsNotRegistered(bookmark *Bookmark) error {
	registered, err := s.r.BookmarkIsURLRegistered(bookmark.UserUUID, bookmark.URL)
	if err != nil {
		return err
	}
	if registered {
		return ErrURLAlreadyRegistered
	}
	return nil
}

func (s *Service) ensureURLIsNotRegisteredToAnotherBookmark(bookmark *Bookmark) error {
	existing, err := s.r.BookmarkGetByURL(bookmark.UserUUID, bookmark.URL)
	if errors.Is(err, ErrNotFound) {
		return nil
	}

	if existing.UserUUID == bookmark.UserUUID && existing.UID == bookmark.UID {
		return nil
	}

	return ErrURLAlreadyRegistered
}

func (s *Service) generateUID(bookmark *Bookmark) error {
	bookmark.UID = ksuid.New().String()
	return nil
}

func (s *Service) normalizeDescription(bookmark *Bookmark) error {
	bookmark.Description = strings.TrimSpace(bookmark.Description)
	return nil
}

func (s *Service) normalizeTitle(bookmark *Bookmark) error {
	bookmark.Title = strings.TrimSpace(bookmark.Title)
	return nil
}

func (s *Service) normalizeURL(bookmark *Bookmark) error {
	bookmark.URL = strings.TrimSpace(bookmark.URL)
	return nil
}

func (s *Service) refreshUpdatedAt(bookmark *Bookmark) error {
	bookmark.UpdatedAt = time.Now().UTC()
	return nil
}

func (s *Service) requireUID(bookmark *Bookmark) error {
	if bookmark.UID == "" {
		return ErrUIDRequired
	}
	return nil
}

func (s *Service) requireTitle(bookmark *Bookmark) error {
	if bookmark.Title == "" {
		return ErrTitleRequired
	}
	return nil
}

func (s *Service) requireURL(bookmark *Bookmark) error {
	if bookmark.URL == "" {
		return ErrURLRequired
	}
	return nil
}

func (s *Service) requireUserUUID(bookmark *Bookmark) error {
	if bookmark.UserUUID == "" {
		return ErrUserUUIDRequired
	}
	return nil
}

func (s *Service) setCreatedUpdatedAt(bookmark *Bookmark) error {
	now := time.Now().UTC()
	bookmark.CreatedAt = now
	bookmark.UpdatedAt = now

	return nil
}

func (s *Service) validateUID(bookmark *Bookmark) error {
	_, err := ksuid.Parse(bookmark.UID)
	if err != nil {
		return ErrUIDInvalid
	}

	return nil
}

// validationFunc defines a function that can be applied to normalize or
// validate Bookmark data.
type validationFunc func(*Bookmark) error

// runValidationFuncs applies Bookmark normalization and validation functions and
// stops at the first encountered error.
func (s *Service) runValidationFuncs(bookmark *Bookmark, fns ...validationFunc) error {
	for _, fn := range fns {
		if err := fn(bookmark); err != nil {
			return err
		}
	}
	return nil
}
