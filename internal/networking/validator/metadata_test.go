package validator

import (
	"net"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/stretchr/testify/require"
)

func TestEncodeUDPMetadata_loopback(t *testing.T) {
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:10101")
	require.NoError(t, err)

	meta, err := EncodeUDPMetadata(addr)
	require.NoError(t, err)

	got, err := PeerAddressFromMetadata(meta)
	require.NoError(t, err)
	require.Equal(t, "127.0.0.1", got.IP.String())
	require.Equal(t, 10101, got.Port)

	var pub types.Ed25519Public
	pub[0] = 1
	v, err := ValidatorWithUDPAddr(pub, addr)
	require.NoError(t, err)
	require.Equal(t, pub, v.Ed25519)
}
