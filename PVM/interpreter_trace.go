//go:build trace

package PVM

import (
	"github.com/New-JAMneration/JAM-Protocol/PVM/PVMtrace"
)

type Interpreter struct {
	Program    *Program
	Registers  Registers
	Memory     *Memory
	Gas        Gas
	InstrCount uint64
	LastLoad   struct {
		Addr   uint32
		Val    uint64
		Active bool
	}
	LastStore struct {
		Addr   uint32
		Val    uint64
		Active bool
	}
	Trace *PVMtrace.Trace
}
