package core

var zigPattern Block64_32

func init() {
	zigPattern = Block64_32{
		0, 1, 8, 16, 9, 2, 3, 10, 17, 24, 32, 25, 18, 11, 4, 5, 12, 19, 26, 33, 40, 48, 41, 34, 27, 20, 13, 6, 7, 14, 21, 28, 35, 42, 49, 56, 57, 50, 43, 36, 29, 22, 15, 23, 30, 37, 44, 51, 58, 59, 52, 45, 38, 31, 39, 46, 53, 60, 61, 54, 47, 55, 62, 63,
	}
	// do a quick check:
	m := make(map[int]bool)
	for _, v := range zigPattern {
		m[int(v)] = true
	}
	if len(m) != len(zigPattern) {
		panic("zigPattern is wrong")
	}
}

func ZigSlow(b *Block64_32) {
	var tmp Block64_32
	for i := range b {
		tmp[i] = b[i]
	}
	for i, j := range zigPattern {
		b[i] = tmp[j]
	}
}
func ZagSlow(b *Block64_32) {
	var tmp Block64_32
	for i := range b {
		tmp[i] = b[i]
	}
	for i, j := range zigPattern {
		b[j] = tmp[i]
	}
}

func Zig(b *Block64_32) {
	var buf int32
	buf = b[2]
	b[2] = b[8]
	b[8] = b[17]
	b[17] = b[19]
	b[19] = b[33]
	b[33] = b[42]
	b[42] = b[15]
	b[15] = b[5]
	b[5] = buf
	buf = b[3]
	b[3] = b[16]
	b[16] = b[12]
	b[12] = b[18]
	b[18] = b[26]
	b[26] = b[13]
	b[13] = b[11]
	b[11] = b[25]
	b[25] = b[20]
	b[20] = b[40]
	b[40] = b[29]
	b[29] = b[14]
	b[14] = b[4]
	b[4] = b[9]
	b[9] = b[24]
	b[24] = b[27]
	b[27] = b[6]
	b[6] = buf
	buf = b[7]
	b[7] = b[10]
	b[10] = b[32]
	b[32] = b[35]
	b[35] = b[56]
	b[56] = b[53]
	b[53] = b[31]
	b[31] = b[28]
	b[28] = buf
	buf = b[21]
	b[21] = b[48]
	b[48] = b[58]
	b[58] = b[61]
	b[61] = b[55]
	b[55] = b[46]
	b[46] = b[44]
	b[44] = b[30]
	b[30] = buf
	buf = b[22]
	b[22] = b[41]
	b[41] = buf
	buf = b[23]
	b[23] = b[34]
	b[34] = b[49]
	b[49] = b[59]
	b[59] = b[54]
	b[54] = b[39]
	b[39] = b[36]
	b[36] = b[57]
	b[57] = b[60]
	b[60] = b[47]
	b[47] = b[51]
	b[51] = b[45]
	b[45] = b[37]
	b[37] = b[50]
	b[50] = b[52]
	b[52] = b[38]
	b[38] = b[43]
	b[43] = buf
}
func Zag(b *Block64_32) {
	var buf int32
	buf = b[2]
	b[2] = b[5]
	b[5] = b[15]
	b[15] = b[42]
	b[42] = b[33]
	b[33] = b[19]
	b[19] = b[17]
	b[17] = b[8]
	b[8] = buf
	buf = b[3]
	b[3] = b[6]
	b[6] = b[27]
	b[27] = b[24]
	b[24] = b[9]
	b[9] = b[4]
	b[4] = b[14]
	b[14] = b[29]
	b[29] = b[40]
	b[40] = b[20]
	b[20] = b[25]
	b[25] = b[11]
	b[11] = b[13]
	b[13] = b[26]
	b[26] = b[18]
	b[18] = b[12]
	b[12] = b[16]
	b[16] = buf
	buf = b[7]
	b[7] = b[28]
	b[28] = b[31]
	b[31] = b[53]
	b[53] = b[56]
	b[56] = b[35]
	b[35] = b[32]
	b[32] = b[10]
	b[10] = buf
	buf = b[21]
	b[21] = b[30]
	b[30] = b[44]
	b[44] = b[46]
	b[46] = b[55]
	b[55] = b[61]
	b[61] = b[58]
	b[58] = b[48]
	b[48] = buf
	buf = b[22]
	b[22] = b[41]
	b[41] = buf
	buf = b[23]
	b[23] = b[43]
	b[43] = b[38]
	b[38] = b[52]
	b[52] = b[50]
	b[50] = b[37]
	b[37] = b[45]
	b[45] = b[51]
	b[51] = b[47]
	b[47] = b[60]
	b[60] = b[57]
	b[57] = b[36]
	b[36] = b[39]
	b[39] = b[54]
	b[54] = b[59]
	b[59] = b[49]
	b[49] = b[34]
	b[34] = buf
}
