package work_package

import (
	"encoding/hex"
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/service_account"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
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
	*WorkPackageController
	// 加入 CE133 需要的額外參數，例如從 builder 接收的 extrinsic raw bytes
	pa                types.OpaqueHash
	pm                types.ByteSequence
	pc                types.ByteSequence
	Extrinsics        types.ByteSequence
	ErasureStore      *store.SegmentErasureMap
	SegmentRootLookup *store.HashSegmentMap
}

func NewInitialPackageProcessor(wp *types.WorkPackage, extrinsics []byte) *InitialPackageProcessor {
	return &InitialPackageProcessor{
		WorkPackageController: NewWorkPackageController(wp),
		Extrinsics:            extrinsics,
	}
}

func (p *InitialPackageProcessor) Process() error {
	pa, pm, pc, err := VerifyAuthorization(p.WorkPackage)
	if err != nil {
		return err
	}
	p.pa = pa
	p.pm = pm
	p.pc = pc

	specs := FlattenExtrinsicSpecs(p.WorkPackage)
	extrinsicMap, err := ExtractExtrinsics(p.Extrinsics, specs)
	if err != nil {
		return err
	}

	if err := p.fetchImportSegments(); err != nil {
		return err
	}

	if err := p.refine(); err != nil {
		return err
	}

	if err := p.updateSegmentRootDict(); err != nil {
		return err
	}

	if err := p.sendToOtherGuarantors(); err != nil {
		return err
	}

	return nil
}

func (p *InitialPackageProcessor) fetchImportSegments() (map[types.OpaqueHash]map[types.U16]types.ExportSegment, error) {
	lookupDict := p.SegmentRootLookup
	segmentErasureMap := p.ErasureStore
	result := make(map[types.OpaqueHash]map[types.U16]types.ExportSegment)
	for _, item := range p.WorkPackage.Items {
		for _, imp := range item.ImportSegments {
			segmentRoot := lookupDict.Lookup(imp.TreeRoot)

			erasureRoot, err := segmentErasureMap.Get(segmentRoot)
			if err != nil {
				return nil, fmt.Errorf("failed to get erasure root for %s: %w", hex.EncodeToString(segmentRoot[:]), err)
			}

			data, err := FetchFromDA(erasureRoot, imp.Index)
			if err != nil {
				return nil, fmt.Errorf("failed to fetch segment from DA: %w", err)
			}

			if result[segmentRoot] == nil {
				result[segmentRoot] = make(map[types.U16]types.ExportSegment)
			}
			result[segmentRoot][imp.Index] = data

			if err := p.StoreSegment(segmentRoot, imp.Index, data); err != nil {
				return nil, fmt.Errorf("failed to store segment: %w", err)
			}
		}
	}
	return result, nil
}

// 14.9
func VerifyAuthorization(wp *types.WorkPackage) (types.OpaqueHash, types.ByteSequence, types.ByteSequence, error) {
	pa := hash.Blake2bHash(append(wp.Authorizer.CodeHash[:], wp.Authorizer.Params[:]...))
	delta := store.GetInstance().GetPriorStates().GetDelta()
	pm, pc, err := service_account.HistoricalLookupFunction(delta[wp.AuthCodeHost], wp.Context.LookupAnchorSlot, wp.Authorizer.CodeHash)
	if err != nil {
		return types.OpaqueHash{}, types.ByteSequence{}, types.ByteSequence{}, err
	}

	return pa, pm, pc, err
}

func FlattenExtrinsicSpecs(wp *types.WorkPackage) []types.ExtrinsicSpec {
	var allSpecs []types.ExtrinsicSpec
	for _, wi := range wp.Items {
		allSpecs = append(allSpecs, wi.Extrinsic...)
	}
	return allSpecs
}

func ExtractExtrinsics(data types.ByteSequence, specs []types.ExtrinsicSpec) (ExtrinsicLookup, error) {
	var result ExtrinsicLookup
	curr := 0

	for _, spec := range specs {
		length := int(spec.Len)
		if curr+length > len(data) {
			return nil, fmt.Errorf("extrinsic length overflow")
		}

		extracted := data[curr : curr+length]
		if hash.Blake2bHash(extracted) != spec.Hash {
			return nil, fmt.Errorf("extrinsic hash mismatch")
		}

		result[spec.Hash] = append([]byte(nil), extracted...)
		curr += length
	}

	if curr != len(data) {
		return nil, fmt.Errorf("data remains after extrinsics parsed")
	}

	return result, nil
}

type ExtrinsicLookup map[types.OpaqueHash]types.ByteSequence

// ParseExtrinsic takes a raw byte stream and the wx list and returns hash → preimage mapping
func ParseExtrinsic(raw []byte, extrinsicSpecs []types.ExtrinsicSpec) ExtrinsicLookup {
	result := make(map[types.OpaqueHash][]byte)
	offset := 0
	for _, spec := range extrinsicSpecs {
		length := int(spec.Length)
		data := raw[offset : offset+length]
		hash := types.Blake2bHash(data)
		result[hash] = data
		offset += length
	}
	return result
}
