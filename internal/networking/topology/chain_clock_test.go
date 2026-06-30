package topology

import (
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/stretchr/testify/require"
)

func TestReadChainTimeslots_usesHeadAndFinalized(t *testing.T) {
	blockchain.ResetInstance()
	chain := blockchain.GetInstance()
	genesis := types.Block{Header: types.Header{Slot: 0}, Extrinsic: types.Extrinsic{}}
	require.NoError(t, chain.GenerateGenesisBlock(genesis))
	t.Cleanup(func() { blockchain.ResetInstance() })

	ts := readChainTimeslots(chain)
	require.Equal(t, types.TimeSlot(0), ts.BestHead)
	require.Equal(t, types.TimeSlot(0), ts.Finalized)

	chain.AddBlock(types.Block{
		Header: types.Header{Slot: 5, Parent: chain.GenesisBlockHash()},
	})
	ts = readChainTimeslots(chain)
	require.Equal(t, types.TimeSlot(5), ts.BestHead)
}
