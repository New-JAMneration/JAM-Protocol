package merklization

import (
	"testing"

	jamTypes "github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities"
)

func TestStateWrapper(t *testing.T) {
	var stateIndex jamTypes.U8 = 7

	state := StateWrapper{StateIndex: stateIndex}

	expected := jamTypes.OpaqueHash{byte(stateIndex)}
	actual := state.StateKeyConstruct()
	flag := false
	for i := 0; i < len(expected); i++ {
		if expected[i] != actual[i] {
			flag = true
		}
	}
	if flag {
		t.Errorf("Expected %v, got %v", expected, actual)
	}
}

func TestStateServiceWrapper(t *testing.T) {
	var stateIndex jamTypes.U8 = 5
	var serviceIndex jamTypes.U32 = 1000

	stateService := StateServiceWrapper{StateIndex: stateIndex, ServiceIndex: serviceIndex}
	Serialized := utilities.SerializeFixedLength(serviceIndex, 4)

	expected := jamTypes.OpaqueHash{byte(stateIndex), Serialized[0], 0, Serialized[1], 0, Serialized[2], 0, Serialized[3], 0}
	actual := stateService.StateKeyConstruct()
	flag := false
	for i := 0; i < len(expected); i++ {
		if expected[i] != actual[i] {
			flag = true
		}
	}
	if flag {
		t.Errorf("Expected %v, got %v", expected, actual)
	}
}

func TestServiceWrapper(t *testing.T) {
	var serviceIndex jamTypes.U32 = 700
	var hash jamTypes.OpaqueHash = jamTypes.OpaqueHash{
		0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef,
		0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef,
		0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef,
		0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef,
	}
	service := ServiceWrapper{ServiceIndex: serviceIndex, Hash: hash}
	Serialized := utilities.SerializeFixedLength(serviceIndex, 4)
	// t.Log("Serialized :", Serialized)
	// t.Log("Hash :", hash)
	expected := jamTypes.OpaqueHash{
		Serialized[0], hash[0], Serialized[1], hash[1], Serialized[2], hash[2], Serialized[3], hash[3],
		hash[4], hash[5], hash[6], hash[7],
		hash[8], hash[9], hash[10], hash[11],
		hash[12], hash[13], hash[14], hash[15],
		hash[16], hash[17], hash[18], hash[19],
		hash[20], hash[21], hash[22], hash[23],
		hash[24], hash[25], hash[26], hash[27],
	}
	actual := service.StateKeyConstruct()
	flag := false
	for i := 0; i < len(expected); i++ {
		if expected[i] != actual[i] {
			flag = true
		}
	}
	if flag {
		t.Errorf("Expected %v, got %v", expected, actual)
	}
}
