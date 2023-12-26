package cmd

import (
	"encoding/csv"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/haijima/stool/internal"
	"github.com/haijima/stool/internal/graphviz"
	"github.com/haijima/stool/internal/log"
	"github.com/haijima/stool/internal/pattern"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewScenarioCmd returns the scenario command
func NewScenarioCmd(p *internal.ScenarioProfiler, v *viper.Viper, fs afero.Fs) *cobra.Command {
	scenarioCmd := &cobra.Command{}
	scenarioCmd.Use = "scenario"
	scenarioCmd.Short = "Show the access patterns of users"
	scenarioCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		return v.BindPFlags(cmd.Flags())
	}
	scenarioCmd.RunE = func(cmd *cobra.Command, args []string) error {
		return runScenario(cmd, v, fs, p)
	}

	scenarioCmd.Flags().String("format", "dot", "The output format {dot|mermaid|csv}")
	scenarioCmd.Flags().Bool("palette", false, "use color palette for each endpoint")
	scenarioCmd.Flags().StringP("file", "f", "", "access log file to profile")
	scenarioCmd.Flags().StringSliceP("matching_groups", "m", []string{}, "comma-separated list of regular expression patterns to group matched URIs")
	scenarioCmd.Flags().String("time_format", "02/Jan/2006:15:04:05 -0700", "format to parse time field on log file")
	scenarioCmd.Flags().StringToString("log_labels", map[string]string{}, "comma-separated list of key=value pairs to override log labels")
	scenarioCmd.Flags().String("filter", "", "filter log lines by regular expression")
	_ = scenarioCmd.MarkFlagFilename("file", viper.SupportedExts...)

	return scenarioCmd
}

func runScenario(cmd *cobra.Command, v *viper.Viper, fs afero.Fs, p *internal.ScenarioProfiler) error {
	matchingGroups := v.GetStringSlice("matching_groups")
	timeFormat := v.GetString("time_format")
	labels := v.GetStringMapString("log_labels")
	filter := v.GetString("filter")
	format := v.GetString("format")
	palette := v.GetBool("palette")

	slog.Debug(fmt.Sprintf("%+v", v.AllSettings()))

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

	scenarios, err := p.Profile(logReader)
	if err != nil {
		return err
	}

	var printFn printScenarioFunc
	switch strings.ToLower(format) {
	case "dot":
		printFn = createScenarioDot
	case "csv":
		printFn = printScenarioCSV
	case "mermaid":
		printFn = createScenarioMermaid
	default:
		return fmt.Errorf("invalid format flag: %s", format)
	}

	return printFn(cmd, scenarios, palette)
}

type printScenarioFunc = func(*cobra.Command, []internal.ScenarioStruct, bool) error

func printScenarioCSV(cmd *cobra.Command, scenarioStructs []internal.ScenarioStruct, usePalette bool) error {
	writer := csv.NewWriter(cmd.OutOrStdout())

	// header
	_ = writer.Write([]string{"first call[s]", "last call[s]", "count", "scenario node"})

	// data rows
	for _, s := range scenarioStructs {
		_ = writer.Write([]string{strconv.Itoa(s.FirstReq), strconv.Itoa(s.LastReq), strconv.Itoa(s.Count), s.Pattern.String(true)})
	}

	writer.Flush()
	return nil
}

func createScenarioDot(cmd *cobra.Command, scenarioStructs []internal.ScenarioStruct, usePalette bool) error {
	graph := graphviz.NewGraph("root", "stool scenario")
	graph.IsHorizontal = true

	sumCount := 0
	for _, scenario := range scenarioStructs {
		sumCount += scenario.Count
	}

	palette := make(map[string]string, 0)
	if usePalette {
		// create palette
		for _, scenarioStruct := range scenarioStructs {
			for _, s := range pattern.Flatten([]pattern.Node{*scenarioStruct.Pattern}, make([]string, scenarioStruct.Pattern.Leaves())) {
				if _, ok := palette[s]; !ok {
					palette[s] = ""
				}
			}
		}
		p, err := colorful.HappyPalette(len(palette))
		if err != nil {
			return err
		}
		i := 0
		for k := range palette {
			palette[k] = p[i].Hex()
			i++
		}
	}

	for i, scenario := range scenarioStructs {
		subGraphName := fmt.Sprintf("cluster_%d", i)
		subGraphTitle := fmt.Sprintf("Scenario #%d  (count: %s, req: %d - %d [s])", i+1, humanize.Comma(int64(scenario.Count)), scenario.FirstReq, scenario.LastReq)
		subGraph := graphviz.NewGraph(subGraphName, subGraphTitle)
		if err := graph.AddSubGraph(subGraph); err != nil {
			return err
		}

		nodes := map[int]string{}
		edges := make([]edge, 0, scenario.Pattern.Leaves())
		patternToNodeAndEdge(*scenario.Pattern, nodes, &edges, 0)

		for _, v := range []string{"start", "end"} {
			node := graphviz.NewTextNode(fmt.Sprintf("%d-%s", i, v), v)
			if err := subGraph.AddTextNode(node); err != nil {
				return err
			}
		}
		for j, v := range nodes {
			node := graphviz.NewBoxNode(fmt.Sprintf("%d-%d", i, j), v)
			node.SetColorLevel(scenario.Count, sumCount)
			if usePalette {
				node.FillColor = palette[v]
			}
			if err := subGraph.AddBoxNode(node); err != nil {
				return err
			}
		}

		penWidth := float64(scenario.Count)*10/float64(sumCount) + 1

		startEdge := graphviz.NewEdge(fmt.Sprintf("%d-start", i), fmt.Sprintf("%d-%d", i, 0))
		startEdge.SetColorLevel(scenario.Count, sumCount)
		startEdge.PenWidth = penWidth
		startEdge.Weight = 1000
		if err := subGraph.AddEdge(startEdge); err != nil {
			return err
		}
		endEdge := graphviz.NewEdge(fmt.Sprintf("%d-%d", i, scenario.Pattern.Leaves()-1), fmt.Sprintf("%d-end", i))
		endEdge.SetColorLevel(scenario.Count, sumCount)
		endEdge.PenWidth = penWidth
		endEdge.Weight = 1000
		if err := subGraph.AddEdge(endEdge); err != nil {
			return err
		}

		for _, edge := range edges {
			e := graphviz.NewEdge(fmt.Sprintf("%d-%d", i, edge.From), fmt.Sprintf("%d-%d", i, edge.To))
			e.SetColorLevel(scenario.Count, sumCount)
			e.PenWidth = penWidth
			if edge.From < edge.To {
				e.Weight = 1000
			}
			if err := subGraph.AddEdge(e); err != nil {
				return err
			}
		}
	}

	return graph.Write(cmd.OutOrStdout())
}

