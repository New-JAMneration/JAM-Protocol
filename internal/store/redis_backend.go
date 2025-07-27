package store

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis"
)

// RedisClient is a small wrapper used by CE handlers/tests.
// It is intentionally minimal and binary-safe (members are stored as raw bytes via Go strings).
type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(addr string, password string, db int) *RedisClient {
	return &RedisClient{
		client: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
			DB:       db,
		}),
	}
}

func (c *RedisClient) Put(key string, value []byte) error {
	return c.client.Set(key, value, 0).Err()
}

func (c *RedisClient) PutWithTTL(key string, value []byte, ttl time.Duration) error {
	return c.client.Set(key, value, ttl).Err()
}

// Get returns (nil, nil) when key does not exist.
func (c *RedisClient) Get(key string) ([]byte, error) {
	b, err := c.client.Get(key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return b, err
}

func (c *RedisClient) Delete(key string) error {
	return c.client.Del(key).Err()
}

func (c *RedisClient) Exists(key string) (bool, error) {
	n, err := c.client.Exists(key).Result()
	return n > 0, err
}

func (c *RedisClient) SAdd(key string, member []byte) error {
	return c.client.SAdd(key, string(member)).Err()
}

func (c *RedisClient) SRem(key string, member []byte) error {
	return c.client.SRem(key, string(member)).Err()
}

func (c *RedisClient) SMembers(key string) ([][]byte, error) {
	members, err := c.client.SMembers(key).Result()
	if err != nil {
		return nil, err
	}
	out := make([][]byte, 0, len(members))
	for _, m := range members {
		out = append(out, []byte(m))
	}
	return out, nil
}

func (c *RedisClient) SIsMember(key string, member []byte) (bool, error) {
	return c.client.SIsMember(key, string(member)).Result()
}

// RedisBackend is a tiny helper facade used by CE138/141 tests.
type RedisBackend struct {
	client *RedisClient
}

func (r *RedisBackend) GetClient() *RedisClient {
	return r.client
}

func (r *RedisBackend) GetJustification(ctx context.Context, erasureRoot []byte, shardIndex uint32) ([]byte, error) {
	_ = ctx
	key := fmt.Sprintf("ce137_justification:%x:%d", erasureRoot, shardIndex)
	return r.client.Get(key)
}

func (r *RedisBackend) PutJustification(ctx context.Context, erasureRoot []byte, shardIndex uint32, justification []byte) error {
	_ = ctx
	key := fmt.Sprintf("ce137_justification:%x:%d", erasureRoot, shardIndex)
	return r.client.Put(key, justification)
}

var (
	redisOnce      sync.Once
	redisBackend   *RedisBackend
	miniRedis      *miniredis.Miniredis
	redisBackendMu sync.Mutex
)

// GetRedisBackend returns a singleton Redis backend.
//
// - If USE_MINI_REDIS=true, a local in-memory Redis is started automatically (for tests).
// - Otherwise, it requires REDIS_ADDR to be set.
func GetRedisBackend() (*RedisBackend, error) {
	var initErr error
	redisOnce.Do(func() {
		if os.Getenv("USE_MINI_REDIS") == "true" {
			mr, err := miniredis.Run()
			if err != nil {
				initErr = err
				return
			}
			miniRedis = mr
			redisBackend = &RedisBackend{client: NewRedisClient(mr.Addr(), "", 0)}
			return
		}

		addr := os.Getenv("REDIS_ADDR")
		if addr == "" {
			initErr = errors.New("REDIS_ADDR is not set (and USE_MINI_REDIS is not true)")
			return
		}
		redisBackend = &RedisBackend{client: NewRedisClient(addr, os.Getenv("REDIS_PASSWORD"), 0)}
	})

	if initErr != nil {
		return nil, initErr
	}
	if redisBackend == nil {
		return nil, errors.New("redis backend is not initialized")
	}
	return redisBackend, nil
}

func resetRedisBackend() {
	redisBackendMu.Lock()
	defer redisBackendMu.Unlock()

	if miniRedis != nil {
		miniRedis.Close()
		miniRedis = nil
	}
	if redisBackend != nil && redisBackend.client != nil && redisBackend.client.client != nil {
		_ = redisBackend.client.client.Close()
	}
	redisBackend = nil
	redisOnce = sync.Once{}
}
