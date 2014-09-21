package utils

import (
	"image"
)

func Power(a, b *image.RGBA, x0, y0 int, mppp uint64) (pow uint64, over bool) {
	pow = 0
	for y := y0; y < y0+8; y++ {
		offset := a.PixOffset(x0, y)
		pow += PowerLine(a.Pix[offset:], b.Pix[offset:])
	}
	over = pow > mppp*64
	return
}
