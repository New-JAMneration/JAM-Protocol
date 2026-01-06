package store_test

import (
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/database"
	"github.com/New-JAMneration/JAM-Protocol/internal/database/provider/memory"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	"github.com/stretchr/testify/require"
)

var (
	encoder = types.NewEncoder()
)

func TestSaveAndGetBlock(t *testing.T) {
	block := &types.Block{
		Header:    types.Header{Slot: 1},
		Extrinsic: types.Extrinsic{},
	}

	db := memory.NewDatabase()
	repo := store.NewRepository(db)

	encoded, err := encoder.Encode(&block.Header)
	require.NoError(t, err)
	headerHash := types.HeaderHash(hash.Blake2bHash(encoded))

	require.NoError(t, repo.SaveBlock(db, block))
	readBlock, err := repo.GetBlock(db, headerHash, 1)
	require.NoError(t, err)

	// require.Equal(t, block.Header, readBlock.Header)
	require.Equal(t, block.Header.Parent, readBlock.Header.Parent)
	require.Equal(t, block.Extrinsic, readBlock.Extrinsic)
}

func TestMultipleBlocks(t *testing.T) {
	block1 := types.Block{
		Header:    types.Header{Slot: 1},
		Extrinsic: types.Extrinsic{},
	}
	block2 := types.Block{
		Header:    types.Header{Slot: 2},
		Extrinsic: types.Extrinsic{},
	}

	encoded1, err := encoder.Encode(&block1.Header)
	require.NoError(t, err)
	headerHash1 := types.HeaderHash(hash.Blake2bHash(encoded1))

	encoded2, err := encoder.Encode(&block2.Header)
	require.NoError(t, err)
	headerHash2 := types.HeaderHash(hash.Blake2bHash(encoded2))

	db := memory.NewDatabase()
	repo := store.NewRepository(db)

	err = repo.WithBatch(func(batch database.Batch) error {
		require.NoError(t, repo.SaveBlock(batch, &block1))
		require.NoError(t, repo.SaveBlock(batch, &block2))
		return nil
	})
	require.NoError(t, err)

	readBlock1, err := repo.GetBlock(db, headerHash1, 1)
	require.NoError(t, err)
	require.Equal(t, block1, *readBlock1)

	readBlock2, err := repo.GetBlock(db, headerHash2, 2)
	require.NoError(t, err)
	require.Equal(t, block2, *readBlock2)
}

func TestGetNonExistentBlock(t *testing.T) {
	db := memory.NewDatabase()
	repo := store.NewRepository(db)

	headerHash := types.HeaderHash{}
	readBlock, err := repo.GetBlock(db, headerHash, 1)
	require.Error(t, err)
	require.Nil(t, readBlock)
}

func TestDeleteBlock(t *testing.T) {
	block := types.Block{
		Header:    types.Header{Slot: 1},
		Extrinsic: types.Extrinsic{},
	}

	db := memory.NewDatabase()
	repo := store.NewRepository(db)

	encoded, err := encoder.Encode(&block.Header)
	require.NoError(t, err)
	headerHash := types.HeaderHash(hash.Blake2bHash(encoded))

	require.NoError(t, repo.SaveBlock(db, &block))
	require.NoError(t, repo.DeleteBlock(db, headerHash, 1))
	readBlock, err := repo.GetBlock(db, headerHash, 1)
	require.Error(t, err)
	require.Nil(t, readBlock)
}

func TestDeleteNonExistentBlock(t *testing.T) {
	db := memory.NewDatabase()
	repo := store.NewRepository(db)

	headerHash := types.HeaderHash{}
	require.NoError(t, repo.DeleteBlock(db, headerHash, 1))
}
