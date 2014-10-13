package qtree

import (
	"github.com/runningwild/argus/rgb"
	"image"
	"image/color"
)

// 8x8 rgb values
type Block [8 * 8 * 3]byte

type MomentBlocks struct {
	blocks []*Block
}

func (mb *MomentBlocks) AddBlock(b *Block) {
	mb.blocks = append(mb.blocks, b)
}
func (mb *MomentBlocks) NumBlocks() int {
	return len(mb.blocks)
}
func (mb *MomentBlocks) Bounds() image.Rectangle {
	return image.Rect(0, 0, 8, len(mb.blocks)*8)
}
func (mb *MomentBlocks) At(x, y int) color.Color {
	b := mb.blocks[y/8]
	y = y % 8
	return color.RGBA{R: (*b)[y*24+x*3], G: (*b)[y*24+x*3+1], B: (*b)[y*24+x*3+2], A: 255}
}
func (mb *MomentBlocks) ColorModel() color.Model {
	return color.RGBAModel
}

// func ExtractBlock(im *rgb.Image, bounds image.Rectangle) *Block {
// 	var b Block
// 	if bounds.Dx() != 8 || bounds.Dy() != 8 {
// 		panic("Bounds are incorrect")
// 	}
// 	for y := 0; y < 8; y++ {
// 		copy(b[y*24:y*24+24], im.Pix[(y+bounds.Min.Y)*im.Stride+bounds.Min.X*3:])
// 	}
// 	return &b
// }

// // SetToImage sets each node in the tree to match the corresponding region in im.
// func (t *Tree) SetToImage(im *rgb.Image) {
// 	if t.Bounds() != im.Bounds() {
// 		panic("NUB!")
// 	}
// 	t.TraverseBottomUp(func(t *Tree) bool {
// 		if len(t.kids) == 0 {
// 			t.Info.Cache = []*Block{ExtractBlock(im, t.Bounds())}
// 		}
// 		return true
// 	})
// }

func ExtractBlock(im *rgb.Image, bounds image.Rectangle, block *Block) {
	for y := 0; y < 8; y++ {
		copy((*block)[y*24:y*24+24], im.Pix[(y+bounds.Min.Y)*im.Stride+bounds.Min.X*3:])
	}
}

// SetToImage sets each node in the tree to match the corresponding region in im.
func (t *Tree) SetToImage(im *rgb.Image) {
	if t.Bounds() != im.Bounds() {
		panic("NUB!")
	}
	t.TraverseBottomUp(func(t *Tree) bool {
		if len(t.kids) == 0 {
			if t.Info.Cache == nil {
				t.Info.Cache = make([]*Block, 1)
				t.Info.Cache[0] = new(Block)
			}
			ExtractBlock(im, t.Bounds(), t.Info.Cache[0])
		}
		return true
	})
}
