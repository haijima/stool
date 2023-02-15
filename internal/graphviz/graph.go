package graphviz

import (
	"fmt"
	"io"

	"github.com/awalterschulze/gographviz"
)

type Graph struct {
	name         string
	title        string
	IsHorizontal bool
	parent       *Graph
	g            *gographviz.Escape
}

func NewGraph(name string, title string) *Graph {
	return &Graph{name: name, title: title}
}

func (g *Graph) toRootGraphviz() (*gographviz.Escape, error) {
	if g.parent != nil {
		return g.parent.toRootGraphviz()
	}
	if g.g != nil {
		return g.g, nil
	}

	graph := gographviz.NewEscape()
	if err := graph.SetName(g.name); err != nil {
		return nil, err
	}
	if err := graph.AddAttr(g.name, "label", g.title); err != nil {
		return nil, err
	}
	if err := graph.AddAttr(g.name, "tooltip", g.title); err != nil {
		return nil, err
	}
	if err := graph.AddAttr(g.name, "labelloc", "t"); err != nil {
		return nil, err
	}
	if err := graph.AddAttr(g.name, "fontname", "Courier"); err != nil {
		return nil, err
	}
	if err := graph.AddAttr(g.name, "margin", "20"); err != nil {
		return nil, err
	}
	if err := graph.SetDir(true); err != nil {
		return nil, err
	}

	if g.IsHorizontal {
		if err := graph.AddAttr("root", "labeljust", "l"); err != nil {
			return nil, err
		}
		if err := graph.AddAttr("root", "rankdir", "LR"); err != nil {
			return nil, err
		}
	}
	g.g = graph

	return graph, nil
}

func (g *Graph) AddSubGraph(subGraph *Graph) error {
	graph, err := g.toRootGraphviz()
	if err != nil {
		return err
	}

	if err := graph.AddSubGraph(g.name, subGraph.name, map[string]string{
		"label":     subGraph.title,
		"tooltip":   subGraph.title,
		"style":     "filled",
		"labelloc":  "t",
		"labeljust": "l",
		"color":     "#cccccc",
		"fillcolor": "#ffffff",
		"sep":       "20",
	}); err != nil {
		return err
	}

	subGraph.parent = g
	return nil
}

func (g *Graph) AddBoxNode(node *BoxNode) error {
	graph, err := g.toRootGraphviz()
	if err != nil {
		return err
	}

	return graph.AddNode(g.name, node.Name, map[string]string{
		"shape":     "box",
		"style":     "filled",
		"fontname":  "Courier",
		"penwidth":  "1",
		"margin":    "0.2",
		"color":     node.Color,
		"fillcolor": node.FillColor,
		"fontsize":  fmt.Sprintf("%f", node.FontSize),
		"label":     node.Title,
		"tooltip":   node.Title,
	})
}

func (g *Graph) AddTextNode(node *TextNode) error {
	graph, err := g.toRootGraphviz()
	if err != nil {
		return err
	}

	return graph.AddNode(g.name, node.Name, map[string]string{
		"shape":    "plaintext",
		"fontname": "Courier",
		"fontsize": fmt.Sprintf("%f", node.FontSize),
		"label":    node.Title,
		"tooltip":  node.Title,
	})
}

func (g *Graph) AddEdge(edge *Edge) error {
	graph, err := g.toRootGraphviz()
	if err != nil {
		return err
	}

	dir := "forward"
	if edge.fromName == edge.toName {
		dir = "back"
	}

	return graph.AddEdge(edge.fromName, edge.toName, true, map[string]string{
		"color":    edge.Color,
		"penwidth": fmt.Sprintf("%f", edge.PenWidth),
		"weight":   fmt.Sprintf("%f", edge.Weight),
		"dir":      dir,
	})
}

func (g *Graph) Write(w io.Writer) error {
	graph, err := g.toRootGraphviz()
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, graph.String())
	return err
}
