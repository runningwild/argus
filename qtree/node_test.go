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
		t.TraverseTopDown(func(x0, y0, x1, y1 int, leaf bool, info *qtree.Info) bool {
			if !leaf {
				return true
			}
			for x := x0; x < x1; x++ {
				for y := y0; y < y1; y++ {
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
		t.TraverseBottomUp(func(x0, y0, x1, y1 int, leaf bool, info *qtree.Info) bool {
			if !leaf {
				return true
			}
			for x := x0; x < x1; x++ {
				for y := y0; y < y1; y++ {
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
