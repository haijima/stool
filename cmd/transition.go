package cmd

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"
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
			return _runE(cmd, p, v, fs)
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

func _runE(cmd *cobra.Command, p *stool.TransitionProfiler, v *viper.Viper, fs afero.Fs) error {
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
		return _printCsv(cmd, result)
	}
	return fmt.Errorf("invalid format flag: %s", format)
}

func createDot(cmd *cobra.Command, result *stool.Transition, format string, fs afero.Fs) error {
	g := graphviz.New()
	graph, err := g.Graph()
	if err != nil {
		return err
	}
	defer func() {
		if err := graph.Close(); err != nil {
			cobra.CheckErr(err)
		}
		g.Close()
	}()
	graph.SetStart(cgraph.RegularStart)
	graph.SetESep(1)

	eps := result.Endpoints.ToSlice()
	maxSum := 0
	for _, e := range eps {
		s := result.Sum[e]
		if s > maxSum {
			maxSum = s
		}
	}

	nodes := map[string]*cgraph.Node{}
	for _, e := range eps {
		if e == "" {
			continue
		}
		n, err := graph.CreateNode(e)
		if err != nil {
			return err
		}
		nodes[e] = n
		n.SetLabel(fmt.Sprintf("%s\n(Call: %d)", e, result.Sum[e]))

		n.SetShape(cgraph.BoxShape)
		n.SetMargin(0.2)
		fontSize, err := logNorm(result.Sum[e], maxSum, 34)
		if err != nil {
			return err
		}
		n.SetFontSize(fontSize + 8)                     // [8, 42]
		level, err := logNorm(result.Sum[e], maxSum, 5) // level: [0, 5]
		if err != nil {
			return err
		}
		n.SetPenWidth(level + 1)
		n.SafeSet("style", "solid,filled", "")

		n.SetColorScheme("reds6")
		if level >= 3 { // [3, 5]
			n.SetColor(strconv.Itoa(int(math.Round(level) + 1)))
		} else if level >= 2 { // [2, 3)
			n.SetColor("#2f4f4f")
		} else { // [0, 2)
			n.SetColor("#696969")
		}
		if level >= 3 {
			n.SetFillColor(strconv.Itoa(int(math.Round(level))))
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
			e, err := graph.CreateEdge(strconv.Itoa(count), nodes[source], nodes[target])
			if err != nil {
				return err
			}
			level, err := logNorm(count, maxSum, 5) // level: [0, 5]
			if err != nil {
				return err
			}
			e.SetLen(1 / (1 + level))
			e.SetPenWidth(level + 1)
			e.SetColorScheme("reds6")
			if level >= 3 {
				e.SetColor(strconv.Itoa(int(math.Round(level) + 1)))
			} else if level == 2 {
				e.SetColor("#2f4f4f")
			} else {
				e.SetColor("#696969")
			}
		}
	}

	var filename string
	if format == string(graphviz.XDOT) {
		filename = "./graph.dot"
	} else if format == string(graphviz.PNG) {
		filename = "./graph.png"
	} else {
		return fmt.Errorf("invalid format: %s", format)
	}
	var buf bytes.Buffer
	if err := g.Render(graph, graphviz.Format(format), &buf); err != nil {
		return err
	}
	if err := afero.WriteFile(fs, filename, buf.Bytes(), 0644); err != nil {
		return err
	}

	return nil
}

func _printCsv(cmd *cobra.Command, result *stool.Transition) error {
	writer := csv.NewWriter(cmd.OutOrStdout())

	eps := result.Endpoints.ToSlice()

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
