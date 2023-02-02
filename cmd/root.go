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
	rootCmd.AddCommand(NewTransitionCmd(stool.NewTransitionProfiler(), v, fs))
	rootCmd.AddCommand(NewScenarioCmd(stool.NewScenarioProfiler(), v, fs))
	rootCmd.AddCommand(versionCmd)

	// Split commands into main command group and utility command group
	main := cobra.Group{ID: "main", Title: "Available Commands:"}
	util := cobra.Group{ID: "util", Title: "Utility Commands:"}
	rootCmd.AddGroup(&main)
	rootCmd.AddGroup(&util)
	for _, command := range rootCmd.Commands() {
		command.GroupID = main.ID
	}
	rootCmd.SetHelpCommandGroupID(util.ID)
	rootCmd.SetCompletionCommandGroupID(util.ID)
	versionCmd.GroupID = util.ID

	return rootCmd
}
