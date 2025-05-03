package work_package

import (
	"github.com/New-JAMneration/JAM-Protocol/PolkaVM"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

type PVMExecutor interface {
	Psi_I(p types.WorkPackage, c types.CoreIndex, code types.ByteSequence) PolkaVM.Psi_I_ReturnType
	RefineInvoke(input PolkaVM.RefineInput) PolkaVM.RefineOutput
}

type RealPVMExecutor struct{}

func (e *RealPVMExecutor) Psi_I(p types.WorkPackage, c types.CoreIndex, code types.ByteSequence) PolkaVM.Psi_I_ReturnType {
	return PolkaVM.Psi_I(p, c, code)
}
func (e *RealPVMExecutor) RefineInvoke(input PolkaVM.RefineInput) PolkaVM.RefineOutput {
	return PolkaVM.RefineInvoke(input)
}

type DASegmentFetcher interface {
	Fetch(erasureRoot types.OpaqueHash, index types.U16) (types.ExportSegment, []types.OpaqueHash, error)
}

type WorkPackageController struct {
	WorkPackage       *types.WorkPackage
	CoreIndex         types.CoreIndex
	ErasureMap        *store.SegmentErasureMap
	SegmentRootLookup *store.HashSegmentMap
	PVM               PVMExecutor
	Fetcher           DASegmentFetcher

	// different guarantors behavior
	Extrinsics []byte                   // Initial
	Bundle     *types.WorkPackageBundle // Shared
}

func NewInitialController(wp *types.WorkPackage, extrinsics []byte, erasureMap *store.SegmentErasureMap, segmentRootLookup *store.HashSegmentMap, coreIndex types.CoreIndex, fetcher DASegmentFetcher) *WorkPackageController {
	return &WorkPackageController{
		WorkPackage:       wp,
		CoreIndex:         coreIndex,
		ErasureMap:        erasureMap,
		SegmentRootLookup: segmentRootLookup,
		Extrinsics:        extrinsics,
		PVM:               &RealPVMExecutor{},
		Fetcher:           fetcher,
	}
}

func (p *WorkPackageController) Process() (types.WorkReport, error) {
	if err := p.WorkPackage.Validate(); err != nil {
		return types.WorkReport{}, err
	}
	delta := store.GetInstance().GetPriorStates().GetDelta()
	pa, _, pc, err := VerifyAuthorization(p.WorkPackage, delta)
	if err != nil {
		return types.WorkReport{}, err
	}

	specs := FlattenExtrinsicSpecs(p.WorkPackage)
	extrinsicMap, err := ExtractExtrinsics(p.Extrinsics, specs)
	if err != nil {
		return types.WorkReport{}, err
	}

	dict, err := p.SegmentRootLookup.LoadDict()
	if err != nil {
		return types.WorkReport{}, err
	}
	importSegments, importProofs, err := p.fetchImportSegments(dict)
	if err != nil {
		return types.WorkReport{}, err
	}
	// build work package bundle
	workPackgeBundle, err := buildWorkPackageBundle(p.WorkPackage, extrinsicMap, importSegments, importProofs)
	if err != nil {
		return types.WorkReport{}, err
	}

	// Get work package hash
	encoder := types.NewEncoder()
	encodedWorkPackage, err := encoder.Encode(p.WorkPackage)
	if err != nil {
		return types.WorkReport{}, err
	}
	workPackageHash := hash.Blake2bHash(encodedWorkPackage)

	// CE134: send core index, segment lookup dict, wp bundle to other guarantors
	// and wait for them to send back work report hash and ed25519 signature
	// two goroutines to two different guarantors
	report, err := WorkReportCompute(p.WorkPackage, p.CoreIndex, pa, pc, extrinsicMap, importSegments, delta, workPackgeBundle, workPackageHash, p.PVM)
	if err != nil {
		return types.WorkReport{}, err
	}
	newDict, err := p.SegmentRootLookup.SaveWithLimit(workPackageHash, types.OpaqueHash(report.PackageSpec.ExportsRoot))
	if err != nil {
		return types.WorkReport{}, err
	}
	lookup := convertMapToLookup(newDict)
	report.SegmentRootLookup = lookup

	// check if work report is same between all the guarantors

	return report, nil
}

func (p *WorkPackageController) fetchImportSegments(lookupDict map[types.OpaqueHash]types.OpaqueHash) (types.ExportSegmentMatrix, types.OpaqueHashMatrix, error) {
	var segments types.ExportSegmentMatrix
	var proofs types.OpaqueHashMatrix

	for _, item := range p.WorkPackage.Items {
		var rowSegments []types.ExportSegment
		var rowProofs []types.OpaqueHash

		for _, spec := range item.ImportSegments {
			segmentRoot := spec.TreeRoot
			// L (14.12)
			if mapped, ok := lookupDict[spec.TreeRoot]; ok {
				segmentRoot = mapped
			}
			erasureRoot, err := p.ErasureMap.Get(segmentRoot)
			if err != nil {
				return nil, nil, err
			}
			segment, proof, err := p.Fetcher.Fetch(erasureRoot, spec.Index)
			if err != nil {
				return nil, nil, err
			}
			rowSegments = append(rowSegments, segment)
			rowProofs = append(rowProofs, proof...)
		}

		segments = append(segments, rowSegments)
		proofs = append(proofs, rowProofs)
	}
	return segments, proofs, nil
}

// TODO: Change this when implementing CE
type FakeFetcher struct{}

func (f *FakeFetcher) Fetch(erasureRoot types.OpaqueHash, index types.U16) ([]types.ExportSegment, []types.OpaqueHash, error) {
	// need to fetch and reconstruct from DA
	return []types.ExportSegment{}, []types.OpaqueHash{}, nil
}
