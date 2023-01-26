package cmd

import (
	"github.com/haijima/stool"
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
	}

	addLoggingOption(rootCmd, v)
	useConfig(rootCmd, v)

	v.SetFs(fs)

	rootCmd.AddCommand(NewTrendCommand(stool.NewTrendProfiler(), v, fs))

	return rootCmd
}
