package qtree_test

import (
	"github.com/orfjackal/gospec/src/gospec"
	"github.com/runningwild/argus/qtree"
)

func VisitorSpec(c gospec.Context) {
	c.Specify("Make sure that all leaf nodes cover region completely and without overlap", func() {
		dx, dy := 301, 511
		region := make([][]int, dx)
		for i := range region {
			region[i] = make([]int, dy)
		}
		t := qtree.MakeTree(dx, dy, 13)
		t.TraverseTopDown(func(t *qtree.Tree) bool {
			if !t.Leaf() {
				return true
			}
			for y := t.Bounds().Min.Y; y < t.Bounds().Max.Y; y++ {
				for x := t.Bounds().Min.X; x < t.Bounds().Max.X; x++ {
					region[x][y]++
				}
			}
			return false
		})
		for x := 0; x < dx; x++ {
			for y := 0; y < dy; y++ {
				c.Expect(region[x][y], gospec.Equals, 1)
			}
		}
		t.TraverseBottomUp(func(t *qtree.Tree) bool {
			if !t.Leaf() {
				return true
			}
			for y := t.Bounds().Min.Y; y < t.Bounds().Max.Y; y++ {
				for x := t.Bounds().Min.X; x < t.Bounds().Max.X; x++ {
					region[x][y]++
				}
			}
			return false
		})
		for x := 0; x < dx; x++ {
			for y := 0; y < dy; y++ {
				c.Expect(region[x][y], gospec.Equals, 2)
			}
		}
	})
}
