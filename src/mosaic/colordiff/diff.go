package colordiff

import (
	"image/color"
	"math"
)

func Ciede2000(c1, c2 LAB) float64 {
	/**
	 * Implemented as in "The CIEDE2000 Color-Difference Formula:
	 * Implementation Notes, Supplementary Test Data, and Mathematical Observations"
	 * by Gaurav Sharma, Wencheng Wu and Edul N. Dalal.
	 */

	// Get L,a,b values for color 1
	L1 := c1.L
	a1 := c1.A
	b1 := c1.B

	// Get L,a,b values for color 2
	L2 := c2.L
	a2 := c2.A
	b2 := c2.B

	// Weight factors
	kL := 1.0
	kC := 1.0
	kH := 1.0

	/**
	 * Step 1: Calculate C1p, C2p, h1p, h2p
	 */
	C1 := math.Sqrt(math.Pow(a1, 2) + math.Pow(b1, 2)) //(2)
	C2 := math.Sqrt(math.Pow(a2, 2) + math.Pow(b2, 2)) //(2)

	a_C1_C2 := (C1 + C2) / 2.0 //(3)

	G := 0.5 * (1 - math.Sqrt(math.Pow(a_C1_C2, 7.0)/
		(math.Pow(a_C1_C2, 7.0)+math.Pow(25.0, 7.0)))) //(4)

	a1p := (1.0 + G) * a1 //(5)
	a2p := (1.0 + G) * a2 //(5)

	C1p := math.Sqrt(math.Pow(a1p, 2) + math.Pow(b1, 2)) //(6)
	C2p := math.Sqrt(math.Pow(a2p, 2) + math.Pow(b2, 2)) //(6)

	h1p := hp_f(b1, a1p) //(7)
	h2p := hp_f(b2, a2p) //(7)

	/**
	 * Step 2: Calculate dLp, dCp, dHp
	 */
	dLp := L2 - L1   //(8)
	dCp := C2p - C1p //(9)

	dhp := dhp_f(C1, C2, h1p, h2p)                             //(10)
	dHp := 2 * math.Sqrt(C1p*C2p) * math.Sin(radians(dhp)/2.0) //(11)

	/**
	 * Step 3: Calculate CIEDE2000 Color-Difference
	 */
	a_L := (L1 + L2) / 2.0    //(12)
	a_Cp := (C1p + C2p) / 2.0 //(13)

	a_hp := a_hp_f(C1, C2, h1p, h2p) //(14)
	T := 1 - 0.17*math.Cos(radians(a_hp-30)) + 0.24*math.Cos(radians(2*a_hp)) +
		0.32*math.Cos(radians(3*a_hp+6)) - 0.20*math.Cos(radians(4*a_hp-63)) //(15)
	d_ro := 30 * math.Exp(-(math.Pow((a_hp-275)/25, 2)))                                 //(16)
	RC := math.Sqrt((math.Pow(a_Cp, 7.0)) / (math.Pow(a_Cp, 7.0) + math.Pow(25.0, 7.0))) //(17)
	SL := 1 + ((0.015 * math.Pow(a_L-50, 2)) /
		math.Sqrt(20+math.Pow(a_L-50, 2.0))) //(18)
	SC := 1 + 0.045*a_Cp                      //(19)
	SH := 1 + 0.015*a_Cp*T                    //(20)
	RT := -2 * RC * math.Sin(radians(2*d_ro)) //(21)
	dE := math.Sqrt(math.Pow(dLp/(SL*kL), 2) + math.Pow(dCp/(SC*kC), 2) +
		math.Pow(dHp/(SH*kH), 2) + RT*(dCp/(SC*kC))*
		(dHp/(SH*kH))) //(22)
	return dE
}

func Diff(c1, c2 color.Color) float64 {
	return Ciede2000(RgbToLab(c1), RgbToLab(c2))
}

func a_hp_f(C1, C2, h1p, h2p float64) float64 {
	if C1*C2 == 0 {
		return h1p + h2p
	} else if math.Abs(h1p-h2p) <= 180 {
		return (h1p + h2p) / 2.0
	} else if (math.Abs(h1p-h2p) > 180) && ((h1p + h2p) < 360) {
		return (h1p + h2p + 360) / 2.0
	} else if (math.Abs(h1p-h2p) > 180) && ((h1p + h2p) >= 360) {
		return (h1p + h2p - 360) / 2.0
	} else {
		// ERROR
	}
	return 0.0
}

func dhp_f(C1, C2, h1p, h2p float64) float64 {
	if C1*C2 == 0 {
		return 0
	} else if math.Abs(h2p-h1p) <= 180 {
		return h2p - h1p
	} else if (h2p - h1p) > 180 {
		return (h2p - h1p) - 360
	} else if (h2p - h1p) < -180 {
		return (h2p - h1p) + 360
	} else {
		// ERROR:::
	}
	return 0.0
}

func hp_f(x, y float64) float64 {
	if x == 0 && y == 0 {
		return 0
	} else {
		tmphp := degrees(math.Atan2(x, y))
		if tmphp >= 0 {
			return tmphp
		} else {
			return tmphp + 360
		}
	}
}

func degrees(n float64) float64 {
	return n * (180 / math.Pi)
}

func radians(n float64) float64 {
	return n * (math.Pi / 180)
}
