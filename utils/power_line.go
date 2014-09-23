// +build !arm

// Default implementation of PowerLine for architectures that don't have an assumbly version.
package utils

func PowerLine(aRgba, bRgba []byte) uint64 {
	var power uint64 = 0.0
	var diff int64

	diff = int64(int(aRgba[0]) - int(bRgba[0]))
	power += uint64(diff * diff)

	diff = int64(int(aRgba[1]) - int(bRgba[1]))
	power += uint64(diff * diff)

	diff = int64(int(aRgba[2]) - int(bRgba[2]))
	power += uint64(diff * diff)

	diff = int64(int(aRgba[3]) - int(bRgba[3]))
	power += uint64(diff * diff)

	diff = int64(int(aRgba[4]) - int(bRgba[4]))
	power += uint64(diff * diff)

	diff = int64(int(aRgba[5]) - int(bRgba[5]))
	power += uint64(diff * diff)

	diff = int64(int(aRgba[6]) - int(bRgba[6]))
	power += uint64(diff * diff)

	diff = int64(int(aRgba[7]) - int(bRgba[7]))
	power += uint64(diff * diff)

	diff = int64(int(aRgba[8]) - int(bRgba[8]))
	power += uint64(diff * diff)

	diff = int64(int(aRgba[9]) - int(bRgba[9]))
	power += uint64(diff * diff)

	diff = int64(int(aRgba[10]) - int(bRgba[10]))
	power += uint64(diff * diff)

	diff = int64(int(aRgba[11]) - int(bRgba[11]))
	power += uint64(diff * diff)

	diff = int64(int(aRgba[12]) - int(bRgba[12]))
	power += uint64(diff * diff)

	diff = int64(int(aRgba[13]) - int(bRgba[13]))
	power += uint64(diff * diff)

	diff = int64(int(aRgba[14]) - int(bRgba[14]))
	power += uint64(diff * diff)

	diff = int64(int(aRgba[15]) - int(bRgba[15]))
	power += uint64(diff * diff)

	diff = int64(int(aRgba[16]) - int(bRgba[16]))
	power += uint64(diff * diff)

	diff = int64(int(aRgba[17]) - int(bRgba[17]))
	power += uint64(diff * diff)

	diff = int64(int(aRgba[18]) - int(bRgba[18]))
	power += uint64(diff * diff)

	diff = int64(int(aRgba[19]) - int(bRgba[19]))
	power += uint64(diff * diff)

	diff = int64(int(aRgba[20]) - int(bRgba[20]))
	power += uint64(diff * diff)

	diff = int64(int(aRgba[21]) - int(bRgba[21]))
	power += uint64(diff * diff)

	diff = int64(int(aRgba[22]) - int(bRgba[22]))
	power += uint64(diff * diff)

	diff = int64(int(aRgba[23]) - int(bRgba[23]))
	power += uint64(diff * diff)

	return power
}
