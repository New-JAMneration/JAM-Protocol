package work_package

import (
	"encoding/hex"
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/PolkaVM"
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

	importSegments, importProofs, err := p.fetchImportSegments()
	if err != nil {
		return err
	}

	// build work package bundle
	workPackgeBundle, err := buildWorkPackageBundle(p.WorkPackage, extrinsicMap, importSegments, importProofs)
	if err != nil {
		return err
	}

	// CE134: send core index, segment lookup dict, wp bundle to other guarantors
	// and wait for them to send back work report hash and ed25519 signature
	// two goroutines to two different guarantors

	var lookup types.SegmentRootLookup
	_, err = workResultCompute(*p.WorkPackage, p.CoreIndex, pa, pc, extrinsicMap, importSegments, delta, lookup, workPackgeBundle)
	if err != nil {
		return err
	}

	// check if work report is same between all the guarantors

	return nil
}

func buildWorkPackageBundle(
	wp *types.WorkPackage,
	extrinsicMap PolkaVM.ExtrinsicDataMap,
	importSegments [][]types.ExportSegment,
	importProofs [][]types.OpaqueHash,
) ([]byte, error) {
	var extrinsics []types.ExtrinsicData
	for _, item := range wp.Items {
		for _, extrinsic := range item.Extrinsic {
			extrinsics = append(extrinsics, types.ExtrinsicData(extrinsicMap[extrinsic.Hash]))
		}
	}

	output := []byte{}
	encoder := types.NewEncoder()
	encoded, err := encoder.Encode(&wp)
	if err != nil {
		return nil, fmt.Errorf("failed to encode work package: %w", err)
	}
	output = append(output, encoded...)
	encoded, err = encoder.Encode(&extrinsics)
	if err != nil {
		return nil, fmt.Errorf("failed to encode extrinsics: %w", err)
	}
	output = append(output, encoded...)
	encoded, err = encoder.Encode(&importSegments)
	if err != nil {
		return nil, fmt.Errorf("failed to encode import segments: %w", err)
	}
	output = append(output, encoded...)
	encoded, err = encoder.Encode(&importProofs)
	if err != nil {
		return nil, fmt.Errorf("failed to encode import proofs: %w", err)
	}
	output = append(output, encoded...)

	return output, nil
}

func (p *InitialPackageProcessor) fetchImportSegments() ([][]types.ExportSegment, [][]types.OpaqueHash, error) {
	lookupDict := p.SegmentRootLookup
	segmentErasureMap := p.ErasureMap
	result := make([][]types.ExportSegment, len(p.WorkPackage.Items))
	proofs := make([][]types.OpaqueHash, len(p.WorkPackage.Items))
	for itemIndex, item := range p.WorkPackage.Items {
		var segs []types.ExportSegment
		var justifications []types.OpaqueHash

		for _, imp := range item.ImportSegments {
			segmentRoot := lookupDict.Lookup(imp.TreeRoot)
			erasureRoot, err := segmentErasureMap.Get(segmentRoot)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to get erasure root for %s: %w", hex.EncodeToString(segmentRoot[:]), err)
			}

			data, proof, err := fetchFromDA(erasureRoot, imp.Index)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to fetch data from DA for %s: %w", hex.EncodeToString(erasureRoot[:]), err)
			}
			segs = append(segs, data...)
			justifications = append(justifications, proof...)
		}

		result[itemIndex] = segs
		proofs[itemIndex] = justifications
	}

	return result, proofs, nil
}

func fetchFromDA(erasureRoot types.OpaqueHash, index types.U16) ([]types.ExportSegment, []types.OpaqueHash, error) {
	// need to fetch and reconstruct from DA
	return []types.ExportSegment{}, []types.OpaqueHash{}, nil
}
