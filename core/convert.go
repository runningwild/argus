package core

import (
	"math"
)

func RGBtoLABfast(r, g, b byte) (L, A, B byte) {
	i := int((uint32(r) << 16) | (uint32(g) << 8) | uint32(b))
	L, A, B = rgb2labIndex[i][0], rgb2labIndex[i][1], rgb2labIndex[i][2]
	return
}

func LABtoRGBfast(L, A, B byte) (r, g, b byte) {
	i := int((uint32(L) << 16) | (uint32(A) << 8) | uint32(B))
	r, g, b = lab2rgbIndex[i][0], lab2rgbIndex[i][1], lab2rgbIndex[i][2]
	return
}

var (
	rgb2labIndex [][3]byte
	lab2rgbIndex [][3]byte
)

func init() {
	data := MustAsset("rgb2lab")
	rgb2labIndex = make([][3]byte, 256*256*256)
	lab2rgbIndex = make([][3]byte, 256*256*256)
	if len(data) != 256*256*256*3 {
		panic("error loading rgb2lab data file: incorrect size")
	}
	for i := range rgb2labIndex {
		rgb2labIndex[i][0] = data[0]
		rgb2labIndex[i][1] = data[1]
		rgb2labIndex[i][2] = data[2]
		inv := int((uint32(data[0]) << 16) | (uint32(data[1]) << 8) | uint32(data[0]))
		lab2rgbIndex[inv][0] = byte(i >> 16)
		lab2rgbIndex[inv][1] = byte(i >> 8)
		lab2rgbIndex[inv][2] = byte(i)
		data = data[3:]
	}
}

func RGBtoLAB(r, g, b byte) (L, A, B byte) {
	Lf, Af, Bf := RGBtoLABf(float64(r), float64(g), float64(b))
	return byte(Lf + 0.5), byte(Af + 0.5), byte(Bf + 0.5)
}

func LABtoRGB(L, A, B byte) (r, g, b byte) {
	rf, gf, bf := LABtoRGBf(float64(L), float64(A), float64(B))
	if rf < 0 {
		rf = 0
	} else if rf > 255 {
		rf = 255
	}
	if gf < 0 {
		gf = 0
	} else if gf > 255 {
		gf = 255
	}
	if bf < 0 {
		bf = 0
	} else if bf > 255 {
		bf = 255
	}
	return byte(rf + 0.5), byte(gf + 0.5), byte(bf + 0.5)
}

func RGBtoLABf(r, g, b float64) (L, A, B float64) {
	x := 0.4887180*r + 0.3106803*g + 0.2006017*b
	y := 0.1762044*r + 0.8129847*g + 0.0108109*b
	z := 0.0000000*r + 0.0102048*g + 0.9897952*b

	L, A, B = x, y, z
	if y == 0 {
		L, A, B = 0, 0, 0
		return
	}
	ynorm := y / yn
	ysq := math.Sqrt(ynorm)
	L = kl * ysq
	A = ka * (x/xn - ynorm) / ysq
	B = kb * (ynorm - z/zn) / ysq
	// Jam these in bytes
	//L = argusScale * L
	//A = argusScale * (A + 15)
	//B = argusScale * (B + 93)
	return
}

// r,g,b are in [0,255], L,A,B are in 12.20 fixed point format.
func RGBtoLABi(r, g, b int32) (L, A, B int32) {
	// 12.20 fixed point
	x := 512457*r + 325771*g + 210346*b
	y := 184763*r + 852476*g + 11336*b
	z := 10700*g + 1037875*b

	L, A, B = x, y, z
	ysq := int32(math.Sqrt(float64(y)/float64(1<<20)) * (1 << 20))
	ysq_inv := int32((1.0 / math.Sqrt(float64(y)/float64(1<<20))) * (1 << 20))
	L = (kli >> 10) * ((ysq) >> 10)
	A = (kai >> 10) * (((((x>>10)*(xni_inv>>10) - y) >> 10) * ysq_inv >> 10) >> 10)
	B = (kbi >> 10) * ((((y - (z>>10)*(zni_inv>>10)) >> 10) * (ysq_inv >> 10)) >> 10)
	return
}

const (
	xn         = 0.95047
	yn         = 1.0
	zn         = 1.08883
	kl         = 1
	ka         = 175.0 / 198.04 * (xn + yn)
	kb         = 70.0 / 218.11 * (yn + zn)
	argusScale = 2.45

	// x.20 fixed point
	xni     = 996640
	yni     = 1048576
	zni     = 1141721
	xni_inv = 1103218
	yni_inv = 1048576
	zni_inv = 963030
	kli     = 1048576
	kai     = 1807275
	kbi     = 702951
)

func LABtoRGBf(L, A, B float64) (r, g, b float64) {
	// Rip them out of bytes
	//L = L / argusScale
	//A = A/argusScale - 15
	//B = B/argusScale - 93

	y := L * L * yn / kl / kl
	ynorm := y / yn
	ysq := math.Sqrt(ynorm)
	x := xn * (A/ka*ysq + ynorm)
	z := zn * (-B/kb*ysq + ynorm)
	r = 2.3706743*x - 0.9000405*y - 0.4706338*z
	g = -0.5138850*x + 1.4253036*y + 0.0885814*z
	b = 0.0052982*x - 0.0146949*y + 1.0093968*z
	return
}
