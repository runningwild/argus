package utils_test

import (
	"github.com/runningwild/argus/rgb"
	"github.com/runningwild/argus/utils"
	"image"
	"testing"
)

func TestPowerLine(t *testing.T) {
	a := make([]byte, 24)
	b := make([]byte, 24)
	for i := 0; i < 24; i++ {
		a[i] = (byte)(i*13 + 17)
		b[i] = (byte)(i*253 + 11)
	}
	pow := utils.PowerLine(a, b)
	if pow != 305760 {
		t.Errorf("Expected power of 305760, got %d.\n", pow)
	}
}

func TestPower(t *testing.T) {
	imgA := rgb.Make(image.Rect(0, 0, 10, 10))
	imgB := rgb.Make(image.Rect(0, 0, 10, 10))

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

	imgA = rgb.Make(image.Rect(0, 0, 8, 8))
	imgB = rgb.Make(image.Rect(0, 0, 8, 8))
	var max uint64
	for i := range imgA.Pix {
		imgA.Pix[i] = 255
		max += 255 * 255
	}
	for i := range imgB.Pix {
		imgB.Pix[i] = 0
	}
	pow = utils.PowerLine(imgA.Pix, imgB.Pix)
	if pow != 255*255*3*8 {
		t.Errorf("Expected power of %d, got %d.\n", 255*255*3*8, pow)
	}
	pow, over = utils.Power(imgA, imgB, 0, 0, 1)
	if pow != max {
		t.Errorf("Expected power of %d, got %d.\n", max, pow)
	}
	if !over {
		t.Errorf("Expected to be over.\n")
	}
}

func BenchmarkPower(b *testing.B) {
	imgA := rgb.Make(image.Rect(0, 0, 8, 8))
	imgB := rgb.Make(image.Rect(0, 0, 8, 8))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		utils.Power(imgA, imgB, 0, 0, 100)
	}
}
