package database_test

import (
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/database"
	"github.com/New-JAMneration/JAM-Protocol/internal/database/provider/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAtomicMultiNamespaceBatch(t *testing.T) {
	t.Run("NamespacedBatch", func(t *testing.T) {
		rootDB := memory.NewDatabase()
		defer rootDB.Close()

		headersDB := database.NewNamespaced(rootDB, []byte("headers/"))

		// Create a batch from namespaced database
		batch := headersDB.NewBatch()

		key := []byte("key1")
		value := []byte("value1")

		err := batch.Put(key, value)
		require.NoError(t, err)

		// Key shouldn't exist before commit
		has, err := headersDB.Has(key)
		require.NoError(t, err)
		assert.False(t, has)

		// Commit the batch
		err = batch.Commit()
		require.NoError(t, err)

		// Verify key exists with correct prefix
		actual, found, err := headersDB.Get(key)
		require.NoError(t, err)
		require.True(t, found)
		assert.Equal(t, value, actual)
	})

	t.Run("SharedBatchAcrossNamespaces", func(t *testing.T) {
		rootDB := memory.NewDatabase()
		defer rootDB.Close()

		key := []byte("key")
		headerValue := []byte("header_data")
		blockValue := []byte("block_data")
		serviceValue := []byte("service_data")

		// Create multiple namespaced databases
		headersDB := database.NewNamespaced(rootDB, []byte("headers/"))
		blocksDB := database.NewNamespaced(rootDB, []byte("blocks/"))
		servicesDB := database.NewNamespaced(rootDB, []byte("services/"))

		// Create a single batch from the root database
		rootBatch := rootDB.NewBatch()

		// Wrap the batch with different namespaces
		headersBatch := headersDB.BindBatch(rootBatch)
		blocksBatch := blocksDB.BindBatch(rootBatch)
		servicesBatch := servicesDB.BindBatch(rootBatch)

		// Write to different namespaces using the same underlying batch
		err := headersBatch.Put(key, headerValue)
		require.NoError(t, err)

		err = blocksBatch.Put(key, blockValue)
		require.NoError(t, err)

		err = servicesBatch.Put(key, serviceValue)
		require.NoError(t, err)

		// Verify nothing is committed yet
		has, err := headersDB.Has(key)
		require.NoError(t, err)
		assert.False(t, has, "headers/key should not exist before commit")

		has, err = blocksDB.Has(key)
		require.NoError(t, err)
		assert.False(t, has, "blocks/key should not exist before commit")

		has, err = servicesDB.Has(key)
		require.NoError(t, err)
		assert.False(t, has, "services/key should not exist before commit")

		// Single atomic commit for all namespaces
		err = rootBatch.Commit()
		require.NoError(t, err)

		// Verify all writes are now visible atomically
		val, found, err := headersDB.Get(key)
		require.NoError(t, err)
		require.True(t, found)
		assert.Equal(t, headerValue, val)

		val, found, err = blocksDB.Get(key)
		require.NoError(t, err)
		require.True(t, found)
		assert.Equal(t, blockValue, val)

		val, found, err = servicesDB.Get(key)
		require.NoError(t, err)
		require.True(t, found)
		assert.Equal(t, serviceValue, val)
	})

	t.Run("MixedOperationsAcrossNamespaces", func(t *testing.T) {
		rootDB := memory.NewDatabase()
		defer rootDB.Close()

		key1 := []byte("key1")
		headerValue1Before := []byte("header_data")
		headerValue1After := []byte("header_data_updated")
		blockValue1Before := []byte("block_data")
		blockValue1After := []byte("block_data_updated")

		key2 := []byte("key2")
		headerValue2 := []byte("header_data_2")
		blockValue2 := []byte("block_data_2")

		// Create multiple namespaced databases
		headersDB := database.NewNamespaced(rootDB, []byte("headers/"))
		blocksDB := database.NewNamespaced(rootDB, []byte("blocks/"))

		// Setup initial data
		err := headersDB.Put(key1, headerValue1Before)
		require.NoError(t, err)

		err = blocksDB.Put(key1, blockValue1Before)
		require.NoError(t, err)

		// Create a shared batch
		rootBatch := rootDB.NewBatch()
		headersBatch := headersDB.BindBatch(rootBatch)
		blocksBatch := blocksDB.BindBatch(rootBatch)

		err = headersBatch.Put(key1, headerValue1After)
		require.NoError(t, err)

		err = blocksBatch.Put(key1, blockValue1After)
		require.NoError(t, err)

		err = headersBatch.Put(key2, headerValue2)
		require.NoError(t, err)

		err = blocksBatch.Put(key2, blockValue2)
		require.NoError(t, err)

		// Verify old data still exists before commit
		value, found, err := headersDB.Get(key1)
		require.NoError(t, err)
		require.True(t, found)
		assert.Equal(t, headerValue1Before, value)

		value, found, err = blocksDB.Get(key1)
		require.NoError(t, err)
		require.True(t, found)
		assert.Equal(t, blockValue1Before, value)

		// Commit atomically
		err = rootBatch.Commit()
		require.NoError(t, err)

		// Verify all changes applied
		value, found, err = headersDB.Get(key1)
		require.NoError(t, err)
		require.True(t, found)
		assert.Equal(t, headerValue1After, value)

		value, found, err = blocksDB.Get(key1)
		require.NoError(t, err)
		require.True(t, found)
		assert.Equal(t, blockValue1After, value)

		value, found, err = headersDB.Get(key2)
		require.NoError(t, err)
		require.True(t, found)
		assert.Equal(t, headerValue2, value)

		value, found, err = blocksDB.Get(key2)
		require.NoError(t, err)
		require.True(t, found)
		assert.Equal(t, blockValue2, value)
	})

	t.Run("DeleteAcrossNamespaces", func(t *testing.T) {
		rootDB := memory.NewDatabase()
		defer rootDB.Close()

		// Create namespaced databases
		headersDB := database.NewNamespaced(rootDB, []byte("headers/"))
		blocksDB := database.NewNamespaced(rootDB, []byte("blocks/"))

		headers := [][]byte{[]byte("key1"), []byte("key2"), []byte("key3"), []byte("key4")}
		for _, header := range headers {
			err := headersDB.Put(header, []byte("data"))
			require.NoError(t, err)
		}

		blocks := [][]byte{[]byte("key1"), []byte("key2"), []byte("key3"), []byte("key4")}
		for _, block := range blocks {
			err := blocksDB.Put(block, []byte("data"))
			require.NoError(t, err)
		}

		// Create shared batch and delete range [key1, key3) in both namespaces
		rootBatch := rootDB.NewBatch()
		headersBatch := headersDB.BindBatch(rootBatch)
		blocksBatch := blocksDB.BindBatch(rootBatch)

		err := headersBatch.DeleteRange([]byte("key1"), []byte("key3"))
		require.NoError(t, err)

		err = blocksBatch.DeleteRange([]byte("key1"), []byte("key3"))
		require.NoError(t, err)

		err = rootBatch.Commit()
		require.NoError(t, err)

		// Verify key1 and key2 are deleted in both namespaces
		for _, key := range [][]byte{[]byte("key1"), []byte("key2")} {
			has, err := headersDB.Has(key)
			require.NoError(t, err)
			assert.False(t, has, "headers/%s should be deleted", key)

			has, err = blocksDB.Has(key)
			require.NoError(t, err)
			assert.False(t, has, "blocks/%s should be deleted", key)
		}

		// Verify key3 and key4 still exist in both namespaces
		for _, key := range [][]byte{[]byte("key3"), []byte("key4")} {
			has, err := headersDB.Has(key)
			require.NoError(t, err)
			assert.True(t, has, "headers/%s should still exist", key)

			has, err = blocksDB.Has(key)
			require.NoError(t, err)
			assert.True(t, has, "blocks/%s should still exist", key)
		}
	})
}
