package work_package

import (
	"errors"
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/PVM"
	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/service_account"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/merkle_tree"
	erasurecoding "github.com/New-JAMneration/JAM-Protocol/pkg/erasure_coding"
)

// (14.9)
func VerifyAuthorization(wp *types.WorkPackage, delta types.ServiceAccountState) (types.OpaqueHash, types.ByteSequence, types.ByteSequence, error) {
	pa := hash.Blake2bHash(append(wp.AuthCodeHash[:], wp.AuthorizerConfig[:]...))
	bytes := service_account.HistoricalLookup(delta[wp.AuthCodeHost], wp.Context.LookupAnchorSlot, wp.AuthCodeHash)
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

func ExtractExtrinsics(data types.ByteSequence, specs []types.ExtrinsicSpec) (PVM.ExtrinsicDataMap, error) {
	result := make(PVM.ExtrinsicDataMap)
	curr := 0

	for _, spec := range specs {
		length := int(spec.Len)
		if curr+length > len(data) {
			return nil, errors.New("extrinsic length overflow")
		}

		extracted := data[curr : curr+length]
		if hash.Blake2bHash(extracted) != spec.Hash {
			return nil, errors.New("extrinsic hash mismatch")
		}

		result[spec.Hash] = append([]byte(nil), extracted...)
		curr += length
	}

	if curr != len(data) {
		return nil, errors.New("data remains after extrinsics parsed")
	}

	return result, nil
}

// (14.15) second param: E(p,X#(pw),S#(pw),J#(pw))
func buildWorkPackageBundle(
	wp *types.WorkPackage,
	extrinsicMap PVM.ExtrinsicDataMap,
	importSegments types.ExportSegmentMatrix,
	importProofs types.OpaqueHashMatrix,
) ([]byte, error) {
	var extrinsics types.ExtrinsicDataList
	for _, item := range wp.Items {
		for _, extrinsic := range item.Extrinsic {
			extrinsics = append(extrinsics, types.ExtrinsicData(extrinsicMap[extrinsic.Hash]))
		}
	}

	bundle := types.WorkPackageBundle{
		Package:        *wp,
		Extrinsics:     extrinsics,
		ImportSegments: importSegments,
		ImportProofs:   importProofs,
	}
	output := []byte{}

	redisBackend, err := blockchain.GetRedisBackend()
	if err != nil {
		return nil, fmt.Errorf("failed to get redis backend: %w", err)
	}
	hashSegmentMap, err := redisBackend.GetHashSegmentMap()
	if err != nil {
		return nil, fmt.Errorf("failed to get hash segment map: %w", err)
	}
	encoder := types.NewEncoder()
	encoder.SetHashSegmentMap(hashSegmentMap)
	encoded, err := encoder.Encode(&bundle)
	if err != nil {
		return nil, fmt.Errorf("failed to encode work package bundle: %w", err)
	}
	output = append(output, encoded...)

	return output, nil
}

// Xi (14.11)
func WorkReportCompute(
	workPackage *types.WorkPackage,
	coreIndex types.CoreIndex,
	pa types.OpaqueHash,
	pc types.ByteSequence,
	extrinsicMap PVM.ExtrinsicDataMap,
	importSegments [][]types.ExportSegment,
	delta types.ServiceAccountState,
	workPackgeBundle []byte,
	workPackageHash types.OpaqueHash,
	pvm PVMExecutor,
) (types.WorkReport, error) {
	returnType := pvm.Psi_I(*workPackage, coreIndex, pc)
	o := returnType.WorkOutput
	g := returnType.Gas
	if returnType.WorkExecResult != types.WorkExecResultOk || len(o) > types.WorkReportOutputBlobsMaximumSize {
		return types.WorkReport{}, fmt.Errorf("work item execution failed: %v", returnType.WorkExecResult)
	}

	results := make([]types.WorkResult, 0, len(workPackage.Items))
	exports := make([][]types.ExportSegment, 0, len(workPackage.Items))

	rSum := 0
	for j, item := range workPackage.Items {
		r, u, e := I(*workPackage, j, o, importSegments, extrinsicMap, delta, pvm, rSum, coreIndex)
		rSum += len(r)
		result := C(item, r, u)
		results = append(results, result)
		exports = append(exports, e)
	}

	var exportsData []types.ExportSegment
	for _, export := range exports {
		exportsData = append(exportsData, export...)
	}
	s, err := A(workPackageHash, workPackgeBundle, exportsData)
	if err != nil {
		return types.WorkReport{}, fmt.Errorf("failed to create work package spec: %w", err)
	}
	return types.WorkReport{
		PackageSpec:    s,
		Context:        workPackage.Context,
		CoreIndex:      coreIndex,
		AuthorizerHash: pa,
		AuthOutput:     o,
		Results:        results,
		AuthGasUsed:    g,
	}, nil
}

func I(workPackage types.WorkPackage, j int, o types.ByteSequence, imports [][]types.ExportSegment, extrinsicMap PVM.ExtrinsicDataMap, delta types.ServiceAccountState, pvm PVMExecutor, rSum int, coreIndex types.CoreIndex) (types.WorkExecResult, types.Gas, []types.ExportSegment) {
	workItem := workPackage.Items[j]
	expectedCount := workItem.ExportCount
	lSum := 0
	for k := 0; k < j; k++ {
		lSum += int(workPackage.Items[k].ExportCount)
	}

	refineInput := PVM.RefineInput{
		CoreIndex:           coreIndex,
		WorkItemIndex:       uint(j),
		WorkPackage:         workPackage,
		AuthOutput:          o,
		ImportSegments:      imports,
		ExportSegmentOffset: uint(lSum),
		ServiceAccounts:     delta,
		ExtrinsicDataMap:    extrinsicMap,
	}

	refineOuput := pvm.RefineInvoke(refineInput)
	r := refineOuput.RefineOutput
	e := refineOuput.ExportSegment
	u := refineOuput.Gas
	z := len(o) + rSum
	if len(r)+z > types.WorkReportOutputBlobsMaximumSize {
		emptyExport := make([]types.ExportSegment, expectedCount)
		return types.WorkExecResult{
			types.WorkExecResultReportOversize: nil,
		}, u, emptyExport
	} else if len(e) != int(workItem.ExportCount) {
		emptyExport := make([]types.ExportSegment, expectedCount)
		return types.WorkExecResult{
			types.WorkExecResultBadExports: nil,
		}, u, emptyExport
	} else if refineOuput.WorkResult != types.WorkExecResultOk {
		emptyExport := make([]types.ExportSegment, expectedCount)
		return types.WorkExecResult{
			refineOuput.WorkResult: r,
		}, u, emptyExport
	} else {
		return types.WorkExecResult{
			refineOuput.WorkResult: r,
		}, u, e
	}
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
			GasUsed:        gas,
			Imports:        importCount,
			ExtrinsicCount: item.ExportCount,
			ExtrinsicSize:  extrinsicSize,
			Exports:        zSum,
		},
	}
}

