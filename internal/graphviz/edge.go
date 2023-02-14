package graphviz

import (
	"fmt"
	"github.com/awalterschulze/gographviz"
)

func AddEdge(graph *gographviz.Escape, fromName, toName, color string, penWidth, weight float64) error {
	dir := "forward"
	if fromName == toName {
		dir = "back"
	}

	return graph.AddEdge(fromName, toName, true, map[string]string{
		"color":    color,
		"penwidth": fmt.Sprintf("%f", penWidth),
		"weight":   fmt.Sprintf("%f", weight),
		"dir":      dir,
	})
}
