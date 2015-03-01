package core_test

import (
	"fmt"
	"github.com/runningwild/argus/core"
	"github.com/runningwild/cmwc"
	. "github.com/smartystreets/goconvey/convey"
	"math/rand"
	"testing"
)

func shouldBeSorted(actual interface{}, expected ...interface{}) string {
	b, ok := actual.(core.Block64_32)
	if !ok {
		return fmt.Sprintf("Expected %T, got %T.", b, actual)
	}
	for i := 1; i < len(b); i++ {
		if b[i-1] > b[i] {
			return fmt.Sprintf("Not sorted")
		}
	}
	return ""
}

func TestZigAndZagWork(t *testing.T) {
	var diag core.Block64_32

	// Set up a block so that lower values are in the upper-left and higher values are in the lower
	// right.  A zig should sort this.
	for x := 0; x < 8; x++ {
		for y := 0; y < 8; y++ {
			diag[x+y*8] = int32(x + y)
		}
	}

	Convey("Slow versions are correct.", t, func() {
		var sorted core.Block64_32
		copy(sorted[:], diag[:])
		core.ZigSlow(&sorted)
		So(sorted, shouldBeSorted)
		core.ZagSlow(&sorted)
		So(sorted, ShouldResemble, diag)
	})

	Convey("Fast versions match the slow versions.", t, func() {
		c := cmwc.MakeGoodCmwc()
		for i := 0; i < 100; i++ {
			c.Seed(int64(i))
			rng := rand.New(c)
			var b0, b1, b2 core.Block64_32
			for i := range b0 {
				b0[i] = rng.Int31()
			}
			copy(b1[:], b0[:])
			copy(b2[:], b0[:])
			So(b1, ShouldResemble, b0)
			So(b2, ShouldResemble, b0)
			core.ZigSlow(&b0)
			core.Zig(&b1)
			So(b1, ShouldResemble, b0)
			core.ZagSlow(&b0)
			core.Zag(&b1)
			So(b1, ShouldResemble, b0)
			So(b1, ShouldResemble, b2)
		}
	})
}

func BenchmarkFdct(b *testing.B) {
	var block core.Block64_32
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		core.Fdct(&block)
	}
}
func BenchmarkIdct(b *testing.B) {
	var block core.Block64_32
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		core.Idct(&block)
	}
}
func BenchmarkZig(b *testing.B) {
	var block core.Block64_32
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		core.Zig(&block)
	}
}
func BenchmarkZag(b *testing.B) {
	var block core.Block64_32
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		core.Zag(&block)
	}
}
func BenchmarkZigSlow(b *testing.B) {
	var block core.Block64_32
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		core.ZigSlow(&block)
	}
}
func BenchmarkZagSlow(b *testing.B) {
	var block core.Block64_32
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		core.ZagSlow(&block)
	}
}
