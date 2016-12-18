// +build !arm
// +build !amd64

// Default implementations for architectures without assembly versions.
package core

func Power(a, b *Block8RGB) (pow uint64) {
	return PowerSlow(a, b)
}
