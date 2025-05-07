package work_package

import (
	"fmt"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/PolkaVM"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/require"
	"github.com/test-go/testify/mock"
)

type MockPVMExecutor struct {
	mock.Mock
}

func (m *MockPVMExecutor) Psi_I(wp types.WorkPackage, core types.CoreIndex, pc types.ByteSequence) PolkaVM.Psi_I_ReturnType {
	args := m.Called(wp, core, pc)
	return args.Get(0).(PolkaVM.Psi_I_ReturnType)
}

func (m *MockPVMExecutor) RefineInvoke(input PolkaVM.RefineInput) PolkaVM.RefineOutput {
	args := m.Called(input)
	return args.Get(0).(PolkaVM.RefineOutput)
}

type MockFetcher struct {
	mock.Mock
}

func (m *MockFetcher) Fetch(erasureRoot types.OpaqueHash, index types.U16) (types.ExportSegment, []types.OpaqueHash, error) {
	args := m.Called(erasureRoot, index)
	return args.Get(0).(types.ExportSegment), args.Get(1).([]types.OpaqueHash), args.Error(2)
}

func TestFetchImportSegments(t *testing.T) {
	// run miniredis server
	rdb, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to init miniredis %v:", err)
	}
	defer rdb.Close()

	client := store.NewRedisClient(rdb.Addr(), "", 0)
	erasureMap := store.NewSegmentErasureMap(client)
	erasureMap.Save(types.OpaqueHash{0x01}, types.OpaqueHash{0x02})
	erasureMap.Save(types.OpaqueHash{0x03}, types.OpaqueHash{0x04})

	wp := &types.WorkPackage{
		Items: []types.WorkItem{
			{
				ImportSegments: []types.ImportSpec{
					{TreeRoot: types.OpaqueHash{0x01}, Index: 0},
					{TreeRoot: types.OpaqueHash{0x03}, Index: 1},
				},
			},
		},
	}

	var fakeSeg types.ExportSegment
	copy(fakeSeg[:], []byte("fake-segment"))
	fakeProof := []types.OpaqueHash{
		hash.Blake2bHash([]byte("proof-1")),
		hash.Blake2bHash([]byte("proof-2")),
	}

	// mock fetcher
	mockFetcher := new(MockFetcher)
	mockFetcher.
		On("Fetch", types.OpaqueHash{0x02}, types.U16(0)).
		Return(fakeSeg, fakeProof, nil)
	mockFetcher.
		On("Fetch", types.OpaqueHash{0x04}, types.U16(1)).
		Return(fakeSeg, fakeProof, nil)

	// Initialize controller
	controller := &WorkPackageController{
		WorkPackage: wp,
		ErasureMap:  erasureMap,
		Fetcher:     mockFetcher,
	}

	segments, proofs, err := controller.fetchImportSegments(map[types.OpaqueHash]types.OpaqueHash{})
	require.NoError(t, err)

	require.Len(t, segments, 1)
	require.Len(t, segments[0], 2)
	for _, seg := range segments[0] {
		require.Equal(t, fakeSeg, seg)
	}

	require.Len(t, proofs, 1)
	require.Len(t, proofs[0], 4)
	require.Equal(t, fakeProof[0], proofs[0][0])
	require.Equal(t, fakeProof[1], proofs[0][1])

	mockFetcher.AssertExpectations(t)
}

