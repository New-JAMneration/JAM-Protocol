package up

import (
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/stretchr/testify/require"
)

func TestBuildHandshake(t *testing.T) {
	blocks, finalized := testChain(t)
	hs, err := BuildHandshake(blocks, finalized)
	require.NoError(t, err)
	require.Equal(t, types.TimeSlot(10), hs.Final.Slot)
	require.Len(t, hs.Leaves, 2)
}
