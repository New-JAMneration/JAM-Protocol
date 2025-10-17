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

	t.Run("BatchWrite", func(t *testing.T) {
		batch := store.NewBatch()

		// Add multiple operations to the batch
		keys := [][]byte{[]byte("key1"), []byte("key2"), []byte("key3")}
		values := [][]byte{[]byte("value1"), []byte("value2"), []byte("value3")}

		for i, key := range keys {
			err := batch.Put(key, values[i])
			require.NoError(t, err)
		}

		// Verify keys don't exist before commit
		for _, key := range keys {
			has, err := store.Has(key)
			require.NoError(t, err)
			assert.False(t, has, "key %s should not exist before batch commit", key)
		}

		// Commit the batch
		err := batch.Commit()
		require.NoError(t, err)

		// Verify all keys now exist
		for i, key := range keys {
			gotValue, found, err := store.Get(key)
			require.NoError(t, err)
			assert.True(t, found, "key %s should exist after batch commit", key)
			assert.True(t, bytes.Equal(gotValue, values[i]))
		}
	})

	t.Run("BatchDelete", func(t *testing.T) {
		// First, put some keys
		keys := [][]byte{[]byte("del1"), []byte("del2"), []byte("del3")}
		value := []byte("value")

		for _, key := range keys {
			err := store.Put(key, value)
			require.NoError(t, err)
		}

		// Create a batch and delete the keys
		batch := store.NewBatch()

		for _, key := range keys {
			err := batch.Delete(key)
			require.NoError(t, err)
		}

		// Verify keys still exist before commit
		for _, key := range keys {
			has, err := store.Has(key)
			require.NoError(t, err)
			assert.True(t, has, "key %s should still exist before batch commit", key)
		}

		// Commit the batch
		err := batch.Commit()
		require.NoError(t, err)

		// Verify all keys are deleted
		for _, key := range keys {
			has, err := store.Has(key)
			require.NoError(t, err)
			assert.False(t, has, "key %s should be deleted after batch commit", key)
		}
	})

	t.Run("BatchMixedOperations", func(t *testing.T) {
		batch := store.NewBatch()

		// Put some initial data
		existingKey := []byte("existing")
		err := store.Put(existingKey, []byte("old_value"))
		require.NoError(t, err)

		// Mix of operations in batch
		err = batch.Put([]byte("new1"), []byte("value1"))
		require.NoError(t, err)

		err = batch.Put(existingKey, []byte("new_value"))
		require.NoError(t, err)

		err = batch.Delete([]byte("new1"))
		require.NoError(t, err)

		err = batch.Put([]byte("new2"), []byte("value2"))
		require.NoError(t, err)

		// Commit the batch
		err = batch.Commit()
		require.NoError(t, err)

		// Verify results
		has, err := store.Has([]byte("new1"))
		require.NoError(t, err)
		assert.False(t, has, "new1 should not exist (put then deleted)")

		gotValue, found, err := store.Get([]byte("new2"))
		require.NoError(t, err)
		assert.True(t, found)
		assert.True(t, bytes.Equal(gotValue, []byte("value2")))

		gotValue, found, err = store.Get(existingKey)
		require.NoError(t, err)
		assert.True(t, found)
		assert.True(t, bytes.Equal(gotValue, []byte("new_value")))
	})
}
