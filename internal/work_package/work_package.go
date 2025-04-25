package work_package

import (
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/PolkaVM"
	"github.com/New-JAMneration/JAM-Protocol/internal/service_account"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/merkle_tree"
)

// Xi (14.11)
func workResultCompute(workPackage types.WorkPackage, coreIndex types.CoreIndex, pa types.OpaqueHash, pc types.ByteSequence, extrinsicMap PolkaVM.ExtrinsicDataMap, importSegments [][]types.ExportSegment, delta types.ServiceAccountState, lookup types.SegmentRootLookup) (types.WorkReport, error) {
	resultType, o, g := PolkaVM.Psi_I(workPackage, coreIndex)
	if resultType != types.WorkExecResultOk {
		return types.WorkReport{}, fmt.Errorf("work item execution failed: %v", resultType)
	}

	var results []types.WorkResult
	var exports [][]types.ExportSegment

	for j, item := range workPackage.Items {
		r, u, e := I(workPackage, j, o, importSegments, extrinsicMap, delta)
		result := C(item, r, u)
		results = append(results, result)
		exports = append(exports, e)
	}

	encoder := types.NewEncoder()
	encodedWorkPackage, err := encoder.Encode(&workPackage)
	if err != nil {
		return types.WorkReport{}, fmt.Errorf("failed to encode work package: %w", err)
	}
	workPackageHash := hash.Blake2bHash(encodedWorkPackage)
	var workPackgeBundle []byte
	var exportsData []types.ExportSegment
	for _, export := range exports {
		exportsData = append(exportsData, export...)
	}
	s, err := A(workPackageHash, workPackgeBundle, exportsData)
	if err != nil {
		return types.WorkReport{}, fmt.Errorf("failed to create work package spec: %w", err)
	}

	return types.WorkReport{
		PackageSpec:       s,
		Context:           workPackage.Context,
		CoreIndex:         coreIndex,
		AuthorizerHash:    pa,
		AuthOutput:        o,
		SegmentRootLookup: lookup,
		Results:           results,
		AuthGasUsed:       types.U64(g),
	}, nil
}

func A(workPackageHash types.OpaqueHash, workPackgeBundle []byte, exportsData []types.ExportSegment) (types.WorkPackageSpec, error) {
	var exports []types.ByteSequence
	for _, export := range exportsData {
		exports = append(exports, types.ByteSequence(export[:]))
	}
	exportsRoot := merkle_tree.M(exports, hash.Blake2bHash)
	erasureRoot, err := ComputeErasureRoot(workPackgeBundle, exportsData)
	if err != nil {
		return types.WorkPackageSpec{}, fmt.Errorf("failed to compute erasure root: %w", err)
	}
	return types.WorkPackageSpec{
		Hash:         types.WorkPackageHash(workPackageHash),
		Length:       types.U32(len(workPackgeBundle)),
		ErasureRoot:  types.ErasureRoot(erasureRoot),
		ExportsRoot:  types.ExportsRoot(exportsRoot),
		ExportsCount: types.U16(len(exportsData)),
	}, nil
}

func ComputeErasureRoot(bundle []byte, exportsData []types.ExportSegment) (types.OpaqueHash, error) {
	padded := PadToMultiple(bundle, types.ECBasicSize)

	shards, err := ErasureCode(padded, (len(bundle)+683)/684)
	if err != nil {
		return types.OpaqueHash{}, err
	}

	var hashedShards []types.OpaqueHash //bCloud
	for _, shard := range shards {
		hashedShards = append(hashedShards, hash.Blake2bHash(types.ByteSequence(shard)))
	}

	pagedProof, err := PagedProofs(exportsData)
	if err != nil {
		return types.OpaqueHash{}, err
	}
	fullSegments := append(exportsData, pagedProof...)
	var groupShards [][][]byte
	for i := 0; i < len(fullSegments); i++ {
		segmentShards, err := ErasureCode(fullSegments[i][:], 6)
		if err != nil {
			return types.OpaqueHash{}, err
		}
		groupShards = append(groupShards, segmentShards)
	}

	// transposed := Transpose(groupShards)
	// sCloud := for loop merkle_tree.Mb transposed
	// for loop Transpose([hashedShards, sCloud]) and concat, then use merkle_tree.Mb to get erasure root

	return types.OpaqueHash{}, nil
}

