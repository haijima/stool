package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// These variables are set in build step
var (
	Version  = "unset"
	Revision = "unset"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show the version of stool command",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintf(cmd.OutOrStdout(), `Version: %s
Revision: %s
OS: %s
Arch: %s
`, Version, Revision, runtime.GOOS, runtime.GOARCH)
	},
}
