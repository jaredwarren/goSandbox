package colordiff

import (
	"fmt"
	"image/color"
)

type ColorDiff struct {
	str1 string
	str2 string
}

func Closest(target color.Color, pallette []color.Color) color.Color {
	key := PaletteMapKey(target)
	result := MapPalette(target, pallette)
	fmt.Println(result)
	return result[key]
}
