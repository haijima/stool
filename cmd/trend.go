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
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/Wing924/ltsv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// trendCmd represents the trend command
var trendCmd = &cobra.Command{
	Use:   "trend",
	Short: "Show the count of accesses for each endpoint over time",
	RunE:  run,
}

func init() {
	rootCmd.AddCommand(trendCmd)

	trendCmd.Flags().IntP("interval", "i", 5, "time (in seconds) of the interval. Access counts are cumulated at each interval.")
	viper.BindPFlag("trend.interval", trendCmd.Flags().Lookup("interval"))
}

func run(cmd *cobra.Command, args []string) error {
	file := viper.GetString("file")
	matchingGroups := viper.GetStringSlice("matching_groups")
	timeFormat := viper.GetString("time_format")
	interval := viper.GetInt("trend.interval")

	var patterns []*regexp.Regexp
	patterns = make([]*regexp.Regexp, len(matchingGroups))
	for i, mg := range matchingGroups {
		p, err := regexp.Compile(mg)
		if err != nil {
			return err
		}
		patterns[i] = p
	}

	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)

	var result = map[string]map[int]int{}
	var startTime time.Time
	var endTime time.Time
	for scanner.Scan() {
		row := make(map[string]string)
		row, err = ltsv.DefaultParser.ParseLineAsMap(scanner.Bytes(), row)
		if err != nil {
			return err
		}

		time, err := time.Parse(timeFormat, row["time"])
		if err != nil {
			return err
		}

		if startTime.IsZero() {
			startTime = time
		}
		endTime = time

		k := key(row["req"], patterns)
		if result[k] == nil {
			result[k] = map[int]int{}
		}
		t := int(time.Sub(startTime).Seconds()) / interval
		result[k][t] += 1
	}

	output(result, startTime, endTime, interval)

	return nil
}

func output(result map[string]map[int]int, startTime, endTime time.Time, interval int) {
	cols := int(endTime.Sub(startTime).Seconds())/interval + 1

	fmt.Fprint(os.Stdout, "Method, Uri")
	for i := 0; i < cols; i++ {
		fmt.Fprintf(os.Stdout, ", %d", i*interval)
	}
	fmt.Fprintln(os.Stdout)

	for k, _ := range result {
		fmt.Fprint(os.Stdout, strings.Replace(k, " ", ", ", 1))
		for i := 0; i < cols; i++ {
			fmt.Fprintf(os.Stdout, ", %d", result[k][i])
		}
		fmt.Fprintln(os.Stdout)
	}
}

func key(req string, patterns []*regexp.Regexp) string {
	splitted := strings.Split(req, " ")
	method := splitted[0]
	uri := strings.Split(splitted[1], "?")[0]

	for _, p := range patterns {
		if p.MatchString(uri) {
			uri = p.String()
		}
	}
	return fmt.Sprintf("%s %s", method, uri)
}
