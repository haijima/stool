package cmd

import (
	"fmt"
	"github.com/haijima/stool/internal/graphviz"
	"github.com/haijima/stool/internal/log"
	"github.com/haijima/stool/internal/pattern"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/haijima/stool/internal"
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
		return printScenarioCSV(scenarios)
	}
	return fmt.Errorf("invalid format flag: %s", format)
}

func printScenarioCSV(scenarioStructs []internal.ScenarioStruct) error {
	fmt.Println("first call[s],last call[s],count,scenario node")
	for _, s := range scenarioStructs {
		fmt.Printf("%d,%d,%d,%s\n", s.FirstReq, s.LastReq, s.Count, s.Pattern.String(true))
	}
	return nil
}

type edge struct {
	From   int
	To     int
	Weight float64
}

func createScenarioDot(cmd *cobra.Command, scenarioStructs []internal.ScenarioStruct, fs afero.Fs, usePalette bool) error {
	graph, err := graphviz.NewEscapedDirectedGraph("stool scenario", true)
	if err != nil {
		return err
	}

	sumCount := 0
	for _, scenario := range scenarioStructs {
		sumCount += scenario.Count
	}

	// create palette
	palette := make(map[string]string, 0)
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

	for i, scenario := range scenarioStructs {
		subGraphName := fmt.Sprintf("cluster_%d", i)
		subGraphTitle := fmt.Sprintf("Scenario #%d  (count: %s, req: %d - %d [s])", i+1, humanize.Comma(int64(scenario.Count)), scenario.FirstReq, scenario.LastReq)
		if err := graphviz.AddSubgraph(graph, subGraphName, subGraphTitle); err != nil {
			return err
		}

		nodes := map[int]string{}
		edges := make([]edge, 0, scenario.Pattern.Leaves())
		patternToNodeAndEdge(*scenario.Pattern, nodes, &edges, 0)

		color := graphviz.Colorize(float64(scenario.Count)/float64(sumCount), false)
		fillColor := graphviz.Colorize(float64(scenario.Count)/float64(sumCount), true)

		for _, v := range []string{"start", "end"} {
			if err := graphviz.AddTextNode(graph, subGraphName, fmt.Sprintf("%d-%s", i, v), v); err != nil {
				return err
			}
		}
		for j, v := range nodes {
			if usePalette {
				color = "#333333"
				fillColor = palette[v]
			}
			if err := graphviz.AddBoxNode(graph, subGraphName, fmt.Sprintf("%d-%d", i, j), v, color, fillColor, 14); err != nil {
				return err
			}
		}

		penWidth := float64(scenario.Count)*10/float64(sumCount) + 1

		if err := graphviz.AddEdge(graph, fmt.Sprintf("%d-start", i), fmt.Sprintf("%d-%d", i, 0), color, penWidth, 1000); err != nil {
			return err
		}
		if err := graphviz.AddEdge(graph, fmt.Sprintf("%d-%d", i, scenario.Pattern.Leaves()-1), fmt.Sprintf("%d-end", i), color, penWidth, 1000); err != nil {
			return err
		}
		for _, edge := range edges {
			if err := graphviz.AddEdge(graph, fmt.Sprintf("%d-%d", i, edge.From), fmt.Sprintf("%d-%d", i, edge.To), color, penWidth, edge.Weight); err != nil {
				return err
			}
		}
	}

	fmt.Fprintln(cmd.OutOrStdout(), graph.String())
	return nil
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
			*edges = append(*edges, edge{From: base + offset - 1, To: base + offset, Weight: 1000})
		} else {
			if base == 0 && !n.IsLeaf() && (n.Degree() > 1 || n.Child(0).IsLeaf()) {
				continue
			}
			*edges = append(*edges, edge{From: base + offset - 1, To: base, Weight: 1})
		}
	}
}
