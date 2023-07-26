package csrf

import (
	"time"

	"golang.org/x/net/xsrftoken"
)

const (
	// CSRF tokens older than this duration will be considered invalid.
	defaultCSRFTokenTimeout time.Duration = 1 * time.Hour
)

// Service handles CSRF token generation and validation operations.
type Service struct {
	key     string
	timeout time.Duration
}

// NewService initializes and returns a new Service.
func NewService(csrfKey string) *Service {
	return &Service{
		key:     csrfKey,
		timeout: defaultCSRFTokenTimeout,
	}
}

// Generate generates a CSRF token for a given user and action.
func (cs *Service) Generate(userUUID string, actionID string) string {
	return xsrftoken.Generate(cs.key, userUUID, actionID)
}

// Validate validates a CSRF token for a given user and action.
func (cs *Service) Validate(token string, userUUID string, actionID string) bool {
	return xsrftoken.ValidFor(token, cs.key, userUUID, actionID, cs.timeout)
}
