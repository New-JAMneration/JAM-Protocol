package work_package

import (
	"github.com/New-JAMneration/JAM-Protocol/PVM"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

type PVMExecutor interface {
	Psi_I(p types.WorkPackage, c types.CoreIndex, code types.ByteSequence) PVM.Psi_I_ReturnType
	RefineInvoke(input PVM.RefineInput) PVM.RefineOutput
}

type RealPVMExecutor struct{}

func (e *RealPVMExecutor) Psi_I(p types.WorkPackage, c types.CoreIndex, code types.ByteSequence) PVM.Psi_I_ReturnType {
	return PVM.Psi_I(p, c, code)
}

func (e *RealPVMExecutor) RefineInvoke(input PVM.RefineInput) PVM.RefineOutput {
	return PVM.RefineInvoke(input)
}

type DASegmentFetcher interface {
	Fetch(erasureRoot types.OpaqueHash, index types.U16) (types.ExportSegment, []types.OpaqueHash, error)
}

type WorkPackageController struct {
	CoreIndex types.CoreIndex
	PVM       PVMExecutor
	Fetcher   DASegmentFetcher

	// different guarantors behavior
	Extrinsics  []byte             // Initial
	WorkPackage *types.WorkPackage // Initial
	Bundle      []byte             // Shared
}

func NewInitialController(wp *types.WorkPackage, extrinsics []byte, coreIndex types.CoreIndex, fetcher DASegmentFetcher) *WorkPackageController {
	return &WorkPackageController{
		WorkPackage: wp,
		CoreIndex:   coreIndex,
		Extrinsics:  extrinsics,
		PVM:         &RealPVMExecutor{},
		Fetcher:     fetcher,
	}
}

func NewSharedController(bundle []byte, coreIndex types.CoreIndex) *WorkPackageController {
	return &WorkPackageController{
		CoreIndex: coreIndex,
		Bundle:    bundle,
		PVM:       &RealPVMExecutor{},
	}
}

func (p *WorkPackageController) Process() (types.WorkReport, error) {
	workPackage, extrinsicMap, importSegments, workPackageBundle, workPackageHash, err := p.prepareInputs()
	if err != nil {
		return types.WorkReport{}, err
	}

	if err := workPackage.Validate(); err != nil {
		return types.WorkReport{}, err
	}
	delta := store.GetInstance().GetPriorStates().GetDelta()
	pa, _, pc, err := VerifyAuthorization(&workPackage, delta)
	if err != nil {
		return types.WorkReport{}, err
	}

	// if p.Bundle == nil => initial controller
	// CE134: send core index, segment lookup dict, wp bundle to other guarantors
	// and wait for them to send back work report hash and ed25519 signature
	// two goroutines to two different guarantors

	report, err := WorkReportCompute(&workPackage, p.CoreIndex, pa, pc, extrinsicMap, importSegments, delta, workPackageBundle, workPackageHash, p.PVM)
	if err != nil {
		return types.WorkReport{}, err
	}
	redisBackend, err := store.GetRedisBackend()
	if err != nil {
		return types.WorkReport{}, err
	}
	// (14.13)
	newDict, err := redisBackend.SetHashSegmentMapWithLimit(workPackageHash, types.OpaqueHash(report.PackageSpec.ExportsRoot))
	if err != nil {
		return types.WorkReport{}, err
	}
	lookup := convertMapToLookup(newDict)
	report.SegmentRootLookup = lookup

	// check if work report is same between all the guarantors

	return report, nil
}

func (p *WorkPackageController) prepareInputs() (types.WorkPackage, PVM.ExtrinsicDataMap, types.ExportSegmentMatrix, []byte, types.OpaqueHash, error) {
	redisBackend, err := store.GetRedisBackend()
	if err != nil {
		return types.WorkPackage{}, nil, nil, nil, types.OpaqueHash{}, err
	}

	dict, err := redisBackend.GetHashSegmentMap()
	if err != nil {
		return types.WorkPackage{}, nil, nil, nil, types.OpaqueHash{}, err
	}
	if p.Bundle == nil {
		// initial controller
		specs := FlattenExtrinsicSpecs(p.WorkPackage)
		extrinsicMap, err := ExtractExtrinsics(p.Extrinsics, specs)
		if err != nil {
			return types.WorkPackage{}, nil, nil, nil, types.OpaqueHash{}, err
		}
		importSegments, importProofs, err := p.fetchImportSegments(dict)
		if err != nil {
			return types.WorkPackage{}, nil, nil, nil, types.OpaqueHash{}, err
		}
		// build work package bundle
		workPackgeBundle, err := buildWorkPackageBundle(p.WorkPackage, extrinsicMap, importSegments, importProofs)
		if err != nil {
			return types.WorkPackage{}, nil, nil, nil, types.OpaqueHash{}, err
		}

		// Get work package hash
		encoder := types.NewEncoder()
		encoder.SetHashSegmentMap(dict)
		encodedWorkPackage, err := encoder.Encode(p.WorkPackage)
		if err != nil {
			return types.WorkPackage{}, nil, nil, nil, types.OpaqueHash{}, err
		}
		workPackageHash := hash.Blake2bHash(encodedWorkPackage)

		return *p.WorkPackage, extrinsicMap, importSegments, workPackgeBundle, workPackageHash, nil
	}
	// shared controller
	// Decode to get work package bundle
	var bundle types.WorkPackageBundle
	decoder := types.NewDecoder()
	decoder.SetHashSegmentMap(dict)
	err = decoder.Decode(p.Bundle, &bundle)
	if err != nil {
		return types.WorkPackage{}, nil, nil, nil, types.OpaqueHash{}, err
	}

	extrinsicMap, err := extractExtrinsicMapFromBundle(&bundle.Package, bundle.Extrinsics)
	if err != nil {
		return types.WorkPackage{}, nil, nil, nil, types.OpaqueHash{}, err
	}
	encoder := types.NewEncoder()
	encoder.SetHashSegmentMap(dict)
	encodedWorkPackage, err := encoder.Encode(&bundle.Package)
	if err != nil {
		return types.WorkPackage{}, nil, nil, nil, types.OpaqueHash{}, err
	}
	workPackageHash := hash.Blake2bHash(encodedWorkPackage)
	// TODO: did not use bundle.ImportProofs now
	return bundle.Package, extrinsicMap, bundle.ImportSegments, p.Bundle, workPackageHash, nil
}

// get (14.14) S & J from DA
func (p *WorkPackageController) fetchImportSegments(lookupDict map[types.OpaqueHash]types.OpaqueHash) (types.ExportSegmentMatrix, types.OpaqueHashMatrix, error) {
	var segments types.ExportSegmentMatrix
	var proofs types.OpaqueHashMatrix
	redisBackend, err := store.GetRedisBackend()
	if err != nil {
		return nil, nil, err
	}
	for _, item := range p.WorkPackage.Items {
		var rowSegments []types.ExportSegment
		var rowProofs []types.OpaqueHash

		for _, spec := range item.ImportSegments {
			segmentRoot := spec.TreeRoot
			// L (14.12)
			if mapped, ok := lookupDict[spec.TreeRoot]; ok {
				segmentRoot = mapped
			}
			erasureRoot, err := redisBackend.GetSegmentErasureMap(segmentRoot)
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
	// need to fetch and reconstruct from DA, including justification check (CE139/140)
	return []types.ExportSegment{}, []types.OpaqueHash{}, nil
}
