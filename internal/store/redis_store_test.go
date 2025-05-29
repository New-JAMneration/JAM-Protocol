package store

import (
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func tearDown() {
	redisInitOnce = sync.Once{}
	CloseMiniRedis()
}

func TestGetRedisBackendSuccess(t *testing.T) {
	defer tearDown()
	os.Setenv("USE_MINI_REDIS", "true")
	_, err := GetRedisBackend()
	assert.NoError(t, err, "should not fail calling GetRedisClient again")
}

func TestGetRedisClient_Idempotent(t *testing.T) {
	defer tearDown()
	os.Setenv("USE_MINI_REDIS", "true")

	backend, err := GetRedisBackend()
	assert.NoError(t, err, "should not fail calling GetRedisClient again")

	backend2, err := GetRedisBackend()
	assert.NoError(t, err, "should not fail calling GetRedisClient again")

	assert.Equal(t, backend, backend2, "expected the same *RedisClient object due to sync.Once")
}
