package validator

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type ValidatorManager struct {
	Grid      *GridMapper
	SelfIndex int
	SelfKey   types.Ed25519Public
}

func (vm *ValidatorManager) GetNeighbors() []types.Validator {
	if vm == nil || vm.Grid == nil {
		return nil
	}
	return vm.Grid.AllNeighborValidators(vm.SelfIndex)
}

func (vm *ValidatorManager) IsNeighbor(key types.Ed25519Public) bool {
	if vm == nil || vm.Grid == nil {
		return false
	}

	if peerIdx, ok := vm.Grid.FindIndex(key); ok {
		return vm.Grid.IsNeighborInEpoch(vm.SelfIndex, peerIdx)
	}
	// Not in current set: may still be a grid neighbour at the same index in Previous/Next epoch.
	return vm.Grid.IsSameIndexCrossEpoch(vm.SelfIndex, key)
}

// P(a, b) \in a, b
/*
	P(a, b) = a  when (a[31] > 127) XOR (b[31] > 127) XOR (a < b)
	P(a, b) = b  otherwise
*/
func PreferredInitiator(a, b types.Ed25519Public) types.Ed25519Public {
	aHighBit := a[31] > 127
	bHighBit := b[31] > 127
	aLessThanB := bytes.Compare(a[:], b[:]) < 0

	if (aHighBit != bHighBit) != aLessThanB {
		return a
	}
	return b
}

func PeerAddressFromMetadata(meta types.ValidatorMetadata) (*net.UDPAddr, error) {
	ip := net.IP(meta[0:16]).To16()
	if ip == nil {
		return nil, fmt.Errorf("invalid ipv6 address in validator metadata")
	}
	port := int(binary.LittleEndian.Uint16(meta[16:18]))
	return &net.UDPAddr{IP: ip, Port: port}, nil
}
