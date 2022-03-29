package command

import (
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/virtualtam/yawbe/pkg/storage/memory"
	"github.com/virtualtam/yawbe/pkg/user"
)

const (
	defaultDebugMode = false
	defaultHMACKey   = "hmac-secret-key"
)

var (
	debugMode   bool
	hmacKey     string
	userService *user.Service
)

// NewRootCommand initializes the main CLI entrypoint and common command flags.
func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "yawbe-srv",
		Short: "Yet Another Web Bookmarking Engine",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if debugMode {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			}

			userRepository := &memory.Repository{}
			userService = user.NewService(userRepository, hmacKey)

			return nil
		},
	}

	cmd.PersistentFlags().BoolVar(
		&debugMode,
		"debug",
		defaultDebugMode,
		"Enable debugging",
	)

	cmd.PersistentFlags().StringVar(
		&hmacKey,
		"hmac-key",
		defaultHMACKey,
		"Secret HMAC key for token hashing",
	)

	return cmd
}
