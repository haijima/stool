package cmd

import (
	"fmt"
	"github.com/haijima/stool/internal/graphviz"
	"github.com/haijima/stool/internal/log"
	"github.com/haijima/stool/internal/pattern"
	"strconv"
	"strings"

	"github.com/awalterschulze/gographviz"
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
	scenarioCmd.PersistentFlags().StringP("file", "f", "", "access log file to profile")
	scenarioCmd.PersistentFlags().StringSliceP("matching_groups", "m", []string{}, "comma-separated list of regular expression patterns to group matched URIs")
	scenarioCmd.PersistentFlags().StringSlice("ignore_patterns", []string{}, "comma-separated list of regular expression patterns to ignore URIs")
	scenarioCmd.PersistentFlags().String("time_format", "02/Jan/2006:15:04:05 -0700", "format to parse time field on log file")
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
	Weight int
}

func createScenarioDot(cmd *cobra.Command, scenarioStructs []internal.ScenarioStruct, fs afero.Fs, usePalette bool) error {
	graph := gographviz.NewEscape()
	if err := graph.SetName("root"); err != nil {
		return err
	}
	if err := graph.SetDir(true); err != nil {
		return err
	}
	if err := graph.AddAttr("root", "label", "stool scenario"); err != nil {
		return err
	}
	if err := graph.AddAttr("root", "tooltip", "stool scenario"); err != nil {
		return err
	}
	if err := graph.AddAttr("root", "labelloc", "t"); err != nil {
		return err
	}
	if err := graph.AddAttr("root", "labeljust", "l"); err != nil {
		return err
	}
	if err := graph.AddAttr("root", "rankdir", "LR"); err != nil {
		return err
	}
	if err := graph.AddAttr("root", "fontname", "Courier"); err != nil {
		return err
	}
	if err := graph.AddAttr("root", "margin", "20"); err != nil {
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
		if err := graph.AddSubGraph("root", subGraphName, map[string]string{
			"label":     fmt.Sprintf("Scenario #%d  (count: %s, req: %d - %d [s])", i+1, humanize.Comma(int64(scenario.Count)), scenario.FirstReq, scenario.LastReq),
			"tooltip":   fmt.Sprintf("Scenario #%d  (count: %s, req: %d - %d [s])", i+1, humanize.Comma(int64(scenario.Count)), scenario.FirstReq, scenario.LastReq),
			"style":     "filled",
			"labelloc":  "t",
			"labeljust": "l",
			"color":     "#cccccc",
			"fillcolor": "#ffffff",
			"sep":       "20",
			"rank":      "min",
		}); err != nil {
			return err
		}

		nodes := map[int]string{}
		edges := make([]edge, 0, scenario.Pattern.Leaves())
		patternToNodeAndEdge(*scenario.Pattern, nodes, &edges, 0)

		color := graphviz.Colorize(float64(scenario.Count)/float64(sumCount), false)
		fillcolor := graphviz.Colorize(float64(scenario.Count)/float64(sumCount), true)

		for _, v := range []string{"start", "end"} {
			if err := graph.AddNode(subGraphName, fmt.Sprintf("%d-%s", i, v), map[string]string{
				"shape":    "plaintext",
				"fontname": "Courier",
				"label":    v,
			}); err != nil {
				return err
			}
		}
		for j, v := range nodes {
			if usePalette {
				color = "#333333"
				fillcolor = palette[v]
			}
			if err := graph.AddNode(subGraphName, fmt.Sprintf("%d-%d", i, j), map[string]string{
				"shape":     "box",
				"style":     "filled",
				"color":     color,
				"fillcolor": fillcolor,
				"fontname":  "Courier",
				"label":     v,
				"tooltip":   v,
			}); err != nil {
				return err
			}
		}

		penWidth := int(float64(scenario.Count)*10/float64(sumCount)) + 1
		if err := graph.AddEdge(fmt.Sprintf("%d-start", i), fmt.Sprintf("%d-%d", i, 0), true, map[string]string{
			"penwidth": strconv.Itoa(penWidth),
			"weight":   strconv.Itoa(1000),
			"color":    color,
		}); err != nil {
			return err
		}
		if err := graph.AddEdge(fmt.Sprintf("%d-%d", i, scenario.Pattern.Leaves()-1), fmt.Sprintf("%d-end", i), true, map[string]string{
			"penwidth": strconv.Itoa(penWidth),
			"weight":   strconv.Itoa(1000),
			"color":    color,
		}); err != nil {
			return err
		}
		for _, edge := range edges {
			dir := "forward"
			if edge.From == edge.To {
				dir = "back"
			}
			err := graph.AddEdge(
				fmt.Sprintf("%d-%d", i, edge.From),
				fmt.Sprintf("%d-%d", i, edge.To),
				true,
				map[string]string{
					"dir":      dir,
					"penwidth": strconv.Itoa(penWidth),
					"weight":   strconv.Itoa(edge.Weight),
					"color":    color,
				})
			if err != nil {
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
