package cmd

import (
	"runtime/debug"

	"github.com/haijima/cobrax"
	"github.com/haijima/stool/internal"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

// NewRootCmd returns the base command used when called without any subcommands
func NewRootCmd(v *viper.Viper, fs afero.Fs) *cobrax.Command {
	rootCmd := cobrax.NewCommand(v, fs)
	rootCmd.Use = "stool"
	rootCmd.Short = "stool is access log profiler"
	rootCmd.Version = version()

	rootCmd.PersistentFlags().StringP("file", "f", "", "access log file to profile")
	rootCmd.PersistentFlags().StringSliceP("matching_groups", "m", []string{}, "comma-separated list of regular expression patterns to group matched URIs")
	rootCmd.PersistentFlags().StringSlice("ignore_patterns", []string{}, "comma-separated list of regular expression patterns to ignore URIs")
	rootCmd.PersistentFlags().String("time_format", "02/Jan/2006:15:04:05 -0700", "format to parse time field on log file")
	_ = rootCmd.MarkFlagFilename("file", viper.SupportedExts...)
	rootCmd.PersistentFlags().Bool("version", false, "Show the version of this command")

	rootCmd.AddCommand(NewTrendCommand(internal.NewTrendProfiler(), v, fs))
	rootCmd.AddCommand(NewTransitionCmd(internal.NewTransitionProfiler(), v, fs))
	rootCmd.AddCommand(NewScenarioCmd(internal.NewScenarioProfiler(), v, fs))
	rootCmd.AddCommand(NewAaCommand(v, fs))

	return rootCmd
}

// Version is set in build step
var Version = ""

func version() string {
	if Version != "" {
		return Version
	}
	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		return buildInfo.Main.Version
	}
	return "(devel)"
}
