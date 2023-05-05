package stat

import "golang.org/x/exp/slices"

type Sort int

const (
	AscSorted Sort = iota
	DescSorted
	Unsorted
)

// Gini returns Gini coefficient of passed slice
func Gini(vs []int, sort Sort) float64 {
	n := len(vs)
	if n == 0 || n == 1 {
		return 0
	}

	if sort == Unsorted {
		slices.Sort(vs)
	}

	var g, s int
	for i, v := range vs {
		if sort == DescSorted {
			g += (n - 1 - i) * v
		} else {
			g += i * v
		}
		s += v
	}
	g *= 2
	g -= (n - 1) * s
	return float64(g) / float64(n*s)
}
