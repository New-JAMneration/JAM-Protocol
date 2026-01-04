package repository_test

import (
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/database"
	"github.com/New-JAMneration/JAM-Protocol/internal/database/provider/memory"
	"github.com/New-JAMneration/JAM-Protocol/internal/store/repository"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/stretchr/testify/require"
)

func TestSaveAndGetStateRootByHeaderHash(t *testing.T) {
	db := memory.NewDatabase()
	repo := repository.NewRepository(db)

	headerHash := types.HeaderHash([32]byte{1, 2, 3, 4, 5})
	stateRoot := types.StateRoot([32]byte{10, 20, 30, 40, 50})

	require.NoError(t, repo.SaveStateRootByHeaderHash(db, headerHash, stateRoot))
	retrievedStateRoot, err := repo.GetStateRootByHeaderHash(db, headerHash)
	require.NoError(t, err)
	require.Equal(t, stateRoot, retrievedStateRoot)
}

func TestGetNonExistentStateRoot(t *testing.T) {
	db := memory.NewDatabase()
	repo := repository.NewRepository(db)

	headerHash := types.HeaderHash([32]byte{1, 2, 3, 4, 5})
	_, err := repo.GetStateRootByHeaderHash(db, headerHash)
	require.Error(t, err)
}

func TestSaveAndGetStateData(t *testing.T) {
	db := memory.NewDatabase()
	repo := repository.NewRepository(db)

	stateRoot := types.StateRoot([32]byte{10, 20, 30, 40, 50})
	stateData := types.StateKeyVals{
		{
			Key:   types.StateKey([31]byte{1, 2, 3}),
			Value: []byte("value1"),
		},
		{
			Key:   types.StateKey([31]byte{4, 5, 6}),
			Value: []byte("value2"),
		},
	}

	require.NoError(t, repo.SaveStateData(db, stateRoot, stateData))
	retrievedStateData, err := repo.GetStateData(db, stateRoot)
	require.NoError(t, err)
	require.Equal(t, stateData, retrievedStateData)
}

func TestGetNonExistentStateData(t *testing.T) {
	db := memory.NewDatabase()
	repo := repository.NewRepository(db)

	stateRoot := types.StateRoot([32]byte{10, 20, 30, 40, 50})
	_, err := repo.GetStateData(db, stateRoot)
	require.Error(t, err)
}

func TestGetStateDataByHeaderHash(t *testing.T) {
	db := memory.NewDatabase()
	repo := repository.NewRepository(db)

	headerHash := types.HeaderHash([32]byte{1, 2, 3, 4, 5})
	stateRoot := types.StateRoot([32]byte{10, 20, 30, 40, 50})
	stateData := types.StateKeyVals{
		{
			Key:   types.StateKey([31]byte{7, 8, 9}),
			Value: []byte("test value"),
		},
	}

	require.NoError(t, repo.SaveStateRootByHeaderHash(db, headerHash, stateRoot))
	require.NoError(t, repo.SaveStateData(db, stateRoot, stateData))

	retrievedStateData, err := repo.GetStateDataByHeaderHash(db, headerHash)
	require.NoError(t, err)
	require.Equal(t, stateData, retrievedStateData)
}

func TestGetStateDataByNonExistentHeaderHash(t *testing.T) {
	db := memory.NewDatabase()
	repo := repository.NewRepository(db)

	headerHash := types.HeaderHash([32]byte{1, 2, 3, 4, 5})
	_, err := repo.GetStateDataByHeaderHash(db, headerHash)
	require.Error(t, err)
}

func TestMultipleStateRoots(t *testing.T) {
	db := memory.NewDatabase()
	repo := repository.NewRepository(db)

	headerHash1 := types.HeaderHash([32]byte{1, 1, 1, 1, 1})
	stateRoot1 := types.StateRoot([32]byte{10, 10, 10, 10, 10})

	headerHash2 := types.HeaderHash([32]byte{2, 2, 2, 2, 2})
	stateRoot2 := types.StateRoot([32]byte{20, 20, 20, 20, 20})

	err := repo.WithBatch(func(batch database.Batch) error {
		err := repo.SaveStateRootByHeaderHash(db, headerHash1, stateRoot1)
		if err != nil {
			return err
		}
		err = repo.SaveStateRootByHeaderHash(db, headerHash2, stateRoot2)
		if err != nil {
			return err
		}
		return nil
	})
	require.NoError(t, err)

	retrieved1, err := repo.GetStateRootByHeaderHash(db, headerHash1)
	require.NoError(t, err)
	require.Equal(t, stateRoot1, retrieved1)

	retrieved2, err := repo.GetStateRootByHeaderHash(db, headerHash2)
	require.NoError(t, err)
	require.Equal(t, stateRoot2, retrieved2)
}

func TestSaveStateDataWithEmptyStateKeyVals(t *testing.T) {
	db := memory.NewDatabase()
	repo := repository.NewRepository(db)

	stateRoot := types.StateRoot([32]byte{10, 20, 30, 40, 50})
	stateData := types.StateKeyVals{}

	require.NoError(t, repo.SaveStateData(db, stateRoot, stateData))
	retrievedStateData, err := repo.GetStateData(db, stateRoot)
	require.NoError(t, err)
	require.True(t, len(retrievedStateData) == 0)
}

func TestSaveStateDataWithMultipleKeyVals(t *testing.T) {
	db := memory.NewDatabase()
	repo := repository.NewRepository(db)

	stateRoot := types.StateRoot([32]byte{30, 40, 50, 60, 70})
	stateData := types.StateKeyVals{
		{
			Key:   types.StateKey([31]byte{1}),
			Value: []byte("value one"),
		},
		{
			Key:   types.StateKey([31]byte{2}),
			Value: []byte("value two"),
		},
		{
			Key:   types.StateKey([31]byte{3}),
			Value: []byte{255, 254, 253, 252},
		},
	}

	require.NoError(t, repo.SaveStateData(db, stateRoot, stateData))
	retrievedStateData, err := repo.GetStateData(db, stateRoot)
	require.NoError(t, err)
	require.Equal(t, len(stateData), len(retrievedStateData))
	for i := range stateData {
		require.Equal(t, stateData[i].Key, retrievedStateData[i].Key)
		require.Equal(t, stateData[i].Value, retrievedStateData[i].Value)
	}
}
