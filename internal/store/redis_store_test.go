package store

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func tearDown() {
	globalRedisClient = nil
	redisInitOnce = sync.Once{}
}

func TestGetRedisClient_Success(t *testing.T) {
	defer tearDown()
	// This test ensures we can successfully connect to miniredis
	rdb, cleanup := setupTestRedis(t)
	defer cleanup()

	// We already connected in setupTestRedis. Now let's just confirm we have a client.
	assert.NotNil(t, rdb, "Expected a non-nil RedisClient")

	_, err := GetRedisClient()
	assert.NoError(t, err, "should not fail calling GetRedisClient again")
}

func TestGetRedisClient_Idempotent(t *testing.T) {
	defer tearDown()

	// Show that subsequent calls to GetRedisClient return the same pointer
	_, cleanup := setupTestRedis(t)
	defer cleanup()

	rdb, err := GetRedisClient()
	assert.NoError(t, err, "should not fail calling GetRedisClient again")

	rdb2, err := GetRedisClient()
	assert.NoError(t, err, "should not fail calling GetRedisClient again")

	assert.Equal(t, rdb, rdb2, "expected the same *RedisClient object due to sync.Once")
}
