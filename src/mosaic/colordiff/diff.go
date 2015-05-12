package colordiff

import (
	"math"
)

/*func getPatchData(image image.Image, col, row int, patchWidth, patchHeight int) []color.Color {
	//patchData := make([]color.Color, patchWidth*patchHeight)
	patchData := []color.Color{}
	xOfset := col * patchWidth
	yOfset := row * patchHeight
	for y := 0; y < patchHeight; y++ {
		for x := 0; x < patchWidth; x++ {
			rgbaPix := image.At(int(xOfset+x), int(yOfset+y))
			patchData = append(patchData, rgbaPix)
		}
	}
	return patchData
}
*/
/**
 * Implementation of "The CIEDE2000 Color-Difference Formula: Implementation Notes, Supplementary Test Data, and Mathematical Observations"
 * by Gaurav Sharma, Wencheng Wu and Edul N. Dalal
 * http://www.ece.rochester.edu/~gsharma/ciede2000/ciede2000noteCRNA.pdf
 */

// Weight factors
const k_L = 1.0
const k_C = 1.0
const k_H = 1.0

func Ciede2000(c1, c2 LAB) float64 {
	LStar_1, aStar_1, bStar_1 := c1.L, c1.A, c1.B
	LStar_2, aStar_2, bStar_2 := c2.L, c2.A, c2.B

	// 1. Calculate C'_i, h'_i
	CStar_1_ab := math.Sqrt(math.Pow(aStar_1, 2) + math.Pow(bStar_1, 2))
	CStar_2_ab := math.Sqrt(math.Pow(aStar_2, 2) + math.Pow(bStar_2, 2))

	CBarStar_ab := (CStar_1_ab + CStar_2_ab) / 2.0

	G := 0.5 * (1 - math.Sqrt(math.Pow(CBarStar_ab, 7.0)/(math.Pow(CBarStar_ab, 7.0)+math.Pow(25.0, 7.0))))

	aPrime_1 := (1.0 + G) * aStar_1
	aPrime_2 := (1.0 + G) * aStar_2

	CPrime_1 := math.Sqrt(math.Pow(aPrime_1, 2) + math.Pow(bStar_1, 2))
	CPrime_2 := math.Sqrt(math.Pow(aPrime_2, 2) + math.Pow(bStar_2, 2))

	hPrime_1 := hPrime_i(bStar_1, aPrime_1)
	hPrime_2 := hPrime_i(bStar_2, aPrime_2)

	// 2. Calculate ΔL', ΔC', ΔH'
	ΔLPrime := LStar_2 - LStar_1
	ΔCPrime := CPrime_2 - CPrime_1

	ΔhPrime := 0.0
	if CStar_1_ab*CStar_2_ab == 0 {
		ΔhPrime = 0
	} else if math.Abs(hPrime_2-hPrime_1) <= 180 {
		ΔhPrime = hPrime_2 - hPrime_1
	} else if (hPrime_2 - hPrime_1) > 180 {
		ΔhPrime = (hPrime_2 - hPrime_1) - 360
	} else if (hPrime_2 - hPrime_1) < -180 {
		ΔhPrime = (hPrime_2 - hPrime_1) + 360
	} else {
		ΔhPrime = 0.0
	}

	ΔHPrime := 2 * math.Sqrt(CPrime_1*CPrime_2) * math.Sin(radians(ΔhPrime)/2.0)

	// 3. Calculate CIEDE2000 Color-Difference ΔE_00
	LBarPrime := (LStar_1 + LStar_2) / 2.0
	CBarPrime := (CPrime_1 + CPrime_2) / 2.0

	hBarPrime := 0.0
	if CStar_1_ab*CStar_2_ab == 0 {
		hBarPrime = hPrime_1 + hPrime_2
	} else if math.Abs(hPrime_1-hPrime_2) <= 180 {
		hBarPrime = (hPrime_1 + hPrime_2) / 2.0
	} else if (math.Abs(hPrime_1-hPrime_2) > 180) && ((hPrime_1 + hPrime_2) < 360) {
		hBarPrime = (hPrime_1 + hPrime_2 + 360) / 2.0
	} else if (math.Abs(hPrime_1-hPrime_2) > 180) && ((hPrime_1 + hPrime_2) >= 360) {
		hBarPrime = (hPrime_1 + hPrime_2 - 360) / 2.0
	} else {
		hBarPrime = 0.0
	}

	T := 1 - 0.17*math.Cos(radians(hBarPrime-30)) + 0.24*math.Cos(radians(2*hBarPrime)) + 0.32*math.Cos(radians(3*hBarPrime+6)) - 0.20*math.Cos(radians(4*hBarPrime-63))
	Δθ := 30 * math.Exp(-(math.Pow((hBarPrime-275)/25, 2)))
	R_C := math.Sqrt((math.Pow(CBarPrime, 7.0)) / (math.Pow(CBarPrime, 7.0) + math.Pow(25.0, 7.0)))
	S_L := 1 + ((0.015 * math.Pow(LBarPrime-50, 2)) / math.Sqrt(20+math.Pow(LBarPrime-50, 2.0)))
	S_C := 1 + 0.045*CBarPrime
	S_H := 1 + 0.015*CBarPrime*T
	R_T := -1 * math.Sin(radians(2*Δθ)) * R_C
	ΔE12_00 := math.Sqrt(math.Pow(ΔLPrime/(S_L*k_L), 2) + math.Pow(ΔCPrime/(S_C*k_C), 2) + math.Pow(ΔHPrime/(S_H*k_H), 2) + R_T*(ΔCPrime/(S_C*k_C))*(ΔHPrime/(S_H*k_H)))
	return ΔE12_00
}
func hPrime_i(bStar_i, aPrime_i float64) float64 {
	if bStar_i == 0 && aPrime_i == 0 {
		return 0
	} else {
		hPrime_i := degrees(math.Atan2(bStar_i, aPrime_i))
		if hPrime_i >= 0 {
			return hPrime_i
		} else {
			return hPrime_i + 360
		}
	}
}

func degrees(n float64) float64 {
	return n * (180 / math.Pi)
}

func radians(n float64) float64 {
	return n * (math.Pi / 180)
}
