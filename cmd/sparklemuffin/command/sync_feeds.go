// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package command

import (
	"github.com/carlmjohnson/versioninfo"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// NewSyncFeedsCommand initializes and returns a new CLI command to synchronize syndication feeds.
func NewSyncFeedsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync-feeds",
		Short: "Synchronize syndication feeds",
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Info().
				Str("log_level", logLevelValue).
				Str("version", versioninfo.Short()).
				Msg("feeds: synchronizing")

			return feedSynchronizingService.Run()
		},
	}

	return cmd
}
