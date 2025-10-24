package store_test

import (
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/database/provider/memory"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	"github.com/stretchr/testify/require"
)

func TestSaveAndGetBlock(t *testing.T) {
	block := types.Block{
		Header:    types.Header{Slot: 1},
		Extrinsic: types.Extrinsic{},
	}

	db := memory.NewDatabase()

	encoder := types.NewEncoder()
	encoded, err := encoder.Encode(&block.Header)
	require.NoError(t, err)
	headerHash := types.HeaderHash(hash.Blake2bHash(encoded))

	require.NoError(t, store.SaveBlock(db, &block))
	readBlock, found, err := store.GetBlock(db, headerHash, 1)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, block, *readBlock)
}

func TestGetNonExistentBlock(t *testing.T) {
	db := memory.NewDatabase()

	headerHash := types.HeaderHash{}
	readBlock, found, err := store.GetBlock(db, headerHash, 1)
	require.NoError(t, err)
	require.False(t, found)
	require.Nil(t, readBlock)
}

func TestDeleteBlock(t *testing.T) {
	block := types.Block{
		Header:    types.Header{Slot: 1},
		Extrinsic: types.Extrinsic{},
	}

	db := memory.NewDatabase()
	encoder := types.NewEncoder()
	encoded, err := encoder.Encode(&block.Header)
	require.NoError(t, err)
	headerHash := types.HeaderHash(hash.Blake2bHash(encoded))

	require.NoError(t, store.SaveBlock(db, &block))
	require.NoError(t, store.DeleteBlock(db, headerHash, 1))
	readBlock, found, err := store.GetBlock(db, headerHash, 1)
	require.NoError(t, err)
	require.False(t, found)
	require.Nil(t, readBlock)
}

func TestDeleteNonExistentBlock(t *testing.T) {
	db := memory.NewDatabase()
	headerHash := types.HeaderHash{}
	require.NoError(t, store.DeleteBlock(db, headerHash, 1))
}
