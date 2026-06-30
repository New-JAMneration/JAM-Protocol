package ce

import (
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/stretchr/testify/require"
)

func TestCERequestHandler_BlockRequestRoundtrip(t *testing.T) {
	h := NewDefaultCERequestHandler()
	want := &CE128Payload{
		HeaderHash: types.HeaderHash{1, 2, 3},
		Direction:  0,
		MaxBlocks:  7,
	}

	body, err := h.Encode(BlockRequest, want)
	require.NoError(t, err)

	gotBody, err := h.DecodePayload(BlockRequest, body)
	require.NoError(t, err)
	got, ok := gotBody.(*CE128Payload)
	require.True(t, ok)
	require.Equal(t, want.HeaderHash, got.HeaderHash)
	require.Equal(t, want.Direction, got.Direction)
	require.Equal(t, want.MaxBlocks, got.MaxBlocks)

	full, err := prependRequestID(BlockRequest, body)
	require.NoError(t, err)
	id, decoded, err := h.Decode(full)
	require.NoError(t, err)
	require.Equal(t, BlockRequest, id)
	require.Equal(t, want, decoded)
}

func TestCERequestHandler_StateRequestRoundtrip(t *testing.T) {
	h := NewDefaultCERequestHandler()
	want := &CE129Payload{
		HeaderHash: types.HeaderHash{9},
		KeyStart:   types.StateKey{1},
		KeyEnd:     types.StateKey{2},
		MaxSize:    4096,
	}

	body, err := h.Encode(StateRequest, want)
	require.NoError(t, err)

	gotBody, err := h.DecodePayload(StateRequest, body)
	require.NoError(t, err)
	got, ok := gotBody.(*CE129Payload)
	require.True(t, ok)
	require.Equal(t, want.HeaderHash, got.HeaderHash)
	require.Equal(t, want.KeyStart, got.KeyStart)
	require.Equal(t, want.KeyEnd, got.KeyEnd)
	require.Equal(t, want.MaxSize, got.MaxSize)
}

func TestCERequestHandler_BundleRequestRoundtrip(t *testing.T) {
	h := NewDefaultCERequestHandler()
	root := make([]byte, HashSize)
	for i := range root {
		root[i] = byte(i)
	}
	want := &CE147Payload{ErasureRoot: root}

	body, err := h.Encode(BundleRequest, want)
	require.NoError(t, err)

	gotBody, err := h.DecodePayload(BundleRequest, body)
	require.NoError(t, err)
	got, ok := gotBody.(*CE147Payload)
	require.True(t, ok)
	require.Equal(t, want.ErasureRoot, got.ErasureRoot)
}

func TestCERequestHandler_DecodeRejectsUnknownType(t *testing.T) {
	h := NewDefaultCERequestHandler()
	full, err := prependRequestID(CERequestID(130), []byte{1})
	require.NoError(t, err)
	_, _, err = h.Decode(full)
	require.Error(t, err)
}
