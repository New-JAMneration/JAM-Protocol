package work

// work item (14.3)
type WorkItem struct {
	Service            uint32   // service index
	CodeHash           [32]byte // code hash of the service
	Payload            []byte
	RefineGasLimit     uint64
	AccumulateGasLimit uint64
	ImportSegments     []ImportSegment
	Extrinsic          []ExtrinsicElement
	ExportCount        uint // number of export segments
}

type ImportSegment struct {
	Hash  [32]byte // hash of segment root or work package
	Index uint     // index of prior exported segments
}

type ExtrinsicElement struct {
	Hash   [32]byte
	Length uint
}
