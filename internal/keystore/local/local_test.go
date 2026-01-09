package local

import (
	"os"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/keystore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLocalKeyStore(t *testing.T) {
	dir := t.TempDir()
	ks, err := NewLocalKeyStore(dir)
	require.NoError(t, err)
	assert.NotNil(t, ks)

	// Verify directory was created
	info, err := os.Stat(dir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestGenerate(t *testing.T) {
	dir := t.TempDir()
	ks, err := NewLocalKeyStore(dir)
	require.NoError(t, err)

	kp, err := ks.Generate(keystore.KeyTypeEd25519)
	require.NoError(t, err)
	assert.NotNil(t, kp)
	assert.Equal(t, keystore.KeyTypeEd25519, kp.Type())
}

func TestImport(t *testing.T) {
	dir := t.TempDir()
	ks, err := NewLocalKeyStore(dir)
	require.NoError(t, err)

	// Create a seed for testing
	seed := make([]byte, 32)
	for i := 0; i < 32; i++ {
		seed[i] = byte(i)
	}

	kp, err := ks.Import(keystore.KeyTypeEd25519, seed)
	require.NoError(t, err)
	assert.NotNil(t, kp)
	assert.Equal(t, keystore.KeyTypeEd25519, kp.Type())
}

func TestContains(t *testing.T) {
	dir := t.TempDir()
	ks, err := NewLocalKeyStore(dir)
	require.NoError(t, err)

	kp, err := ks.Generate(keystore.KeyTypeEd25519)
	require.NoError(t, err)

	exists, err := ks.Contains(keystore.KeyTypeEd25519, kp.PublicKey())
	require.NoError(t, err)
	assert.True(t, exists)

	// Test with non-existent key
	fakeKey := make([]byte, 32)
	exists, err = ks.Contains(keystore.KeyTypeEd25519, fakeKey)
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestGet(t *testing.T) {
	dir := t.TempDir()
	ks, err := NewLocalKeyStore(dir)
	require.NoError(t, err)

	kp1, err := ks.Generate(keystore.KeyTypeEd25519)
	require.NoError(t, err)

	kp2, err := ks.Get(keystore.KeyTypeEd25519, kp1.PublicKey())
	require.NoError(t, err)
	assert.Equal(t, kp1.PublicKey(), kp2.PublicKey())
	assert.Equal(t, kp1.PrivateKey(), kp2.PrivateKey())
}

func TestList(t *testing.T) {
	dir := t.TempDir()
	ks, err := NewLocalKeyStore(dir)
	require.NoError(t, err)

	// Generate multiple keys
	_, err = ks.Generate(keystore.KeyTypeEd25519)
	require.NoError(t, err)
	_, err = ks.Generate(keystore.KeyTypeEd25519)
	require.NoError(t, err)
	_, err = ks.Generate(keystore.KeyTypeEd25519)
	require.NoError(t, err)

	keys, err := ks.List(keystore.KeyTypeEd25519)
	require.NoError(t, err)
	assert.Len(t, keys, 3)
}

func TestDelete(t *testing.T) {
	dir := t.TempDir()
	ks, err := NewLocalKeyStore(dir)
	require.NoError(t, err)

	kp, err := ks.Generate(keystore.KeyTypeEd25519)
	require.NoError(t, err)

	// Verify key exists
	exists, err := ks.Contains(keystore.KeyTypeEd25519, kp.PublicKey())
	require.NoError(t, err)
	assert.True(t, exists)

	// Delete key
	err = ks.Delete(keystore.KeyTypeEd25519, kp.PublicKey())
	require.NoError(t, err)

	// Verify key is deleted
	exists, err = ks.Contains(keystore.KeyTypeEd25519, kp.PublicKey())
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestPersistence(t *testing.T) {
	dir := t.TempDir()
	var pubKey []byte

	// Create keystore and generate key
	{
		ks, err := NewLocalKeyStore(dir)
		require.NoError(t, err)

		kp, err := ks.Generate(keystore.KeyTypeEd25519)
		require.NoError(t, err)
		pubKey = kp.PublicKey()
	}

	// Create new keystore instance and verify key exists
	{
		ks, err := NewLocalKeyStore(dir)
		require.NoError(t, err)

		exists, err := ks.Contains(keystore.KeyTypeEd25519, pubKey)
		require.NoError(t, err)
		assert.True(t, exists)

		kp, err := ks.Get(keystore.KeyTypeEd25519, pubKey)
		require.NoError(t, err)
		assert.Equal(t, pubKey, kp.PublicKey())
	}
}

func TestSign(t *testing.T) {
	dir := t.TempDir()
	ks, err := NewLocalKeyStore(dir)
	require.NoError(t, err)

	kp, err := ks.Generate(keystore.KeyTypeEd25519)
	require.NoError(t, err)

	message := []byte("test message")
	signature, err := kp.Sign(message)
	require.NoError(t, err)
	assert.NotNil(t, signature)

	// Verify signature
	verified := kp.Verify(message, signature)
	assert.True(t, verified)
}

func TestGenerateBandersnatch(t *testing.T) {
	dir := t.TempDir()
	ks, err := NewLocalKeyStore(dir)
	require.NoError(t, err)

	kp, err := ks.Generate(keystore.KeyTypeBandersnatch)
	require.NoError(t, err)
	assert.NotNil(t, kp)
	assert.Equal(t, keystore.KeyTypeBandersnatch, kp.Type())
}

func TestImportBandersnatch(t *testing.T) {
	dir := t.TempDir()
	ks, err := NewLocalKeyStore(dir)
	require.NoError(t, err)

	// Create a seed for testing
	seed := make([]byte, 32)
	for i := 0; i < 32; i++ {
		seed[i] = byte(i)
	}

	kp, err := ks.Import(keystore.KeyTypeBandersnatch, seed)
	require.NoError(t, err)
	assert.NotNil(t, kp)
	assert.Equal(t, keystore.KeyTypeBandersnatch, kp.Type())
}

func TestGenerateAllKeyTypes(t *testing.T) {
	dir := t.TempDir()
	ks, err := NewLocalKeyStore(dir)
	require.NoError(t, err)

	// Generate Ed25519
	edKP, err := ks.Generate(keystore.KeyTypeEd25519)
	require.NoError(t, err)
	require.NotNil(t, edKP)

	// Generate Bandersnatch
	bnKP, err := ks.Generate(keystore.KeyTypeBandersnatch)
	require.NoError(t, err)
	require.NotNil(t, bnKP)

	// Contains
	ok, err := ks.Contains(keystore.KeyTypeEd25519, edKP.PublicKey())
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = ks.Contains(keystore.KeyTypeBandersnatch, bnKP.PublicKey())
	require.NoError(t, err)
	assert.True(t, ok)

	// Get
	gotEd, err := ks.Get(keystore.KeyTypeEd25519, edKP.PublicKey())
	require.NoError(t, err)
	assert.Equal(t, edKP.PublicKey(), gotEd.PublicKey())
	assert.Equal(t, edKP.PrivateKey(), gotEd.PrivateKey())

	gotBn, err := ks.Get(keystore.KeyTypeBandersnatch, bnKP.PublicKey())
	require.NoError(t, err)
	assert.Equal(t, bnKP.PublicKey(), gotBn.PublicKey())
	assert.Equal(t, bnKP.PrivateKey(), gotBn.PrivateKey())

	// List (should be 1 each)
	eds, err := ks.List(keystore.KeyTypeEd25519)
	require.NoError(t, err)
	assert.Len(t, eds, 1)

	bns, err := ks.List(keystore.KeyTypeBandersnatch)
	require.NoError(t, err)
	assert.Len(t, bns, 1)

	// Delete + check not exists
	err = ks.Delete(keystore.KeyTypeEd25519, edKP.PublicKey())
	require.NoError(t, err)
	err = ks.Delete(keystore.KeyTypeBandersnatch, bnKP.PublicKey())
	require.NoError(t, err)

	ok, err = ks.Contains(keystore.KeyTypeEd25519, edKP.PublicKey())
	require.NoError(t, err)
	assert.False(t, ok)

	ok, err = ks.Contains(keystore.KeyTypeBandersnatch, bnKP.PublicKey())
	require.NoError(t, err)
	assert.False(t, ok)
}

// TestJIP5ImportWithLocalKeyStore verifies JIP-5 import and covers main keystore operations
func TestJIP5ImportWithLocalKeyStore(t *testing.T) {
	dir := t.TempDir()
	ks, err := NewLocalKeyStore(dir)
	require.NoError(t, err)

	// 32-byte seed
	seed := make([]byte, 32)
	for i := range seed {
		seed[i] = byte(i)
	}

	// Using JIP-5 to derive Ed25519 and Bandersnatch secret seeds and public keys
	edSeed, bnSeed, edPub, bnPub, err := keystore.DeriveValidatorKeys(seed)
	require.NoError(t, err)

	// Import to keystore (Import expects a 32-byte seed, not the full private key)
	edKP, err := ks.Import(keystore.KeyTypeEd25519, edSeed)
	require.NoError(t, err)
	require.NotNil(t, edKP)

	bnKP, err := ks.Import(keystore.KeyTypeBandersnatch, bnSeed)
	require.NoError(t, err)
	require.NotNil(t, bnKP)

	// Contains
	ok, err := ks.Contains(keystore.KeyTypeEd25519, edPub[:])
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = ks.Contains(keystore.KeyTypeBandersnatch, bnPub[:])
	require.NoError(t, err)
	assert.True(t, ok)

	// Get
	gotEd, err := ks.Get(keystore.KeyTypeEd25519, edPub[:])
	require.NoError(t, err)
	assert.Equal(t, edKP.PublicKey(), gotEd.PublicKey())
	assert.Equal(t, edKP.PrivateKey(), gotEd.PrivateKey())

	gotBn, err := ks.Get(keystore.KeyTypeBandersnatch, bnPub[:])
	require.NoError(t, err)
	assert.Equal(t, bnKP.PublicKey(), gotBn.PublicKey())
	assert.Equal(t, bnKP.PrivateKey(), gotBn.PrivateKey())

	// List (should be 1 each)
	eds, err := ks.List(keystore.KeyTypeEd25519)
	require.NoError(t, err)
	assert.Len(t, eds, 1)

	bns, err := ks.List(keystore.KeyTypeBandersnatch)
	require.NoError(t, err)
	assert.Len(t, bns, 1)

	// Delete + check not exists
	err = ks.Delete(keystore.KeyTypeEd25519, edPub[:])
	require.NoError(t, err)
	err = ks.Delete(keystore.KeyTypeBandersnatch, bnPub[:])
	require.NoError(t, err)

	ok, err = ks.Contains(keystore.KeyTypeEd25519, edPub[:])
	require.NoError(t, err)
	assert.False(t, ok)

	ok, err = ks.Contains(keystore.KeyTypeBandersnatch, bnPub[:])
	require.NoError(t, err)
	assert.False(t, ok)
}

// TestImportValidatorKeysFromSeed tests the convenience function for importing JIP-5 keys
func TestImportValidatorKeysFromSeed(t *testing.T) {
	dir := t.TempDir()
	ks, err := NewLocalKeyStore(dir)
	require.NoError(t, err)

	// 32-byte seed
	seed := make([]byte, 32)
	for i := range seed {
		seed[i] = byte(i)
	}

	// Use the convenience function to import both keys
	err = ks.ImportValidatorKeysFromSeed(seed)
	require.NoError(t, err)

	// Verify keys were imported by deriving and checking
	_, _, edPub, bnPub, err := keystore.DeriveValidatorKeys(seed)
	require.NoError(t, err)

	// Check Ed25519
	ok, err := ks.Contains(keystore.KeyTypeEd25519, edPub[:])
	require.NoError(t, err)
	assert.True(t, ok)

	// Check Bandersnatch
	ok, err = ks.Contains(keystore.KeyTypeBandersnatch, bnPub[:])
	require.NoError(t, err)
	assert.True(t, ok)

	// Verify we can get the keys and use them for signing
	edKP, err := ks.Get(keystore.KeyTypeEd25519, edPub[:])
	require.NoError(t, err)

	message := []byte("test message")
	signature, err := edKP.Sign(message)
	require.NoError(t, err)
	assert.True(t, edKP.Verify(message, signature))
}
