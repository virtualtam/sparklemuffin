package view

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"
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

// flash represents a message that wilshl be displayed to the user when
// rendering a View.
type flash struct {
	Level   flashLevel `json:"level"`
	Message string     `json:"message"`
}

func (f *flash) base64URLDecode(src string) error {
	flashBytes, err := base64.URLEncoding.DecodeString(src)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(flashBytes, f); err != nil {
		return err
	}

	return nil
}

func (f *flash) base64URLEncode() (string, error) {
	flashJSON, err := json.Marshal(f)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(flashJSON), nil
}

// putFlash sets a session flash message to be displayed by the next rendered
// view.
func putFlash(w http.ResponseWriter, level flashLevel, message string) {
	flash := flash{
		Level:   level,
		Message: message,
	}

	encoded, err := flash.base64URLEncode()
	if err != nil {
		log.Error().Err(err).Msg("failed to put flash cookie")
		return
	}

	cookie := &http.Cookie{
		Name:     flashCookieName,
		Path:     "/",
		Value:    encoded,
		HttpOnly: true,
	}

	http.SetCookie(w, cookie)
}

// PutFlashError sets a Flash that will be rendered as an error message.
func PutFlashError(w http.ResponseWriter, message string) {
	putFlash(w, flashLevelError, fmt.Sprintf("Error: %s", message))
}

// PutFlashInfo sets a Flash that will be rendered as an information message.
func PutFlashInfo(w http.ResponseWriter, message string) {
	putFlash(w, flashLevelInfo, message)
}

// PutFlashSuccess sets a Flash that will be rendered as a success message.
func PutFlashSuccess(w http.ResponseWriter, message string) {
	putFlash(w, flashLevelSuccess, message)
}

// PutFlashWarning sets a Flash that will be rendered as a warning message.
func PutFlashWarning(w http.ResponseWriter, message string) {
	putFlash(w, flashLevelWarning, fmt.Sprintf("Warning: %s", message))
}
