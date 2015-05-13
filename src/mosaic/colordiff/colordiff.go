package colordiff

import (
	"fmt"
	"image/color"
)

func Closest(target color.Color, pallette []color.Color) color.Color {
	key := PaletteMapKey(target)
	result := MapPalette(target, pallette)
	fmt.Println(result)
	return result[key]
}

func Diff(c1, c2 color.Color) float64 {
	return Ciede2000(RgbToLab(c1), RgbToLab(c2))
}

func Diff2(c1, c2 LAB) float64 {
	return Ciede2000(c1, c2)
}
