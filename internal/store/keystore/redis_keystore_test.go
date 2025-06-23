package keystore

import (
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/require"
)

func setupTestRedis(t *testing.T) *store.RedisClient {
	s := miniredis.RunT(t)
	rdb := store.NewRedisClient(s.Addr(), "", 0)
	err := rdb.Ping()
	require.NoError(t, err)
	return rdb
}

func TestRedisKeyStore_GenerateAndContains(t *testing.T) {
	rdb := setupTestRedis(t)
	ks := NewRedisKeyStore(rdb)

	kp, err := ks.Generate(KeyTypeEd25519)
	require.NoError(t, err)
	require.NotNil(t, kp)

	ok, err := ks.Contains(KeyTypeEd25519, kp.PublicKey())
	require.NoError(t, err)
	require.True(t, ok)
}

func TestRedisKeyStore_ImportAndGet(t *testing.T) {
	rdb := setupTestRedis(t)
	ks := NewRedisKeyStore(rdb)

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

func TestRedisKeyStore_List(t *testing.T) {
	rdb := setupTestRedis(t)
	ks := NewRedisKeyStore(rdb)

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

func TestRedisKeyStore_GetNotExist(t *testing.T) {
	rdb := setupTestRedis(t)
	ks := NewRedisKeyStore(rdb)

	_, err := ks.Get(KeyTypeEd25519, []byte("nonexistent"))
	require.Error(t, err)
}

func TestRedisKeyStore_ContainsFalse(t *testing.T) {
	rdb := setupTestRedis(t)
	ks := NewRedisKeyStore(rdb)

	ok, err := ks.Contains(KeyTypeEd25519, []byte("nonexistent"))
	require.NoError(t, err)
	require.False(t, ok)
}

func TestRedisKeyStore_Delete(t *testing.T) {
	rdb := setupTestRedis(t)
	ks := NewRedisKeyStore(rdb)
	kp, err := ks.Generate(KeyTypeEd25519)
	require.NoError(t, err)

	err = ks.Delete(KeyTypeEd25519, kp.PublicKey())
	require.NoError(t, err)

	exist, err := ks.Contains(KeyTypeEd25519, kp.PublicKey())
	require.NoError(t, err)
	require.False(t, exist)
}
