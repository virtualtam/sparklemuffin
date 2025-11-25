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

				if _, err := fmt.Fprintf(tw, "Version:\t%s\n", versionDetails.Short); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(tw, "Revision:\t%s\n", versionDetails.Revision); err != nil {
					return err
				}

				if versionDetails.CommittedAt != nil && !versionDetails.CommittedAt.IsZero() {
					if _, err := fmt.Fprintf(tw, "Committed At:\t%s\n", versionDetails.CommittedAt.Format(time.UnixDate)); err != nil {
						return err
					}
				}

				if _, err := fmt.Fprintf(tw, "Dirty Build:\t%t\n", versionDetails.DirtyBuild); err != nil {
					return err
				}

				if err := tw.Flush(); err != nil {
					return err
				}

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
