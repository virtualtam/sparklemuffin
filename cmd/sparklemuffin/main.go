// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package main

import (
	"github.com/spf13/cobra"
	"github.com/virtualtam/sparklemuffin/cmd/sparklemuffin/command"
)

func main() {
	rootCommand := command.NewRootCommand()

	commands := []*cobra.Command{
		command.NewCreateAdminUserCommand(),
		command.NewMigrateCommand(),
		command.NewRunCommand(),
		command.NewSyncFeedsCommand(),
		command.NewVersionCommand(),
	}

	rootCommand.AddCommand(commands...)

	cobra.CheckErr(rootCommand.Execute())
}
