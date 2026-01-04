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
