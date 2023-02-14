package cmd

import (
	"encoding/csv"
	"fmt"
	"github.com/haijima/stool/internal/log"
	"strconv"
	"strings"

	"github.com/haijima/stool/internal"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// NewTrendCommand returns the trend command
func NewTrendCommand(p *internal.TrendProfiler, v *viper.Viper, fs afero.Fs) *cobra.Command {
	trendCmd := &cobra.Command{
		Use:   "trend",
		Short: "Show the count of accesses for each endpoint over time",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTrend(cmd, p, v, fs)
		},
	}

	trendCmd.PersistentFlags().StringP("file", "f", "", "access log file to profile")
	trendCmd.PersistentFlags().StringSliceP("matching_groups", "m", []string{}, "comma-separated list of regular expression patterns to group matched URIs")
	trendCmd.PersistentFlags().String("time_format", "02/Jan/2006:15:04:05 -0700", "format to parse time field on log file")
	trendCmd.PersistentFlags().IntP("interval", "i", 5, "time (in seconds) of the interval. Access counts are cumulated at each interval.")
	_ = v.BindPFlags(trendCmd.PersistentFlags())
	v.SetFs(fs)

	return trendCmd
}

func runTrend(cmd *cobra.Command, p *internal.TrendProfiler, v *viper.Viper, fs afero.Fs) error {
	file := v.GetString("file")
	matchingGroups := v.GetStringSlice("matching_groups")
	timeFormat := v.GetString("time_format")
	interval := v.GetInt("interval")
	zap.L().Debug(fmt.Sprintf("%+v", v.AllSettings()))

	if interval <= 0 {
		return fmt.Errorf("interval flag should be positive. but: %d", interval)
	}

	f, err := fs.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()
	logReader, err := log.NewLTSVReader(f, log.LTSVReadOpt{
		MatchingGroups: matchingGroups,
		TimeFormat:     timeFormat,
	})
	if err != nil {
		return err
	}

	result, err := p.Profile(logReader, interval)
	if err != nil {
		return err
	}

	return printTrendCsv(cmd, result)
}

func printTrendCsv(cmd *cobra.Command, result *internal.Trend) error {
	writer := csv.NewWriter(cmd.OutOrStdout())

	header := make([]string, 0)
	header = append(header, "Method", "Uri")
	for i := 0; i < result.Step; i++ {
		header = append(header, strconv.Itoa(i*result.Interval))
	}
	_ = writer.Write(header)

	// data rows for each endpoint
	for _, endpoint := range result.Endpoints() {
		row := make([]string, 0)
		row = append(row, strings.SplitN(endpoint, " ", 2)...) // split into Method and Uri
		for _, count := range result.Counts(endpoint) {
			row = append(row, strconv.Itoa(count))
		}
		_ = writer.Write(row)
	}
	writer.Flush()
	return nil
}
