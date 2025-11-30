package repository_test

import (
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/database"
	"github.com/New-JAMneration/JAM-Protocol/internal/database/provider/memory"
	"github.com/New-JAMneration/JAM-Protocol/internal/store/repository"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	"github.com/test-go/testify/require"
)

func TestSaveAndGetCanonicalHash(t *testing.T) {
	db := memory.NewDatabase()
	repo := repository.NewRepository(db)

	slot := types.TimeSlot(1)
	headerHash := types.HeaderHash{0x01, 0x02, 0x03}

	err := repo.SaveCanonicalHash(db, headerHash, slot)
	require.NoError(t, err)

	readHash, err := repo.GetCanonicalHash(db, slot)
	require.NoError(t, err)
	require.Equal(t, headerHash, readHash)
}

func TestSaveAndGetHeader(t *testing.T) {
	db := memory.NewDatabase()
	repo := repository.NewRepository(db)

	header := &types.Header{
		Slot: 1,
	}

	encoded, err := encoder.Encode(header)
	require.NoError(t, err)
	headerHash := types.HeaderHash(hash.Blake2bHash(encoded))

	_, err = repo.SaveHeader(db, header)
	require.NoError(t, err)

	readHeader, err := repo.GetHeader(db, headerHash, header.Slot)
	require.NoError(t, err)
	require.Equal(t, header, readHeader)
}

func TestDeleteHeader(t *testing.T) {
	db := memory.NewDatabase()
	repo := repository.NewRepository(db)

	header := &types.Header{
		Slot: 1,
	}
	encoded, err := encoder.Encode(header)
	require.NoError(t, err)
	headerHash := types.HeaderHash(hash.Blake2bHash(encoded))

	_, err = repo.SaveHeader(db, header)
	require.NoError(t, err)

	err = repo.DeleteHeader(db, headerHash, header.Slot)
	require.NoError(t, err)

	readHeader, err := repo.GetHeader(db, headerHash, header.Slot)
	require.Error(t, err)
	require.Nil(t, readHeader)
}

func TestGetHeaderTimeSlot(t *testing.T) {
	db := memory.NewDatabase()
	repo := repository.NewRepository(db)

	header := &types.Header{
		Slot: 1,
	}
	encoded, err := encoder.Encode(header)
	require.NoError(t, err)
	headerHash := types.HeaderHash(hash.Blake2bHash(encoded))

	_, err = repo.SaveHeader(db, header)
	require.NoError(t, err)

	slot, err := repo.GetHeaderTimeSlot(db, headerHash)
	require.NoError(t, err)
	require.Equal(t, header.Slot, slot)
}

func TestGetHeaderHashesByTimeSlot(t *testing.T) {
	db := memory.NewDatabase()
	repo := repository.NewRepository(db)

	slot := types.TimeSlot(1)
	header1 := &types.Header{
		Slot:   slot,
		Parent: [32]byte{0x00},
	}
	header2 := &types.Header{
		Slot:   slot,
		Parent: [32]byte{0x01},
	}

	slot2 := types.TimeSlot(11)
	header3 := &types.Header{
		Slot:   slot2,
		Parent: [32]byte{0x02},
	}

	err := repo.WithBatch(func(batch database.Batch) error {
		_, err := repo.SaveHeader(db, header1)
		if err != nil {
			return err
		}
		_, err = repo.SaveHeader(db, header2)
		if err != nil {
			return err
		}
		_, err = repo.SaveHeader(db, header3)
		if err != nil {
			return err
		}
		return nil
	})
	require.NoError(t, err)

	encoded1, err := encoder.Encode(header1)
	require.NoError(t, err)
	headerHash1 := types.HeaderHash(hash.Blake2bHash(encoded1))

	encoded2, err := encoder.Encode(header2)
	require.NoError(t, err)
	headerHash2 := types.HeaderHash(hash.Blake2bHash(encoded2))

	// Ensure only headers for the specified slot are returned
	hashes, err := repo.GetHeaderHashesByTimeSlot(db, slot)
	require.NoError(t, err)
	require.Len(t, hashes, 2)
	require.Contains(t, hashes, headerHash1)
	require.Contains(t, hashes, headerHash2)
}
