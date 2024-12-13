package work

import "fmt"

// work package (14.2)
type WorkPackage struct {
	Authorization []byte        `json:"authorization"`  // authorization token
	AuthCodeHost  uint32        `json:"auth_code_host"` // host service index
	Authorizer    Authorizer    `json:"authorizer"`
	Context       RefineContext `json:"context"`
	WorkItems     []WorkItem    `json:"items"`
}

type Authorizer struct {
	CodeHash [32]byte `json:"code_hash"` // authorization code hash
	Params   []byte   `json:"params"`    // parameterization blob
}

const (
	MaxTotalSize     = 12 * 1024 * 1024 // 12 MB (14.6)
	MaxRefineGas     = 500000000
	MaxAccumulateGas = 100000
	MaxSegments      = 2048 // import/export segment total limit 2^11 (14.4)
	SegmentSize      = 4104 // size of segment
)

func (wp *WorkPackage) Validate() error {
	totalSize := len(wp.Authorization) + len(wp.Authorizer.Params)
	totalImportSegments := 0
	totalExportSegments := 0

	for _, item := range wp.WorkItems {
		totalSize += len(item.Payload)

		totalImportSegments += len(item.ImportSegments)
		totalSize += len(item.ImportSegments) * SegmentSize

		for _, extrinsic := range item.Extrinsic {
			totalSize += int(extrinsic.Length)
		}

		totalExportSegments += int(item.ExportCount)
	}

	// total size check (14.5)
	if totalSize > MaxTotalSize {
		return fmt.Errorf("total size exceeds %d bytes", MaxTotalSize)
	}

	// import/export segment count check （14.4)
	if totalImportSegments+totalExportSegments > MaxSegments {
		return fmt.Errorf("total import and export segments exceed %d", MaxSegments)
	}

	// gas limit check (14.7)
	var totalRefineGas, totalAccumulateGas uint64
	for _, item := range wp.WorkItems {
		totalRefineGas += item.RefineGasLimit
		totalAccumulateGas += item.AccumulateGasLimit
	}

	if totalRefineGas > MaxRefineGas {
		return fmt.Errorf("refine gas limit exceeds %d", MaxRefineGas)
	}
	if totalAccumulateGas > MaxAccumulateGas {
		return fmt.Errorf("accumulate gas limit exceeds %d", MaxAccumulateGas)
	}

	return nil
}
