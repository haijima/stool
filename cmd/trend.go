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

	trendCmd.Flags().String("format", "table", "The output format (table, md, csv)")
	trendCmd.Flags().IntP("interval", "i", 5, "time (in seconds) of the interval. Access counts are cumulated at each interval.")
	trendCmd.Flags().Bool("no_color", false, "disable colorized output")

	return trendCmd
}

func runTrend(cmd *cobrax.Command, p *internal.TrendProfiler) error {
	matchingGroups := cmd.Viper().GetStringSlice("matching_groups")
	timeFormat := cmd.Viper().GetString("time_format")
	labels := cmd.Viper().GetStringMapString("log_labels")
	filter := cmd.Viper().GetString("filter")
	format := cmd.Viper().GetString("format")
	interval := cmd.Viper().GetInt("interval")
	noColor := cmd.Viper().GetBool("no_color")
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

	result, err := p.Profile(logReader, interval)
	if err != nil {
		return err
	}

	if format == "table" {
		return printTrendTable(cmd, result, false, noColor)
	} else if format == "md" {
		return printTrendTable(cmd, result, true, noColor)
	} else if format == "csv" {
		return printTrendCsv(cmd, result, noColor)
	}
	return fmt.Errorf("unknown format: %s", format)
}

func printTrendTable(cmd *cobrax.Command, result *internal.Trend, markdown, noColor bool) error {
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
	table.AppendBulk(resultToRows(result, true, noColor))
	table.Render()
	return nil
}

func printTrendCsv(cmd *cobrax.Command, result *internal.Trend, noColor bool) error {
	writer := csv.NewWriter(cmd.OutOrStdout())

	if err := writer.Write(resultToHeader(result)); err != nil {
		return err
	}

	if err := writer.WriteAll(resultToRows(result, false, noColor)); err != nil {
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

func resultToRows(result *internal.Trend, humanized, noColor bool) [][]string {
	rows := make([][]string, 0, len(result.Endpoints()))
	for _, endpoint := range result.Endpoints() {
		row := make([]string, 0)
		row = append(row, strings.SplitN(endpoint, " ", 2)...) // split into Method and Uri
		for i, count := range result.Counts(endpoint) {
			s := strconv.Itoa(count)
			if humanized {
				s = humanize.Comma(int64(count))
			}
			if !noColor {
				if i > 0 && count*2 > result.Counts(endpoint)[i-1]*3 {
					s = color.GreenString(s)
				} else if i > 0 && count*3 < result.Counts(endpoint)[i-1]*2 {
					s = color.RedString(s)
				}
			}
			row = append(row, s)
		}
		rows = append(rows, row)
	}
	return rows
}
