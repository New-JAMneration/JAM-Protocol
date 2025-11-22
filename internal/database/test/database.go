package database_test

import (
	"bytes"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatabase(t *testing.T, New func() database.Database) {
	t.Run("BasicOps", func(t *testing.T) {
		store := New()
		defer store.Close()

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

	t.Run("BatchWrite", func(t *testing.T) {
		store := New()
		defer store.Close()

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
		store := New()
		defer store.Close()

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
		store := New()
		defer store.Close()

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

	t.Run("IteratorAllKeys", func(t *testing.T) {
		store := New()
		defer store.Close()

		// Put some test data with a specific prefix to avoid conflicts with other tests
		keys := [][]byte{[]byte("allrange_key1"), []byte("allrange_key2"), []byte("allrange_key3")}
		values := [][]byte{[]byte("value1"), []byte("value2"), []byte("value3")}

		for i, key := range keys {
			err := store.Put(key, values[i])
			require.NoError(t, err)
		}

		// Iterate through keys with specific range prefix
		iter, err := store.NewIterator(nil, nil)
		require.NoError(t, err)
		defer iter.Close()

		iteratedKeys := make([][]byte, 0)
		iteratedValues := make([][]byte, 0)
		for iter.Next() {
			key := iter.Key()
			value := iter.Value()
			assert.NotNil(t, key)
			assert.NotNil(t, value)

			// Copy key and value since Iterator.Next() returns pointers to internal data
			keyCopy := make([]byte, len(key))
			copy(keyCopy, key)
			valueCopy := make([]byte, len(value))
			copy(valueCopy, value)
			iteratedKeys = append(iteratedKeys, keyCopy)
			iteratedValues = append(iteratedValues, valueCopy)
		}

		assert.NoError(t, iter.Error())
		assert.Equal(t, len(keys), len(iteratedKeys), "should iterate through all keys")

		// Verify length matches
		assert.Equal(t, len(keys), len(iteratedKeys), "should iterate exactly the expected number of keys")
		assert.Equal(t, len(values), len(iteratedValues), "should iterate exactly the expected number of values")

		// Verify keys and values at same indices
		for i := 0; i < len(iteratedKeys); i++ {
			assert.True(t, bytes.Equal(iteratedKeys[i], keys[i]), "key at index %d should match", i)
			assert.True(t, bytes.Equal(iteratedValues[i], values[i]), "value at index %d should match", i)
		}
	})

	t.Run("IteratorWithPrefix", func(t *testing.T) {
		store := New()
		defer store.Close()

		// Put test data
		keys := [][]byte{[]byte("key"), []byte("prefix_x"), []byte("prefix_y"), []byte("prefix_z")}
		values := [][]byte{[]byte("value"), []byte("prefix_val_x"), []byte("prefix_val_y"), []byte("prefix_val_z")}

		for i, key := range keys {
			err := store.Put(key, values[i])
			require.NoError(t, err)
		}

		iter, err := store.NewIterator([]byte("prefix"), nil)
		require.NoError(t, err)
		defer iter.Close()

		var iteratedKeys [][]byte
		var iteratedValues [][]byte
		for iter.Next() {
			key := iter.Key()
			value := iter.Value()

			// Copy key and value since Iterator.Next() returns pointers to internal data
			keyCopy := make([]byte, len(key))
			copy(keyCopy, key)
			valueCopy := make([]byte, len(value))
			copy(valueCopy, value)
			iteratedKeys = append(iteratedKeys, keyCopy)
			iteratedValues = append(iteratedValues, valueCopy)
		}

		assert.NoError(t, iter.Error())

		// Expected keys before "end_y": end_x
		expectedKeys := [][]byte{[]byte("prefix_x"), []byte("prefix_y"), []byte("prefix_z")}
		expectedValues := [][]byte{[]byte("prefix_val_x"), []byte("prefix_val_y"), []byte("prefix_val_z")}

		// Verify length matches
		assert.Equal(t, len(expectedKeys), len(iteratedKeys), "should iterate exactly the expected number of keys")
		assert.Equal(t, len(expectedValues), len(iteratedValues), "should iterate exactly the expected number of values")

		// Verify keys and values at same indices
		for i := 0; i < len(iteratedKeys); i++ {
			assert.True(t, bytes.Equal(iteratedKeys[i], expectedKeys[i]), "key at index %d should match", i)
			assert.True(t, bytes.Equal(iteratedValues[i], expectedValues[i]), "value at index %d should match", i)
		}
	})

	t.Run("IteratorWithPrefixAndStart", func(t *testing.T) {
		store := New()
		defer store.Close()

		// Put test data
		keys := [][]byte{[]byte("prefix_key"), []byte("prefix_start_a"), []byte("prefix_start_b"), []byte("prefix_start_c")}
		values := [][]byte{[]byte("val"), []byte("val_a"), []byte("val_b"), []byte("val_c")}

		for i, key := range keys {
			err := store.Put(key, values[i])
			require.NoError(t, err)
		}

		// Iterate from "range_b" to end
		iter, err := store.NewIterator([]byte("prefix"), []byte("_start"))
		require.NoError(t, err)
		defer iter.Close()

		var iteratedKeys [][]byte
		var iteratedValues [][]byte
		for iter.Next() {
			key := iter.Key()
			value := iter.Value()
			// Copy key and value since Iterator.Next() returns pointers to internal data
			keyCopy := make([]byte, len(key))
			copy(keyCopy, key)
			valueCopy := make([]byte, len(value))
			copy(valueCopy, value)
			iteratedKeys = append(iteratedKeys, keyCopy)
			iteratedValues = append(iteratedValues, valueCopy)
		}

		assert.NoError(t, iter.Error())

		// Expected keys from "range_b" onwards: range_b, range_c, range_d, range_e
		expectedKeys := [][]byte{[]byte("prefix_start_a"), []byte("prefix_start_b"), []byte("prefix_start_c")}
		expectedValues := [][]byte{[]byte("val_a"), []byte("val_b"), []byte("val_c")}

		// Verify length matches
		assert.Equal(t, len(expectedKeys), len(iteratedKeys), "should iterate exactly the expected number of keys")
		assert.Equal(t, len(expectedValues), len(iteratedValues), "should iterate exactly the expected number of values")

		// Verify keys and values at same indices
		for i := 0; i < len(iteratedKeys); i++ {
			assert.True(t, bytes.Equal(iteratedKeys[i], expectedKeys[i]), "key at index %d should match", i)
			assert.True(t, bytes.Equal(iteratedValues[i], expectedValues[i]), "value at index %d should match", i)
		}
	})

	t.Run("IteratorNoMatchingRange", func(t *testing.T) {
		store := New()
		defer store.Close()

		// Put some data
		err := store.Put([]byte("nomatch_abc"), []byte("value"))
		require.NoError(t, err)

		// Iterate with a range that doesn't match
		iter, err := store.NewIterator([]byte("nomatch_xyz"), []byte("nomatch_zzz"))
		require.NoError(t, err)
		defer iter.Close()

		assert.False(t, iter.Next(), "should not find keys outside range")
		assert.NoError(t, iter.Error())
	})
}
