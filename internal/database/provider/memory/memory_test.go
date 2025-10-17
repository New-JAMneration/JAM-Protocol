package memory

import (
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/database"
	testsuite "github.com/New-JAMneration/JAM-Protocol/internal/database/test"
)

func TestMemoryKVDB(t *testing.T) {
	memoryStore := NewDatabase()
	testsuite.TestDatabaseSuite(t, func() database.Database {
		return memoryStore
	})
}
