package ce

import (
	"bytes"

	"github.com/New-JAMneration/JAM-Protocol/PVM"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	"github.com/New-JAMneration/JAM-Protocol/internal/work_package"
)

func CreateTestWorkPackageBundle() *types.WorkPackageBundle {
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
				ExportCount:        types.U16(3),
				ImportSegments: []types.ImportSpec{
					{TreeRoot: types.OpaqueHash{0x1C, 0x1D, 0x1E}, Index: types.U16(1)},
				},
				Extrinsic: []types.ExtrinsicSpec{
					{Hash: hash.Blake2bHash([]byte("test")), Len: 4},
				},
			},
		},
	}

	extrinsicMap := PVM.ExtrinsicDataMap{
		hash.Blake2bHash([]byte("test")): []byte("test"),
	}

	segment := make([]byte, 4104)
	copy(segment, bytes.Repeat([]byte("mock_segment_data"), 200))
	importSegments := types.ExportSegmentMatrix{
		{types.ExportSegment(segment)},
	}

	importProofs := types.OpaqueHashMatrix{
		{types.OpaqueHash{0x01}},
	}

	bundleBytes, err := work_package.BuildWorkPackageBundle(wp, extrinsicMap, importSegments, importProofs)
	if err != nil {
		return &types.WorkPackageBundle{
			Package:        *wp,
			Extrinsics:     types.ExtrinsicDataList{types.ExtrinsicData(extrinsicMap[hash.Blake2bHash([]byte("test"))])},
			ImportSegments: importSegments,
			ImportProofs:   importProofs,
		}
	}

	var bundle types.WorkPackageBundle
	decoder := types.NewDecoder()
	if err := decoder.Decode(bundleBytes, &bundle); err != nil {
		return &types.WorkPackageBundle{
			Package:        *wp,
			Extrinsics:     types.ExtrinsicDataList{types.ExtrinsicData(extrinsicMap[hash.Blake2bHash([]byte("test"))])},
			ImportSegments: importSegments,
			ImportProofs:   importProofs,
		}
	}

	return &bundle
}

// CreateTestWorkPackageBundleWithCustomExtrinsics creates a test bundle with custom extrinsic data
func CreateTestWorkPackageBundleWithCustomExtrinsics(extrinsicData map[string][]byte) *types.WorkPackageBundle {
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
				ExportCount:        types.U16(3),
				ImportSegments: []types.ImportSpec{
					{TreeRoot: types.OpaqueHash{0x1C, 0x1D, 0x1E}, Index: types.U16(1)},
				},
				Extrinsic: []types.ExtrinsicSpec{
					{Hash: hash.Blake2bHash([]byte("abc")), Len: 3},
					{Hash: hash.Blake2bHash([]byte("def")), Len: 3},
				},
			},
		},
	}

	extrinsicMap := PVM.ExtrinsicDataMap{
		hash.Blake2bHash([]byte("abc")): bytes.Repeat([]byte("abc"), 1000),
		hash.Blake2bHash([]byte("def")): bytes.Repeat([]byte("def"), 1000),
	}

	for key, data := range extrinsicData {
		extrinsicMap[hash.Blake2bHash([]byte(key))] = data
	}

	segment := make([]byte, 4104)
	copy(segment, bytes.Repeat([]byte("mock_segment_data"), 200))
	importSegments := types.ExportSegmentMatrix{
		{types.ExportSegment(segment)},
	}

	importProofs := types.OpaqueHashMatrix{
		{types.OpaqueHash{0x01}},
	}

	bundleBytes, err := work_package.BuildWorkPackageBundle(wp, extrinsicMap, importSegments, importProofs)
	if err != nil {
		return &types.WorkPackageBundle{
			Package:        *wp,
			Extrinsics:     types.ExtrinsicDataList{types.ExtrinsicData(extrinsicMap[hash.Blake2bHash([]byte("abc"))]), types.ExtrinsicData(extrinsicMap[hash.Blake2bHash([]byte("def"))])},
			ImportSegments: importSegments,
			ImportProofs:   importProofs,
		}
	}

	var bundle types.WorkPackageBundle
	decoder := types.NewDecoder()
	if err := decoder.Decode(bundleBytes, &bundle); err != nil {
		return &types.WorkPackageBundle{
			Package:        *wp,
			Extrinsics:     types.ExtrinsicDataList{types.ExtrinsicData(extrinsicMap[hash.Blake2bHash([]byte("abc"))]), types.ExtrinsicData(extrinsicMap[hash.Blake2bHash([]byte("def"))])},
			ImportSegments: importSegments,
			ImportProofs:   importProofs,
		}
	}

	return &bundle
}

func CreateTestWorkPackageBundleForCE134(extrinsicHash types.OpaqueHash, extrinsicData []byte) ([]byte, error) {
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
					{Hash: extrinsicHash, Len: types.U32(len(extrinsicData))},
				},
			},
		},
	}

	extrinsicMap := PVM.ExtrinsicDataMap{
		extrinsicHash: extrinsicData,
	}

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

	return work_package.BuildWorkPackageBundle(wp, extrinsicMap, importSegments, importProofs)
}

// encodeLE16 encodes a uint16 as little-endian bytes
func encodeLE16(value uint16) []byte {
	return []byte{
		byte(value),
		byte(value >> 8),
	}
}

func encodeLE32(value uint32) []byte {
	return []byte{
		byte(value),
		byte(value >> 8),
		byte(value >> 16),
		byte(value >> 24),
	}
}
