package utils

func PowerLine(aRgba, bRgba []byte) uint64

func PowerLineNormal(aRgba, bRgba []byte) uint64 {
	var power uint64 = 0.0
	var diff int64

	diff = int64(int8(aRgba[0] - bRgba[0]))
	power += uint64(diff * diff)
	diff = int64(int8(aRgba[1] - bRgba[1]))
	power += uint64(diff * diff)
	diff = int64(int8(aRgba[2] - bRgba[2]))
	power += uint64(diff * diff)
	diff = int64(int8(aRgba[4] - bRgba[4]))
	power += uint64(diff * diff)
	diff = int64(int8(aRgba[5] - bRgba[5]))
	power += uint64(diff * diff)
	diff = int64(int8(aRgba[6] - bRgba[6]))
	power += uint64(diff * diff)
	diff = int64(int8(aRgba[8] - bRgba[8]))
	power += uint64(diff * diff)
	diff = int64(int8(aRgba[9] - bRgba[9]))
	power += uint64(diff * diff)
	diff = int64(int8(aRgba[10] - bRgba[10]))
	power += uint64(diff * diff)
	diff = int64(int8(aRgba[12] - bRgba[12]))
	power += uint64(diff * diff)
	diff = int64(int8(aRgba[13] - bRgba[13]))
	power += uint64(diff * diff)
	diff = int64(int8(aRgba[14] - bRgba[14]))
	power += uint64(diff * diff)
	diff = int64(int8(aRgba[16] - bRgba[16]))
	power += uint64(diff * diff)
	diff = int64(int8(aRgba[17] - bRgba[17]))
	power += uint64(diff * diff)
	diff = int64(int8(aRgba[18] - bRgba[18]))
	power += uint64(diff * diff)
	diff = int64(int8(aRgba[20] - bRgba[20]))
	power += uint64(diff * diff)
	diff = int64(int8(aRgba[21] - bRgba[21]))
	power += uint64(diff * diff)
	diff = int64(int8(aRgba[22] - bRgba[22]))
	power += uint64(diff * diff)
	diff = int64(int8(aRgba[24] - bRgba[24]))
	power += uint64(diff * diff)
	diff = int64(int8(aRgba[25] - bRgba[25]))
	power += uint64(diff * diff)
	diff = int64(int8(aRgba[26] - bRgba[26]))
	power += uint64(diff * diff)
	diff = int64(int8(aRgba[28] - bRgba[28]))
	power += uint64(diff * diff)
	diff = int64(int8(aRgba[29] - bRgba[29]))
	power += uint64(diff * diff)
	diff = int64(int8(aRgba[30] - bRgba[30]))
	power += uint64(diff * diff)

	return power
}

