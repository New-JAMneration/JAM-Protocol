package ce

import (
	"crypto/ed25519"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/PVM"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/store/keystore"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/work_package"
	"github.com/alicebob/miniredis/v2"
)

// FakePVMExecutor mocks the PVMExecutor interface for unit testing
type FakePVMExecutor struct{}

func (f *FakePVMExecutor) Psi_I(p types.WorkPackage, c types.CoreIndex, code types.ByteSequence) PVM.Psi_I_ReturnType {
	return PVM.Psi_I_ReturnType{
		WorkExecResult: types.WorkExecResultOk,
		WorkOutput:     []byte("auth output"),
		Gas:            types.Gas(10),
	}
}
func (f *FakePVMExecutor) RefineInvoke(input PVM.RefineInput) PVM.RefineOutput {
	return PVM.RefineOutput{
		WorkResult:   types.WorkExecResultOk,
		RefineOutput: []byte("refine output"),
		ExportSegment: []types.ExportSegment{
			[4104]byte{0x1F, 0x20, 0x21},
		},
		Gas: types.Gas(10),
	}
}

func TestHandleWorkPackageShare(t *testing.T) {
	// Prepare test data
	coreIndex := types.CoreIndex(1)
	coreIndexBytes := make([]byte, 2)
	coreIndexBytes[0] = byte(coreIndex)
	coreIndexBytes[1] = byte(coreIndex >> 8)

	// One mapping
	mappingCount := byte(1)
	wpHash := make([]byte, 32)
	for i := range wpHash {
		wpHash[i] = byte(i)
	}
	segRoot := make([]byte, 32)
	for i := range segRoot {
		segRoot[i] = byte(100 + i)
	}

	// Build a minimal valid work-package bundle
	// Use similar logic as TestBuildWorkPackageBundle in work_package_test.go
	extrinsicHash1 := [32]byte{}
	for i := range extrinsicHash1 {
		extrinsicHash1[i] = byte(i)
	}
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
					{TreeRoot: types.OpaqueHash{0x1C, 0x1D, 0x1E}, Index: types.U16(1)},
				},
				Extrinsic: []types.ExtrinsicSpec{
					{Hash: extrinsicHash1, Len: 5},
				},
			},
		},
	}
	extrinsicMap := PVM.ExtrinsicDataMap{}
	extrinsicMap[types.OpaqueHash(extrinsicHash1)] = []byte("abcde")
	segment := [4104]byte{}
	copy(segment[:], []byte("segment1"))
	importSegments := types.ExportSegmentMatrix{
		{
			types.ExportSegment(segment),
		},
	}
	importProofs := types.OpaqueHashMatrix{
		{
			[32]byte{1},
		},
	}
	bundle, err := work_package.BuildWorkPackageBundle(wp, extrinsicMap, importSegments, importProofs)
	if err != nil {
		t.Fatalf("failed to build work-package bundle: %v", err)
	}

	// Compose input: coreIndex ++ mappingCount ++ mapping ++ bundle
	input := append(coreIndexBytes, mappingCount)
	input = append(input, wpHash...)
	input = append(input, segRoot...)
	input = append(input, bundle...)

	// Set up service account state so the handler can verify the work-package
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

	stream := newMockStream(input)

	// Generate Ed25519 keypair
	pub, priv, _ := ed25519.GenerateKey(nil)
	keypair, _ := keystore.FromEd25519PrivateKey(priv)

	rdb, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to init miniredis: %v", err)
	}
	defer rdb.Close()
	client := store.NewRedisClient(rdb.Addr(), "", 0)
	segmentMap := store.NewHashSegmentMap(client)
	erasureMap := store.NewSegmentErasureMap(client)
	fakePVM := &FakePVMExecutor{}
	err = HandleWorkPackageShare(nil, stream, keypair, fakePVM, erasureMap, segmentMap) // Use mockStream directly for test
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	// Check output: should be 32 bytes hash + 64 bytes signature
	resp := stream.w.Bytes()
	if len(resp) != 96 {
		t.Fatalf("expected 96 bytes response, got %d", len(resp))
	}
	workReportHash := resp[:32]
	sig := resp[32:]
	if !ed25519.Verify(pub, workReportHash, sig) {
		t.Errorf("signature verification failed")
	}
}
