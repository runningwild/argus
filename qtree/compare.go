package qtree

import (
	"github.com/runningwild/argus/rgb"
	"image"
	"io"
)

func extractBlock(im *rgb.Image, bounds image.Rectangle) Block {
	var b Block
	if bounds.Dx() != 8 || bounds.Dy() != 8 {
		panic("Bounds are incorrect")
	}
	for y := 0; y < 8; y++ {
		copy(b[y*24:y*24+24], im.Pix[(y+bounds.Min.Y)*im.Stride+bounds.Min.X:])
	}
	return b
}

// returns true iff anything was written to out
func (t *Tree) injectImageNode(im *rgb.Image, out io.Writer) bool {
	if len(t.kids) > 0 {
		var written bool
		for _, kid := range t.kids {
			if kid.injectImageNode(im, out) {
				written = true
			}
		}
		return written
	}

	return false
}

func (t *Tree) InjectImage(im *rgb.Image, out io.Writer) {
	if im.Bounds() != t.Bounds() {
		panic("Suck it nubs")
	}
	t.injectImageNode(im, out)
}
