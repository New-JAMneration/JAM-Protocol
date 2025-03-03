package PolkaVM

import "github.com/New-JAMneration/JAM-Protocol/internal/types"

type Psi_H_ReturnType struct {
	Pagefault        bool            // exit reason is page fault (for handling multiple return type)
	ExitReason       ExitReasonTypes // exit reason
	Counter          uint32          // new instruction counter
	Gas              uint64          // gas remain
	Reg              Registers       // new registers
	Ram              PageMap         // new memory
	Addition         any             // addition host-call context
	PagefaultAddress uint64          // page fault address, only use for page fault
}

// OperationType Enum
type OperationType int

const (
	GasOp              OperationType = iota // gas = 0
	LookupOp                                // lookup = 1
	ReadOp                                  // read = 2
	WriteOp                                 // write = 3
	InfoOp                                  // info = 4
	BlessOp                                 // bless = 5
	AssignOp                                // assign = 6
	DesignateOp                             // designate = 7
	CheckpointOp                            // checkpoint = 8
	NewOp                                   // new = 9
	UpgradeOp                               // upgrade = 10
	TransferOp                              // transfer = 11
	EjectOp                                 // eject = 12
	QueryOp                                 // query = 13
	SolicitOp                               // solicit = 14
	ForgetOp                                // forget = 15
	YieldOp                                 // yield = 16
	HistoricalLookupOp                      // historical_lookup = 17
	FetchOp                                 // fetch = 18
	ExportOp                                // export = 19
	MachineOp                               // machine = 20
	PeekOp                                  // peek = 21
	PokeOp                                  // poke = 22
	ZeroOp                                  // zero = 23
	VoidOp                                  // void = 24
	InvokeOp                                // invoke = 25
	ExpungeOp                               // expunge = 26
)

type HistoryState struct {
	PreviousGas       Gas
	PreviousRegisters Registers
	PreviousMemory    PageMap
	PreviousAccount   types.ServiceAccountState
}

type OmegaOutput struct {
	ExitReason      error                     // Exit reason
	NewGas          Gas                       // New Gas
	NewRegisters    Registers                 // New Register
	NewMemory       PageMap                   // New Memory
	NewAccountState types.ServiceAccountState // New State
	Counter         uint32                    // For calculate run count
	History         *HistoryState             // For roll back
	Addition        []any                     // addition host-call context
}

type GeneralFunction func(OmegaInput) OmegaOutput

type OmegaInput struct {
	Instruction  uint64                    // opcode
	Operation    OperationType             // operation type
	Gas          Gas                       // gas counter
	Registers    Registers                 // PVM registers
	Memory       PageMap                   // memory
	AccountState types.ServiceAccountState // State
	History      *HistoryState             // For storing history state
	Args         []any                     // Extra parameter for each host-call function
}
