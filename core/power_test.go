package core_test

import (
	"github.com/runningwild/argus/core"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestPower(t *testing.T) {
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

	Convey("Blocks have zero power relative to themselves", t, func() {
		So(core.Power(&b0, &b0), ShouldEqual, uint64(0))
		So(core.Power(&b1, &b1), ShouldEqual, uint64(0))
		So(core.Power(&b2, &b2), ShouldEqual, uint64(0))
		So(core.Power(&b3, &b3), ShouldEqual, uint64(0))
	})
	Convey("Simple manual power check", t, func() {
		So(core.Power(&b0, &b1), ShouldEqual, uint64(600))
		So(core.Power(&b1, &b0), ShouldEqual, uint64(600))
	})
	Convey("Simple manual power check", t, func() {
		So(core.Power(&b0, &b2), ShouldEqual, uint64(9100))
		So(core.Power(&b2, &b0), ShouldEqual, uint64(9100))
	})
	Convey("Simple manual power check", t, func() {
		So(core.Power(&b1, &b2), ShouldEqual, uint64(13900))
		So(core.Power(&b2, &b1), ShouldEqual, uint64(13900))
	})
	Convey("Simple manual power check - zero vs zero", t, func() {
		So(core.Power(&b2, &b2), ShouldEqual, uint64(0))
	})
	Convey("Power function should cover every pixel in a block", t, func() {
		So(core.Power(&b3, &b4), ShouldEqual, uint64(192))
		So(core.Power(&b4, &b3), ShouldEqual, uint64(192))
	})
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
