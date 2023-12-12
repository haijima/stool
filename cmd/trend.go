package cmd

import (
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
	"github.com/haijima/cobrax"
	"github.com/haijima/stool/internal"
	"github.com/haijima/stool/internal/log"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
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
	trendCmd.Args = cobra.NoArgs

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

func runTrend(cmd *cobrax.Command, p *internal.TrendProfiler) error {
	matchingGroups := cmd.Viper().GetStringSlice("matching_groups")
	timeFormat := cmd.Viper().GetString("time_format")
	labels := cmd.Viper().GetStringMapString("log_labels")
	filter := cmd.Viper().GetString("filter")
	format := cmd.Viper().GetString("format")
	sortKeys := cmd.Viper().GetStringSlice("sort")
	interval := cmd.Viper().GetInt("interval")
	cmd.V.Printf("%+v", cmd.Viper().AllSettings())

	if interval <= 0 {
		return fmt.Errorf("interval flag should be positive. but: %d", interval)
	}

	f, err := cmd.OpenOrStdIn(cmd.Viper().GetString("file"))
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

	if format == "table" {
		return printTrendTable(cmd, result, false)
	} else if format == "md" {
		return printTrendTable(cmd, result, true)
	} else if format == "csv" {
		return printTrendCsv(cmd, result)
	}
	return fmt.Errorf("unknown format: %s", format)
}

func printTrendTable(cmd *cobrax.Command, result *internal.Trend, markdown bool) error {
	table := tablewriter.NewWriter(cmd.OutOrStdout())
	if markdown {
		table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
		table.SetCenterSeparator("|")
	}

	header := resultToHeader(result)
	table.SetHeader(header)

	aligns := make([]int, 0, len(header))
	aligns = append(aligns, tablewriter.ALIGN_LEFT)
	aligns = append(aligns, tablewriter.ALIGN_LEFT)
	for i := 2; i < len(header); i++ {
		aligns = append(aligns, tablewriter.ALIGN_RIGHT)
	}
	table.SetColumnAlignment(aligns)
	table.AppendBulk(resultToRows(result, true))
	table.Render()
	return nil
}

func printTrendCsv(cmd *cobrax.Command, result *internal.Trend) error {
	writer := csv.NewWriter(cmd.OutOrStdout())

	if err := writer.Write(resultToHeader(result)); err != nil {
		return err
	}

	if err := writer.WriteAll(resultToRows(result, false)); err != nil {
		return err
	}
	return nil
}

func resultToHeader(result *internal.Trend) []string {
	header := make([]string, 0, result.Step+2)
	header = append(header, "Method", "Uri")
	for i := 0; i < result.Step; i++ {
		header = append(header, strconv.Itoa(i*result.Interval))
	}
	return header
}

func resultToRows(result *internal.Trend, humanized bool) [][]string {
	rows := make([][]string, 0, len(result.Endpoints()))
	for _, endpoint := range result.Endpoints() {
		row := make([]string, 0)
		row = append(row, strings.SplitN(endpoint, " ", 2)...) // split into Method and Uri
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
