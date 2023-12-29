package cmd

import (
	"encoding/csv"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/haijima/stool/internal"
	"github.com/haijima/stool/internal/graphviz"
	"github.com/haijima/stool/internal/log"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewTransitionCmd returns the transition command
func NewTransitionCmd(p *internal.TransitionProfiler, v *viper.Viper, fs afero.Fs) *cobra.Command {
	transitionCmd := &cobra.Command{}
	transitionCmd.Use = "transition"
	transitionCmd.Short = "Show the transition between endpoints"
	transitionCmd.RunE = func(cmd *cobra.Command, args []string) error {
		return runTransition(cmd, v, fs, p)
	}

	transitionCmd.Flags().String("format", "dot", "The output format {dot|mermaid|csv}")
	transitionCmd.Flags().StringP("file", "f", "", "access log file to profile")
	transitionCmd.Flags().StringSliceP("matching_groups", "m", []string{}, "comma-separated list of regular expression patterns to group matched URIs")
	transitionCmd.Flags().String("time_format", "02/Jan/2006:15:04:05 -0700", "format to parse time field on log file")
	transitionCmd.Flags().StringToString("log_labels", map[string]string{}, "comma-separated list of key=value pairs to override log labels")
	transitionCmd.Flags().String("filter", "", "filter log lines by regular expression")
	_ = transitionCmd.MarkFlagFilename("file", viper.SupportedExts...)

	return transitionCmd
}

func runTransition(cmd *cobra.Command, v *viper.Viper, fs afero.Fs, p *internal.TransitionProfiler) error {
	matchingGroups := v.GetStringSlice("matching_groups")
	timeFormat := v.GetString("time_format")
	labels := v.GetStringMapString("log_labels")
	filter := v.GetString("filter")
	format := v.GetString("format")

	f, err := OpenOrStdIn(v.GetString("file"), fs, cmd.InOrStdin())
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

	result, err := p.Profile(logReader)
	if err != nil {
		return err
	}

	var printFn printTransitionFunc
	switch strings.ToLower(format) {
	case "dot":
		printFn = createTransitionDot
	case "mermaid":
		printFn = createTransitionMermaid
	case "csv":
		printFn = printTransitionCsv
	default:
		return fmt.Errorf("invalid format flag: %s", format)
	}

	return printFn(cmd, result)
}

type printTransitionFunc = func(*cobra.Command, *internal.Transition) error

func createTransitionDot(cmd *cobra.Command, result *internal.Transition) error {
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

func createTransitionMermaid(cmd *cobra.Command, result *internal.Transition) error {
	cmd.Println("---")
	cmd.Println("title: stool transition")
	cmd.Println("---")
	cmd.Println("stateDiagram-v2")
	cmd.Println("direction TB")

	eps := result.Endpoints
	sort.Strings(eps)

	// Calculate the total of calls for each endpoint
	totalSum := 0
	for _, e := range eps {
		totalSum += result.Sum[e]
	}

	// Add each endpoint as a node
	cmd.Println("\t[*]")
	for i, e := range eps {
		if e == "" {
			continue
		}
		sum := result.Sum[e]
		cmd.Printf("\ts%d : %s Call %s (%s%%)\n", i, e, humanize.Comma(int64(sum)), humanize.FtoaWithDigits(100*float64(sum)/float64(totalSum), 2))
	}

	for i, source := range eps {
		for j, target := range eps {
			if result.Data[source] == nil {
				continue
			}
			count := result.Data[source][target]
			if count == 0 {
				continue
			}
			s := fmt.Sprintf("s%d", i)
			if source == "" {
				s = "[*]"
			}
			t := fmt.Sprintf("s%d", j)
			if target == "" {
				t = "[*]"
			}

			cmd.Printf("\t%s --> %s: %d\n", s, t, count)
		}
	}

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