func Transpose[T any](input [][]T) [][]T {
	if len(input) == 0 || len(input[0]) == 0 {
		return [][]T{}
	}

	rows := len(input)
	cols := len(input[0])
	output := make([][]T, cols)
	for i := range output {
		output[i] = make([]T, rows)
		for j := range input {
			output[i][j] = input[j][i]
		}
	}
	return output
}

// mock function for ErasureCode
func ErasureCode(data []byte, k int) ([][]byte, error) {
	return nil, nil
}

// 14.9
func VerifyAuthorization(wp *types.WorkPackage, delta types.ServiceAccountState) (types.OpaqueHash, types.ByteSequence, types.ByteSequence, error) {
	pa := hash.Blake2bHash(append(wp.Authorizer.CodeHash[:], wp.Authorizer.Params[:]...))
	bytes := service_account.HistoricalLookup(delta[wp.AuthCodeHost], wp.Context.LookupAnchorSlot, wp.Authorizer.CodeHash)
	pm, pc, err := service_account.DecodeMetaCode(bytes)
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

func ExtractExtrinsics(data types.ByteSequence, specs []types.ExtrinsicSpec) (PolkaVM.ExtrinsicDataMap, error) {
	var result PolkaVM.ExtrinsicDataMap
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

// C (14.8)
func C(item types.WorkItem, result types.WorkExecResult, gas types.Gas) types.WorkResult {
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
	return result, nil
}

// (14.17)
func PadToMultiple(x []byte, n int) []byte {
	padLen := (n - (len(x) % n)) % n
	padding := make([]byte, padLen)
	return append(x, padding...)
}

func I(workPackage types.WorkPackage, j int, o types.ByteSequence, imports [][]types.ExportSegment, extrinsicMap PolkaVM.ExtrinsicDataMap, delta types.ServiceAccountState) (types.WorkExecResult, types.Gas, []types.ExportSegment) {
	workItem := workPackage.Items[j]
	expectedCount := workItem.ExportCount
	lSum := types.U16(0)
	for k := 0; k < j; k++ {
		lSum += workPackage.Items[k].ExportCount
	}

	refineInput := PolkaVM.RefineInput{
		WorkItemIndex:       uint(j),
		WorkPackage:         workPackage,
		AuthOutput:          o,
		ImportSegments:      imports,
		ExportSegmentOffset: uint(lSum),
		ServiceAccounts:     delta,
		ExtrinsicDataMap:    extrinsicMap,
	}

	refineOuput := PolkaVM.Psi_R(refineInput)
	r := refineOuput.RefineOutput
	e := refineOuput.EportSegment
	u := refineOuput.Gas

	if len(e) == int(expectedCount) {
		return types.WorkExecResult{
			refineOuput.WorkResult: r,
		}, u, e
	} else if refineOuput.WorkResult != types.WorkExecResultOk {
		emptyExport := make([]types.ExportSegment, expectedCount)
		return types.WorkExecResult{
			refineOuput.WorkResult: r,
		}, u, emptyExport
	} else {
		emptyExport := make([]types.ExportSegment, expectedCount)
		return types.WorkExecResult{
			types.WorkExecResultBadExports: nil,
		}, u, emptyExport
	}
}

func fetchFromDA(erasureRoot types.OpaqueHash, index types.U16) ([]types.ExportSegment, error) {
	// need to fetch and reconstruct from DA
	return []types.ExportSegment{}, nil
}
