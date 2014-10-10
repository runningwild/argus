package qtree_test

import (
	"github.com/orfjackal/gospec/src/gospec"
	"github.com/runningwild/argus/qtree"
	"github.com/runningwild/argus/rgb"
	"image"
	"image/color"
	"image/draw"

	// These are here for debugging
	"fmt"
	"image/png"
	"os"
)

type randomImage struct {
	dx, dy int
}

func (r randomImage) Bounds() image.Rectangle {
	return image.Rect(0, 0, r.dx, r.dy)
}
func (r randomImage) At(x, y int) color.Color {
	return color.RGBA{R: byte(x*7 + y*8), G: byte(x*3 + y*9), B: byte(x*1 + y*8), A: 255}
}
func (r randomImage) ColorModel() color.Model {
	return color.RGBAModel
}

func MomentBlocksSpec(c gospec.Context) {
	c.Specify("Make sure that moment blocks encode properly", func() {
		// Make some randomish image
		r := randomImage{16, 16}
		canvas0 := rgb.Make(image.Rect(0, 0, 16, 16))
		draw.Draw(canvas0, canvas0.Bounds(), r, image.Point{}, draw.Over)
		var mb qtree.MomentBlocks
		mb.AddBlock(qtree.ExtractBlock(canvas0, image.Rect(0, 0, 8, 8)))
		mb.AddBlock(qtree.ExtractBlock(canvas0, image.Rect(8, 0, 16, 8)))
		mb.AddBlock(qtree.ExtractBlock(canvas0, image.Rect(0, 8, 8, 16)))
		mb.AddBlock(qtree.ExtractBlock(canvas0, image.Rect(8, 8, 16, 16)))

		canvas1 := rgb.Make(image.Rect(0, 0, 16, 16))
		draw.Draw(canvas1, image.Rect(0, 0, 8, 8), &mb, image.Point{0, 0}, draw.Over)
		draw.Draw(canvas1, image.Rect(8, 0, 16, 8), &mb, image.Point{0, 8}, draw.Over)
		draw.Draw(canvas1, image.Rect(0, 8, 8, 16), &mb, image.Point{0, 16}, draw.Over)
		draw.Draw(canvas1, image.Rect(8, 8, 16, 16), &mb, image.Point{0, 24}, draw.Over)

		// These images should be equal
		for i := range canvas1.Pix {
			fmt.Printf("%d %d %d\n", i, canvas0.Pix[i], canvas1.Pix[i])
			c.Expect(canvas1.Pix[i], gospec.Equals, canvas0.Pix[i])
		}

		// Enable this block to get pngs of the images in this test
		if false {
			f0, err := os.Create("canvas0.png")
			if err != nil {
				fmt.Printf("Failed to open canvas0.png: %v\n", err)
			} else {
				err = png.Encode(f0, canvas0)
				f0.Close()
				if err != nil {
					fmt.Printf("Failed to encode canvas0.png: %v\n", err)
				}
			}
			f1, err := os.Create("canvas1.png")
			if err != nil {
				fmt.Printf("Failed to open canvas1.png: %v\n", err)
			} else {
				err = png.Encode(f1, canvas1)
				f1.Close()
				if err != nil {
					fmt.Printf("Failed to encode canvas1.png: %v\n", err)
				}
			}
			f2, err := os.Create("mb.png")
			if err != nil {
				fmt.Printf("Failed to open mb.png: %v\n", err)
			} else {
				err = png.Encode(f2, &mb)
				f2.Close()
				if err != nil {
					fmt.Printf("Failed to encode mb.png: %v\n", err)
				}
			}
		}
	})
}
