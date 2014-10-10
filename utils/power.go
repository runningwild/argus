// +build !arm

// Default implementations for architectures without assembly versions.
package utils

import (
	"github.com/runningwild/argus/qtree"
)

func Power(a, b *qtree.Block) (pow uint64) {
	var power uint64 = 0
	for i := range a {
		diff := int64(int64((*a)[i]) - int64((*b)[i]))
		power += uint64(diff * diff)
	}
	return power
}
