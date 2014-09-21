package utils_test

import (
	"github.com/runningwild/argus/utils"
	"image"
	"testing"
)

func BenchmarkPower(b *testing.B) {
	imgA := image.NewRGBA(image.Rect(0, 0, 8, 8))
	imgB := image.NewRGBA(image.Rect(0, 0, 8, 8))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		utils.Power(imgA, imgB, 0, 0, 100)
	}
}
