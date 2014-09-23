TEXT	·PowerLine+0(SB),16,$36-32
	MOVW	$·aRgba+0(FP),R0
	MOVW	0(R0), R0		// R0 holds the address of the first element in array A

	MOVW	$·bRgba+12(FP),R1
	MOVW	0(R1), R1		// R1 holds the address of the first element in array B

	// MOVW	$0, R8			// R8 will accumulate the power (RAWR!!)

        MOVW    0(R0), R2       // Store 4 bytes of A in R2
        MOVW    0(R1), R3       // Store 4 bytes of B in R2
        AND     $255, R2, R4    // Put the lowest byte of A in R4
        AND     $255, R3, R5    // Put the lowest byte of B in R5
        SUB     R4, R5, R4      // Compute the power and add it to R8
        MUL     R4, R4, R8      // (This first one is can just be stored directly into R8)
        // ADD     R4, R8, R8      

        MOVW    R2>>8, R2       // Shift by 8 to get the next byte and repeat for all bytes
        MOVW    R3>>8, R3
        AND     $255, R2, R4
        AND     $255, R3, R5
        SUB     R4, R5, R4
        MUL     R4, R4, R4      // Unlike above we have to store the temporary result and
        ADD     R4, R8, R8      // add it to R8

        MOVW    R2>>8, R2
        MOVW    R3>>8, R3
        AND     $255, R2, R4
        AND     $255, R3, R5
        SUB     R4, R5, R4
        MUL     R4, R4, R4
        ADD     R4, R8, R8

        MOVW    R2>>8, R4
        MOVW    R3>>8, R5
        SUB     R4, R5, R4
        MUL     R4, R4, R4
        ADD     R4, R8, R8

        MOVW    4(R0), R2       // Grab the next 4 bytes and repeat
        MOVW    4(R1), R3
        AND     $255, R2, R4
        AND     $255, R3, R5
        SUB     R4, R5, R4
        MUL     R4, R4, R4
        ADD     R4, R8, R8

        MOVW    R2>>8, R2
        MOVW    R3>>8, R3
        AND     $255, R2, R4
        AND     $255, R3, R5
        SUB     R4, R5, R4
        MUL     R4, R4, R4
        ADD     R4, R8, R8

        MOVW    R2>>8, R2
        MOVW    R3>>8, R3
        AND     $255, R2, R4
        AND     $255, R3, R5
        SUB     R4, R5, R4
        MUL     R4, R4, R4
        ADD     R4, R8, R8

        MOVW    R2>>8, R4
        MOVW    R3>>8, R5
        SUB     R4, R5, R4
        MUL     R4, R4, R4
        ADD     R4, R8, R8

        MOVW    8(R0), R2
        MOVW    8(R1), R3
        AND     $255, R2, R4
        AND     $255, R3, R5
        SUB     R4, R5, R4
        MUL     R4, R4, R4
        ADD     R4, R8, R8

        MOVW    R2>>8, R2
        MOVW    R3>>8, R3
        AND     $255, R2, R4
        AND     $255, R3, R5
        SUB     R4, R5, R4
        MUL     R4, R4, R4
        ADD     R4, R8, R8

        MOVW    R2>>8, R2
        MOVW    R3>>8, R3
        AND     $255, R2, R4
        AND     $255, R3, R5
        SUB     R4, R5, R4
        MUL     R4, R4, R4
        ADD     R4, R8, R8

        MOVW    R2>>8, R4
        MOVW    R3>>8, R5
        SUB     R4, R5, R4
        MUL     R4, R4, R4
        ADD     R4, R8, R8

        MOVW    12(R0), R2
        MOVW    12(R1), R3
        AND     $255, R2, R4
        AND     $255, R3, R5
        SUB     R4, R5, R4
        MUL     R4, R4, R4
        ADD     R4, R8, R8

        MOVW    R2>>8, R2
        MOVW    R3>>8, R3
        AND     $255, R2, R4
        AND     $255, R3, R5
        SUB     R4, R5, R4
        MUL     R4, R4, R4
        ADD     R4, R8, R8

        MOVW    R2>>8, R2
        MOVW    R3>>8, R3
        AND     $255, R2, R4
        AND     $255, R3, R5
        SUB     R4, R5, R4
        MUL     R4, R4, R4
        ADD     R4, R8, R8

        MOVW    R2>>8, R4
        MOVW    R3>>8, R5
        SUB     R4, R5, R4
        MUL     R4, R4, R4
        ADD     R4, R8, R8

        MOVW    16(R0), R2
        MOVW    16(R1), R3
        AND     $255, R2, R4
        AND     $255, R3, R5
        SUB     R4, R5, R4
        MUL     R4, R4, R4
        ADD     R4, R8, R8

        MOVW    R2>>8, R2
        MOVW    R3>>8, R3
        AND     $255, R2, R4
        AND     $255, R3, R5
        SUB     R4, R5, R4
        MUL     R4, R4, R4
        ADD     R4, R8, R8

        MOVW    R2>>8, R2
        MOVW    R3>>8, R3
        AND     $255, R2, R4
        AND     $255, R3, R5
        SUB     R4, R5, R4
        MUL     R4, R4, R4
        ADD     R4, R8, R8

        MOVW    R2>>8, R4
        MOVW    R3>>8, R5
        SUB     R4, R5, R4
        MUL     R4, R4, R4
        ADD     R4, R8, R8

        MOVW    20(R0), R2
        MOVW    20(R1), R3
        AND     $255, R2, R4
        AND     $255, R3, R5
        SUB     R4, R5, R4
        MUL     R4, R4, R4
        ADD     R4, R8, R8

        MOVW    R2>>8, R2
        MOVW    R3>>8, R3
        AND     $255, R2, R4
        AND     $255, R3, R5
        SUB     R4, R5, R4
        MUL     R4, R4, R4
        ADD     R4, R8, R8

        MOVW    R2>>8, R2
        MOVW    R3>>8, R3
        AND     $255, R2, R4
        AND     $255, R3, R5
        SUB     R4, R5, R4
        MUL     R4, R4, R4
        ADD     R4, R8, R8

        MOVW    R2>>8, R4
        MOVW    R3>>8, R5
        SUB     R4, R5, R4
        MUL     R4, R4, R4
        ADD     R4, R8, R8

	MOVW	$0, R1
	MOVW	R8,·r2+24(FP)		// FP+24 is the low part of the return value
	MOVW	R1,·r2+28(FP)		// FP+28 is the high part of the return value
	RET

