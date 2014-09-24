// Simple image class that implements image.Image and draw.Iamge, but doesn't support an alpha
// channel.
package rgb

import (
	"image"
	"image/color"
)

func Make(r image.Rectangle) *Image {
	return &Image{
		Pix:    make([]byte, r.Dx()*r.Dy()*3),
		Stride: r.Dx() * 3,
		Rect:   image.Rect(0, 0, r.Dx(), r.Dy()),
	}
}

func MakeWithData(r image.Rectangle, pix []byte) *Image {
	return &Image{
		Pix:    pix,
		Stride: r.Dx() * 3,
		Rect:   image.Rect(0, 0, r.Dx(), r.Dy()),
	}
}

type Image struct {
	Pix    []byte
	Stride int
	Rect   image.Rectangle
}

func (im *Image) At(x, y int) color.Color {
	offset := im.PixOffset(x, y)
	return color.RGBA{im.Pix[offset], im.Pix[offset+1], im.Pix[offset+2], 255}
}

func (im *Image) Set(x, y int, c color.Color) {
	if !(image.Point{x, y}).In(im.Rect) {
		return
	}
	cr, cg, cb, _ := c.RGBA()
	index := im.PixOffset(x, y)
	im.Pix[index+0] = (byte)(cr >> 8)
	im.Pix[index+1] = (byte)(cg >> 8)
	im.Pix[index+2] = (byte)(cb >> 8)
}

func (im *Image) Bounds() image.Rectangle {
	return im.Rect
}

func (im *Image) ColorModel() color.Model {
	return color.RGBAModel
}

func (im *Image) PixOffset(x, y int) int {
	return (x + y*im.Rect.Dx()) * 3
}
