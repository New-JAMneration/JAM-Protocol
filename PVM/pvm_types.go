package PVM

type (
	Registers      [13]uint64
	ProgramCode    []byte // program blob "p" in A.2
	Instruction    []byte // instruction "c" in A.2
	Gas            int64
	ProgramCounter uint32 // "ı" in GP
)

/*
registers name and index
RA = 0, SP = 1, T0 = 2, T1 = 3, T2 = 4, S0 = 5, S1 = 6,
A0 = 7, A1 = 8, A2 = 9, A3 = 10, A4 = 11, A5 = 12
*/

var RegName = [13]string{
	"ra", "sp", "t0", "t1", "t2", "s0", "s1", "a0", "a1", "a2", "a3", "a4", "a5",
}

type MemoryChunk struct {
	Address  uint32 `json:"address"`
	Contents []byte `json:"contents"`
}

type MemoryChunks []MemoryChunk

type PageMap struct {
	Address    uint32 `json:"address"`
	Length     uint32 `json:"length"`
	IsWritable bool   `json:"is-writable"`
}

type PageMaps []PageMap

type InstructionTestCase struct {
	Name                  string         `json:"name"`
	InitialRegisters      Registers      `json:"initial-regs"`
	InitialProgramCounter ProgramCounter `json:"initial-pc"`
	InitialPageMap        PageMaps       `json:"initial-page-map"`
	InitialMemory         MemoryChunks   `json:"initial-memory"`
	InitialGas            Gas            `json:"initial-gas"`
	ProgramBlob           []byte         `json:"program"`

	// expected-status -> panic, halt, page-fault
	ExpectedStatus           string         `json:"expected-status"`
	ExpectedRegisters        Registers      `json:"expected-regs"`
	ExpectedProgramCounter   ProgramCounter `json:"expected-pc"`
	ExpectedMemory           MemoryChunks   `json:"expected-memory"`
	ExpectedGas              Gas            `json:"expected-gas"`
	ExpectedPageFaultAddress uint32         `json:"expected-page-fault-address,omitempty"`
}

// HostCallResultConstants
const (
	OK   uint64 = 0
	HUH  uint64 = ^uint64(8)
	LOW  uint64 = ^uint64(7)
	CASH uint64 = ^uint64(6)
	CORE uint64 = ^uint64(5)
	FULL uint64 = ^uint64(4)
	WHO  uint64 = ^uint64(3)
	OOB  uint64 = ^uint64(2)
	WHAT uint64 = ^uint64(1)
	NONE uint64 = ^uint64(0) // 2^64 - 1
)

// Inner PVM invocations called by Omega_k(invoke)
const (
	INNERHALT uint64 = iota
	INNERPANIC
	INNERFAULT
	INNERHOST
	INNEROOG
)
