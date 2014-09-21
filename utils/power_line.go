package utils

func PowerLine(aRgba, bRgba []byte) uint64 {
	var power uint64 = 0.0
	for i := 0; i < 32; i++ {
		diff := int64(int8(aRgba[i] - bRgba[i]))
		power += uint64(diff * diff)
		i++
		diff = int64(int8(aRgba[i] - bRgba[i]))
		power += uint64(diff * diff)
		i++
		diff = int64(int8(aRgba[i] - bRgba[i]))
		power += uint64(diff * diff)

		// This extra increment will skip the alpha channel
		i++
	}
	return power
}
