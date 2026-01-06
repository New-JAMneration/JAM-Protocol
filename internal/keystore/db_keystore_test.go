package keystore

import (
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/database/provider/memory"
	"github.com/stretchr/testify/require"
)

// setupTestStorage creates a test storage using memory database
func setupTestStorage(t *testing.T) Storage {
	db := memory.NewDatabase()
	return NewDatabaseStorage(db)
}

func TestDBKeyStore_GenerateAndContains(t *testing.T) {
	storage := setupTestStorage(t)
	ks := NewDBKeyStore(storage)

	kp, err := ks.Generate(KeyTypeEd25519)
	require.NoError(t, err)
	require.NotNil(t, kp)

	ok, err := ks.Contains(KeyTypeEd25519, kp.PublicKey())
	require.NoError(t, err)
	require.True(t, ok)
}

func TestDBKeyStore_ImportAndGet(t *testing.T) {
	storage := setupTestStorage(t)
	ks := NewDBKeyStore(storage)

	seed := make([]byte, 32)
	for i := range seed {
		seed[i] = byte(i)
	}

	kp1, err := ks.Import(KeyTypeEd25519, seed)
	require.NoError(t, err)
	require.NotNil(t, kp1)

	kp2, err := ks.Get(KeyTypeEd25519, kp1.PublicKey())
	require.NoError(t, err)
	require.NotNil(t, kp2)
	require.Equal(t, kp1.PublicKey(), kp2.PublicKey())
	require.Equal(t, kp1.PrivateKey(), kp2.PrivateKey())
}

func TestDBKeyStore_List(t *testing.T) {
	storage := setupTestStorage(t)
	ks := NewDBKeyStore(storage)

	_, err := ks.Generate(KeyTypeEd25519)
	require.NoError(t, err)
	_, err = ks.Generate(KeyTypeEd25519)
	require.NoError(t, err)

	list, err := ks.List(KeyTypeEd25519)
	require.NoError(t, err)
	require.Len(t, list, 2)

	for _, kp := range list {
		require.NotEmpty(t, kp.PublicKey())
		require.NotEmpty(t, kp.PrivateKey())
	}
}

func TestDBKeyStore_GetNotExist(t *testing.T) {
	storage := setupTestStorage(t)
	ks := NewDBKeyStore(storage)

	_, err := ks.Get(KeyTypeEd25519, []byte("nonexistent"))
	require.Error(t, err)
}

func TestDBKeyStore_ContainsFalse(t *testing.T) {
	storage := setupTestStorage(t)
	ks := NewDBKeyStore(storage)

	ok, err := ks.Contains(KeyTypeEd25519, []byte("nonexistent"))
	require.NoError(t, err)
	require.False(t, ok)
}

func TestDBKeyStore_Delete(t *testing.T) {
	storage := setupTestStorage(t)
	ks := NewDBKeyStore(storage)
	kp, err := ks.Generate(KeyTypeEd25519)
	require.NoError(t, err)

	err = ks.Delete(KeyTypeEd25519, kp.PublicKey())
	require.NoError(t, err)

	exist, err := ks.Contains(KeyTypeEd25519, kp.PublicKey())
	require.NoError(t, err)
	require.False(t, exist)
}

