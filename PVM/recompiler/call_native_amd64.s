//go:build linux && amd64

#include "textflag.h"

// func callNative(guestBase uintptr, blockAddr uintptr, trampolineAddr uintptr)
//
// Assembly functions use ABI0 (stack-based arguments). The Go compiler
// generates a register-to-stack wrapper automatically, so we must
// explicitly load arguments from the stack into the registers the
// entry trampoline expects:
//   RAX = guestBase    (base address of guest memory / R15 source)
//   RBX = blockAddr    (target native code address)
//   CX  = trampoline   (entry trampoline address, called via CALL)
TEXT ·callNative(SB), NOSPLIT, $0-24
	MOVQ guestBase+0(FP), AX
	MOVQ blockAddr+8(FP), BX
	MOVQ trampolineAddr+16(FP), CX
	CALL CX
	RET
