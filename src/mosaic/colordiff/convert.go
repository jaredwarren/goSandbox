package colordiff

import (
	"image/color"
	"math"
)

type LAB struct {
	L float64
	A float64
	B float64
}

type XYZ struct {
	X float64
	Y float64
	Z float64
}

func RgbToLab(c color.Color) LAB {
	return XyzToLab(RgbToXyz(c))
}

const RefX = 95.047
const RefY = 100.000
const RefZ = 108.883

func RgbToXyz(c color.Color) XYZ {
	R, G, B, _ := c.RGBA()
	r := float64(uint8(R>>8)) / 255
	g := float64(uint8(G>>8)) / 255
	b := float64(uint8(B>>8)) / 255
	if r > 0.04045 {
		r = math.Pow(((r + 0.055) / 1.055), 2.4)
	} else {
		r = r / 12.92
	}
	if g > 0.04045 {
		g = math.Pow(((g + 0.055) / 1.055), 2.4)
	} else {
		g = g / 12.92
	}
	if b > 0.04045 {
		b = math.Pow(((b + 0.055) / 1.055), 2.4)
	} else {
		b = b / 12.92
	}

	r *= 100
	g *= 100
	b *= 100

	x := r*0.4124 + g*0.3576 + b*0.1805
	y := r*0.2126 + g*0.7152 + b*0.0722
	z := r*0.0193 + g*0.1192 + b*0.9505
	return XYZ{X: x, Y: y, Z: z}
}

func XyzToLab(c XYZ) LAB {
	x := c.X / RefX
	y := c.Y / RefY
	z := c.Z / RefZ
	if x > 0.008856 {
		x = math.Pow(x, 0.333333)
	} else {
		x = (7.787 * x) + (16 / 116)
	}
	if y > 0.008856 {
		y = math.Pow(y, 0.333333)
	} else {
		y = (7.787 * y) + (16 / 116)
	}
	if z > 0.008856 {
		z = math.Pow(z, 0.333333)
	} else {
		z = (7.787 * z) + (16 / 116)
	}
	l := (116 * y) - 16
	a := 500 * (x - y)
	b := 200 * (y - z)
	return LAB{L: l, A: a, B: b}
}
