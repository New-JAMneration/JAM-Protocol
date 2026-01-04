package memory

import (
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/database"
	testcase "github.com/New-JAMneration/JAM-Protocol/internal/database/test"
)

func TestMemoryDB(t *testing.T) {
	t.Run("DatabaseSuite", func(t *testing.T) {
		testcase.TestDatabase(t, func() database.Database {
			memoryDB := NewDatabase()
			return memoryDB
		})
	})
}
