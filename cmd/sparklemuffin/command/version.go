// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package command

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

const (
	versionCmdName string = "version"
)

var (
	versionVerbose    bool
	versionFormatJSON bool
)

// NewVersionCommand initializes and returns a CLI command to display the program version.
func NewVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   versionCmdName,
		Short: "Display the prorgam version",
		RunE: func(cmd *cobra.Command, args []string) error {
			if versionFormatJSON {
				detailsBytes, err := json.Marshal(versionDetails)
				if err != nil {
					return fmt.Errorf("failed to marshal version details as JSON: %w", err)
				}

				fmt.Println(string(detailsBytes))

				return nil
			}

			if versionVerbose {
				tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)

				fmt.Fprintf(tw, "Version:\t%s\n", versionDetails.Short)
				fmt.Fprintf(tw, "Revision:\t%s\n", versionDetails.Revision)

				if versionDetails.CommittedAt != nil && !versionDetails.CommittedAt.IsZero() {
					fmt.Fprintf(tw, "Committed At:\t%s\n", versionDetails.CommittedAt.Format(time.UnixDate))
				}

				fmt.Fprintf(tw, "Dirty Build:\t%t\n", versionDetails.DirtyBuild)

				tw.Flush()

				return nil
			}

			fmt.Println(rootCmdName, "version", versionDetails.Short)

			return nil
		},
	}

	cmd.Flags().BoolVar(
		&versionFormatJSON,
		"json",
		false,
		"Format version information as JSON",
	)

	cmd.Flags().BoolVarP(
		&versionVerbose,
		"verbose",
		"v",
		false,
		"Display detailed version information",
	)

	return cmd
}