func TestWorkPackageController_InitialProcess(t *testing.T) {
	// run miniredis server
	rdb, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to init miniredis %v:", err)
	}
	defer rdb.Close()

	client := store.NewRedisClient(rdb.Addr(), "", 0)

	segmentMap := store.NewHashSegmentMap(client)
	erasureMap := store.NewSegmentErasureMap(client)
	erasureMap.Save(types.OpaqueHash{0x01}, types.OpaqueHash{0x02})
	erasureMap.Save(types.OpaqueHash{0x03}, types.OpaqueHash{0x04})

	inputDelta := types.ServiceAccountState{
		types.ServiceId(1): {
			ServiceInfo: types.ServiceInfo{
				CodeHash:   types.OpaqueHash{0x04, 0x05, 0x06},
				Balance:    1000,
				MinItemGas: types.Gas(100),
				MinMemoGas: types.Gas(100),
				Bytes:      types.U64(1),
				Items:      types.U32(1),
			},
			PreimageLookup: nil,
			LookupDict:     nil,
			StorageDict:    nil,
		},
	}
	s := store.GetInstance()
	s.GetPriorStates().SetDelta(inputDelta)

	wp := &types.WorkPackage{
		Authorization: types.ByteSequence{0x01, 0x02, 0x03},
		AuthCodeHost:  types.ServiceId(1),
		Authorizer: types.Authorizer{
			CodeHash: types.OpaqueHash{0x04, 0x05, 0x06},
			Params:   types.ByteSequence{0x07, 0x08, 0x09},
		},
		Context: types.RefineContext{
			Anchor:           types.HeaderHash{0x0A, 0x0B, 0x0C},
			StateRoot:        types.StateRoot{0x0D, 0x0E, 0x0F},
			BeefyRoot:        types.BeefyRoot{0x10, 0x11, 0x12},
			LookupAnchor:     types.HeaderHash{0x13, 0x14, 0x15},
			LookupAnchorSlot: types.TimeSlot(12345),
			Prerequisites:    nil,
		},
		Items: []types.WorkItem{
			{
				Service:            types.ServiceId(1),
				CodeHash:           types.OpaqueHash{0x16, 0x17, 0x18},
				Payload:            types.ByteSequence{0x19, 0x1A, 0x1B},
				RefineGasLimit:     types.Gas(1000),
				AccumulateGasLimit: types.Gas(2000),
				ExportCount:        types.U16(1),
				ImportSegments: []types.ImportSpec{
					{TreeRoot: types.OpaqueHash{0x01}, Index: 0},
					{TreeRoot: types.OpaqueHash{0x03}, Index: 1},
				},
				Extrinsic: []types.ExtrinsicSpec{
					{Hash: hash.Blake2bHash([]byte("abc")), Len: 3},
					{Hash: hash.Blake2bHash([]byte("def")), Len: 3},
				},
			},
		},
	}

	extrinsics := []byte("abcdef")
	coreIndex := types.CoreIndex(0)

	// mock PVM
	mockPVM := new(MockPVMExecutor)
	mockPVM.On("Psi_I", mock.Anything, mock.Anything, mock.Anything).Return(PolkaVM.Psi_I_ReturnType{
		WorkExecResult: types.WorkExecResultOk,
		WorkOutput:     []byte("auth output"),
		Gas:            types.Gas(10),
	})
	mockPVM.On("RefineInvoke", mock.Anything).Return(PolkaVM.RefineOutput{
		WorkResult:   types.WorkExecResultOk,
		RefineOutput: []byte("refine output"),
		ExportSegment: []types.ExportSegment{
			[4104]byte{0x1F, 0x20, 0x21},
		},
		Gas: types.Gas(10),
	})

	// Mock fetch DA
	var fakeSeg types.ExportSegment
	copy(fakeSeg[:], []byte("fake-segment"))

	fakeProof := []types.OpaqueHash{
		hash.Blake2bHash([]byte("fake-proof-1")),
		hash.Blake2bHash([]byte("fake-proof-2")),
	}
	mockFetcher := new(MockFetcher)
	mockFetcher.
		On("Fetch", types.OpaqueHash{0x02}, types.U16(0)).
		Return(fakeSeg, fakeProof, nil)
	mockFetcher.
		On("Fetch", types.OpaqueHash{0x04}, types.U16(1)).
		Return(fakeSeg, fakeProof, nil)

	// Initialize the controller
	controller := NewInitialController(wp, extrinsics, erasureMap, segmentMap, coreIndex, mockFetcher)
	controller.PVM = mockPVM

	fmt.Println("Processing work package...")
	report, err := controller.Process()
	require.NoError(t, err)
	require.Equal(t, report.CoreIndex, coreIndex)

	require.Equal(t, report.AuthOutput, types.ByteSequence("auth output"))

	require.Equal(t, report.Results[0].Result[types.WorkExecResultOk], []byte("refine output"))

	encode := types.NewEncoder()
	encoded, err := encode.Encode(wp)
	require.NoError(t, err)
	workPackageHash := hash.Blake2bHash(encoded)
	require.Equal(t, report.PackageSpec.Hash, types.WorkPackageHash(workPackageHash))

	// Check the local map with report
	dict, err := segmentMap.LoadDict()
	for _, lookupItem := range report.SegmentRootLookup {
		key := types.OpaqueHash(lookupItem.WorkPackageHash[:])
		require.Equal(t, dict[key], lookupItem.SegmentTreeRoot)
	}
	require.NoError(t, err)
	require.Greater(t, len(dict), 0, "segment root dict should be updated")

	// TODO: only check few things now, can check more with test package and report
}

