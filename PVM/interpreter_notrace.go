//go:build !pvmtrace

package PVM

// Interpreter without trace-only fields (!trace build = zero struct bloat on hot path).
type Interpreter struct {
	Program    *Program
	Registers  Registers
	Memory     *Memory
	Gas        Gas
	InstrCount uint64
	GasCharged bool // formula A.7 (0.8.0): true when current block's gas is pre-charged
}
