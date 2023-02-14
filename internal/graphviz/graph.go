package graphviz

import (
	"github.com/awalterschulze/gographviz"
)

func NewEscapedDirectedGraph(title string, isHorizontal bool) (*gographviz.Escape, error) {
	graph := gographviz.NewEscape()
	if err := graph.SetName("root"); err != nil {
		return nil, err
	}
	if err := graph.AddAttr("root", "label", title); err != nil {
		return nil, err
	}
	if err := graph.AddAttr("root", "tooltip", title); err != nil {
		return nil, err
	}
	if err := graph.AddAttr("root", "labelloc", "t"); err != nil {
		return nil, err
	}
	if err := graph.AddAttr("root", "fontname", "Courier"); err != nil {
		return nil, err
	}
	if err := graph.AddAttr("root", "margin", "20"); err != nil {
		return nil, err
	}
	if err := graph.SetDir(true); err != nil {
		return nil, err
	}

	if isHorizontal {
		if err := graph.AddAttr("root", "labeljust", "l"); err != nil {
			return nil, err
		}
		if err := graph.AddAttr("root", "rankdir", "LR"); err != nil {
			return nil, err
		}
	}

	return graph, nil
}

func AddSubgraph(parent *gographviz.Escape, subGraphName string, subGraphTitle string) error {
	return parent.AddSubGraph(parent.Name, subGraphName, map[string]string{
		"label":     subGraphTitle,
		"tooltip":   subGraphTitle,
		"style":     "filled",
		"labelloc":  "t",
		"labeljust": "l",
		"color":     "#cccccc",
		"fillcolor": "#ffffff",
		"sep":       "20",
	})
}
