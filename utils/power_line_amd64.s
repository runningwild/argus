TEXT	·PowerLineAsm+0(SB),4,$16-56
	MOVQ	·aRgba+0(FP),DX
	MOVQ	·bRgba+24(FP),CX

	MOVSD	$(0.0),X0
	MOVSD	$(0.0),X1

	MOVQ	$0,AX

L2:	LEAQ	(DX)(AX*1),BX
	MOVBQZX	(BX),BX
	LEAQ	(CX)(AX*1),BP
	MOVBQZX	(BP),BP

	SUBQ	BP,BX
	MOVBLSX	BL,BP
	CVTSL2SD	BP,X0
	MULSD	X0,X0
	ADDSD	X1,X0
	MOVAPD	X0,X3
	INCQ	,AX
	LEAQ	(DX)(AX*1),BX
	MOVBQZX	(BX),BX
	LEAQ	(CX)(AX*1),BP
	MOVBQZX	(BP),BP
	SUBQ	BP,BX
	MOVBLSX	BL,BP
	CVTSL2SD	BP,X0
	MULSD	X0,X0
	ADDSD	X3,X0
	MOVAPD	X0,X3
	INCQ	,AX
	LEAQ	(DX)(AX*1),BX
	MOVBQZX	(BX),BX
	LEAQ	(CX)(AX*1),BP
	MOVBQZX	(BP),BP
	SUBQ	BP,BX
	MOVBLSX	BL,BP
	CVTSL2SD	BP,X0
	MULSD	X0,X0
	ADDSD	X3,X0
	MOVAPD	X0,X1
	ADDQ	$2, AX
	CMPQ	AX,$32
	JLT	$0,L2

L1:	MOVSD	X1,·noname+48(FP)
	RET	,
