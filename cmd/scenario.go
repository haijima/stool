package cmd

import (
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewScenarioCmd returns the scenario command
func NewScenarioCmd(v *viper.Viper, fs afero.Fs) *cobra.Command {
	var scenarioCmd = &cobra.Command{
		Use:   "scenario",
		Short: "Show the access patterns of users",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runScenario(cmd, v, fs)
		},
	}
	return scenarioCmd
}

func runScenario(cmd *cobra.Command, v *viper.Viper, fs afero.Fs) error {
	cmd.Println("scenario called")
	return nil
}
