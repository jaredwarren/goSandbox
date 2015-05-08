package colordiff

import (
	"fmt"
	"image/color"
	"math"
)

func PaletteMapKey(c color.Color) string {
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("R%dG%dB%d", r, g, b)
}

func MapPalette(targetColor color.Color, palette []color.Color) map[string]color.Color {
	c := make(map[string]color.Color)
	var bestColor color.Color
	bestColorDiff := math.MaxFloat64
	for _, currentColor := range palette {
		currentColorDiff := Diff(targetColor, currentColor)
		if currentColorDiff < bestColorDiff {
			bestColor = currentColor
			bestColorDiff = currentColorDiff
		}
	}
	c[PaletteMapKey(targetColor)] = bestColor
	return c
}