// A (14.16)
func A(workPackageHash types.OpaqueHash, workPackgeBundle []byte, exportsData []types.ExportSegment) (types.WorkPackageSpec, error) {
	var exports []types.ByteSequence
	for _, export := range exportsData {
		exports = append(exports, types.ByteSequence(export[:]))
	}
	exportsRoot := merkle_tree.M(exports, hash.Blake2bHash)
	erasureRoot, err := computeErasureRoot(workPackgeBundle, exportsData)
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

func computeErasureRoot(bundle []byte, exportsData []types.ExportSegment) (types.OpaqueHash, error) {
	bcloud, err := buildBCloud(bundle)
	if err != nil {
		return types.OpaqueHash{}, err
	}
	scloud, err := buildSCloud(exportsData)
	if err != nil {
		return types.OpaqueHash{}, err
	}
	erasureRoot := mergeBCloudSCloud(bcloud, scloud)
	return erasureRoot, nil
}

func buildBCloud(bundle []byte) ([]types.OpaqueHash, error) {
	padded := PadToMultiple(bundle, types.ECBasicSize)

	shards, err := erasurecoding.EncodeDataShards(padded, types.DataShards, types.TotalShards-types.DataShards)
	if err != nil {
		return nil, err
	}

	hashedShards := make([]types.OpaqueHash, len(shards))
	for i, shard := range shards {
		hashedShards[i] = hash.Blake2bHash(types.ByteSequence(shard))
	}

	return hashedShards, nil
}

func buildSCloud(exports []types.ExportSegment) ([]types.OpaqueHash, error) {
	pagedProof, err := PagedProofs(exports)
	if err != nil {
		return nil, err
	}
	fullSegments := append(exports, pagedProof...)

	groupShards := make([][][]byte, len(fullSegments))
	for i := range fullSegments {
		shards, err := erasurecoding.EncodeDataShards(fullSegments[i][:], types.DataShards, types.TotalShards-types.DataShards)
		if err != nil {
			return nil, err
		}
		groupShards[i] = shards
	}
	transposed := Transpose(groupShards)
	merkleResult := make([]types.OpaqueHash, len(transposed))
	for i := range transposed {
		byteSequences := make([]types.ByteSequence, len(transposed[i]))
		for j, shard := range transposed[i] {
			byteSequences[j] = types.ByteSequence(shard)
		}
		merkleResult[i] = merkle_tree.Mb(byteSequences, hash.Blake2bHash)
	}
	return merkleResult, nil
}

func mergeBCloudSCloud(bcloud []types.OpaqueHash, scloud []types.OpaqueHash) types.OpaqueHash {
	merged := make([]types.ByteSequence, len(bcloud))
	for i := range bcloud {
		pair := append(bcloud[i][:], scloud[i][:]...)
		merged[i] = types.ByteSequence(pair)
	}
	return merkle_tree.Mb(merged, hash.Blake2bHash)
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

func convertMapToLookup(m map[types.OpaqueHash]types.OpaqueHash) types.SegmentRootLookup {
	lookup := make(types.SegmentRootLookup, 0, len(m))
	for wpHash, segmentRoot := range m {
		lookup = append(lookup, types.SegmentRootLookupItem{
			WorkPackageHash: types.WorkPackageHash(wpHash),
			SegmentTreeRoot: segmentRoot,
		})
	}
	return lookup
}

func extractExtrinsicMapFromBundle(workPackage *types.WorkPackage, extrinsics types.ExtrinsicDataList) (PVM.ExtrinsicDataMap, error) {
	specs := FlattenExtrinsicSpecs(workPackage)

	if len(specs) != len(extrinsics) {
		return nil, fmt.Errorf("extrinsic count mismatch: %d specs vs %d data", len(specs), len(extrinsics))
	}

	result := make(PVM.ExtrinsicDataMap)
	for i, spec := range specs {
		result[spec.Hash] = PVM.ExtrinsicData(extrinsics[i])
	}
	return result, nil
}
