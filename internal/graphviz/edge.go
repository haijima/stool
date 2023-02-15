package graphviz

type Edge struct {
	fromName string
	toName   string
	Color    string
	PenWidth float64
	Weight   float64
}

const DefaultEdgeColor = "black"
const DefaultPenWidth = 1
const DefaultWeight = 1

func NewEdge(fromName string, toName string) *Edge {
	return &Edge{fromName: fromName, toName: toName, Color: DefaultEdgeColor, PenWidth: DefaultPenWidth, Weight: DefaultWeight}
}

func (e *Edge) SetColorLevel(level, maxLevel int) {
	e.Color = Colorize(float64(level)/float64(maxLevel), false)
}
