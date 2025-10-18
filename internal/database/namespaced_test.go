package database_test

import (
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/database"
	"github.com/New-JAMneration/JAM-Protocol/internal/database/provider/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNamespacedDatabase(t *testing.T) {
	rootDB := memory.NewDatabase()
	defer rootDB.Close()

	headersDB := database.NewNamespaced(rootDB, []byte("headers/"))
	blocksDB := database.NewNamespaced(rootDB, []byte("blocks/"))

	t.Run("IsolatedNamespaces", func(t *testing.T) {
		// Write to namespace 1
		err := headersDB.Put([]byte("key"), []byte("data1"))
		require.NoError(t, err)

		// Write to namespace 2
		err = blocksDB.Put([]byte("key"), []byte("data2"))
		require.NoError(t, err)

		// Verify isolation - same key in different namespaces
		val1, found, err := headersDB.Get([]byte("key"))
		require.NoError(t, err)
		require.True(t, found)
		assert.Equal(t, []byte("data1"), val1)

		val2, found, err := blocksDB.Get([]byte("key"))
		require.NoError(t, err)
		require.True(t, found)
		assert.Equal(t, []byte("data2"), val2)

		// Verify the actual keys in root DB have prefixes
		rootVal1, found, err := rootDB.Get([]byte("headers/key"))
		require.NoError(t, err)
		require.True(t, found)
		assert.Equal(t, []byte("data1"), rootVal1)

		rootVal2, found, err := rootDB.Get([]byte("blocks/key"))
		require.NoError(t, err)
		require.True(t, found)
		assert.Equal(t, []byte("data2"), rootVal2)
	})
}
