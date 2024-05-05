package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
	"github.com/haijima/cobrax"
	"github.com/haijima/stool/internal"
	"github.com/haijima/stool/internal/log"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewTrendCmd returns the trend command
func NewTrendCmd(p *internal.TrendProfiler, v *viper.Viper, fs afero.Fs) *cobra.Command {
	trendCmd := &cobra.Command{}
	trendCmd.Use = "trend"
	trendCmd.Short = "Show the count of accesses for each endpoint over time"
	trendCmd.RunE = func(cmd *cobra.Command, args []string) error {
		return runTrend(cmd, v, fs, p)
	}

	trendCmd.Flags().String("format", "table", "The output format {table|md|csv}")
	trendCmd.Flags().IntP("interval", "i", 5, "time (in seconds) of the interval. Access counts are cumulated at each interval.")
	trendCmd.Flags().StringSlice("sort", []string{"sum:desc"}, "comma-separated list of \"<sort keys>:<order>\" Sort keys are {method|uri|sum|count0|count1|countN}. Orders are [asc|desc]. e.g. \"sum:desc,count0:asc\"")
	trendCmd.Flags().StringP("file", "f", "", "access log file to profile")
	trendCmd.Flags().StringSliceP("matching_groups", "m", []string{}, "comma-separated list of regular expression patterns to group matched URIs")
	trendCmd.Flags().String("time_format", "02/Jan/2006:15:04:05 -0700", "format to parse time field on log file")
	trendCmd.Flags().StringToString("log_labels", map[string]string{}, "comma-separated list of key=value pairs to override log labels")
	trendCmd.Flags().String("filter", "", "filter log lines by regular expression")
	_ = trendCmd.MarkFlagFilename("file", viper.SupportedExts...)

	return trendCmd
}

func runTrend(cmd *cobra.Command, v *viper.Viper, fs afero.Fs, p *internal.TrendProfiler) error {
	matchingGroups := v.GetStringSlice("matching_groups")
	timeFormat := v.GetString("time_format")
	labels := v.GetStringMapString("log_labels")
	filter := v.GetString("filter")
	format := v.GetString("format")
	sortKeys := v.GetStringSlice("sort")
	interval := v.GetInt("interval")

	if interval <= 0 {
		return fmt.Errorf("interval flag should be positive. but: %d", interval)
	}
	if format != "table" && format != "md" && format != "csv" {
		return errors.Newf("unknown format: %s", format)
	}

	f, err := cobrax.OpenOrStdIn(v.GetString("file"), fs, cobrax.WithStdin(cmd.InOrStdin()))
	if err != nil {
		return err
	}
	defer f.Close()
	logReader, err := log.NewLTSVReader(f, log.LTSVReadOpt{
		MatchingGroups: matchingGroups,
		TimeFormat:     timeFormat,
		Labels:         labels,
		Filter:         filter,
	})
	if err != nil {
		return err
	}

	result, err := p.Profile(logReader, interval, sortKeys)
	if err != nil {
		return err
	}

	return printTrendTable(cmd, result, format)
}

func printTrendTable(cmd *cobra.Command, result *internal.Trend, format string) error {
	t := table.NewWriter()
	t.SetOutputMirror(cmd.OutOrStdout())

	header := table.Row{"Method", "Uri"}
	for i := 0; i < result.Step; i++ {
		header = append(header, strconv.Itoa(i*result.Interval))
	}
	t.AppendHeader(header)

	aligns := make([]table.ColumnConfig, 0, len(header)-2)
	for i := 3; i <= len(header); i++ {
		aligns = append(aligns, table.ColumnConfig{Number: i, Align: text.AlignRight})
	}
	t.SetColumnConfigs(aligns)

	switch format {
	case "table":
		t.AppendRows(resultToRows(result, true))
		t.Render()
	case "md":
		t.AppendRows(resultToRows(result, true))
		t.RenderMarkdown()
	case "csv":
		t.AppendRows(resultToRows(result, false))
		t.RenderCSV()
	default:
		return nil // unreachable
	}
	return nil
}

func resultToRows(result *internal.Trend, humanized bool) []table.Row {
	rows := make([]table.Row, 0, len(result.Endpoints()))
	for _, endpoint := range result.Endpoints() {
		methodAndUri := strings.SplitN(endpoint, " ", 2) // split into Method and Uri
		row := table.Row{methodAndUri[0], methodAndUri[1]}
		for i, count := range result.Counts(endpoint) {
			s := strconv.Itoa(count)
			if humanized {
				s = humanize.Comma(int64(count))
			}
			if i > 0 && count*2 > result.Counts(endpoint)[i-1]*3 {
				s = color.GreenString(s)
			} else if count == 0 {
				s = color.HiBlackString(s)
			} else if i > 0 && count*3 < result.Counts(endpoint)[i-1]*2 {
				s = color.RedString(s)
			}
			row = append(row, s)
		}
		rows = append(rows, row)
	}
	return rows
}
