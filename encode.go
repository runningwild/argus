package main

import (
	"fmt"
	"github.com/runningwild/argus/qtree"
	"image"
)

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

// Need to fill in this function with something that takes into account that a larger region should
// have a lower thresholdPerPixel than a smaller region
func thresholdForSize() {

}

func doDiff(a, b image.Image) {
	// If a cell needs to be replaced, then it should be removed from its parent.
	// If all of a cell's children need to be replaced then we just replace that cell as a whole.
}

func main() {
	fmt.Printf("Hello, world!\n")
	t := qtree.MakeTree(100, 160, 16)
	t.Traverse(func(x0, y0, x1, y1 int, leaf bool, info *qtree.Info) bool {
		if !leaf {
			return true
		}
		fmt.Printf("%d %d %d %d %v\n", x0, y0, x1, y1, info)
		return true
	})
}
