package www

import (
	"encoding/base64"
	"encoding/json"
)

// flashLevel represents the severity level of a Flash message that will be
// displayed to the user.
type flashLevel string

const (
	flashLevelError   flashLevel = "error"
	flashLevelInfo    flashLevel = "info"
	flashLevelSuccess flashLevel = "success"
	flashLevelWarning flashLevel = "warning"
)

const (
	flashCookieName string = "flash"
)

// Flash represents a message that wilshl be displayed to the user when
// rendering a View.
type Flash struct {
	Level   flashLevel `json:"level"`
	Message string     `json:"message"`
}

func (f *Flash) base64URLDecode(src string) error {
	flashBytes, err := base64.URLEncoding.DecodeString(src)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(flashBytes, f); err != nil {
		return err
	}

	return nil
}

func (f *Flash) base64URLEncode() (string, error) {
	flashJSON, err := json.Marshal(f)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(flashJSON), nil
}