func TestPrepareInputs_Shared(t *testing.T) {
	// run miniredis server
	rdb, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to init miniredis %v:", err)
	}
	defer rdb.Close()

	client := store.NewRedisClient(rdb.Addr(), "", 0)

	segmentMap := store.NewHashSegmentMap(client)
	erasureMap := store.NewSegmentErasureMap(client)
	erasureMap.Save(types.OpaqueHash{0x01}, types.OpaqueHash{0x02})
	erasureMap.Save(types.OpaqueHash{0x03}, types.OpaqueHash{0x04})

	inputDelta := types.ServiceAccountState{
		types.ServiceId(1): {
			ServiceInfo: types.ServiceInfo{
				CodeHash:   types.OpaqueHash{0x04, 0x05, 0x06},
				Balance:    1000,
				MinItemGas: types.Gas(100),
				MinMemoGas: types.Gas(100),
				Bytes:      types.U64(1),
				Items:      types.U32(1),
			},
			PreimageLookup: nil,
			LookupDict:     nil,
			StorageDict:    nil,
		},
	}
	s := store.GetInstance()
	s.GetPriorStates().SetDelta(inputDelta)

	wp := &types.WorkPackage{
		Authorization: types.ByteSequence{0x01, 0x02, 0x03},
		AuthCodeHost:  types.ServiceId(1),
		Authorizer: types.Authorizer{
			CodeHash: types.OpaqueHash{0x04, 0x05, 0x06},
			Params:   types.ByteSequence{0x07, 0x08, 0x09},
		},
		Context: types.RefineContext{
			Anchor:           types.HeaderHash{0x0A, 0x0B, 0x0C},
			StateRoot:        types.StateRoot{0x0D, 0x0E, 0x0F},
			BeefyRoot:        types.BeefyRoot{0x10, 0x11, 0x12},
			LookupAnchor:     types.HeaderHash{0x13, 0x14, 0x15},
			LookupAnchorSlot: types.TimeSlot(12345),
			Prerequisites:    nil,
		},
		Items: []types.WorkItem{
			{
				Service:            types.ServiceId(1),
				CodeHash:           types.OpaqueHash{0x16, 0x17, 0x18},
				Payload:            types.ByteSequence{0x19, 0x1A, 0x1B},
				RefineGasLimit:     types.Gas(1000),
				AccumulateGasLimit: types.Gas(2000),
				ExportCount:        types.U16(1),
				ImportSegments: []types.ImportSpec{
					{TreeRoot: types.OpaqueHash{0x01}, Index: 0},
					{TreeRoot: types.OpaqueHash{0x03}, Index: 1},
				},
				Extrinsic: []types.ExtrinsicSpec{
					{Hash: hash.Blake2bHash([]byte("abc")), Len: 3},
					{Hash: hash.Blake2bHash([]byte("def")), Len: 3},
				},
			},
		},
	}

	coreIndex := types.CoreIndex(0)

	// mock PVM
	mockPVM := new(MockPVMExecutor)
	mockPVM.On("Psi_I", mock.Anything, mock.Anything, mock.Anything).Return(PolkaVM.Psi_I_ReturnType{
		WorkExecResult: types.WorkExecResultOk,
		WorkOutput:     []byte("auth output"),
		Gas:            types.Gas(10),
	})
	mockPVM.On("RefineInvoke", mock.Anything).Return(PolkaVM.RefineOutput{
		WorkResult:   types.WorkExecResultOk,
		RefineOutput: []byte("refine output"),
		ExportSegment: []types.ExportSegment{
			[4104]byte{0x1F, 0x20, 0x21},
		},
		Gas: types.Gas(10),
	})

	extrinsics := types.ExtrinsicDataList{
		[]byte("abc"),
		[]byte("def"),
	}

	fakeSegment := types.ExportSegment{}
	copy(fakeSegment[:], []byte("seg"))

	bundle := &types.WorkPackageBundle{
		Package:        *wp,
		Extrinsics:     extrinsics,
		ImportSegments: types.ExportSegmentMatrix{{fakeSegment, fakeSegment}},
		ImportProofs:   types.OpaqueHashMatrix{{types.OpaqueHash{0x01}}},
	}

	// encode bundle
	encoder := types.NewEncoder()
	data, err := encoder.Encode(bundle)
	require.NoError(t, err)

	controller := NewSharedController(data, erasureMap, segmentMap, coreIndex)
	controller.PVM = mockPVM

	fmt.Println("Processing work package...")
	report, err := controller.Process()

	require.NoError(t, err)
	require.Equal(t, report.CoreIndex, coreIndex)

	require.Equal(t, report.AuthOutput, types.ByteSequence("auth output"))

	require.Equal(t, report.Results[0].Result[types.WorkExecResultOk], []byte("refine output"))

	encode := types.NewEncoder()
	encoded, err := encode.Encode(wp)
	require.NoError(t, err)
	workPackageHash := hash.Blake2bHash(encoded)
	require.Equal(t, report.PackageSpec.Hash, types.WorkPackageHash(workPackageHash))

	// Check the local map with report
	dict, err := segmentMap.LoadDict()
	for _, lookupItem := range report.SegmentRootLookup {
		key := types.OpaqueHash(lookupItem.WorkPackageHash[:])
		require.Equal(t, dict[key], lookupItem.SegmentTreeRoot)
	}
	require.NoError(t, err)
	require.Greater(t, len(dict), 0, "segment root dict should be updated")

	// TODO: only check few things now, can check more with test package and report
}
