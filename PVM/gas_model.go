package PVM

// A.9 ROB gas model (GP v0.8.0 sec:gascostmodel).

const (
	MaxROB = 32

	DecodeWidth    = 4 // d.^(0)
	ExecutionWidth = 5 // e.^(0)
)

type ExecUnits struct {
	A, L, S, M, D int
}

var InitialUnits = ExecUnits{A: 4, L: 4, S: 4, M: 1, D: 1} // x.^(0)

type Reg uint8

// BlockState — x (gas-sim state).
type BlockState struct {
	Iota           ProgramCounter
	Cyc            Gas
	DecodeSlots    int
	ExecutionSlots int
	UnitsAvail     ExecUnits
	ROB            []robEntry
}

// InitBlockState returns Ξ₀(ι) — initial gas simulation state (A.9).
func InitBlockState(startIota ProgramCounter) *BlockState {
	return &BlockState{
		Iota:           startIota,
		Cyc:            0,
		DecodeSlots:    DecodeWidth,
		ExecutionSlots: ExecutionWidth,
		UnitsAvail:     InitialUnits,
		ROB:            make([]robEntry, 0, MaxROB),
	}
}
