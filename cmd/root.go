package cmd

import (
	"github.com/haijima/stool/internal"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewRootCmd returns the base command used when called without any subcommands
func NewRootCmd(v *viper.Viper, fs afero.Fs) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:          "stool",
		Short:        "stool is access log profiler",
		SilenceUsage: true, // don't show help content when error occurred
		Version:      version(),
	}

	rootCmd.Flags().Bool("version", false, "Show the version of this command")
	addLoggingOption(rootCmd, v)
	useConfig(rootCmd, v)

	v.SetFs(fs)

	rootCmd.AddCommand(NewTrendCommand(internal.NewTrendProfiler(), v, fs))
	rootCmd.AddCommand(NewTransitionCmd(internal.NewTransitionProfiler(), v, fs))
	rootCmd.AddCommand(NewScenarioCmd(internal.NewScenarioProfiler(), v, fs))
	rootCmd.AddCommand(NewAaCommand(v, fs))

	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})

	return rootCmd
}
