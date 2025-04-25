package work_package

import (
	"encoding/hex"
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type WorkPackageProcessor interface {
	Process() error
}

type WorkPackageController struct {
	WorkPackage *types.WorkPackage
}

// NewWorkPackageController creates a new WorkPackageController
func NewWorkPackageController(workPackage *types.WorkPackage) *WorkPackageController {
	return &WorkPackageController{
		WorkPackage: workPackage,
	}
}

type InitialPackageProcessor struct {
	WorkPackageController
	CoreIndex         types.CoreIndex
	Extrinsics        types.ByteSequence
	ErasureMap        *store.SegmentErasureMap
	SegmentRootLookup *store.HashSegmentMap
}

func NewInitialPackageProcessor(wp *types.WorkPackage, extrinsics []byte) *InitialPackageProcessor {
	return &InitialPackageProcessor{
		WorkPackageController: *NewWorkPackageController(wp),
		Extrinsics:            extrinsics,
	}
}

func (p *InitialPackageProcessor) Process() error {
	if err := p.WorkPackage.Validate(); err != nil {
		return err
	}
	delta := store.GetInstance().GetPriorStates().GetDelta()
	pa, _, pc, err := VerifyAuthorization(p.WorkPackage, delta)
	if err != nil {
		return err
	}

	specs := FlattenExtrinsicSpecs(p.WorkPackage)
	extrinsicMap, err := ExtractExtrinsics(p.Extrinsics, specs)
	if err != nil {
		return err
	}

	importSegments, err := p.fetchImportSegments()
	if err != nil {
		return err
	}

	// CE134: send core index, segment lookup dict, wp bundle to other guarantors
	// and wait for them to send back work report hash and ed25519 signature
	// two goroutines to two different guarantors

	var lookup types.SegmentRootLookup
	_, err = workResultCompute(*p.WorkPackage, p.CoreIndex, pa, pc, extrinsicMap, importSegments, delta, lookup)
	if err != nil {
		return err
	}

	// check if work report is same between all the guarantors

	// if err := p.updateSegmentRootDict(); err != nil {
	// 	return err
	// }

	return nil
}

func (p *InitialPackageProcessor) fetchImportSegments() ([][]types.ExportSegment, error) {

	lookupDict := p.SegmentRootLookup
	segmentErasureMap := p.ErasureMap
	result := make([][]types.ExportSegment, len(p.WorkPackage.Items))
	for _, item := range p.WorkPackage.Items {
		for _, imp := range item.ImportSegments {
			segmentRoot := lookupDict.Lookup(imp.TreeRoot)
			erasureRoot, err := segmentErasureMap.Get(segmentRoot)
			if err != nil {
				return nil, fmt.Errorf("failed to get erasure root for %s: %w", hex.EncodeToString(segmentRoot[:]), err)
			}

			data, err := fetchFromDA(erasureRoot, imp.Index)
			if err != nil {
				return nil, fmt.Errorf("failed to fetch data from DA for %s: %w", hex.EncodeToString(erasureRoot[:]), err)
			}
			result = append(result, data)
		}
	}
	return result, nil
}
