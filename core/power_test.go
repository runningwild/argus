package core_test

import (
	"github.com/orfjackal/gospec/src/gospec"
	"github.com/runningwild/argus/core"
	"testing"
)

func PowerSpec(c gospec.Context) {
	var b0, b1, b2, b3, b4 core.Block8
	b0[0] = 10
	b0[1] = 20
	b0[2] = 30
	b0[3] = 40
	b0[4] = 50
	b0[5] = 60
	b1[0] = 20
	b1[1] = 30
	b1[2] = 40
	b1[3] = 50
	b1[4] = 60
	b1[5] = 70

	for i := range b3 {
		// Something randomish
		b3[i] = byte(i*10 + i*i + 3)
		if b3[i] == 0 {
			b3[i] = 1
		}
		b4[i] = b3[i] - 1
	}

	c.Specify("Blocks have zero power relative to themselves", func() {
		c.Expect(core.Power(&b0, &b0), gospec.Equals, uint64(0))
		c.Expect(core.Power(&b1, &b1), gospec.Equals, uint64(0))
		c.Expect(core.Power(&b2, &b2), gospec.Equals, uint64(0))
		c.Expect(core.Power(&b3, &b3), gospec.Equals, uint64(0))
	})
	c.Specify("Simple manual power check", func() {
		c.Expect(core.Power(&b0, &b1), gospec.Equals, uint64(600))
		c.Expect(core.Power(&b1, &b0), gospec.Equals, uint64(600))
	})
	c.Specify("Simple manual power check", func() {
		c.Expect(core.Power(&b0, &b2), gospec.Equals, uint64(9100))
		c.Expect(core.Power(&b2, &b0), gospec.Equals, uint64(9100))
	})
	c.Specify("Simple manual power check", func() {
		c.Expect(core.Power(&b1, &b2), gospec.Equals, uint64(13900))
		c.Expect(core.Power(&b2, &b1), gospec.Equals, uint64(13900))
	})
	c.Specify("Simple manual power check - zero vs zero", func() {
		c.Expect(core.Power(&b2, &b2), gospec.Equals, uint64(0))
	})
	c.Specify("Power function should cover every pixel in a block", func() {
		c.Expect(core.Power(&b3, &b4), gospec.Equals, uint64(192))
		c.Expect(core.Power(&b4, &b3), gospec.Equals, uint64(192))
	})
	for i := 0; i < 100; i++ {
		var b0, b1 core.Block8
		for i := 0; i < 100; i++ {
			core.Power(&b0, &b1)
		}
	}
}

func BenchmarkPowerAllSame(b *testing.B) {
	var b0, b1 core.Block8
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		core.Power(&b0, &b1)
	}
}

func BenchmarkPowerAllDifferent(b *testing.B) {
	var b0, b1 core.Block8
	for i := range b1 {
		b1[i] = byte(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		core.Power(&b0, &b1)
	}
}
