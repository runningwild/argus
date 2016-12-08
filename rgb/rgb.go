// Simple image class that implements image.Image and draw.Iamge, but doesn't support an alpha
// channel.
package rgb

import (
	"image"
	"image/color"

	"github.com/runningwild/argus/core"
)

type Image struct {
	blocks []core.Block8RGB

	// Width and Height in blocks, which may be larger than the actual width and height
	blockDx, blockDy int

	dx, dy int
}

func Make(r image.Rectangle) *Image {
	return &Image{
		blocks:  make([]core.Block8RGB, (r.Dx()+7)*(r.Dy()+7)/64),
		blockDx: (r.Dx() + 7) / 8,
		blockDy: (r.Dy() + 7) / 8,
		dx:      r.Dx(),
		dy:      r.Dy(),
	}
}

func (im *Image) At(x, y int) color.Color {
	if x < 0 || y < 0 || x >= im.dx || y >= im.dy {
		return color.Black
	}
	boff, poff := im.blockAndPixOffset(x, y)
	return color.RGBA{im.blocks[boff][poff+0], im.blocks[boff][poff+1], im.blocks[boff][poff+2], 255}
}

func (im *Image) Set(x, y int, c color.Color) {
	if x < 0 || y < 0 || x >= im.dx || y >= im.dy {
		return
	}
	boff, poff := im.blockAndPixOffset(x, y)
	cr, cg, cb, _ := c.RGBA()
	im.blocks[boff][poff+0] = byte(cr >> 8)
	im.blocks[boff][poff+1] = byte(cg >> 8)
	im.blocks[boff][poff+2] = byte(cb >> 8)
}

func (im *Image) Bounds() image.Rectangle {
	return image.Rect(0, 0, im.dx, im.dy)
}

func (im *Image) ColorModel() color.Model {
	return color.RGBAModel
}

func (im *Image) blockAndPixOffset(x, y int) (int, int) {
	return im.blockDx*(y/8) + (x / 8), 3 * ((y%8)*8 + x%8)
}
