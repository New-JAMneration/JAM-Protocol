package memory

import (
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/database"
	testsuite "github.com/New-JAMneration/JAM-Protocol/internal/database/test"
)

func TestMemoryDB(t *testing.T) {
	testsuite.TestDatabaseSuite(t, func() database.Database {
		memoryDB := NewDatabase()
		return memoryDB
	})
}
