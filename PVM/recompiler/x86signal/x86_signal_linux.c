#define _GNU_SOURCE
#include <signal.h>
#include <ucontext.h>
#include <stdint.h>
#include <string.h>

#include "x86_signal_linux.h"

static __thread uint8_t *guest_base_ptr;
static __thread uintptr_t jit_code_start;
static __thread uintptr_t jit_code_end;

static struct sigaction old_sigsegv;
static struct sigaction old_sigbus;
static struct sigaction old_sigfpe;

// Must match recompiler.pvmRegSlot (control-region slot per PVM register index).
static const int pvm_reg_slot[PVM_REG_COUNT] = {
    10, 11, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 12,
};

static const int pvm_to_greg[PVM_REG_COUNT] = {
    REG_RAX,  // RA=0
    REG_RDX,  // SP=1
    REG_RBX,  // T0=2
    REG_RSI,  // T1=3
    REG_RDI,  // T2=4
    REG_R8,   // S0=5
    REG_R9,   // S1=6
    REG_R10,  // A0=7
    REG_R11,  // A1=8
    REG_R12,  // A2=9
    REG_R13,  // A3=10
    REG_R14,  // A4=11
    REG_RBP,  // A5=12
};

static int is_jit_fault(uint8_t *base, uintptr_t fault_addr) {
    if (base == NULL) return 0;
    uintptr_t lo = (uintptr_t)base;
    uintptr_t hi = lo + (4ULL * 1024 * 1024 * 1024) + 4096;
    return fault_addr >= lo && fault_addr < hi;
}

static int is_jit_code(uintptr_t rip) {
    return jit_code_start != 0 && rip >= jit_code_start && rip < jit_code_end;
}

static void chain_to_old(int sig, siginfo_t *info, void *uctx) {
    struct sigaction *old;
    switch (sig) {
    case SIGSEGV: old = &old_sigsegv; break;
    case SIGBUS:  old = &old_sigbus;  break;
    case SIGFPE:  old = &old_sigfpe;  break;
    default:      old = NULL;         break;
    }

    if (old != NULL && (old->sa_flags & SA_SIGINFO)) {
        old->sa_sigaction(sig, info, uctx);
    } else if (old != NULL && old->sa_handler != SIG_DFL && old->sa_handler != SIG_IGN) {
        old->sa_handler(sig);
    } else {
        struct sigaction sa;
        memset(&sa, 0, sizeof(sa));
        sa.sa_handler = SIG_DFL;
        sigaction(sig, &sa, NULL);
    }
}

static void store_pvm_regs(uint8_t *base, greg_t *gregs) {
    for (int i = 0; i < PVM_REG_COUNT; i++) {
        uint64_t *slot = (uint64_t *)(base - OFF_REGISTERS + (size_t)pvm_reg_slot[i] * 8);
        *slot = (uint64_t)gregs[pvm_to_greg[i]];
    }
}

static void jit_exit_panic(uint8_t *base, greg_t *gregs) {
    store_pvm_regs(base, gregs);
    *(uint64_t *)(base - OFF_EXIT_REASON) = EXIT_PANIC_TYPE << 56;

    uintptr_t return_addr  = *(uintptr_t *)(base - OFF_RETURN_ADDR);
    uintptr_t return_stack = *(uintptr_t *)(base - OFF_RETURN_STACK);
    gregs[REG_RIP] = (greg_t)return_addr;
    gregs[REG_RSP] = (greg_t)return_stack;
}

static void signal_handler(int sig, siginfo_t *info, void *uctx_void) {
    ucontext_t *uc = (ucontext_t *)uctx_void;
    greg_t *gregs = uc->uc_mcontext.gregs;
    uint8_t *base = guest_base_ptr;

    if (sig == SIGFPE) {
        uintptr_t rip = (uintptr_t)gregs[REG_RIP];
        if (!is_jit_code(rip) || base == NULL) {
            chain_to_old(sig, info, uctx_void);
            return;
        }
        jit_exit_panic(base, gregs);
        return;
    }

    uintptr_t fault_addr = (uintptr_t)info->si_addr;
    if (!is_jit_fault(base, fault_addr)) {
        chain_to_old(sig, info, uctx_void);
        return;
    }

    uint32_t guest_addr = (uint32_t)(fault_addr - (uintptr_t)base);
    if (guest_addr < 0x10000u) {
        jit_exit_panic(base, gregs);
    } else {
        store_pvm_regs(base, gregs);
        uint64_t exit_reason = (EXIT_PAGE_FAULT_TYPE << 56) | (uint64_t)guest_addr;
        *(uint64_t *)(base - OFF_EXIT_REASON) = exit_reason;

        uintptr_t return_addr  = *(uintptr_t *)(base - OFF_RETURN_ADDR);
        uintptr_t return_stack = *(uintptr_t *)(base - OFF_RETURN_STACK);
        gregs[REG_RIP] = (greg_t)return_addr;
        gregs[REG_RSP] = (greg_t)return_stack;
    }
}

void setup_signal_handler(void) {
    struct sigaction sa;
    memset(&sa, 0, sizeof(sa));
    sa.sa_sigaction = signal_handler;
    sa.sa_flags = SA_SIGINFO | SA_NODEFER | SA_ONSTACK;
    sigemptyset(&sa.sa_mask);

    sigaction(SIGSEGV, &sa, &old_sigsegv);
    sigaction(SIGBUS, &sa, &old_sigbus);
    sigaction(SIGFPE, &sa, &old_sigfpe);
}

void set_vmctx(void *gb) {
    guest_base_ptr = (uint8_t *)gb;
}

void set_fault_window(void *gb, void *code_start, void *code_end) {
    guest_base_ptr = (uint8_t *)gb;
    jit_code_start = (uintptr_t)code_start;
    jit_code_end   = (uintptr_t)code_end;
}

void clear_fault_window(void) {
    guest_base_ptr = NULL;
    jit_code_start = 0;
    jit_code_end   = 0;
}
