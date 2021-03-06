package core_test

import (
	"testing"

	"github.com/runningwild/argus/core"
	. "github.com/smartystreets/goconvey/convey"
	"fmt"
)

func TestFixedPointResemblesFloatingPoint(t *testing.T) {
	Convey("f", t, func() {
		for ri := 0; ri < 256; ri++ {
			for gi := 0; gi < 256; gi++ {
				for bi := 0; bi < 256; bi++ {
					Lf, Af, Bf := core.RGBtoLABf(float64(ri), float64(gi), float64(bi))
					Li, Ai, Bi := core.RGBtoLABi(int32(ri), int32(gi), int32(bi))
					fmt.Printf("%0.2f %0.2f %0.2f    %0.2f %0.2f %0.2f\n", Lf, Af, Bf, float64(Li)/(1<<20), float64(Ai)/(1<<20), float64(Bi)/(1<<20))
					dL := (Lf - float64(Li)/(1<<20))
					dA := (Af - float64(Ai)/(1<<20))
					dB := (Bf - float64(Bi)/(1<<20))
					diff := dL*dL + dA*dA + dB*dB
					So(diff, ShouldBeLessThan, 10)
				}}}
	})

}


func TestRGBToLABToRGB(t *testing.T) {
	Convey("f", t, func() {
		worst := 0.0
		var labs [3][2]float64
		for ri := 0; ri < 256; ri++ {
			for gi := 0; gi < 256; gi++ {
				for bi := 0; bi < 256; bi++ {
					lf, af, bf := core.RGBtoLABf(float64(ri), float64(gi), float64(bi))
					if lf < labs[0][0] {
						labs[0][0] = lf
					}
					if lf > labs[0][1] {
						labs[0][1] = lf
					}
					if af < labs[1][0] {
						labs[1][0] = af
					}
					if af > labs[1][1] {
						labs[1][1] = af
					}
					if bf < labs[2][0] {
						labs[2][0] = bf
					}
					if bf > labs[2][1] {
						labs[2][1] = bf
					}
					r, g, b := byte(ri), byte(gi), byte(bi)
					R, G, B := core.LABtoRGB(core.RGBtoLAB(r, g, b))
					dr := float64(r) - float64(R)
					dg := float64(g) - float64(G)
					db := float64(b) - float64(B)
					pow := dr*dr + dg*dg + db*db
					if pow > worst {
						worst = pow
						fmt.Printf("Worst: %v: %02x%02x%02x %02x%02x%02x\n", worst, r, g, b, R, G, B)
						Lb,Ab,Bb := core.RGBtoLAB(r, g, b)
						fmt.Printf("%d %d %d -> %d %d %d -> %d %d %d\n", r, g, b,Lb,Ab,Bb, R, G, B)
						{
							lf, af, bf := core.RGBtoLABf(float64(ri), float64(gi), float64(bi))
							rfx,gfx,bfx := core.LABtoRGBf(lf, af, bf)
							rf,gf,bf := core.LABtoRGBf(float64(byte(lf)), float64(byte(af)), float64(byte(bf)))
							fmt.Printf("Raw %d %d %d -> %f %f %f\n", ri, gi, bi, rfx, gfx, bfx)
							fmt.Printf("%d %d %d -> %f %f %f -> %f %f %f\n", ri, gi, bi, lf, af, bf, rf, gf, bf)
						}
					}
					So(pow, ShouldBeLessThan, 900)
				}
			}
		}
		fmt.Printf("\n\n%f %f\n%f %f\n%f %f\n", labs[0][0], labs[0][1], labs[1][0], labs[1][1], labs[2][0], labs[2][1])
	})
}

func BenchmarkLABtoRGB(b *testing.B) {
	for i:=0;i<b.N;i++ {
		core.LABtoRGB(0,0,0)
	}
}

func BenchmarkRGBtoLAB(b *testing.B) {
	for i:=0;i<b.N;i++ {
		core.RGBtoLAB(0,0,0)
	}
}

func BenchmarkLABtoRGBf(b *testing.B) {
	for i:=0;i<b.N;i++ {
		core.LABtoRGBf(0,0,0)
	}
}

func BenchmarkRGBtoLABf(b *testing.B) {
	for i:=0;i<b.N;i++ {
		core.RGBtoLABf(0,0,0)
	}
}

func BenchmarkLABtoRGBfast(b *testing.B) {
	for i:=0;i<b.N;i++ {
		core.LABtoRGBfast(0,0,0)
	}
}

func BenchmarkRGBtoLABfast(b *testing.B) {
	for i:=0;i<b.N;i++ {
		core.RGBtoLABfast(0,0,0)
	}
}
