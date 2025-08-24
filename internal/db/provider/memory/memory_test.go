package memory

import (
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/db"
	testsuite "github.com/New-JAMneration/JAM-Protocol/internal/db/test"
)

func TestMemoryKVDB(t *testing.T) {
	memoryStore := New()
	testsuite.TestDatabaseSuite(t, func() db.KeyValueDB {
		return memoryStore
	})
}
