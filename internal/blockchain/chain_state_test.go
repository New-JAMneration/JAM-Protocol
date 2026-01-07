package blockchain_test

import (
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	"github.com/test-go/testify/require"
)

func TestStorePersistsBlocks(t *testing.T) {
	blockchain.ResetInstance()

	store := blockchain.GetInstance()
	block := types.Block{
		Header: types.Header{
			Slot:   123,
			Parent: types.HeaderHash{},
			EpochMark: &types.EpochMark{
				Entropy: types.Entropy{1},
				Validators: []types.EpochMarkValidatorKeys{
					{
						Bandersnatch: types.BandersnatchPublic{1},
						Ed25519:      types.Ed25519Public{1},
					},
					{
						Bandersnatch: types.BandersnatchPublic{2},
						Ed25519:      types.Ed25519Public{2},
					},
					{
						Bandersnatch: types.BandersnatchPublic{3},
						Ed25519:      types.Ed25519Public{3},
					},
					{
						Bandersnatch: types.BandersnatchPublic{4},
						Ed25519:      types.Ed25519Public{4},
					},
					{
						Bandersnatch: types.BandersnatchPublic{5},
						Ed25519:      types.Ed25519Public{5},
					},
					{
						Bandersnatch: types.BandersnatchPublic{6},
						Ed25519:      types.Ed25519Public{6},
					},
				},
			},
		},
	}

	store.AddBlock(block)

	headerHash, err := hash.ComputeBlockHeaderHash(block.Header)
	require.NoError(t, err, "computeBlockHeaderHash should succeed")

	retrieved, err := store.GetBlockByHash(headerHash)
	require.NoError(t, err, "GetBlockByHash should succeed")

	require.Equal(t, block.Header.Slot, retrieved.Header.Slot, "retrieved block should match stored block")
	require.Equal(t, block.Header.Parent, retrieved.Header.Parent, "retrieved block should match stored block")
	require.Equal(t, block.Header.EpochMark, retrieved.Header.EpochMark, "retrieved block should match stored block")
}
