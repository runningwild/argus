// func Power(a, b *[192]byte) uint64
TEXT	·Power+0(SB),4,$24-24
	//SUBQ	$24,SP
	MOVQ	$0, DX			// DX: The accumulator
	MOVQ	$0, CX			// CX: Loop counter
	MOVQ	·a+0(FP),R8		// R8: Address of block A
	MOVQ	·b+8(FP),R9		// R9: Address of block B
	LEAQ	(R8)(CX*1),AX	// AX: Addr A value
	LEAQ	(R9)(CX*1),BX	// BX: Addr B value

	MOVQ	$24, SI

LOOPINC:
	CMPQ	CX, SI
	JGE		$0, DONE
	INCQ	CX
	MOVQ	(AX),R8			// 8 byte value from block A
	MOVQ	(BX),R9			// 8 byte value from block B
	ADDQ	$8, AX
	ADDQ	$8, BX

	// Compare the values from A and B, if equal then skip this round
	CMPQ	R8, R9
	JEQ		$0, LOOPINC

	// No need to SHR on the first one.
	MOVQ    R8, R10
	ANDQ	$255, R10
	MOVQ    R9, R11
	ANDQ	$255, R11
	SUBQ	R10, R11
	IMULQ	R11, R11
	ADDQ	R11, DX

	MOVQ    R8, R10
	SHRQ	$8, R10
	ANDQ	$255, R10
	MOVQ    R9, R11
	SHRQ	$8, R11
	ANDQ	$255, R11
	SUBQ	R10, R11
	IMULQ	R11, R11
	ADDQ	R11, DX

	MOVQ    R8, R10
	SHRQ	$16, R10
	ANDQ	$255, R10
	MOVQ    R9, R11
	SHRQ	$16, R11
	ANDQ	$255, R11
	SUBQ	R10, R11
	IMULQ	R11, R11
	ADDQ	R11, DX

	MOVQ    R8, R10
	SHRQ	$24, R10
	ANDQ	$255, R10
	MOVQ    R9, R11
	SHRQ	$24, R11
	ANDQ	$255, R11
	SUBQ	R10, R11
	IMULQ	R11, R11
	ADDQ	R11, DX

	MOVQ    R8, R10
	SHRQ	$32, R10
	ANDQ	$255, R10
	MOVQ    R9, R11
	SHRQ	$32, R11
	ANDQ	$255, R11
	SUBQ	R10, R11
	IMULQ	R11, R11
	ADDQ	R11, DX

	MOVQ    R8, R10
	SHRQ	$40, R10
	ANDQ	$255, R10
	MOVQ    R9, R11
	SHRQ	$40, R11
	ANDQ	$255, R11
	SUBQ	R10, R11
	IMULQ	R11, R11
	ADDQ	R11, DX

	MOVQ    R8, R10
	SHRQ	$48, R10
	ANDQ	$255, R10
	MOVQ    R9, R11
	SHRQ	$48, R11
	ANDQ	$255, R11
	SUBQ	R10, R11
	IMULQ	R11, R11
	ADDQ	R11, DX

	// No need to AND on the last one because there are no other bytes left
	MOVQ    R8, R10
	SHRQ	$56, R10
	MOVQ    R9, R11
	SHRQ	$56, R11
	SUBQ	R10, R11
	IMULQ	R11, R11
	ADDQ	R11, DX
	JMP		LOOPINC

DONE:
	MOVQ	DX, ·pow+16(FP)
	RET
