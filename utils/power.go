package utils

import (
	"github.com/runningwild/argus/rgb"
)

func Power(a, b *rgb.Image, x0, y0 int, mppp uint64) (pow uint64, over bool) {
	pow = 0
	offset := a.PixOffset(x0, y0)
	pow += PowerLine(a.Pix[offset:], b.Pix[offset:])
	offset += a.Stride
	pow += PowerLine(a.Pix[offset:], b.Pix[offset:])
	offset += a.Stride
	pow += PowerLine(a.Pix[offset:], b.Pix[offset:])
	offset += a.Stride
	pow += PowerLine(a.Pix[offset:], b.Pix[offset:])
	offset += a.Stride
	pow += PowerLine(a.Pix[offset:], b.Pix[offset:])
	offset += a.Stride
	pow += PowerLine(a.Pix[offset:], b.Pix[offset:])
	offset += a.Stride
	pow += PowerLine(a.Pix[offset:], b.Pix[offset:])
	offset += a.Stride
	pow += PowerLine(a.Pix[offset:], b.Pix[offset:])
	over = pow > mppp*64
	return
}
