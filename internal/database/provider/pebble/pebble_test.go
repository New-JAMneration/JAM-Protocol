package pebble

import (
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/database"
	testsuite "github.com/New-JAMneration/JAM-Protocol/internal/database/test"
	"github.com/cockroachdb/pebble"
	"github.com/cockroachdb/pebble/vfs"
	"github.com/test-go/testify/require"
)

func TestWithPebbleDB(t *testing.T) {
	t.Run("DatabaseSuite", func(t *testing.T) {
		testsuite.TestDatabaseSuite(t, func() database.Database {
			db, err := pebble.Open("", &pebble.Options{
				FS: vfs.NewMem(),
			})
			require.NoError(t, err)

			return &pebbleDB{
				inner: db,
			}
		})
	})
}
