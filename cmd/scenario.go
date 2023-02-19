package cmd

import (
	"encoding/csv"
	"fmt"
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
	"go.uber.org/zap"
)

// NewScenarioCmd returns the scenario command
func NewScenarioCmd(p *internal.ScenarioProfiler, v *viper.Viper, fs afero.Fs) *cobra.Command {
	var scenarioCmd = &cobra.Command{
		Use:   "scenario",
		Short: "Show the access patterns of users",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runScenario(cmd, p, v, fs)
		},
	}
	scenarioCmd.PersistentFlags().String("format", "dot", "The output format (dot, csv)")
	scenarioCmd.PersistentFlags().Bool("palette", false, "use color palette for each endpoint")
	_ = v.BindPFlags(scenarioCmd.PersistentFlags())
	v.SetFs(fs)

	return scenarioCmd
}

func runScenario(cmd *cobra.Command, p *internal.ScenarioProfiler, v *viper.Viper, fs afero.Fs) error {
	file := v.GetString("file")
	matchingGroups := v.GetStringSlice("matching_groups")
	ignorePatterns := v.GetStringSlice("ignore_patterns")
	timeFormat := v.GetString("time_format")
	format := v.GetString("format")
	palette := v.GetBool("palette")
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

	scenarios, err := p.Profile(logReader)
	if err != nil {
		return err
	}

	switch strings.ToLower(format) {
	case "dot":
		return createScenarioDot(cmd, scenarios, fs, palette)
	case "csv":
		return printScenarioCSV(cmd, scenarios)
	}
	return fmt.Errorf("invalid format flag: %s", format)
}

func printScenarioCSV(cmd *cobra.Command, scenarioStructs []internal.ScenarioStruct) error {
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

func createScenarioDot(cmd *cobra.Command, scenarioStructs []internal.ScenarioStruct, fs afero.Fs, usePalette bool) error {
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
