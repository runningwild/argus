package core

import (
	"math"
)

func RGBtoLAB(r, g, b byte) (L, A, B byte) {
	Lf, Af, Bf := RGBtoLABf(float64(r), float64(g), float64(b))
	return byte(Lf+0.5), byte(Af+0.5), byte(Bf+0.5)
}

func LABtoRGB(L, A, B byte) (r, g, b byte) {
	rf, gf, bf := LABtoRGBf(float64(L), float64(A), float64(B))
	if rf < 0 { rf = 0 } else if rf > 255 { rf = 255 }
	if gf < 0 { gf = 0 } else if gf > 255 { gf = 255 }
	if bf < 0 { bf = 0 } else if bf > 255 { bf = 255 }
	return byte(rf+0.5), byte(gf+0.5), byte(bf+0.5)
}

func RGBtoLABf(r, g, b float64) (L, A, B float64) {
	x := 0.4887180*r + 0.3106803*g + 0.2006017*b
	y := 0.1762044*r + 0.8129847*g + 0.0108109*b
	z := 0.0000000*r + 0.0102048*g + 0.9897952*b
	L, A, B = x, y, z
	L = kl* math.Sqrt(y / yn)
	A = ka * (x/xn - y/yn) / math.Sqrt(y/yn)
	B = kb * (y/yn - z/zn) / math.Sqrt(y/yn)

	// Jam these in bytes
	L = argusScale * L
	A = argusScale * (A + 15)
	B = argusScale * (B + 93)
	return
}

const (
	xn = 0.95047
	yn = 1.0
	zn = 1.08883
	kl = 1
	ka = 175.0 / 198.04 * (xn + yn)
	kb = 70.0 / 218.11 * (yn + zn)
	argusScale = 2.45
)

func LABtoRGBf(L, A, B float64) (r, g, b float64) {
	// Rip them out of bytes
	L = L/argusScale
	A = A/argusScale - 15
	B = B/argusScale - 93

	y := L * L * yn / kl / kl
	x := xn * (A/ka*math.Sqrt(y/yn) + y/yn)
	z := zn * (-B/kb*math.Sqrt(y/yn) + y/yn)
	r = 2.3706743*x - 0.9000405*y - 0.4706338*z
	g = -0.5138850*x + 1.4253036*y + 0.0885814*z
	b = 0.0052982*x - 0.0146949*y + 1.0093968*z
	return
}
