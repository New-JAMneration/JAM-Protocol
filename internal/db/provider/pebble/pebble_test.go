package pebble

import (
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/db"
	testsuite "github.com/New-JAMneration/JAM-Protocol/internal/db/test"
	"github.com/cockroachdb/pebble"
	"github.com/cockroachdb/pebble/vfs"
	"github.com/test-go/testify/require"
)

func TestWithPebbleDB(t *testing.T) {
	t.Run("DatabaseSuite", func(t *testing.T) {
		pebbleDB, err := pebble.Open("", &pebble.Options{
			FS: vfs.NewMem(),
		})
		require.NoError(t, err)

		testsuite.TestDatabaseSuite(t, func() db.KeyValueDB {
			return &KVStore{
				inner: pebbleDB,
			}
		})
	})
}
