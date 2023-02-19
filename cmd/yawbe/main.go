package main

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/virtualtam/yawbe/cmd/yawbe/command"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})

	rootCommand := command.NewRootCommand()

	commands := []*cobra.Command{
		command.NewMigrateCommand(),
		command.NewRunCommand(),
	}

	for _, cmd := range commands {
		rootCommand.AddCommand(cmd)
	}

	cobra.CheckErr(rootCommand.Execute())
}
