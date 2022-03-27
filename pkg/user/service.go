package user

import "errors"

// Service handles operations for the user domain.
type Service struct {
	r Repository
}

// NewService initializes and returns a User Repository.
func NewService(r Repository) *Service {
	return &Service{
		r: r,
	}
}

func (s *Service) AuthenticateUser(email, password string) (User, error) {
	return User{}, errors.New("not implemented")
}
