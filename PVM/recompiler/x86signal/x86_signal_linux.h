#ifndef X86_SIGNAL_LINUX_H
#define X86_SIGNAL_LINUX_H

#include <stdint.h>

// Control region field offsets (negative from guest_base = R15).
// Must stay in sync with context.go constants.
#define OFF_RETURN_STACK   8
#define OFF_RETURN_ADDR   16
#define OFF_HEAP_POINTER  24
#define OFF_EXIT_PC       32
#define OFF_EXIT_REASON   40
#define OFF_GAS           48
#define OFF_REGISTERS    152

// PVM register count
#define PVM_REG_COUNT 13

// ExitReason encoding: type in top byte (bits 63..56), payload in lower bytes.
// PAGE_FAULT type = 4, so ExitPageFault = (4ULL << 56) | fault_addr
#define EXIT_PAGE_FAULT_TYPE 4ULL
#define EXIT_PANIC_TYPE      2ULL

// Indices into the ucontext gregset matching PVMToX86 register mapping.
// PVM reg -> x86 reg -> ucontext REG_Rxx index (from <sys/ucontext.h>)
//   RA=0  -> RAX -> REG_RAX=13
//   SP=1  -> RDX -> REG_RDX=12
//   T0=2  -> RBX -> REG_RBX=11
//   T1=3  -> RSI -> REG_RSI=9
//   T2=4  -> RDI -> REG_RDI=8
//   S0=5  -> R8  -> REG_R8=0
//   S1=6  -> R9  -> REG_R9=1
//   A0=7  -> R10 -> REG_R10=2
//   A1=8  -> R11 -> REG_R11=3
//   A2=9  -> R12 -> REG_R12=4
//   A3=10 -> R13 -> REG_R13=5
//   A4=11 -> R14 -> REG_R14=6
//   A5=12 -> RBP -> REG_RBP=10

void setup_signal_handler(void);
void set_vmctx(void *guest_base);
void set_fault_window(void *guest_base, void *code_start, void *code_end);
void clear_fault_window(void);

#endif // X86_SIGNAL_LINUX_H
