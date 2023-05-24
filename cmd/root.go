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
	"github.com/withfig/autocomplete-tools/integrations/cobra"
)

// NewRootCmd returns the base command used when called without any subcommands
func NewRootCmd(v *viper.Viper, fs afero.Fs) *cobrax.Command {
	rootCmd := cobrax.NewCommand(v, fs)
	rootCmd.Use = "stool"
	rootCmd.Short = "stool is access log profiler"
	rootCmd.Args = cobra.NoArgs

	rootCmd.PersistentFlags().StringP("file", "f", "", "access log file to profile")
	rootCmd.PersistentFlags().StringSliceP("matching_groups", "m", []string{}, "comma-separated list of regular expression patterns to group matched URIs")
	rootCmd.PersistentFlags().String("time_format", "02/Jan/2006:15:04:05 -0700", "format to parse time field on log file")
	rootCmd.PersistentFlags().StringToString("log_labels", map[string]string{}, "comma-separated list of key=value pairs to override log labels")
	rootCmd.PersistentFlags().String("filter", "", "filter log lines by regular expression")
	rootCmd.PersistentFlags().Bool("no_color", false, "disable colorized output")
	_ = rootCmd.MarkFlagFilename("file", viper.SupportedExts...)

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
	rootCmd.AddCommand(cobrax.Wrap(cobracompletefig.CreateCompletionSpecCommand(), v, fs))

	return rootCmd
}
