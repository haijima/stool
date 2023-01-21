/*
Copyright © 2023 haijima

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/haijima/stool"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewTrendCmd returns the trend command
func NewTrendCmd() *cobra.Command {
	trendCmd := &cobra.Command{
		Use:   "trend",
		Short: "Show the count of accesses for each endpoint over time",
		RunE:  run,
	}
	trendCmd.Flags().IntP("interval", "i", 5, "time (in seconds) of the interval. Access counts are cumulated at each interval.")
	_ = viper.BindPFlag("trend.interval", trendCmd.Flags().Lookup("interval"))

	return trendCmd
}

func run(cmd *cobra.Command, args []string) error {
	file := viper.GetString("file")
	matchingGroups := viper.GetStringSlice("matching_groups")
	timeFormat := viper.GetString("time_format")
	interval := viper.GetInt("trend.interval")

	opt := stool.TrendOption{
		MatchingGroups: matchingGroups,
		File:           file,
		TimeFormat:     timeFormat,
		Interval:       interval,
	}

	result, err := stool.Trend(opt)
	if err != nil {
		return err
	}

	return printCsv(result, os.Stdout)
}

func printCsv(result stool.TrendResult, out io.Writer) error {
	// header
	fmt.Fprint(out, "Method, Uri")
	for i := 0; i < result.Step(); i++ {
		fmt.Fprintf(out, ", %d", i*result.Interval())
	}
	fmt.Fprintln(out)

	// data rows for each endpoint
	for _, endpoint := range result.Endpoints() {
		fmt.Fprint(out, strings.Replace(endpoint, " ", ", ", 1)) // split into Method and Uri
		for _, count := range result.Counts(endpoint) {
			fmt.Fprintf(out, ", %d", count)
		}
		fmt.Fprintln(out)
	}
	return nil
}
