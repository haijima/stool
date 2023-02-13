package cmd

import (
	"encoding/csv"
	"fmt"
	"github.com/haijima/stool/internal/graphviz"
	"github.com/haijima/stool/internal/log"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/awalterschulze/gographviz"
	"github.com/dustin/go-humanize"
	"github.com/haijima/stool/internal"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// NewTransitionCmd returns the transition command
func NewTransitionCmd(p *internal.TransitionProfiler, v *viper.Viper, fs afero.Fs) *cobra.Command {
	var transitionCmd = &cobra.Command{
		Use:   "transition",
		Short: "Show the transition between endpoints",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTransition(cmd, p, v, fs)
		},
	}

	transitionCmd.PersistentFlags().StringP("file", "f", "", "access log file to profile")
	transitionCmd.PersistentFlags().StringSliceP("matching_groups", "m", []string{}, "comma-separated list of regular expression patterns to group matched URIs")
	transitionCmd.PersistentFlags().StringSlice("ignore_patterns", []string{}, "comma-separated list of regular expression patterns to ignore URIs")
	transitionCmd.PersistentFlags().String("time_format", "02/Jan/2006:15:04:05 -0700", "format to parse time field on log file")
	transitionCmd.PersistentFlags().String("format", "dot", "The output format (dot, csv)")
	_ = v.BindPFlags(transitionCmd.PersistentFlags())
	v.SetFs(fs)

	return transitionCmd
}

func runTransition(cmd *cobra.Command, p *internal.TransitionProfiler, v *viper.Viper, fs afero.Fs) error {
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
	logReader, err := log.NewLTSVReader(f, log.LTSVReadOpt{
		MatchingGroups: matchingGroups,
		IgnorePatterns: ignorePatterns,
		TimeFormat:     timeFormat,
	})
	if err != nil {
		return err
	}

	result, err := p.Profile(logReader)
	if err != nil {
		return err
	}

	format := v.GetString("format")
	switch strings.ToLower(format) {
	case "dot":
		return createTransitionDot(cmd, result, fs)
	case "csv":
		return printTransitionCsv(cmd, result)
	}
	return fmt.Errorf("invalid format flag: %s", format)
}

func createTransitionDot(cmd *cobra.Command, result *internal.Transition, fs afero.Fs) error {
	graph := gographviz.NewEscape()
	if err := graph.SetName("root"); err != nil {
		return err
	}
	if err := graph.AddAttr("root", "label", "stool transition"); err != nil {
		return err
	}
	if err := graph.AddAttr("root", "tooltip", "stool transition"); err != nil {
		return err
	}
	if err := graph.AddAttr("root", "labelloc", "t"); err != nil {
		return err
	}
	if err := graph.AddAttr("root", "fontname", "Courier"); err != nil {
		return err
	}
	if err := graph.AddAttr("root", "margin", "20"); err != nil {
		return err
	}
	if err := graph.SetDir(true); err != nil {
		return err
	}

	eps := result.Endpoints
	sort.Strings(eps)

	// Calculate the total of calls for each endpoint
	totalSum := 0
	for _, e := range eps {
		totalSum += result.Sum[e]
	}

	// Add "start" and "end" nodes
	for _, name := range []string{"start", "end"} {
		if err := graph.AddNode("", name, map[string]string{
			"shape":    "plaintext",
			"fontname": "Courier",
			"label":    name,
		}); err != nil {
			return err
		}
	}
	// Add each endpoint as a node
	for _, e := range eps {
		if e == "" {
			continue
		}

		sum := result.Sum[e]

		shape := "box"
		penwidth := "1"
		margin := "0.2"
		fontSize, _ := logNorm(sum, totalSum, 16)
		fontSize += 8

		err := graph.AddNode("", e, map[string]string{
			"shape":     shape,
			"style":     "\"solid,filled\"",
			"color":     graphviz.Colorize(float64(sum)/float64(totalSum), false),
			"fillcolor": graphviz.Colorize(float64(sum)/float64(totalSum), true),
			"fontsize":  strconv.Itoa(int(fontSize)),
			"fontname":  "Courier",
			"penwidth":  penwidth,
			"margin":    margin,
			"label":     fmt.Sprintf("%s\nCall: %s (%s%%)", e, humanize.Comma(int64(sum)), humanize.FtoaWithDigits(100*float64(sum)/float64(totalSum), 2)),
			"tooltip":   fmt.Sprintf("%s\nCall: %s (%s%%)", e, humanize.Comma(int64(sum)), humanize.FtoaWithDigits(100*float64(sum)/float64(totalSum), 2)),
		})
		if err != nil {
			return err
		}
	}

	for _, source := range eps {
		for _, target := range eps {
			if result.Data[source] == nil {
				continue
			}
			count := result.Data[source][target]
			if count == 0 {
				continue
			}

			s := source
			if source == "" {
				s = "start"
			}
			t := target
			if target == "" {
				t = "end"
			}

			weight, _ := logNorm(count, totalSum, 1000)
			weight += 1
			width, _ := logNorm(count, totalSum, 7)
			width += 1
			err := graph.AddEdge(s, t, true, map[string]string{
				"color":    graphviz.Colorize(float64(count)/float64(totalSum), false),
				"penwidth": strconv.Itoa(int(width)),
				"weight":   strconv.Itoa(int(weight)),
				"len":      "3",
				"label":    humanize.Comma(int64(count)),
			})
			if err != nil {
				return err
			}
		}
	}

	fmt.Fprintln(cmd.OutOrStdout(), graph.String())
	return nil
}

func printTransitionCsv(cmd *cobra.Command, result *internal.Transition) error {
	writer := csv.NewWriter(cmd.OutOrStdout())

	eps := result.Endpoints
	sort.Strings(eps)

	// header
	header := []string{""}
	header = append(header, eps...)
	_ = writer.Write(header)

	// data rows
	var row []string
	for _, e := range eps {
		row = []string{e}
		for _, e2 := range eps {
			row = append(row, strconv.Itoa(result.Data[e][e2]))
		}
		_ = writer.Write(row)
	}

	writer.Flush()
	return nil
}

// logNorm maps num where it is mapped from [1, s] to [0, t] on a logarithmic scale.
func logNorm(num, src, target int) (float64, error) {
	if num <= 0 {
		return 0, fmt.Errorf("num should be more than 0 but: %d", num)
	}
	if src <= 1 {
		return 0, fmt.Errorf("src should be more than 1 but: %d", src)
	}
	if target <= 0 {
		return 0, fmt.Errorf("target should be more than 0 but: %d", target)
	}
	if num == 1 {
		return 0, nil
	}
	base := math.Pow(float64(src), 1/float64(target))

	return math.Log(float64(num)) / math.Log(base), nil
}
