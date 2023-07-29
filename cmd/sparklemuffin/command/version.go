package command

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/carlmjohnson/versioninfo"
	"github.com/spf13/cobra"
)

const (
	versionCmdName string = "version"
)

var (
	versionVerbose    bool
	versionFormatJson bool
)

type versionDetails struct {
	Short       string     `json:"short"`
	Revision    string     `json:"revision"`
	CommittedAt *time.Time `json:"last_commit,omitempty"`
	DirtyBuild  bool       `json:"dirty_build"`
}

func newVersionDetails() *versionDetails {
	v := &versionDetails{
		Short:      versioninfo.Short(),
		Revision:   versioninfo.Revision,
		DirtyBuild: versioninfo.DirtyBuild,
	}

	if !versioninfo.LastCommit.IsZero() {
		v.CommittedAt = &versioninfo.LastCommit
	}

	return v
}

// NewVersionCommand initializes and returns a CLI command to display the program version.
func NewVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   versionCmdName,
		Short: "Display the prorgam version",
		RunE: func(cmd *cobra.Command, args []string) error {
			details := newVersionDetails()

			if versionFormatJson {
				detailsBytes, err := json.Marshal(details)
				if err != nil {
					return fmt.Errorf("failed to marshal version details as JSON: %w", err)
				}

				fmt.Println(string(detailsBytes))

				return nil
			}

			if versionVerbose {
				tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)

				fmt.Fprintf(tw, "Version:\t%s\n", details.Short)
				fmt.Fprintf(tw, "Revision:\t%s\n", details.Revision)

				if details.CommittedAt != nil && !details.CommittedAt.IsZero() {
					fmt.Fprintf(tw, "Committed At:\t%s\n", details.CommittedAt.Format(time.UnixDate))
				}

				fmt.Fprintf(tw, "Dirty Build:\t%t\n", details.DirtyBuild)

				tw.Flush()

				return nil
			}

			fmt.Println(rootCmdName, "version", details.Short)

			return nil
		},
	}

	cmd.Flags().BoolVar(
		&versionFormatJson,
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
