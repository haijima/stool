package graphviz

import (
	"fmt"
	"github.com/awalterschulze/gographviz"
)

func AddBoxNode(graph *gographviz.Escape, parentName string, nodeName, nodeTitle, color, fillColor string, fontSize float64) error {
	return graph.AddNode(parentName, nodeName, map[string]string{
		"shape":     "box",
		"style":     "filled",
		"fontname":  "Courier",
		"penwidth":  "1",
		"margin":    "0.2",
		"color":     color,
		"fillcolor": fillColor,
		"fontsize":  fmt.Sprintf("%f", fontSize),
		"label":     nodeTitle,
		"tooltip":   nodeTitle,
	})
}

func AddTextNode(graph *gographviz.Escape, parentName string, nodeName, nodeTitle string) error {
	return graph.AddNode(parentName, nodeName, map[string]string{
		"shape":    "plaintext",
		"fontname": "Courier",
		"label":    nodeTitle,
		"tooltip":  nodeTitle,
	})
}
