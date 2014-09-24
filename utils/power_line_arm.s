// WORD $0xe0288494;       // R8 = R8 + R4 *R4
// 1110 -> cond = always execute
// 000000 -> empty
// 1 -> multiply and accumulate
// 0 -> do not alter condition codes
// regs 8 8 4 (1001) 4  (1001 is the MLA code)
// 1110 0000 - 0010 1000 - 1000 0100 - 1001 0100

TEXT	·PowerLine+0(SB),16,$36-32
	MOVW	$·aRgba+0(FP),R0
	MOVW	0(R0), R0		// R0 holds the address of the first element in array A

	MOVW	$·bRgba+12(FP),R1
	MOVW	0(R1), R1		// R1 holds the address of the first element in array B

        MOVW	$0, R8			// R8 will accumulate the power (RAWR!!)

        // R6 is the inner loop variable
        MOVW    $0, R6

        START:
        ADD     $1, R6, R6
        CMP     $7, R6
        BEQ     DONE

        MOVW    0(R0), R2       // Store 4 bytes of A in R2
        MOVW    0(R1), R3       // Store 4 bytes of B in R2
        AND     $255, R2, R4    // Put the lowest byte of A in R4
        AND     $255, R3, R5    // Put the lowest byte of B in R5
        SUB     R4, R5, R4      // Compute the power and add it to R8

        WORD $0xe0288494;       // R8 = R8 + R4 *R4

        MOVW    R2>>8, R2       // Shift by 8 to get the next byte and repeat for all bytes
        MOVW    R3>>8, R3
        AND     $255, R2, R4
        AND     $255, R3, R5
        SUB     R4, R5, R4
        WORD $0xe0288494;       // R8 = R8 + R4 *R4

        MOVW    R2>>8, R2
        MOVW    R3>>8, R3
        AND     $255, R2, R4
        AND     $255, R3, R5
        SUB     R4, R5, R4
        WORD $0xe0288494;       // R8 = R8 + R4 *R4

        MOVW    R2>>8, R4
        MOVW    R3>>8, R5
        SUB     R4, R5, R4
        WORD $0xe0288494;       // R8 = R8 + R4 *R4

        ADD     $4, R0, R0
        ADD     $4, R1, R1

        B       START


        DONE:
	MOVW	$0, R1
	MOVW	R8,·r2+24(FP)		// FP+24 is the low part of the return value
	MOVW	R1,·r2+28(FP)		// FP+28 is the high part of the return value
	RET


