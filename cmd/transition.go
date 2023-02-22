package cmd

import (
	"encoding/csv"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/haijima/cobrax"
	"github.com/haijima/stool/internal"
	"github.com/haijima/stool/internal/graphviz"
	"github.com/haijima/stool/internal/log"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

// NewTransitionCmd returns the transition command
func NewTransitionCmd(p *internal.TransitionProfiler, v *viper.Viper, fs afero.Fs) *cobrax.Command {
	var transitionCmd = cobrax.NewCommand(v, fs)
	transitionCmd.Use = "transition"
	transitionCmd.Short = "Show the transition between endpoints"
	transitionCmd.RunE = func(cmd *cobrax.Command, args []string) error {
		return runTransition(cmd, p)
	}

	transitionCmd.PersistentFlags().String("format", "dot", "The output format (dot, csv)")
	transitionCmd.BindPersistentFlags()

	return transitionCmd
}

func runTransition(cmd *cobrax.Command, p *internal.TransitionProfiler) error {
	matchingGroups := cmd.Viper().GetStringSlice("matching_groups")
	ignorePatterns := cmd.Viper().GetStringSlice("ignore_patterns")
	timeFormat := cmd.Viper().GetString("time_format")
	cmd.V.Printf("%+v", cmd.Viper().AllSettings())

	f, err := cmd.ReadFileOrStdIn("file")
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

	var printFn printTransitionFunc
	format := cmd.Viper().GetString("format")
	switch strings.ToLower(format) {
	case "dot":
		printFn = createTransitionDot
	case "csv":
		printFn = printTransitionCsv
	default:
		return fmt.Errorf("invalid format flag: %s", format)
	}

	return printFn(cmd, result)
}

type printTransitionFunc = func(*cobrax.Command, *internal.Transition) error

func createTransitionDot(cmd *cobrax.Command, result *internal.Transition) error {
	graph := graphviz.NewGraph("root", "stool transition")

	eps := result.Endpoints
	sort.Strings(eps)

	// Calculate the total of calls for each endpoint
	totalSum := 0
	for _, e := range eps {
		totalSum += result.Sum[e]
	}

	// Add "start" and "end" nodes
	for _, name := range []string{"start", "end"} {
		node := graphviz.NewTextNode(name, name)
		if err := graph.AddTextNode(node); err != nil {
			return err
		}
	}
	// Add each endpoint as a node
	for _, e := range eps {
		if e == "" {
			continue
		}

		sum := result.Sum[e]
		fontSize, _ := logNorm(sum, totalSum, 16)
		fontSize += 8

		nodeTitle := fmt.Sprintf("%s\nCall: %s (%s%%)", e, humanize.Comma(int64(sum)), humanize.FtoaWithDigits(100*float64(sum)/float64(totalSum), 2))
		node := graphviz.NewBoxNode(e, nodeTitle)
		node.SetColorLevel(sum, totalSum)
		node.FontSize = fontSize
		if err := graph.AddBoxNode(node); err != nil {
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
			penWidth, _ := logNorm(count, totalSum, 7)
			penWidth += 1
			edge := graphviz.NewEdge(s, t)
			edge.SetColorLevel(count, totalSum)
			edge.PenWidth = penWidth
			edge.Weight = weight
			if err := graph.AddEdge(edge); err != nil {
				return err
			}
		}
	}

	return graph.Write(cmd.OutOrStdout())
}

func printTransitionCsv(cmd *cobrax.Command, result *internal.Transition) error {
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
