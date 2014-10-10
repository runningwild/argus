package qtree

import (
	"github.com/runningwild/argus/rgb"
	"io"
)

// returns true iff anything was written to out
func (t *Tree) injectImageNode(im *rgb.Image, out io.Writer, mb *MomentBlocks) bool {
	if len(t.kids) > 0 {
		var written bool
		for _, kid := range t.kids {
			if kid.injectImageNode(im, out, mb) {
				written = true
			}
		}
		return written
	}
	// block := ExtractBlock(im, t.Bounds())

	return false
}

func (t *Tree) InjectImage(im *rgb.Image, out io.Writer, mb *MomentBlocks) {
	if im.Bounds() != t.Bounds() {
		panic("Suck it nubs")
	}
	t.injectImageNode(im, out, mb)
}
