package node

import (
	"errors"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/stretchr/testify/require"
)

func TestImportBlock_parentMismatch(t *testing.T) {
	blockchain.ResetInstance()
	chain := blockchain.GetInstance()
	genesis := types.Block{Header: types.Header{Slot: 0}, Extrinsic: types.Extrinsic{}}
	require.NoError(t, chain.GenerateGenesisBlock(genesis))
	t.Cleanup(func() { blockchain.ResetInstance() })

	_, err := ImportBlock(chain, types.Block{
		Header: types.Header{
			Slot:   1,
			Parent: types.HeaderHash{0xff},
		},
	})
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrParentMismatch))
}
