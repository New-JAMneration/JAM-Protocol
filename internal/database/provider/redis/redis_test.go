package redis

import (
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/database"
	database_test "github.com/New-JAMneration/JAM-Protocol/internal/database/test"
	"github.com/alicebob/miniredis/v2"
)

func TestRedisDatabase(t *testing.T) {
	database_test.TestDatabase(t, func() database.Database {
		mr, err := miniredis.Run()
		if err != nil {
			t.Fatalf("failed to start miniredis: %v", err)
		}

		t.Cleanup(func() {
			mr.Close()
		})

		return NewDatabase(mr.Addr(), "", 0)
	})
}
