package PolkaVM

type Registers [13]uint64
type ProgramCode []byte // program blob "p" in A.2
type Instruction []byte // instruction "c" in A.2
type Gas int64
type ProgramCounter uint32 // "ı" in GP

type MemoryChunk struct {
	Addres   uint32 `json:"address"`
	Contents []byte `json:"contents"`
}

type MemoryChunks []MemoryChunk

type PageMap struct {
	Address    uint32 `json:"address"`
	Length     uint32 `json:"length"`
	IsWritable bool   `json:"is-writable"`
}

type PageMaps []PageMap

type TestCase struct {
	Name                  string      `json:"name"`
	InitialRegisters      [13]uint64  `json:"initial-regs"`
	InitialProgramCounter uint32      `json:"initial-pc"`
	InitialPageMap        PageMap     `json:"initial-page-map"`
	InitialMemory         MemoryChunk `json:"initial-memory"`
	InitialGas            Gas         `json:"initial-gas"`
	ProgramBlob           uint8       `json:"program"`

	// expected-status -> panic, halt, page-fault
	ExpectedStatus string `json:"expected-status"`

	ExpectedRegisters        [13]uint64  `json:"expected-regs"`
	ExpectedProgramCounter   uint32      `json:"expected-pc"`
	ExpectedMemory           MemoryChunk `json:"expected-memory"`
	ExpectedGas              int64       `json:"expected-gas"`
	ExpectedPageFaultAddress uint32      `json:"expected-page-fault-address,omitempty"`
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
