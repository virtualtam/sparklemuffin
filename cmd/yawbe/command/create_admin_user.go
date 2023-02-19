package command

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/virtualtam/yawbe/internal/rand"
	"github.com/virtualtam/yawbe/pkg/user"
)

const (
	adminUserPasswordNBytes = 32
)

var (
	adminUserEmail       string
	adminUserNickname    string
	adminUserDisplayName string
)

// NewCreateAdminUserCommand initializes a CLI command to create a user
// with administration privileges.
func NewCreateAdminUserCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "createadmin",
		Short: "Create a user with administration privileges",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: generate strong password
			adminUserPassword, err := rand.RandomBase64URLString(adminUserPasswordNBytes)
			if err != nil {
				log.Error().Err(err).Msg("failed to generate admin user password")
				return err
			}

			adminUser := user.User{
				Email:       adminUserEmail,
				NickName:    adminUserNickname,
				DisplayName: adminUserDisplayName,
				Password:    adminUserPassword,
				IsAdmin:     true,
			}

			if err := userService.Add(adminUser); err != nil {
				log.Error().Err(err).Msg("failed to create admin user")
				return err
			}

			log.Info().
				Str("email", adminUserEmail).
				Str("nickname", adminUserNickname).
				Msg("admin user successfully created")

			// Display generated password on the CLI
			fmt.Println("Generated password:", adminUserPassword)

			return nil
		},
	}

	cmd.Flags().StringVar(
		&adminUserEmail,
		"email",
		"",
		"Email address",
	)
	cmd.MarkFlagRequired("email")

	cmd.Flags().StringVar(
		&adminUserNickname,
		"nickname",
		"",
		"User nickname",
	)
	cmd.MarkFlagRequired("nickname")

	cmd.Flags().StringVar(
		&adminUserDisplayName,
		"displayname",
		"",
		"User display name",
	)
	cmd.MarkFlagRequired("displayname")

	return cmd
}
