package utils

// func PowerLine(aRgba, bRgba [32]byte) float64

func PowerLineAsm(aRgba, bRgba []byte) float64

func PowerLine(aRgba, bRgba []byte) float64 {
	var power float64 = 0.0
	for i := 0; i < 32; i++ {
		diff := float64(int8(aRgba[i] - bRgba[i]))
		power += diff * diff
		i++
		diff = float64(int8(aRgba[i] - bRgba[i]))
		power += diff * diff
		i++
		diff = float64(int8(aRgba[i] - bRgba[i]))
		power += diff * diff

		// This extra increment will skip the alpha channel
		i++
	}
	return power
}

// func PowerLineFixed(aRgba, bRgba [32]byte) float64

// func PowerLineFixed(aRgba, bRgba [32]byte) float64 {
// 	var power float64 = 0.0
// 	for i := 0; i < 32; i++ {
// 		diff := float64(aRgba[i] - bRgba[i])
// 		power += diff * diff
// 		i++
// 		diff = float64(aRgba[i] - bRgba[i])
// 		power += diff * diff
// 		i++
// 		diff = float64(aRgba[i] - bRgba[i])
// 		power += diff * diff

// 		// This extra increment will skip the alpha channel
// 		i++
// 	}
// 	return power
// }
