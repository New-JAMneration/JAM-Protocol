package validator

import (
	"encoding/binary"
	"fmt"
	"net"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// EncodeUDPMetadata packs an IPv6-mapped or native IPv6 address and UDP port into validator metadata.
func EncodeUDPMetadata(addr *net.UDPAddr) (types.ValidatorMetadata, error) {
	var meta types.ValidatorMetadata
	if addr == nil {
		return meta, fmt.Errorf("nil udp address")
	}
	ip6 := addr.IP.To16()
	if ip6 == nil {
		return meta, fmt.Errorf("invalid ip for metadata: %s", addr.IP)
	}
	copy(meta[0:16], ip6)
	binary.LittleEndian.PutUint16(meta[16:18], uint16(addr.Port))
	return meta, nil
}

// ValidatorWithUDPAddr builds a validator entry with QUIC listen metadata.
func ValidatorWithUDPAddr(pub types.Ed25519Public, addr *net.UDPAddr) (types.Validator, error) {
	meta, err := EncodeUDPMetadata(addr)
	if err != nil {
		return types.Validator{}, err
	}
	return types.Validator{Ed25519: pub, Metadata: meta}, nil
}
