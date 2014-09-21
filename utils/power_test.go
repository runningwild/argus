package utils_test

import (
	"github.com/runningwild/argus/utils"
	"image"
	"testing"
)

func BenchmarkSimple(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		utils.Foo(i)
	}
}

func TestSimple(t *testing.T) {
	foo1 := utils.Foo(1)
	if foo1 != 2 {
		t.Errorf("Foo(1) returned %v failed", foo1)
	}
	foo12 := utils.Foo(12)
	if foo12 != 13 {
		t.Errorf("Foo(12) returned %v failed", foo12)
	}
	foom100 := utils.Foo(-100)
	if foom100 != -99 {
		t.Errorf("Foo(-100) returned %v failed", foom100)
	}
	foo7777 := utils.Foo(7777)
	if foo7777 != 7778 {
		t.Errorf("Foo(7777) returned %v failed", foo7777)
	}
}

func TestPower(t *testing.T) {
	imgA := image.NewRGBA(image.Rect(0, 0, 10, 10))
	imgB := image.NewRGBA(image.Rect(0, 0, 10, 10))

	pow, over := utils.Power(imgA, imgB, 0, 0, 1)
	if pow != 0 {
		t.Errorf("Expected power of 0, got %v.\n", pow)
	}
	if over {
		t.Errorf("Expected not to be over.\n")
	}

	offset := imgA.PixOffset(0, 0)
	imgA.Pix[offset+0] = 10
	imgA.Pix[offset+1] = 10
	imgA.Pix[offset+2] = 10
	pow, over = utils.Power(imgA, imgB, 0, 0, 1)
	if pow != 300 {
		t.Errorf("Expected power of 300, got %v.\n", pow)
	}
	if !over {
		t.Errorf("Expected to be over.\n")
	}

	pow, over = utils.Power(imgA, imgB, 1, 1, 1)
	if pow != 0 {
		t.Errorf("Expected power of 0, got %v.\n", pow)
	}
	if over {
		t.Errorf("Expected not to be over.\n")
	}

	offset = imgA.PixOffset(0, 0)
	imgB.Pix[offset+0] = 10
	imgB.Pix[offset+1] = 10
	imgB.Pix[offset+2] = 10
	offset = imgA.PixOffset(9, 9)
	imgB.Pix[offset+0] = 0
	imgB.Pix[offset+1] = 10
	imgB.Pix[offset+2] = 20
	pow, over = utils.Power(imgA, imgB, 0, 0, 1)
	if pow != 0 {
		t.Errorf("Expected power of 0, got %v.\n", pow)
	}
	if over {
		t.Errorf("Expected not to be over.\n")
	}
	pow, over = utils.Power(imgA, imgB, 1, 1, 1)
	if pow != 0 {
		t.Errorf("Expected power of 0, got %v.\n", pow)
	}
	if over {
		t.Errorf("Expected not to be over.\n")
	}
	pow, over = utils.Power(imgA, imgB, 2, 2, 1)
	if pow != 500 {
		t.Errorf("Expected power of 500, got %v.\n", pow)
	}
	if !over {
		t.Errorf("Expected to be over.\n")
	}
}

func TestPowerAsm(t *testing.T) {
	imgA := image.NewRGBA(image.Rect(0, 0, 10, 10))
	imgB := image.NewRGBA(image.Rect(0, 0, 10, 10))

	pow, over := utils.PowerAsm(imgA, imgB, 0, 0, 1)
	if pow != 0 {
		t.Errorf("Expected power of 0, got %v.\n", pow)
	}
	if over {
		t.Errorf("Expected not to be over.\n")
	}

	offset := imgA.PixOffset(0, 0)
	imgA.Pix[offset+0] = 10
	imgA.Pix[offset+1] = 10
	imgA.Pix[offset+2] = 10
	pow, over = utils.PowerAsm(imgA, imgB, 0, 0, 1)
	if pow != 300 {
		t.Errorf("Expected power of 300, got %v.\n", pow)
	}
	if !over {
		t.Errorf("Expected to be over.\n")
	}

	pow, over = utils.PowerAsm(imgA, imgB, 1, 1, 1)
	if pow != 0 {
		t.Errorf("Expected power of 0, got %v.\n", pow)
	}
	if over {
		t.Errorf("Expected not to be over.\n")
	}

	offset = imgA.PixOffset(0, 0)
	imgB.Pix[offset+0] = 10
	imgB.Pix[offset+1] = 10
	imgB.Pix[offset+2] = 10
	offset = imgA.PixOffset(9, 9)
	imgB.Pix[offset+0] = 0
	imgB.Pix[offset+1] = 10
	imgB.Pix[offset+2] = 20
	pow, over = utils.PowerAsm(imgA, imgB, 0, 0, 1)
	if pow != 0 {
		t.Errorf("Expected power of 0, got %v.\n", pow)
	}
	if over {
		t.Errorf("Expected not to be over.\n")
	}
	pow, over = utils.PowerAsm(imgA, imgB, 1, 1, 1)
	if pow != 0 {
		t.Errorf("Expected power of 0, got %v.\n", pow)
	}
	if over {
		t.Errorf("Expected not to be over.\n")
	}
	pow, over = utils.PowerAsm(imgA, imgB, 2, 2, 1)
	if pow != 500 {
		t.Errorf("Expected power of 500, got %v.\n", pow)
	}
	if !over {
		t.Errorf("Expected to be over.\n")
	}
}

func BenchmarkPower(b *testing.B) {
	imgA := image.NewRGBA(image.Rect(0, 0, 8, 8))
	imgB := image.NewRGBA(image.Rect(0, 0, 8, 8))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		utils.Power(imgA, imgB, 0, 0, 100)
	}
}

func BenchmarkPowerAsm(b *testing.B) {
	imgA := image.NewRGBA(image.Rect(0, 0, 8, 8))
	imgB := image.NewRGBA(image.Rect(0, 0, 8, 8))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		utils.PowerAsm(imgA, imgB, 0, 0, 100)
	}
}
