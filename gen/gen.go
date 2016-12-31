package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/jkl1337/go-chromath"
	"golang.org/x/crypto/sha3"
)

var (
	rgb2xyz = chromath.NewRGBTransformer(&chromath.SpaceSRGB, nil, nil, nil, 0.0, nil)
	lab2xyz = chromath.NewLabTransformer(nil)
)

func convert(r, g, b byte) (L, A, B byte) {
	lab := lab2xyz.Invert(rgb2xyz.Convert(chromath.RGB{float64(r) / 256, float64(g) / 256, float64(b) / 256}))

	// These numbers shift LAB so that everything is within [0, 255] without scaling any channel by
	// a different amount than the others.
	Lf := lab.L() / maxRange * 256
	Af := (lab.A() + 87.64619389265587) / maxRange * 256
	Bf := (lab.B() + 125.98932719947872) / maxRange * 256
	L = byte(constrain(Lf, 0, 255))
	A = byte(constrain(Af, 0, 255))
	B = byte(constrain(Bf, 0, 255))
	return
}

const (
	scale    = 1.0
	maxRange = 210.477981223
)

func constrain(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func invert(L, A, B byte) (r, g, b byte) {
	lab := chromath.Lab{
		(float64(L) / 256 * maxRange),
		(float64(A)/256*maxRange - 87.64619389265587),
		(float64(B)/256*maxRange - 125.98932719947872),
	}
	rgb := rgb2xyz.Invert(lab2xyz.Convert(lab))
	r = byte(constrain(rgb.R()*256, 0, 255))
	g = byte(constrain(rgb.G()*256, 0, 255))
	b = byte(constrain(rgb.B()*256, 0, 255))
	return
}

func main() {
	var worst float64
	buf := bytes.NewBuffer(nil)
	for r := 0; r < 256; r++ {
		for g := 0; g < 256; g++ {
			for b := 0; b < 256; b++ {
				L, A, B := convert(byte(r), byte(g), byte(b))
				r1, g1, b1 := invert(L, A, B)
				dr := float64(r1) - float64(r)
				dg := float64(g1) - float64(g)
				db := float64(b1) - float64(b)
				pow := dr*dr + dg*dg + db*db
				if pow > worst {
					worst = pow
					// fmt.Printf("%v: %02x%02x%02x -> %02x%02x%02x\n", worst, r, g, b, r1, g1, b1)
				}
				buf.Write([]byte{L, A, B})
			}
		}
	}
	hash := fmt.Sprintf("%x", sha3.Sum256(buf.Bytes()))
	if hash != "3dd8b13140f13b8f34ec1bb8f061bdf933519f3c742a436e36951cf9fc306bda" {
		log.Fatal("hashes don't match, maybe floating point error is to blame?\n")
	}
	ioutil.WriteFile("rgb2lab", buf.Bytes(), 0664)
}
