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
	"encoding/csv"
	"strconv"
	"strings"

	"github.com/haijima/stool"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewTrendCommand returns the trend command
func NewTrendCommand(p stool.TrendProfiler, v *viper.Viper, fs afero.Fs) *cobra.Command {
	trendCmd := &cobra.Command{
		Use:   "trend",
		Short: "Show the count of accesses for each endpoint over time",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runE(cmd, p, v, fs)
		},
	}

	trendCmd.PersistentFlags().StringP("file", "f", "", "access log file to profile")
	trendCmd.PersistentFlags().StringSliceP("matching_groups", "m", []string{}, "comma-separated list of regular expression patterns to group matched URIs")
	trendCmd.PersistentFlags().String("time_format", "02/Jan/2006:15:04:05 -0700", "format to parse time field on log file")
	trendCmd.PersistentFlags().IntP("interval", "i", 5, "time (in seconds) of the interval. Access counts are cumulated at each interval.")
	//trendCmd.SetGlobalNormalizationFunc(func(f *pflag.FlagSet, name string) pflag.NormalizedName {
	//	return pflag.NormalizedName("trend." + name)
	//})
	v.BindPFlags(trendCmd.PersistentFlags())
	v.SetFs(fs)

	return trendCmd
}

func runE(cmd *cobra.Command, p stool.TrendProfiler, v *viper.Viper, fs afero.Fs) error {
	file := v.GetString("file")
	matchingGroups := v.GetStringSlice("matching_groups")
	timeFormat := v.GetString("time_format")
	interval := v.GetInt("interval")

	f, err := fs.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	opt := stool.TrendOption{
		MatchingGroups: matchingGroups,
		TimeFormat:     timeFormat,
		Interval:       interval,
	}

	result, err := p.Profile(f, opt)
	if err != nil {
		return err
	}

	return printCsv(cmd, result)
}

func printCsv(cmd *cobra.Command, result stool.Trend) error {
	writer := csv.NewWriter(cmd.OutOrStdout())

	header := make([]string, 0)
	header = append(header, "Method", "Uri")
	for i := 0; i < result.Step; i++ {
		header = append(header, strconv.Itoa(i*result.Interval))
	}
	writer.Write(header)

	// data rows for each endpoint
	for _, endpoint := range result.Endpoints() {
		row := make([]string, 0)
		row = append(row, strings.SplitN(endpoint, " ", 2)...) // split into Method and Uri
		for _, count := range result.Counts(endpoint) {
			row = append(row, strconv.Itoa(count))
		}
		writer.Write(row)
	}
	writer.Flush()
	return nil
}
