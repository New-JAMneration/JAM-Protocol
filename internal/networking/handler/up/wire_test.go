package up

import (
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/stretchr/testify/require"
)

func TestHandshakeRoundTrip(t *testing.T) {
	var finalHash, leafHash types.HeaderHash
	finalHash[0] = 0xfa
	leafHash[0] = 0x1e

	hs := Handshake{
		Final: BlockRef{Hash: finalHash, Slot: 100},
		Leaves: []BlockRef{
			{Hash: leafHash, Slot: 102},
		},
	}

	encoded, err := EncodeHandshake(hs)
	require.NoError(t, err)

	decoded, err := DecodeHandshake(encoded)
	require.NoError(t, err)
	require.Equal(t, hs, decoded)
}

func TestAnnouncementRoundTrip(t *testing.T) {
	var finalHash types.HeaderHash
	finalHash[1] = 0xbb

	a := Announcement{
		Header: types.Header{Slot: 42, Parent: types.HeaderHash{3: 1}},
		Final:  BlockRef{Hash: finalHash, Slot: 40},
	}

	encoded, err := EncodeAnnouncement(a)
	require.NoError(t, err)

	decoded, err := DecodeAnnouncement(encoded)
	require.NoError(t, err)
	require.Equal(t, a.Header.Slot, decoded.Header.Slot)
	require.Equal(t, a.Header.Parent, decoded.Header.Parent)
	require.Equal(t, a.Final, decoded.Final)
}
