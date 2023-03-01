package cmd

import (
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"

	"github.com/haijima/cobrax"
	"github.com/haijima/stool/internal"
	"github.com/haijima/stool/internal/log"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

// NewTrendCommand returns the trend command
func NewTrendCommand(p *internal.TrendProfiler, v *viper.Viper, fs afero.Fs) *cobrax.Command {
	trendCmd := cobrax.NewCommand(v, fs)
	trendCmd.Use = "trend"
	trendCmd.Short = "Show the count of accesses for each endpoint over time"
	trendCmd.RunE = func(cmd *cobrax.Command, args []string) error {
		return runTrend(cmd, p)
	}

	trendCmd.PersistentFlags().IntP("interval", "i", 5, "time (in seconds) of the interval. Access counts are cumulated at each interval.")
	_ = trendCmd.BindPersistentFlags()

	return trendCmd
}

func runTrend(cmd *cobrax.Command, p *internal.TrendProfiler) error {
	matchingGroups := cmd.Viper().GetStringSlice("matching_groups")
	timeFormat := cmd.Viper().GetString("time_format")
	interval := cmd.Viper().GetInt("interval")
	cmd.V.Printf("%+v", cmd.Viper().AllSettings())

	if interval <= 0 {
		return fmt.Errorf("interval flag should be positive. but: %d", interval)
	}

	f, err := cmd.ReadFileOrStdIn("file")
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

func printTrendCsv(cmd *cobrax.Command, result *internal.Trend) error {
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
