package repository_test

// import (
// "testing"

// "github.com/New-JAMneration/JAM-Protocol/internal/database/provider/memory"
// "github.com/New-JAMneration/JAM-Protocol/internal/store/store"
// "github.com/New-JAMneration/JAM-Protocol/internal/types"
// "github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
// "github.com/test-go/testify/require"
// )

// func TestSaveAndGetCanonicalHash(t *testing.T) {
// db := memory.NewDatabase()

// slot := types.TimeSlot(1)
// headerHash := types.HeaderHash{0x01, 0x02, 0x03}

// err := store.SaveCanonicalHash(db, headerHash, slot)
// require.NoError(t, err)

// readHash, found, err := store.GetCanonicalHash(db, slot)
// require.NoError(t, err)
// require.True(t, found)
// require.Equal(t, headerHash, readHash)
// }

// func TestSaveAndGetHeader(t *testing.T) {
// db := memory.NewDatabase()

// header := &types.Header{
// Slot: 1,
// }

// encoder := types.NewEncoder()
// encoded, err := encoder.Encode(header)
// require.NoError(t, err)
// headerHash := types.HeaderHash(hash.Blake2bHash(encoded))

// err = store.SaveHeader(db, header)
// require.NoError(t, err)

// readHeader, found, err := store.GetHeader(db, headerHash, header.Slot)
// require.NoError(t, err)
// require.True(t, found)
// require.Equal(t, header, readHeader)
// }

// func TestDeleteHeader(t *testing.T) {
// db := memory.NewDatabase()

// header := &types.Header{
// Slot: 1,
// }
// encoder := types.NewEncoder()
// encoded, err := encoder.Encode(header)
// require.NoError(t, err)
// headerHash := types.HeaderHash(hash.Blake2bHash(encoded))

// err = store.SaveHeader(db, header)
// require.NoError(t, err)

// err = store.DeleteHeader(db, headerHash, header.Slot)
// require.NoError(t, err)

// readHeader, found, err := store.GetHeader(db, headerHash, header.Slot)
// require.NoError(t, err)
// require.False(t, found)
// require.Nil(t, readHeader)
// }

// func TestGetHeaderTimeSlot(t *testing.T) {
// db := memory.NewDatabase()

// header := &types.Header{
// Slot: 1,
// }
// encoder := types.NewEncoder()
// encoded, err := encoder.Encode(header)
// require.NoError(t, err)
// headerHash := types.HeaderHash(hash.Blake2bHash(encoded))

// err = store.SaveHeader(db, header)
// require.NoError(t, err)

// slot, found, err := store.GetHeaderTimeSlot(db, headerHash)
// require.NoError(t, err)
// require.True(t, found)
// require.Equal(t, header.Slot, slot)
// }

// func TestGetHeaderHashesByTimeSlot(t *testing.T) {
// db := memory.NewDatabase()

// slot := types.TimeSlot(1)
// header1 := &types.Header{
// Slot:   slot,
// Parent: [32]byte{0x00},
// }
// header2 := &types.Header{
// Slot:   slot,
// Parent: [32]byte{0x01},
// }

// encoder := types.NewEncoder()
// encoded1, err := encoder.Encode(header1)
// require.NoError(t, err)
// headerHash1 := types.HeaderHash(hash.Blake2bHash(encoded1))

// encoded2, err := encoder.Encode(header2)
// require.NoError(t, err)
// headerHash2 := types.HeaderHash(hash.Blake2bHash(encoded2))

// err = store.SaveHeader(db, header1)
// require.NoError(t, err)
// err = store.SaveHeader(db, header2)
// require.NoError(t, err)

// hashes, err := store.GetHeaderHashesByTimeSlot(db, slot)
// require.NoError(t, err)
// require.Len(t, hashes, 2)
// require.Contains(t, hashes, headerHash1)
// require.Contains(t, hashes, headerHash2)
// }
