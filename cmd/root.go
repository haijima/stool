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
	rootCmd := &cobra.Command{}
	rootCmd.Use = "stool"
	rootCmd.Short = "stool is access log profiler"
	rootCmd.Args = cobra.NoArgs
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})
	rootCmd.SilenceUsage = true  // don't show help content when error occurred
	rootCmd.SilenceErrors = true // Print error by own slog logger
	rootCmd.Version = cobrax.VersionFunc()
	rootCmd.PersistentPreRunE = setup(v)

	rootCmd.PersistentFlags().BoolP("version", "V", false, "Print version information and quit")
	rootCmd.PersistentFlags().String("config", "", "config file (default is $XDG_CONFIG_HOME/.stool.yaml)")
	rootCmd.PersistentFlags().Bool("no_color", false, "disable colorized output")
	_ = v.BindPFlag("no_color", rootCmd.PersistentFlags().Lookup("no_color"))
	rootCmd.PersistentFlags().CountP("verbose", "v", "More output per occurrence. (Use -vvvv or --verbose 4 for max verbosity)")
	_ = v.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "Silence all output")
	_ = v.BindPFlag("quiet", rootCmd.PersistentFlags().Lookup("quiet"))
	rootCmd.MarkFlagsMutuallyExclusive("verbose", "quiet")

	rootCmd.AddCommand(NewTrendCmd(internal.NewTrendProfiler(), v, fs))
	rootCmd.AddCommand(NewTransitionCmd(internal.NewTransitionProfiler(), v, fs))
	rootCmd.AddCommand(NewScenarioCmd(internal.NewScenarioProfiler(), v, fs))
	rootCmd.AddCommand(NewParamCmd(internal.NewParamProfiler(), v, fs))
	rootCmd.AddCommand(NewAaCmd(v, fs))
	rootCmd.AddCommand(NewGenConfCmd(v, fs))

	return rootCmd
}

func setup(v *viper.Viper) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// Colorize settings (Do before logger setup)
		color.NoColor = color.NoColor || v.GetBool("no_color")
		// Set Logger
		l := Logger(v)
		slog.SetDefault(l)
		cobrax.SetLogger(l)
		// Read config file
		cfg, err := cmd.Flags().GetString("config")
		if err != nil {
			return err
		}
		if err := cobrax.ReadConfigFile(v, cfg, true, cmd.Name()); err != nil {
			return err
		}
		// Bind flags (flags of the command to be executed)
		if err := v.BindPFlags(cmd.Flags()); err != nil {
			return err
		}
		// Print config values
		slog.Debug(cobrax.DebugViper(v))
		return nil
	}
}
