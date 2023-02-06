package cmd

import (
	"fmt"

	"github.com/haijima/stool/internal"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// NewScenarioCmd returns the scenario command
func NewScenarioCmd(p *internal.ScenarioProfiler, v *viper.Viper, fs afero.Fs) *cobra.Command {
	var scenarioCmd = &cobra.Command{
		Use:   "scenario",
		Short: "Show the access patterns of users",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runScenario(cmd, p, v, fs)
		},

		Hidden: true,
	}
	scenarioCmd.PersistentFlags().StringP("file", "f", "", "access log file to profile")
	scenarioCmd.PersistentFlags().StringSliceP("matching_groups", "m", []string{}, "comma-separated list of regular expression patterns to group matched URIs")
	scenarioCmd.PersistentFlags().StringSlice("ignore_patterns", []string{}, "comma-separated list of regular expression patterns to ignore URIs")
	scenarioCmd.PersistentFlags().String("time_format", "02/Jan/2006:15:04:05 -0700", "format to parse time field on log file")
	_ = v.BindPFlags(scenarioCmd.PersistentFlags())
	v.SetFs(fs)

	return scenarioCmd
}

func runScenario(cmd *cobra.Command, p *internal.ScenarioProfiler, v *viper.Viper, fs afero.Fs) error {
	file := v.GetString("file")
	matchingGroups := v.GetStringSlice("matching_groups")
	ignorePatterns := v.GetStringSlice("ignore_patterns")
	timeFormat := v.GetString("time_format")
	zap.L().Debug(fmt.Sprintf("%+v", v.AllSettings()))

	f, err := fs.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	opt := internal.ScenarioOption{
		MatchingGroups: matchingGroups,
		IgnorePatterns: ignorePatterns,
		TimeFormat:     timeFormat,
	}

	_, err = p.Profile(f, opt)
	if err != nil {
		return err
	}

	return nil
}
