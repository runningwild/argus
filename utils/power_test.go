package utils_test

import (
	"github.com/runningwild/argus/utils"
	"image"
	"testing"
)

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

func BenchmarkPower(b *testing.B) {
	imgA := image.NewRGBA(image.Rect(0, 0, 8, 8))
	imgB := image.NewRGBA(image.Rect(0, 0, 8, 8))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		utils.Power(imgA, imgB, 0, 0, 100)
	}
}
