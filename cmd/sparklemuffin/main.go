// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package main

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/virtualtam/sparklemuffin/cmd/sparklemuffin/command"
)

func main() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	})

	rootCommand := command.NewRootCommand()

	commands := []*cobra.Command{
		command.NewCreateAdminUserCommand(),
		command.NewMigrateCommand(),
		command.NewRunCommand(),
		command.NewVersionCommand(),
	}

	rootCommand.AddCommand(commands...)

	cobra.CheckErr(rootCommand.Execute())
}
