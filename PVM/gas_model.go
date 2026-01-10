package PVM

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
type RegSlices [][]Reg

type BlockState struct {
	// ι^(n): next opcode index
	Iota ProgramCounter

	// c.^(n): cycle counter
	Cyc Gas

	// n.^(n): ROB index count
	Next int

	// d.^(n): remaining decode slots in current cycle
	DecodeSlots int

	// e.^(n): remaining execution slots in current cycle
	ExecutionSlots int

	// ROB vectors
	S []uint8     // s->: instruction state: 1. decoded, 2. pending, 3. executing, 4. finished
	C []int       // c->: remaining exec cycles
	P RegSlices   // p->: pending dependencies
	R RegSlices   // r->: destination registers
	X []ExecUnits // x->: units required to start

	// x.^(n): available execution units in current cycle
	UnitsAvail ExecUnits

	// step index n
	Step int
}

// A.51: Ξ₀(ι)
func InitBlockState(startIota ProgramCounter) *BlockState {
	return &BlockState{
		Iota:           startIota,      // ι^(0)=ι
		Cyc:            0,              // c.^(0)=0
		Next:           0,              // n.^(0)=0
		DecodeSlots:    DecodeWidth,    // d.^(0)=4
		ExecutionSlots: ExecutionWidth, // e.^(0)=5

		S: make([]uint8, 0, MaxROB),
		C: make([]int, 0, MaxROB),
		P: make(RegSlices, 0, MaxROB),
		R: make(RegSlices, 0, MaxROB),
		X: make([]ExecUnits, 0, MaxROB),

		UnitsAvail: InitialUnits, // x.^(0)
		Step:       0,
	}
}

func (r *RegSlices) remove1D(j int) {
	if r == nil || j < 0 || j >= len(*r) {
		return
	}

	s := *r
	copy(s[j:], s[j+1:])
	s[len(s)-1] = nil
	*r = s[:len(s)-1]
}

func (r *RegSlices) remove2D(j, k int) {
	if r == nil || j < 0 || j >= len(*r) {
		return
	}

	inner := (*r)[j]
	if k < 0 || k >= len(inner) {
		return
	}

	copy(inner[k:], inner[k+1:])
	inner[len(inner)-1] = 0
	(*r)[j] = inner[:len(inner)-1]
}
