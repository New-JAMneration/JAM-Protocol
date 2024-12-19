package work

// work item (14.3)
type WorkItem struct {
	Service            uint32             `json:"service"`   // service index
	CodeHash           [32]byte           `json:"code_hash"` // code hash of the service
	Payload            []byte             `json:"payload"`
	RefineGasLimit     uint64             `json:"refine_gas_limit"`
	AccumulateGasLimit uint64             `json:"accumulate_gas_limit"`
	ImportSegments     []ImportSegment    `json:"import_segments"`
	Extrinsic          []ExtrinsicElement `json:"extrinsic"`
	ExportCount        uint16             `json:"export_count"` // number of export segments
}

type ImportSegment struct {
	Hash  [32]byte `json:"tree_root"` // hash of segment root or work package
	Index uint16   `json:"index"`     // index of prior exported segments
}

type ExtrinsicElement struct {
	Hash   [32]byte `json:"hash"`
	Length uint32   `json:"len"`
}
