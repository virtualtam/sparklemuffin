package www

// flashLevel represents the severity level of a Flash message that will be
// displayed to the user.
type flashLevel string

const (
	flashLevelError   flashLevel = "error"
	flashLevelInfo    flashLevel = "info"
	flashLevelSuccess flashLevel = "success"
	flashLevelWarning flashLevel = "warning"
)

// Flash represents a message that wilshl be displayed to the user when
// rendering a View.
type Flash struct {
	Level   flashLevel
	Message string
}
