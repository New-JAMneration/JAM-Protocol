package blockchain_test

import (
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	"github.com/stretchr/testify/assert"
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

func tearDown() {
	blockchain.ResetInstance()
}

func TestGetInstanceSuccess(t *testing.T) {
	defer tearDown()
	cs := blockchain.GetInstance()
	assert.NotNil(t, cs, "should not fail calling GetInstance")
}

func TestGetInstanceIdempotent(t *testing.T) {
	defer tearDown()

	cs1 := blockchain.GetInstance()
	assert.NotNil(t, cs1, "should not fail calling GetInstance")

	cs2 := blockchain.GetInstance()
	assert.NotNil(t, cs2, "should not fail calling GetInstance again")

	assert.Equal(t, cs1, cs2, "expected the same ChainState object due to sync.Once")
}

func TestStorePersistsBlocksInPersistent(t *testing.T) {
	defer tearDown()

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