func createScenarioMermaid(cmd *cobra.Command, scenarioStructs []internal.ScenarioStruct, usePalette bool) error {
	cmd.Println("---")
	cmd.Println("title: stool scenario")
	cmd.Println("---")
	cmd.Println("flowchart LR")

	graph := graphviz.NewGraph("root", "stool scenario")
	graph.IsHorizontal = true

	sumCount := 0
	for _, scenario := range scenarioStructs {
		sumCount += scenario.Count
	}

	palette := make(map[string]string, 0)
	if usePalette {
		// create palette
		for _, scenarioStruct := range scenarioStructs {
			for _, s := range pattern.Flatten([]pattern.Node{*scenarioStruct.Pattern}, make([]string, scenarioStruct.Pattern.Leaves())) {
				if _, ok := palette[s]; !ok {
					palette[s] = ""
				}
			}
		}
		p, err := colorful.HappyPalette(len(palette))
		if err != nil {
			return err
		}
		i := 0
		for k := range palette {
			palette[k] = p[i].Hex()
			i++
		}
	}

	for i, scenario := range scenarioStructs {
		subGraphName := fmt.Sprintf("cluster_%d", i)
		subGraphTitle := fmt.Sprintf("Scenario #%d  (count: %s, req: %d - %d [s])", i+1, humanize.Comma(int64(scenario.Count)), scenario.FirstReq, scenario.LastReq)
		subGraph := graphviz.NewGraph(subGraphName, subGraphTitle)
		if err := graph.AddSubGraph(subGraph); err != nil {
			return err
		}

		nodes := map[int]string{}
		edges := make([]edge, 0, scenario.Pattern.Leaves())
		patternToNodeAndEdge(*scenario.Pattern, nodes, &edges, 0)

		for _, v := range []string{"start", "end"} {
			node := graphviz.NewTextNode(fmt.Sprintf("%d-%s", i, v), v)
			if err := subGraph.AddTextNode(node); err != nil {
				return err
			}
		}
		for j, v := range nodes {
			node := graphviz.NewBoxNode(fmt.Sprintf("%d-%d", i, j), v)
			node.SetColorLevel(scenario.Count, sumCount)
			if usePalette {
				node.FillColor = palette[v]
			}
			if err := subGraph.AddBoxNode(node); err != nil {
				return err
			}
		}

		penWidth := float64(scenario.Count)*10/float64(sumCount) + 1

		startEdge := graphviz.NewEdge(fmt.Sprintf("%d-start", i), fmt.Sprintf("%d-%d", i, 0))
		startEdge.SetColorLevel(scenario.Count, sumCount)
		startEdge.PenWidth = penWidth
		startEdge.Weight = 1000
		if err := subGraph.AddEdge(startEdge); err != nil {
			return err
		}
		endEdge := graphviz.NewEdge(fmt.Sprintf("%d-%d", i, scenario.Pattern.Leaves()-1), fmt.Sprintf("%d-end", i))
		endEdge.SetColorLevel(scenario.Count, sumCount)
		endEdge.PenWidth = penWidth
		endEdge.Weight = 1000
		if err := subGraph.AddEdge(endEdge); err != nil {
			return err
		}

		for _, edge := range edges {
			e := graphviz.NewEdge(fmt.Sprintf("%d-%d", i, edge.From), fmt.Sprintf("%d-%d", i, edge.To))
			e.SetColorLevel(scenario.Count, sumCount)
			e.PenWidth = penWidth
			if edge.From < edge.To {
				e.Weight = 1000
			}
			if err := subGraph.AddEdge(e); err != nil {
				return err
			}
		}
	}

	return graph.Write(cmd.OutOrStdout())
}

type edge struct {
	From int
	To   int
}

func patternToNodeAndEdge(n pattern.Node, nodes map[int]string, edges *[]edge, base int) {
	if n.IsLeaf() {
		nodes[base] = n.Value()
		return
	}

	offset := 0
	for i, child := range n.Children() {
		patternToNodeAndEdge(child, nodes, edges, base+offset)
		offset += child.Leaves()
		if i < n.Degree()-1 {
			*edges = append(*edges, edge{From: base + offset - 1, To: base + offset})
		} else {
			if base == 0 && !n.IsLeaf() && (n.Degree() > 1 || n.Child(0).IsLeaf()) {
				continue
			}
			*edges = append(*edges, edge{From: base + offset - 1, To: base})
		}
	}
}
