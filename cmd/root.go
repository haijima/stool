package cmd

import (
	"io"
	"log"

	"github.com/fatih/color"
	"github.com/haijima/cobrax"
	"github.com/haijima/stool/internal"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewRootCmd returns the base command used when called without any subcommands
func NewRootCmd(v *viper.Viper, fs afero.Fs) *cobrax.Command {
	rootCmd := cobrax.NewCommand(v, fs)
	rootCmd.Use = "stool"
	rootCmd.Short = "stool is access log profiler"
	rootCmd.Args = cobra.NoArgs

	rootCmd.PersistentFlags().Bool("no_color", false, "disable colorized output")

	rootCmd.PersistentPreRunE = func(cmd *cobrax.Command, args []string) error {
		if cmd.Viper().GetBool("debug") {
			log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)
			log.SetOutput(cmd.ErrOrStderr())
		} else if cmd.Viper().GetBool("verbose") {
			log.SetOutput(cmd.ErrOrStderr())
		} else {
			log.SetOutput(io.Discard)
		}

		color.NoColor = color.NoColor || cmd.Viper().GetBool("no_color")

		return nil
	}

	rootCmd.AddCommand(NewTrendCommand(internal.NewTrendProfiler(), v, fs))
	rootCmd.AddCommand(NewTransitionCmd(internal.NewTransitionProfiler(), v, fs))
	rootCmd.AddCommand(NewScenarioCmd(internal.NewScenarioProfiler(), v, fs))
	rootCmd.AddCommand(NewParamCommand(internal.NewParamProfiler(), v, fs))
	rootCmd.AddCommand(NewAaCommand(v, fs))
	rootCmd.AddCommand(NewGenConfCmd(v, fs))

	return rootCmd
}
