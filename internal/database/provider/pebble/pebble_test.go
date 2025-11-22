package pebble_test

import (
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/database"
	pebbledb "github.com/New-JAMneration/JAM-Protocol/internal/database/provider/pebble"
	testcase "github.com/New-JAMneration/JAM-Protocol/internal/database/test"
	"github.com/test-go/testify/require"
)

func TestWithPebbleDB(t *testing.T) {
	t.Run("DatabaseSuite", func(t *testing.T) {
		testcase.TestDatabase(t, func() database.Database {
			db, err := pebbledb.NewTestDatabase()
			require.NoError(t, err)
			return db
		})
	})
}
