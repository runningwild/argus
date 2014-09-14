package main

import (
	"flag"
	"fmt"
	"github.com/runningwild/argus/qtree"
	"image"
	"image/color"
	"image/jpeg"
	"math"
	"os"
)

var in0 = flag.String("in0", "", "Image 0")
var in1 = flag.String("in1", "", "Image 1")

// File format
// Dims
// Kayframe (jpeg or png)
// for each frame:
// timestamp
// quad-tree representation of changed cells
// for each changed cell:
// a jpeg or png replacement

// No cell will ever get smaller than minDim on a side
const minDim = 16

const maxPowerPerPixel = 1.0

func maxPowerForRegion(dx, dy int) float64 {
	region := float64(dx * dy)
	return maxPowerPerPixel * (math.Pow(region, 1.8))
}

func power(a, b color.Color) float64 {
	var power float64
	ar, ag, ab, _ := a.RGBA()
	br, bg, bb, _ := b.RGBA()
	ar = ar >> 8
	ag = ag >> 8
	ab = ab >> 8
	br = br >> 8
	bg = bg >> 8
	bb = bb >> 8
	power += float64((ar - br) * (ar - br))
	power += float64((ag - bg) * (ag - bg))
	power += float64((ab - bb) * (ab - bb))
	return power
}

func doDiff(q *qtree.Tree, a, b image.Image) {
	if !q.Bounds().Eq(a.Bounds()) || !a.Bounds().Eq(b.Bounds()) {
		panic("Cannot diff two images with different bounds.")
	}
	// If a cell needs to be replaced, then it should be removed from its parent.
	// If all of a cell's children need to be replaced then we just replace that cell as a whole.

	q.TraverseBottomUp(func(t *qtree.Tree) bool {
		t.Info = qtree.Info{}
		if t.Leaf() {
			for y := t.Bounds().Min.Y; y < t.Bounds().Max.Y; y++ {
				for x := t.Bounds().Min.X; x < t.Bounds().Max.X; x++ {
					t.Info.Power += power(a.At(x, y), b.At(x, y))
				}
			}
		} else {
			if t.Child(0).Info.Over && t.Child(1).Info.Over && t.Child(2).Info.Over && t.Child(3).Info.Over {
				t.Info.Over = true
			} else {
				t.Info.Power = t.Child(0).Info.Power + t.Child(1).Info.Power + t.Child(2).Info.Power + t.Child(3).Info.Power
			}
		}
		if t.Info.Power > maxPowerForRegion(t.Bounds().Dx(), t.Bounds().Dy()) {
			t.Info.Over = true
			t.Info.Power = 0
		}
		return true
	})
}

func main() {
	// for i := 0; i < 20; i++ {
	// 	mp := maxPowerForRegion(1, 1<<uint(i))
	// 	fmt.Printf("%d: %0.2v, %0.2v\n", i, mp, mp/float64(uint(1)<<uint(i)))
	// }
	// return
	flag.Parse()
	if *in0 == "" || *in1 == "" {
		fmt.Printf("Must specify both input files\n")
		os.Exit(1)
	}
	fmt.Printf("Hello, world!\n")
	f0, err := os.Open(*in0)
	if err != nil {
		fmt.Printf("Failed to open file %q: %v\n", *in0, err)
		os.Exit(1)
	}
	defer f0.Close()
	f1, err := os.Open(*in1)
	if err != nil {
		fmt.Printf("Failed to open file %q: %v\n", *in1, err)
		os.Exit(1)
	}
	defer f1.Close()

	im0, _, err := image.Decode(f0)
	if err != nil {
		fmt.Printf("Unable to decode %q: %v\n", *in0, err)
		os.Exit(1)
	}
	im1, _, err := image.Decode(f1)
	if err != nil {
		fmt.Printf("Unable to decode %q: %v\n", *in1, err)
		os.Exit(1)
	}

	t := qtree.MakeTree(im0.Bounds().Dx(), im0.Bounds().Dy(), minDim)
	doDiff(t, im0, im1)
	fmt.Printf("Root: %v\n", t.Info)
	for i := 0; i < 4; i++ {
		fmt.Printf("Child(%d): %v\n", i, t.Child(i).Info)
	}
	t.TraverseTopDown(func(t *qtree.Tree) bool {
		if t.Info.Over {
			fmt.Printf("OVER %v: %f\n", t.Bounds(), t.Info.Power)
		} else {
			fmt.Printf("under %v: %f\n", t.Bounds(), t.Info.Power)
		}
		return true
	})
	out := image.NewRGBA(t.Bounds())
	colors := []color.Color{
		color.RGBA{255, 0, 255, 255},
		color.RGBA{255, 0, 0, 255},
		color.RGBA{255, 255, 0, 255},
		color.RGBA{255, 255, 255, 255},
		color.RGBA{0, 0, 255, 255},
		color.RGBA{0, 255, 0, 255},
		color.RGBA{0, 255, 255, 255},
	}
	t.TraverseTopDown(func(t *qtree.Tree) bool {
		height := 0
		c := t
		for !c.Leaf() {
			height++
			c = c.Child(0)
		}
		if t.Info.Over {
			for y := t.Bounds().Min.Y; y < t.Bounds().Max.Y; y++ {
				for x := t.Bounds().Min.X; x < t.Bounds().Max.X; x++ {
					out.Set(x, y, colors[height%len(colors)])
				}
			}
		}
		return !t.Info.Over
	})
	f, err := os.Create("output.jpg")
	if err != nil {
		fmt.Printf("Unable to make output file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()
	err = jpeg.Encode(f, out, nil)
	if err != nil {
		fmt.Printf("Unable to encode output image: %v\n", err)
		os.Exit(1)
	}
}
