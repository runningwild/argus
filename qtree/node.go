package qtree

import (
	"image"
	"math"
)

type params struct {
	maxBufferLength int
}

type Info struct {
	Over      bool
	AboveOver bool
	Power     uint64
	Cache     []Block
}

type Tree struct {
	params *params

	// top-left to bottom-right
	x0, y0, x1, y1 int

	// kids in clockwise order starting from the top-left
	kids []*Tree

	maxPower uint64

	Info Info
}

func align8(n int) int {
	return (n / 8) * 8
}

func (t *Tree) divide(maxPowerPerPixel uint64) {
	region := float64(t.Bounds().Dx() * t.Bounds().Dy())
	t.maxPower = uint64(float64(maxPowerPerPixel) * (math.Pow(region, 1.4)))

	midx := align8((t.x1 + t.x0) / 2)
	midy := align8((t.y1 + t.y0) / 2)
	splitx := midx != t.x0
	splity := midy != t.y0
	switch {
	case !splitx && !splity:
		return

	case splitx && splity:
		t.kids = []*Tree{
			&Tree{params: t.params, x0: t.x0, y0: t.y0, x1: midx, y1: midy},
			&Tree{params: t.params, x0: t.x0, y0: midy, x1: midx, y1: t.y1},
			&Tree{params: t.params, x0: midx, y0: t.y0, x1: t.x1, y1: midy},
			&Tree{params: t.params, x0: midx, y0: midy, x1: t.x1, y1: t.y1},
		}

	case splitx:
		t.kids = []*Tree{
			&Tree{params: t.params, x0: t.x0, y0: midy, x1: midx, y1: t.y1},
			&Tree{params: t.params, x0: midx, y0: midy, x1: t.x1, y1: t.y1},
		}

	case splity:
		t.kids = []*Tree{
			&Tree{params: t.params, x0: midx, y0: t.y0, x1: t.x1, y1: midy},
			&Tree{params: t.params, x0: midx, y0: midy, x1: t.x1, y1: t.y1},
		}
	}

	for _, kid := range t.kids {
		kid.divide(maxPowerPerPixel)
	}
}

// Visitor is applied used in the Traverse* functions.  In the case of TraverseTopDown, the return
// value from Visitor will indicate if the children should be visited.  The return value is ignored
// in TraverseBottomUp.
type Visitor func(*Tree) bool

func (t *Tree) Bounds() image.Rectangle {
	return image.Rect(t.x0, t.y0, t.x1, t.y1)
}
func (t *Tree) Leaf() bool {
	return t.kids == nil
}
func (t *Tree) NumChildren() int {
	return len(t.kids)
}
func (t *Tree) Child(n int) *Tree {
	return t.kids[n]
}
func (t *Tree) TraverseTopDown(visitor Visitor) {
	if !visitor(t) {
		return
	}
	for _, kid := range t.kids {
		kid.TraverseTopDown(visitor)
	}
}
func (t *Tree) MaxPower() uint64 {
	return t.maxPower
}

func (t *Tree) TraverseBottomUp(visitor Visitor) {
	for _, kid := range t.kids {
		kid.TraverseBottomUp(visitor)
	}
	visitor(t)
}

func MakeTree(dx, dy int, maxPowerPerPixel uint64, backBufferSize uint8) *Tree {
	t := Tree{params: &params{maxBufferLength: int(backBufferSize)}, x0: 0, y0: 0, x1: dx, y1: dy}
	t.divide(maxPowerPerPixel)
	return &t
}
