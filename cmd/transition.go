package cmd

import (
	"encoding/csv"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/awalterschulze/gographviz"
	"github.com/haijima/stool"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// NewTransitionCmd returns the transition command
func NewTransitionCmd(p *stool.TransitionProfiler, v *viper.Viper, fs afero.Fs) *cobra.Command {
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
	transitionCmd.PersistentFlags().String("format", "dot", "The output format (dot, png, csv)")
	_ = v.BindPFlags(transitionCmd.PersistentFlags())
	v.SetFs(fs)

	return transitionCmd
}

func runTransition(cmd *cobra.Command, p *stool.TransitionProfiler, v *viper.Viper, fs afero.Fs) error {
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

	opt := stool.TransitionOption{
		MatchingGroups: matchingGroups,
		IgnorePatterns: ignorePatterns,
		TimeFormat:     timeFormat,
	}

	result, err := p.Profile(f, opt)
	if err != nil {
		return err
	}

	format := v.GetString("format")
	switch strings.ToLower(format) {
	case "dot":
		return createDot(cmd, result, "dot", fs)
	case "png":
		return createDot(cmd, result, "png", fs)
	case "csv":
		return printTransitionCsv(cmd, result)
	}
	return fmt.Errorf("invalid format flag: %s", format)
}

func createDot(cmd *cobra.Command, result *stool.Transition, format string, fs afero.Fs) error {
	graph := gographviz.NewEscape()
	if err := graph.SetName("stool transition"); err != nil {
		return err
	}
	if err := graph.SetDir(true); err != nil {
		return err
	}

	eps := result.Endpoints.ToSlice()
	sort.Strings(eps)

	for _, e := range eps {
		if e == "" {
			continue
		}
		err := graph.AddNode("G", e, nil)
		if err != nil {
			return err
		}
	}

	for _, source := range eps {
		for _, target := range eps {
			if source == "" || target == "" {
				continue
			}
			if result.Data[source] == nil {
				continue
			}
			count := result.Data[source][target]
			if count == 0 {
				continue
			}
			err := graph.AddEdge(source, target, true, nil)
			if err != nil {
				return err
			}
		}
	}

	if format == "dot" {
		fmt.Fprintln(cmd.OutOrStdout(), graph.String())
	} else if format == "png" {
	} else {
		return fmt.Errorf("invalid format: %s", format)
	}
	return nil
}

func printTransitionCsv(cmd *cobra.Command, result *stool.Transition) error {
	writer := csv.NewWriter(cmd.OutOrStdout())

	eps := result.Endpoints.ToSlice()
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
