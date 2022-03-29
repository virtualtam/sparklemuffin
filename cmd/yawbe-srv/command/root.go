package command

import (
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/virtualtam/yawbe/pkg/storage/postgresql"
	"github.com/virtualtam/yawbe/pkg/user"

	_ "github.com/jackc/pgx/v4/stdlib"
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

			db, err := sqlx.Connect("pgx", "postgres://yawbe:yawbe@localhost:15432/yawbe")
			if err != nil {
				log.Error().Err(err)
				return err
			}
			log.Info().Msg("Successfully connected to PostgreSQL")

			userRepository := postgresql.NewRepository(db)
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
