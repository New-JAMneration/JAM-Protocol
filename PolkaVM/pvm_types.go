package PolkaVM

type Registers [13]uint64
type ProgramCode uint8
type Gas int64
type ProgramCounter uint32

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
