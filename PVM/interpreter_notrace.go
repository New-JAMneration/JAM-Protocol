//go:build !pvmtrace

package PVM

// Interpreter without trace-only fields (!trace build = zero struct bloat on hot path).
type Interpreter struct {
	Program    *Program
	Registers  Registers
	Memory     *Memory
	Gas        Gas
	InstrCount uint64
}
