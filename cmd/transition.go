package cmd

import (
	"encoding/csv"
	"fmt"
	"strconv"

	"github.com/haijima/stool"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// NewTransitionCmd returns the transition command
func NewTransitionCmd(p *stool.TransitionProfiler, v *viper.Viper, fs afero.Fs) *cobra.Command {
	var transitionCmd = &cobra.Command{
		Use:   "transition",
		Short: "Show the transition between endpoints",
		RunE: func(cmd *cobra.Command, args []string) error {
			return _runE(cmd, p, v, fs)
		},
	}

	transitionCmd.PersistentFlags().StringP("file", "f", "", "access log file to profile")
	transitionCmd.PersistentFlags().StringSliceP("matching_groups", "m", []string{}, "comma-separated list of regular expression patterns to group matched URIs")
	transitionCmd.PersistentFlags().StringSlice("ignore_patterns", []string{}, "comma-separated list of regular expression patterns to ignore URIs")
	transitionCmd.PersistentFlags().String("time_format", "02/Jan/2006:15:04:05 -0700", "format to parse time field on log file")
	_ = v.BindPFlags(transitionCmd.PersistentFlags())
	v.SetFs(fs)

	return transitionCmd
}

func _runE(cmd *cobra.Command, p *stool.TransitionProfiler, v *viper.Viper, fs afero.Fs) error {
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

	opt := stool.TransitionOption{
		MatchingGroups: matchingGroups,
		IgnorePatterns: ignorePatterns,
		TimeFormat:     timeFormat,
	}

	result, err := p.Profile(f, opt)
	if err != nil {
		return err
	}

	return _printCsv(cmd, result)
}

func _printCsv(cmd *cobra.Command, result *stool.Transition) error {
	writer := csv.NewWriter(cmd.OutOrStdout())

	eps := result.Endpoints.ToSlice()

	// header
	header := []string{""}
	header = append(header, eps...)
	writer.Write(header)

	// data rows
	var row []string
	for _, e := range eps {
		row = []string{e}
		for _, e2 := range eps {
			row = append(row, strconv.Itoa(result.Data[e][e2]))
		}
		writer.Write(row)
	}

	writer.Flush()
	return nil
}
