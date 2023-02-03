package cmd

import (
	"fmt"
	"runtime"
	"runtime/debug"

	"github.com/spf13/cobra"
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
`, version(), revision(), runtime.GOOS, runtime.GOARCH)
	},
}

func version() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "(devel)"
	}
	return info.Main.Version
}

func revision() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "(devel)"
	}
	for _, setting := range info.Settings {
		if setting.Key == "vcs.revision" {
			return setting.Value
		}
	}
	return "(devel)"
}
