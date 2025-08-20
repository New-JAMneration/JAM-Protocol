package redis

import (
	"os"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/db"
	testsuite "github.com/New-JAMneration/JAM-Protocol/internal/db/test"
)

func TestWithRedis(t *testing.T) {
	t.Run("DatabaseSuite", func(t *testing.T) {
		// Skip test if Redis is not available
		redisAddr := os.Getenv("REDIS_ADDR")
		if redisAddr == "" {
			redisAddr = "localhost:6379"
		}
		store, err := New(redisAddr, "", 0)
		if err != nil {
			t.Skipf("Redis not available: %v", err)
		}

		testsuite.TestDatabaseSuite(t, func() db.KeyValueDB {
			// Skip test if Redis is not available
			return store
		})
	})
}
