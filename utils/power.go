package utils

import (
	"image"
)

func Power(a, b *image.RGBA, x0, y0 int, mppp float64) (pow float64, over bool) {
	var p float64
	for y := y0; y < y0+8; y++ {
		for x := x0; x < x0+8; x++ {
			offset := a.PixOffset(x, y)
			rdiff := float64(a.Pix[offset+0]) - float64(b.Pix[offset+0])
			gdiff := float64(a.Pix[offset+1]) - float64(b.Pix[offset+1])
			bdiff := float64(a.Pix[offset+2]) - float64(b.Pix[offset+2])
			add := rdiff*rdiff + gdiff*gdiff + bdiff*bdiff
			// if add > mppp {
			// 	return add, true
			// }
			p += add
		}
	}
	return p, p > mppp
}
