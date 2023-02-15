package graphviz

const DefaultFontSize = 14
const DefaultNodeColor = "black"
const DefaultNodeFillColor = "lightgrey"

type TextNode struct {
	Name     string
	Title    string
	FontSize float64
}

type BoxNode struct {
	TextNode
	Color     string
	FillColor string
}

func NewTextNode(name string, title string) *TextNode {
	return &TextNode{Name: name, Title: title, FontSize: DefaultFontSize}
}

func NewBoxNode(name string, title string) *BoxNode {
	return &BoxNode{TextNode: TextNode{Name: name, Title: title, FontSize: DefaultFontSize}, Color: DefaultNodeColor, FillColor: DefaultNodeFillColor}
}

func (n *BoxNode) SetColors(color, fillColor string) {
	n.Color = color
	n.FillColor = fillColor
}

func (n *BoxNode) SetColorLevel(level, maxLevel int) {
	n.Color = Colorize(float64(level)/float64(maxLevel), false)
	n.FillColor = Colorize(float64(level)/float64(maxLevel), true)
}
