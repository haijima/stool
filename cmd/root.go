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

// NewRootCmd returns the base command used when called without any subcommands
func NewRootCmd(v *viper.Viper, fs afero.Fs) *cobra.Command {
	rootCmd := cobrax.NewRoot(v)
	rootCmd.Use = "stool"
	rootCmd.Short = "stool is access log profiler"
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// Colorization settings
		color.NoColor = color.NoColor || v.GetBool("no-color")
		// Set Logger
		l := slog.New(internal.NewCliSlogHandler(&slog.HandlerOptions{Level: cobrax.VerbosityLevel(v)}))
		slog.SetDefault(l)
		cobrax.SetLogger(l)

		return cobrax.RootPersistentPreRunE(cmd, v, fs, args)
	}

	rootCmd.AddCommand(NewTrendCmd(internal.NewTrendProfiler(), v, fs))
	rootCmd.AddCommand(NewTransitionCmd(internal.NewTransitionProfiler(), v, fs))
	rootCmd.AddCommand(NewScenarioCmd(internal.NewScenarioProfiler(), v, fs))
	rootCmd.AddCommand(NewParamCmd(internal.NewParamProfiler(), v, fs))
	rootCmd.AddCommand(NewAaCmd(v, fs))
	rootCmd.AddCommand(NewGenConfCmd(v, fs))

	return rootCmd
}
