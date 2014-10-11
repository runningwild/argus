// func Power(a, b *[192]byte) uint64
TEXT	·Power+0(SB),4,$24-24
	//SUBQ	$24,SP
	MOVQ	$0, DX			// DX: The accumulator
	MOVQ	$0, CX			// CX: Loop counter
	MOVQ	·a+0(FP),R8		// R8: Address of block A
	MOVQ	·b+8(FP),R9		// R9: Address of block B
	LEAQ	(R8)(CX*1),AX	// AX: Addr A value
	LEAQ	(R9)(CX*1),BX	// BX: Addr B value

	MOVQ	$48, SI

MONKEY:
	CMPQ	CX, SI
	JGE		$0, DONE
	INCQ	, CX
	MOVQ	(AX),R8			// 4 byte value from block A
	MOVQ	(BX),R9			// 4 byte value from block B
	ADDQ	$4, AX
	ADDQ	$4, BX

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
	JMP		,MONKEY

DONE:
	MOVQ	DX, ·pow+16(FP)
	MOVQ	$0, DX
	RET
