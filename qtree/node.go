package qtree

import (
	"image"
)

type params struct {
	minDim int
}

type Info struct {
	Power float64
}

// NEXT: construct a quad tree for the dims of the image.  The tree will be re-used for each frame.
type Tree struct {
	params *params

	// top-left to bottom-right
	x0, y0, x1, y1 int

	// kids in clockwise order starting from the top-left
	kids []*Tree

	info Info
}

func (t *Tree) divide() {
	if t.x1-t.x0 < t.params.minDim || t.y1-t.y0 < t.params.minDim {
		return
	}
	midx := (t.x1 + t.x0) / 2
	midy := (t.y1 + t.y0) / 2
	t.kids = []*Tree{
		&Tree{params: t.params, x0: t.x0, y0: t.y0, x1: midx, y1: midy},
		&Tree{params: t.params, x0: midx, y0: t.y0, x1: t.x1, y1: midy},
		&Tree{params: t.params, x0: t.x0, y0: midy, x1: midx, y1: t.y1},
		&Tree{params: t.params, x0: midx, y0: midy, x1: t.x1, y1: t.y1},
	}
	for _, kid := range t.kids {
		kid.divide()
	}
}

// Visitor is applied used in the Traverse* functions.  In the case of TraverseTopDown, the return
// value from Visitor will indicate if the children should be visited.  The return value is ignored
// in TraverseBottomUp.
type Visitor func(x0, y0, x1, y1 int, leaf bool, info *Info) bool

func (t *Tree) Bounds() image.Rectangle {
	return image.Rect(t.x0, t.y0, t.x1, t.y1)
}

func (t *Tree) TraverseTopDown(visitor Visitor) {
	if !visitor(t.x0, t.y0, t.x1, t.y1, t.kids == nil, &t.info) {
		return
	}
	for _, kid := range t.kids {
		kid.TraverseTopDown(visitor)
	}
}

func (t *Tree) TraverseBottomUp(visitor Visitor) {
	for _, kid := range t.kids {
		kid.TraverseBottomUp(visitor)
	}
	visitor(t.x0, t.y0, t.x1, t.y1, t.kids == nil, &t.info)
}

func MakeTree(dx, dy, minDim int) *Tree {
	t := Tree{params: &params{minDim: minDim}, x0: 0, y0: 0, x1: dx, y1: dy}
	t.divide()
	return &t
}
