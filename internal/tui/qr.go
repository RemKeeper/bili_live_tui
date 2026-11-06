package tui

import "strings"

// renderQR paints each module column with the same half-block glyph.
// skip2's ToSmallString mixes ASCII spaces with ▀▄█. On a Chinese Windows
// terminal those blocks are a different width from spaces, and centering
// then shifts every row by a different amount, so the code will not scan.
// One glyph plus explicit colors keeps every row the same width on light
// and dark backgrounds.
func renderQR(modules [][]bool) string {
	if len(modules) == 0 || len(modules[0]) == 0 {
		return ""
	}
	height := len(modules)
	width := len(modules[0])
	var b strings.Builder
	for y := 0; y < height; y += 2 {
		if y > 0 {
			b.WriteByte('\n')
		}
		prev := ""
		for x := 0; x < width; x++ {
			top := x < len(modules[y]) && modules[y][x]
			bottom := y+1 < height && x < len(modules[y+1]) && modules[y+1][x]
			seq := qrSGR(top, bottom)
			if seq != prev {
				b.WriteString(seq)
				prev = seq
			}
			b.WriteRune('▀')
		}
		b.WriteString("\x1b[0m")
	}
	return b.String()
}

func qrSGR(topBlack, bottomBlack bool) string {
	fg := "97"
	if topBlack {
		fg = "30"
	}
	bg := "107"
	if bottomBlack {
		bg = "40"
	}
	return "\x1b[" + fg + ";" + bg + "m"
}
