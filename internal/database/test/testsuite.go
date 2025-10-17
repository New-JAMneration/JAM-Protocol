package testsuite

import (
	"bytes"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatabaseSuite(t *testing.T, New func() database.Database) {
	store := New()
	defer store.Close()

	t.Run("BasicOps", func(t *testing.T) {
		key := []byte("key")

		got, err := store.Has(key)
		require.NoError(t, err)
		assert.False(t, got)

		value := []byte("value")
		err = store.Put(key, value)
		require.NoError(t, err)

		got, err = store.Has(key)
		require.NoError(t, err)
		assert.True(t, got)

		gotValue, found, err := store.Get(key)
		require.NoError(t, err)
		assert.True(t, found)
		assert.True(t, bytes.Equal(gotValue, value))

		err = store.Delete(key)
		require.NoError(t, err)

		got, err = store.Has(key)
		require.NoError(t, err)
		assert.False(t, got)
	})

	t.Run("RangeDelete", func(t *testing.T) {
		keys := [][]byte{[]byte("a"), []byte("b"), []byte("c"), []byte("d"), []byte("e")}
		value := []byte("value")

		for _, key := range keys {
			err := store.Put(key, value)
			require.NoError(t, err)
		}

		// expect to delete keys "a", "b", "c"
		// delete [start,end)
		err := store.DeleteRange([]byte("a"), []byte("d"))
		require.NoError(t, err)

		for _, key := range keys {
			has, err := store.Has(key)
			require.NoError(t, err)

			if bytes.Compare(key, []byte("a")) >= 0 && bytes.Compare(key, []byte("d")) < 0 {
				assert.False(t, has, "key %s should be deleted", key)
			} else {
				assert.True(t, has, "key %s should not be deleted", key)
			}
		}
	})
}