func TestGenerateAllKeyTypesWithMemoryStorage(t *testing.T) {
	storage := setupTestStorage(t)
	ks := NewDBKeyStore(storage)

	// Generate Ed25519
	edKP, err := ks.Generate(KeyTypeEd25519)
	require.NoError(t, err)
	require.NotNil(t, edKP)

	// Generate Bandersnatch
	bnKP, err := ks.Generate(KeyTypeBandersnatch)
	require.NoError(t, err)
	require.NotNil(t, bnKP)

	// Contains
	ok, err := ks.Contains(KeyTypeEd25519, edKP.PublicKey())
	require.NoError(t, err)
	require.True(t, ok)

	ok, err = ks.Contains(KeyTypeBandersnatch, bnKP.PublicKey())
	require.NoError(t, err)
	require.True(t, ok)

	// Get
	gotEd, err := ks.Get(KeyTypeEd25519, edKP.PublicKey())
	require.NoError(t, err)
	require.Equal(t, edKP.PublicKey(), gotEd.PublicKey())
	require.Equal(t, edKP.PrivateKey(), gotEd.PrivateKey())

	gotBn, err := ks.Get(KeyTypeBandersnatch, bnKP.PublicKey())
	require.NoError(t, err)
	require.Equal(t, bnKP.PublicKey(), gotBn.PublicKey())
	require.Equal(t, bnKP.PrivateKey(), gotBn.PrivateKey())

	// List (should be 1 each)
	eds, err := ks.List(KeyTypeEd25519)
	require.NoError(t, err)
	require.Len(t, eds, 1)

	bns, err := ks.List(KeyTypeBandersnatch)
	require.NoError(t, err)
	require.Len(t, bns, 1)

	// Delete + check not exists
	err = ks.Delete(KeyTypeEd25519, edKP.PublicKey())
	require.NoError(t, err)
	err = ks.Delete(KeyTypeBandersnatch, bnKP.PublicKey())
	require.NoError(t, err)

	ok, err = ks.Contains(KeyTypeEd25519, edKP.PublicKey())
	require.NoError(t, err)
	require.False(t, ok)

	ok, err = ks.Contains(KeyTypeBandersnatch, bnKP.PublicKey())
	require.NoError(t, err)
	require.False(t, ok)
}

// Verify JIP-5 import and cover main keystore operations
func TestJIP5ImportWithMemoryStorage(t *testing.T) {
	// setup storage & keystore
	db := memory.NewDatabase()
	storage := NewDatabaseStorage(db)
	ks := NewDBKeyStore(storage)
	// 32-byte seed
	seed := make([]byte, 32)
	for i := range seed {
		seed[i] = byte(i)
	}

	// Using JIP-5 to derive Ed25519 and Bandersnatch secret seeds and public keys
	edSeed, bnSeed, edPub, bnPub, err := DeriveValidatorKeys(seed)
	require.NoError(t, err)

	// Import to keystore (Import expects a 32-byte seed, not the full private key)
	edKP, err := ks.Import(KeyTypeEd25519, edSeed)
	require.NoError(t, err)
	require.NotNil(t, edKP)

	bnKP, err := ks.Import(KeyTypeBandersnatch, bnSeed)
	require.NoError(t, err)
	require.NotNil(t, bnKP)

	// Contains
	ok, err := ks.Contains(KeyTypeEd25519, edPub[:])
	require.NoError(t, err)
	require.True(t, ok)

	ok, err = ks.Contains(KeyTypeBandersnatch, bnPub[:])
	require.NoError(t, err)
	require.True(t, ok)

	// Get
	gotEd, err := ks.Get(KeyTypeEd25519, edPub[:])
	require.NoError(t, err)
	require.Equal(t, edKP.PublicKey(), gotEd.PublicKey())
	require.Equal(t, edKP.PrivateKey(), gotEd.PrivateKey())

	gotBn, err := ks.Get(KeyTypeBandersnatch, bnPub[:])
	require.NoError(t, err)
	require.Equal(t, bnKP.PublicKey(), gotBn.PublicKey())
	require.Equal(t, bnKP.PrivateKey(), gotBn.PrivateKey())

	// List (should be 1 each)
	eds, err := ks.List(KeyTypeEd25519)
	require.NoError(t, err)
	require.Len(t, eds, 1)

	bns, err := ks.List(KeyTypeBandersnatch)
	require.NoError(t, err)
	require.Len(t, bns, 1)

	// Delete + check not exists
	err = ks.Delete(KeyTypeEd25519, edPub[:])
	require.NoError(t, err)
	err = ks.Delete(KeyTypeBandersnatch, bnPub[:])
	require.NoError(t, err)

	ok, err = ks.Contains(KeyTypeEd25519, edPub[:])
	require.NoError(t, err)
	require.False(t, ok)

	ok, err = ks.Contains(KeyTypeBandersnatch, bnPub[:])
	require.NoError(t, err)
	require.False(t, ok)
}
