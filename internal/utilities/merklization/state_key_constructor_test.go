package merklization

import (
	"encoding/hex"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func TestStateKeyContractorStateWrapper(t *testing.T) {
	testCases := []struct {
		stateIndex                types.U8
		expectedStateKeyHexString string
	}{
		{1, "0x01000000000000000000000000000000000000000000000000000000000000"},
		{2, "0x02000000000000000000000000000000000000000000000000000000000000"},
		{3, "0x03000000000000000000000000000000000000000000000000000000000000"},
		{4, "0x04000000000000000000000000000000000000000000000000000000000000"},
		{5, "0x05000000000000000000000000000000000000000000000000000000000000"},
		{6, "0x06000000000000000000000000000000000000000000000000000000000000"},
		{7, "0x07000000000000000000000000000000000000000000000000000000000000"},
		{8, "0x08000000000000000000000000000000000000000000000000000000000000"},
		{9, "0x09000000000000000000000000000000000000000000000000000000000000"},
		{10, "0x0a000000000000000000000000000000000000000000000000000000000000"},
		{11, "0x0b000000000000000000000000000000000000000000000000000000000000"},
		{12, "0x0c000000000000000000000000000000000000000000000000000000000000"},
		{13, "0x0d000000000000000000000000000000000000000000000000000000000000"},
		{14, "0x0e000000000000000000000000000000000000000000000000000000000000"},
		{15, "0x0f000000000000000000000000000000000000000000000000000000000000"},
	}

	for _, testCase := range testCases {
		stateWrapper := StateWrapper{StateIndex: testCase.stateIndex}
		actual := stateWrapper.StateKeyConstruct()

		// Decode the expected hex string to a state key
		bytes, err := hex.DecodeString(testCase.expectedStateKeyHexString[2:])
		if err != nil {
			t.Errorf("Error decoding expected state key hex: %v", err)
		}

		expectedStateKey := types.StateKey{}
		copy(expectedStateKey[:], bytes)

		if actual != expectedStateKey {
			t.Errorf("Expected %x, got %x", expectedStateKey, actual)
		} else {
			t.Logf("Test passed for C(%d)", testCase.stateIndex)
		}
	}
}

func TestStateKeyContractorStateServiceWrapper(t *testing.T) {
	testCases := []struct {
		stateIndex                types.U8
		serviceIndex              types.ServiceId
		expectedStateKeyHexString string
	}{
		{255, 0, "0xff000000000000000000000000000000000000000000000000000000000000"},
		{255, 1065941251, "0xff0300f90088003f0000000000000000000000000000000000000000000000"},
		{255, 2953942612, "0xff540096001100b00000000000000000000000000000000000000000000000"},
		{255, 3273088977, "0xffd1005f001700c30000000000000000000000000000000000000000000000"},
		{255, 1343007977, "0xffe900ac000c00500000000000000000000000000000000000000000000000"},
	}

	for _, testCase := range testCases {
		stateServiceWrapper := StateServiceWrapper{
			StateIndex:   testCase.stateIndex,
			ServiceIndex: testCase.serviceIndex,
		}

		actual := stateServiceWrapper.StateKeyConstruct()

		// Decode the expected hex string to a byte array
		bytes, err := hex.DecodeString(testCase.expectedStateKeyHexString[2:])
		if err != nil {
			t.Errorf("Error decoding expected state key hex: %v", err)
		}

		expectedStateKey := types.StateKey{}
		copy(expectedStateKey[:], bytes)

		if actual != expectedStateKey {
			t.Errorf("Expected %x, got %x", expectedStateKey, actual)
		} else {
			t.Logf("Test passed for C(%d, %d)", testCase.stateIndex, testCase.serviceIndex)
		}
	}
}

func TestStateKeyContractorServiceWrapper(t *testing.T) {
	var serviceIndex types.ServiceId = 700
	var h [27]byte = [27]byte{
		0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef,
		0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef,
		0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef,
		0x01, 0x23, 0x45,
	}
	service := ServiceWrapper{ServiceIndex: serviceIndex, h: h}

	encoder := types.NewEncoder()
	n, err := encoder.EncodeUintWithLength(uint64(serviceIndex), 4)
	if err != nil {
		t.Errorf("Error encoding service index: %v", err)
	}

	expectedStateKey := types.StateKey{
		n[0], h[0], n[1], h[1], n[2], h[2], n[3], h[3],
		h[4], h[5], h[6], h[7],
		h[8], h[9], h[10], h[11],
		h[12], h[13], h[14], h[15],
		h[16], h[17], h[18], h[19],
		h[20], h[21], h[22], h[23],
		h[24], h[25], h[26],
	}

	actual := service.StateKeyConstruct()

	if actual != expectedStateKey {
		t.Errorf("Expected %x, got %x", expectedStateKey, actual)
	}
}
