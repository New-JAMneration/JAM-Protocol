package statekey_test

import (
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/statekey"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/merklization"
)

// TestStateKeyEquivalence_Storage verifies that types.BuildStorageStateKey and
// merklization.NewStorageStateKey produce identical results via the shared
// statekey package (eliminating the previous duplication).
func TestStateKeyEquivalence_Storage(t *testing.T) {
	testCases := []struct {
		name      string
		serviceID types.ServiceID
		rawKey    types.ByteSequence
	}{
		{"empty key, service 0", 0, nil},
		{"empty key, service 42", 42, nil},
		{"short key", 100, []byte{0xAA, 0xBB}},
		{"32-byte key", 700, make([]byte, 32)},
		{"long key", 1065941251, []byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xAB, 0xCD, 0xEF}},
		{"max service ID", 4294967295, []byte{0xFF}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fromTypes := types.BuildStorageStateKey(tc.serviceID, tc.rawKey)
			fromMerkl, err := merklization.NewStorageStateKey(tc.serviceID, tc.rawKey)
			if err != nil {
				t.Fatalf("NewStorageStateKey error: %v", err)
			}
			fromDirect := statekey.Storage(uint32(tc.serviceID), tc.rawKey)

			if fromTypes != fromMerkl {
				t.Errorf("types.Build (%x) != merklization.New (%x)", fromTypes, fromMerkl)
			}
			if types.StateKey(fromDirect) != fromTypes {
				t.Errorf("statekey.Storage (%x) != types.Build (%x)", fromDirect, fromTypes)
			}
		})
	}
}

// TestStateKeyEquivalence_PreimageMeta verifies that types.BuildPreimageMetaStateKey
// and merklization.NewPreimageMetaStateKey produce identical results.
func TestStateKeyEquivalence_PreimageMeta(t *testing.T) {
	testCases := []struct {
		name      string
		serviceID types.ServiceID
		length    types.U32
	}{
		{"service 0, length 0", 0, 0},
		{"service 42, length 100", 42, 100},
		{"service 700, length 65535", 700, 65535},
		{"large service ID", 2953942612, 1024},
		{"max service ID, max length", 4294967295, 4294967295},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var hash types.OpaqueHash
			for i := range hash {
				hash[i] = byte(i + int(tc.serviceID%256))
			}

			fromTypes := types.BuildPreimageMetaStateKey(tc.serviceID, hash, tc.length)
			fromMerkl, err := merklization.NewPreimageMetaStateKey(tc.serviceID, hash, tc.length)
			if err != nil {
				t.Fatalf("NewPreimageMetaStateKey error: %v", err)
			}
			fromDirect := statekey.PreimageMeta(uint32(tc.serviceID), [32]byte(hash), uint32(tc.length))

			if fromTypes != fromMerkl {
				t.Errorf("types.Build (%x) != merklization.New (%x)", fromTypes, fromMerkl)
			}
			if types.StateKey(fromDirect) != fromTypes {
				t.Errorf("statekey.PreimageMeta (%x) != types.Build (%x)", fromDirect, fromTypes)
			}
		})
	}
}

// TestStateKeyEquivalence_PreimageLookup verifies merklization.NewPreimageLookupStateKey
// matches statekey.PreimageLookup (no types.Build* equivalent for delta3).
func TestStateKeyEquivalence_PreimageLookup(t *testing.T) {
	testCases := []struct {
		name      string
		serviceID types.ServiceID
	}{
		{"service 0", 0},
		{"service 42", 42},
		{"service 700", 700},
		{"max service ID", 4294967295},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var hash types.OpaqueHash
			for i := range hash {
				hash[i] = byte(i * 3)
			}

			fromMerkl, err := merklization.NewPreimageLookupStateKey(tc.serviceID, hash)
			if err != nil {
				t.Fatalf("NewPreimageLookupStateKey error: %v", err)
			}
			fromDirect := statekey.PreimageLookup(uint32(tc.serviceID), [32]byte(hash))

			if types.StateKey(fromDirect) != fromMerkl {
				t.Errorf("statekey.PreimageLookup (%x) != merklization.New (%x)", fromDirect, fromMerkl)
			}
		})
	}
}
