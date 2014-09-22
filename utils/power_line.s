	TEXT	·PowerLine+0(SB),16,$36-32
	MOVW	$·aRgba+0(FP),R0
	MOVW	0(R0), R0		// R0 holds the address of the first element in array A

	MOVW	$·bRgba+12(FP),R1
	MOVW	0(R1), R1		// R1 holds the address of the first element in array B

	MOVW	$0, R8			// R8 will accumulate the power (RAWR!!)

        MOVB    0(R0), R2
        MOVB    0(R1), R3
        SUB     R3, R2, R2
        MUL     R2, R2, R2
        ADD     R2, R8, R8

        MOVB    1(R0), R2
        MOVB    1(R1), R3
        SUB     R3, R2, R2
        MUL     R2, R2, R2
        ADD     R2, R8, R8

        MOVB    2(R0), R2
        MOVB    2(R1), R3
        SUB     R3, R2, R2
        MUL     R2, R2, R2
        ADD     R2, R8, R8

        MOVB    4(R0), R2
        MOVB    4(R1), R3
        SUB     R3, R2, R2
        MUL     R2, R2, R2
        ADD     R2, R8, R8

        MOVB    5(R0), R2
        MOVB    5(R1), R3
        SUB     R3, R2, R2
        MUL     R2, R2, R2
        ADD     R2, R8, R8

        MOVB    6(R0), R2
        MOVB    6(R1), R3
        SUB     R3, R2, R2
        MUL     R2, R2, R2
        ADD     R2, R8, R8

        MOVB    8(R0), R2
        MOVB    8(R1), R3
        SUB     R3, R2, R2
        MUL     R2, R2, R2
        ADD     R2, R8, R8

        MOVB    9(R0), R2
        MOVB    9(R1), R3
        SUB     R3, R2, R2
        MUL     R2, R2, R2
        ADD     R2, R8, R8

        MOVB    10(R0), R2
        MOVB    10(R1), R3
        SUB     R3, R2, R2
        MUL     R2, R2, R2
        ADD     R2, R8, R8

        MOVB    12(R0), R2
        MOVB    12(R1), R3
        SUB     R3, R2, R2
        MUL     R2, R2, R2
        ADD     R2, R8, R8

        MOVB    13(R0), R2
        MOVB    13(R1), R3
        SUB     R3, R2, R2
        MUL     R2, R2, R2
        ADD     R2, R8, R8

        MOVB    14(R0), R2
        MOVB    14(R1), R3
        SUB     R3, R2, R2
        MUL     R2, R2, R2
        ADD     R2, R8, R8

        MOVB    16(R0), R2
        MOVB    16(R1), R3
        SUB     R3, R2, R2
        MUL     R2, R2, R2
        ADD     R2, R8, R8

        MOVB    17(R0), R2
        MOVB    17(R1), R3
        SUB     R3, R2, R2
        MUL     R2, R2, R2
        ADD     R2, R8, R8

        MOVB    18(R0), R2
        MOVB    18(R1), R3
        SUB     R3, R2, R2
        MUL     R2, R2, R2
        ADD     R2, R8, R8

        MOVB    20(R0), R2
        MOVB    20(R1), R3
        SUB     R3, R2, R2
        MUL     R2, R2, R2
        ADD     R2, R8, R8

        MOVB    21(R0), R2
        MOVB    21(R1), R3
        SUB     R3, R2, R2
        MUL     R2, R2, R2
        ADD     R2, R8, R8

        MOVB    22(R0), R2
        MOVB    22(R1), R3
        SUB     R3, R2, R2
        MUL     R2, R2, R2
        ADD     R2, R8, R8

        MOVB    24(R0), R2
        MOVB    24(R1), R3
        SUB     R3, R2, R2
        MUL     R2, R2, R2
        ADD     R2, R8, R8

        MOVB    25(R0), R2
        MOVB    25(R1), R3
        SUB     R3, R2, R2
        MUL     R2, R2, R2
        ADD     R2, R8, R8

        MOVB    26(R0), R2
        MOVB    26(R1), R3
        SUB     R3, R2, R2
        MUL     R2, R2, R2
        ADD     R2, R8, R8

        MOVB    28(R0), R2
        MOVB    28(R1), R3
        SUB     R3, R2, R2
        MUL     R2, R2, R2
        ADD     R2, R8, R8

        MOVB    29(R0), R2
        MOVB    29(R1), R3
        SUB     R3, R2, R2
        MUL     R2, R2, R2
        ADD     R2, R8, R8

        MOVB    30(R0), R2
        MOVB    30(R1), R3
        SUB     R3, R2, R2
        MUL     R2, R2, R2
        ADD     R2, R8, R8

	MOVW	$0, R1
	MOVW	R8,·r2+24(FP)		// FP+24 is the low part of the return value
	MOVW	R1,·r2+28(FP)		// FP+28 is the high part of the return value
	RET

