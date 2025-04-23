package work_package

import (
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/PolkaVM"
	"github.com/New-JAMneration/JAM-Protocol/internal/service_account"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/merkle_tree"
)

// (14.8)
func ItemToResult(item types.WorkItem, result types.WorkExecResult, gas types.Gas) types.WorkResult {
	payloadHash := hash.Blake2bHash(item.Payload)
	importCount := types.U16(len(item.ImportSegments))
	extrinsicSize := types.U32(len(item.Extrinsic))
	var zSum types.U16
	for _, v := range item.Extrinsic {
		zSum += types.U16(v.Len)
	}
	return types.WorkResult{
		ServiceId:     item.Service,
		CodeHash:      item.CodeHash,
		PayloadHash:   payloadHash,
		AccumulateGas: item.AccumulateGasLimit,
		Result:        result,
		RefineLoad: types.RefineLoad{
			GasUsed:        types.U64(gas),
			Imports:        importCount,
			ExtrinsicCount: item.ExportCount,
			ExtrinsicSize:  extrinsicSize,
			Exports:        zSum,
		},
	}
}

// (14.9)
func GetPaPmPc(accountState types.ServiceAccountState, workPackage types.WorkPackage) (types.OpaqueHash, types.ByteSequence, types.ByteSequence, error) {
	pa := hash.Blake2bHash(append(workPackage.Authorizer.CodeHash[:], workPackage.Authorizer.Params[:]...))
	pm, pc, err := service_account.HistoricalLookupFunction(accountState[workPackage.AuthCodeHost], workPackage.Context.LookupAnchorSlot, workPackage.Authorizer.CodeHash)
	if err != nil {
		return types.OpaqueHash{}, types.ByteSequence{}, types.ByteSequence{}, err
	}
	return pa, pm, pc, nil
}

// (14.10)
func PagedProofs(exportSegments []types.ExportSegment) ([]types.ExportSegment, error) {
	byteSequences := make([]types.ByteSequence, len(exportSegments))
	for i, segment := range exportSegments {
		byteSequences[i] = types.ByteSequence(segment[:])
	}
	result := []types.ExportSegment{}
	maxIndex := (len(exportSegments) + 63) / 64 // ceiling
	for i := 0; i < maxIndex; i++ {
		j6 := merkle_tree.Jx(6, byteSequences, types.U32(i), hash.Blake2bHash)
		l6 := merkle_tree.Lx(6, byteSequences, types.U32(i), hash.Blake2bHash)

		output := []byte{}
		encoder := types.NewEncoder()

		encoded, err := encoder.Encode(&types.SliceHash{
			A: j6,
			B: l6,
		})
		if err != nil {
			return nil, err
		}
		output = append(output, encoded...)
		padded := PadToMultiple(output, types.SegmentSize)
		result = append(result, types.ExportSegment(padded))
	}
	return exportSegments, nil
}

// (14.17)
func PadToMultiple(x []byte, n int) []byte {
	padLen := (n - ((len(x) + n - 1) % n)) % n // 1
	// padLen := (len(x) + n - 1) % n	//
	// sub := (len(seq) + n - 1) % n + 1	//
	padding := make([]byte, padLen)
	return append(x, padding...)
}

type SegmentRootLookupDict map[types.OpaqueHash]types.OpaqueHash
type ExportedSegmentMap map[types.OpaqueHash][]types.ExportSegment

// (14.12)
func LookupSegmentRoot(r types.OpaqueHash, lookupDict SegmentRootLookupDict) types.OpaqueHash {
	if val, ok := lookupDict[r]; ok {
		return val // r is a WP hash (with ⊞), mapped to segment root
	}
	return r // r is already a segment root
}

// (14.11)
func Xi(workPackage types.WorkPackage, coreIndex types.CoreIndex) (types.WorkReport, error) {
	//validate work package
	if err := workPackage.Validate(); err != nil {
		return types.WorkReport{}, err
	}
	// GetPaPmPc
	delta := store.GetInstance().GetPriorStates().GetDelta()
	pa, _, pc, err := GetPaPmPc(delta, workPackage)
	if err != nil {
		return types.WorkReport{}, err
	}

	// get lookup dict from store
	var lookupDict SegmentRootLookupDict

	e, o, g := PolkaVM.Psi_I(workPackage, coreIndex)
	if e != types.WorkExecResultOk {
		return types.WorkReport{}, fmt.Errorf("work item execution failed: %v", e)
	}

	var result [][]any
	for j, item := range workPackage.Items {
		r, u, e := I(workPackage, j, lookupDict, o)
		ItemToResult(item, r, u)
	}
	// r, eBar := TransTranspose()

	return types.WorkReport{
		// PackageSpec: s,
		Context:        workPackage.Context,
		CoreIndex:      coreIndex,
		AuthorizerHash: pa,
		// AuthOutput: o,
		// SegmentRootLookup: l,
		// Results: r,
		// AuthGasUsed: g,
	}, nil
}

// (14.12) Segment root lookup function L(r)
func L(r types.OpaqueHash, lookup SegmentRootLookupDict) types.OpaqueHash {
	if root, ok := lookup[r]; ok {
		return root // r ∈ H⊞ → l[h]
	}
	return r // r ∈ H → r
}

func I(workPackage types.WorkPackage, j int, lookupDict SegmentRootLookupDict, o types.ByteSequence) (types.WorkExecResult, types.Gas) {
	workItem := workPackage.Items[j]
	expectedCount := workItem.ExportCount
	var segmentRootToErasureRoot SegmentRootToErasureRoot
	imports := S(workItem, lookupDict, segmentRootToErasureRoot)
	lSum := types.U16(0)
	for k := 0; k < j; k++ {
		lSum += workPackage.Items[k].ExportCount
	}
	refineOuput := PolkaVM.Psi_R(types.U64(j), workPackage, o, imports, types.U64(lSum))
	r := refineOuput.RefineOutput
	e := refineOuput.EportSegment
	u := refineOuput.Gas

}

type SegmentRootToErasureRoot map[types.OpaqueHash]types.OpaqueHash

// (14.14) S
func S(w types.WorkItem, segmentRootLookupDict SegmentRootLookupDict, segmentRootToErasureRoot SegmentRootToErasureRoot) [][]types.ExportSegment {

	var result [][]types.ExportSegment

	for _, imp := range w.ImportSegments {
		segmentRoot := L(imp.TreeRoot, segmentRootLookupDict)

		erasureRoot, ok := segmentRootToErasureRoot[segmentRoot]
		if !ok {
			panic("erasure root not found for segment root")
		}

		segment, err := fetchFromDA(erasureRoot)
		if err != nil {
			panic("failed to fetch from DA: " + err.Error())
		}

		result = append(result, segment)
	}

	return result
}

func fetchFromDA(erasureRoot types.OpaqueHash) ([]types.ExportSegment, error) {
	// need to fetch and reconstruct from DA
	return []types.ExportSegment{}, nil
}
