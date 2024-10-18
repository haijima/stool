package cmd

import (
	"log/slog"

	"github.com/fatih/color"
	"github.com/haijima/cobrax"
	"github.com/haijima/stool/internal"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var Lv slog.LevelVar

// NewRootCmd returns the base command used when called without any subcommands
func NewRootCmd(v *viper.Viper, fs afero.Fs) *cobra.Command {
	rootCmd := cobrax.NewRoot(v)
	rootCmd.Use = "stool"
	rootCmd.Short = "stool is access log profiler"
	rootCmd.SetGlobalNormalizationFunc(cobrax.SnakeToKebab)
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// Colorization settings
		color.NoColor = color.NoColor || v.GetBool("no-color")
		// Set Log level
		Lv.Set(cobrax.VerbosityLevel(v))

		return cobrax.RootPersistentPreRunE(cmd, v, fs, args)
	}

	rootCmd.PersistentFlags().StringP("file", "f", "", "access log file to profile")
	rootCmd.PersistentFlags().StringSliceP("matching_groups", "m", []string{}, "comma-separated list of regular expression patterns to group matched URIs")
	rootCmd.PersistentFlags().String("time_format", "02/Jan/2006:15:04:05 -0700", "format to parse time field on log file")
	rootCmd.PersistentFlags().StringToString("log_labels", map[string]string{}, "comma-separated list of key=value pairs to override log labels")
	rootCmd.PersistentFlags().String("filter", "", "filter log lines by regular expression")
	_ = rootCmd.MarkFlagFilename("file", viper.SupportedExts...)

	rootCmd.AddCommand(NewTrendCmd(internal.NewTrendProfiler(), v, fs))
	rootCmd.AddCommand(NewTransitionCmd(internal.NewTransitionProfiler(), v, fs))
	rootCmd.AddCommand(NewScenarioCmd(internal.NewScenarioProfiler(), v, fs))
	rootCmd.AddCommand(NewParamCmd(internal.NewParamProfiler(), v, fs))
	rootCmd.AddCommand(NewAaCmd(v, fs))
	rootCmd.AddCommand(NewGenConfCmd(v, fs))

	return rootCmd
}
