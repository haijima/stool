package internal

import "github.com/lucasb-eyer/go-colorful"

func Colorize(score float64, isBackground bool) string {
	if score < 0 {
		score = 0
	}
	if score > 1 {
		score = 1
	}

	//score = score * score
	if score < 0.2 {
		score = 0.2
	}
	if score > 0.8 {
		score = 0.8
	}
	t := (score - 0.2) / (0.8 - 0.2)

	var color colorful.Color

	if isBackground {
		// Keep Hue
		from, _ := colorful.Hex("#ECE4DC")
		to, _ := colorful.Hex("#D28A72")
		color = from.BlendHcl(to, t)
	} else {
		// Keep Luminance
		from, _ := colorful.Hex("#554B40")
		to, _ := colorful.Hex("#9B0016")
		color = from.BlendHcl(to, t)
	}

	return color.Hex()
}
