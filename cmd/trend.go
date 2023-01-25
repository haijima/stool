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
	"strings"

	"github.com/haijima/stool"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type TrendCommand struct {
	IO
	profiler stool.TrendProfiler
}

// NewTrendCommand returns the trend command
func NewTrendCommand(p stool.TrendProfiler) *TrendCommand {
	return &TrendCommand{
		IO:       NewStdIO(),
		profiler: p,
	}
}

func (c *TrendCommand) Cmd() *cobra.Command {
	trendCmd := &cobra.Command{
		Use:   "trend",
		Short: "Show the count of accesses for each endpoint over time",
		RunE:  c.RunE,
	}

	trendCmd.PersistentFlags().StringP("file", "f", "", "access log file to profile")
	trendCmd.PersistentFlags().StringSliceP("matching_groups", "m", []string{}, "comma-separated list of regular expression patterns to group matched URIs")
	trendCmd.PersistentFlags().String("time_format", "02/Jan/2006:15:04:05 -0700", "format to parse time field on log file")
	trendCmd.Flags().IntP("interval", "i", 5, "time (in seconds) of the interval. Access counts are cumulated at each interval.")
	_ = viper.BindPFlag("file", trendCmd.PersistentFlags().Lookup("file"))
	_ = viper.BindPFlag("matching_groups", trendCmd.PersistentFlags().Lookup("matching_groups"))
	_ = viper.BindPFlag("time_format", trendCmd.PersistentFlags().Lookup("time_format"))
	_ = viper.BindPFlag("trend.interval", trendCmd.Flags().Lookup("interval"))

	return trendCmd
}

func (c *TrendCommand) RunE(cmd *cobra.Command, args []string) error {
	file := viper.GetString("file")
	matchingGroups := viper.GetStringSlice("matching_groups")
	timeFormat := viper.GetString("time_format")
	interval := viper.GetInt("trend.interval")

	f, err := c.Fs.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	opt := stool.TrendOption{
		MatchingGroups: matchingGroups,
		TimeFormat:     timeFormat,
		Interval:       interval,
	}

	result, err := c.profiler.Profile(f, opt)
	if err != nil {
		return err
	}

	return c.printCsv(result)
}

func (c *TrendCommand) printCsv(result stool.Trend) error {
	// header
	fmt.Fprint(c.Out, "Method, Uri")
	for i := 0; i < result.Step; i++ {
		fmt.Fprintf(c.Out, ", %d", i*result.Interval)
	}
	fmt.Fprintln(c.Out)

	// data rows for each endpoint
	for _, endpoint := range result.Endpoints() {
		fmt.Fprint(c.Out, strings.Replace(endpoint, " ", ", ", 1)) // split into Method and Uri
		for _, count := range result.Counts(endpoint) {
			fmt.Fprintf(c.Out, ", %d", count)
		}
		fmt.Fprintln(c.Out)
	}
	return nil
}
