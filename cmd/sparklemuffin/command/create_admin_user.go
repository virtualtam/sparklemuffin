// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package command

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/virtualtam/sparklemuffin/internal/rand"
	"github.com/virtualtam/sparklemuffin/pkg/user"
)

// NewCreateAdminUserCommand initializes a CLI command to create a user
// with administration privileges.
func NewCreateAdminUserCommand() *cobra.Command {
	const (
		adminUserPasswordNBytes = 32
	)

	var (
		email       string
		nickname    string
		displayName string
	)

	cmd := &cobra.Command{
		Use:   "createadmin",
		Short: "Create a user with administration privileges",
		RunE: func(cmd *cobra.Command, args []string) error {
			adminUserPassword, err := rand.RandomBase64URLString(adminUserPasswordNBytes)
			if err != nil {
				log.Error().Err(err).Msg("failed to generate admin user password")
				return err
			}

			adminUser, err := user.NewAdminUser(email, nickname, displayName, adminUserPassword)
			if err != nil {
				log.Error().Err(err).Msg("failed to create admin user")
				return err
			}

			if err := userService.Add(context.Background(), adminUser); err != nil {
				log.Error().Err(err).Msg("failed to create admin user")
				return err
			}

			log.Info().
				Str("email", email).
				Str("nickname", nickname).
				Msg("admin user successfully created")

			// Display generated password on the CLI
			fmt.Println("Generated password:", adminUserPassword)

			return nil
		},
	}

	cmd.Flags().StringVar(
		&email,
		"email",
		"",
		"Email address",
	)
	if err := cmd.MarkFlagRequired("email"); err != nil {
		log.Fatal().Err(err).Msg("failed to mark flag as required")
	}

	cmd.Flags().StringVar(
		&nickname,
		"nickname",
		"",
		"User nickname",
	)
	if err := cmd.MarkFlagRequired("nickname"); err != nil {
		log.Fatal().Err(err).Msg("failed to mark flag as required")
	}

	cmd.Flags().StringVar(
		&displayName,
		"displayname",
		"",
		"User display name",
	)
	if err := cmd.MarkFlagRequired("displayname"); err != nil {
		log.Fatal().Err(err).Msg("failed to mark flag as required")
	}

	return cmd
}
