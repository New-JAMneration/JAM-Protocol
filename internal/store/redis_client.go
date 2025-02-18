package store

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/go-redis/redis"
)

// RedisClient wraps the go-redis client.
type RedisClient struct {
	client *redis.Client
}

// NewRedisClient initializes and returns a new Redis client.
func NewRedisClient(addr string, password string, db int) *RedisClient {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return &RedisClient{client: rdb}
}

// Ping can be used to verify the connection to Redis.
func (r *RedisClient) Ping() error {
	return r.client.Ping().Err()
}

// Put sets key -> value in Redis. No expiration is set.
func (r *RedisClient) Put(key string, value []byte) error {
	// For logging/tracing:
	log.Printf("PUT key=%s value(hex)=%s", key, hex.EncodeToString(value))

	// TODO: setup expired time
	err := r.client.Set(key, value, 0).Err()
	if err != nil {
		return fmt.Errorf("Put failed: %w", err)
	}
	return nil
}

// Get fetches the value at a given key. Returns nil if key does not exist.
func (r *RedisClient) Get(key string) ([]byte, error) {
	log.Printf("GET key=%s", key)

	val, err := r.client.Get(key).Bytes()
	if err != nil {
		if err == redis.Nil {
			// Key not found
			return nil, nil
		}
		return nil, fmt.Errorf("Get failed: %w", err)
	}
	return val, nil
}

func (r *RedisClient) Exists(key string) (bool, error) {
	_, err := r.client.Get(key).Bytes()
	if err != nil {
		if err == redis.Nil {
			// Key not found
			return false, nil
		}
		return false, fmt.Errorf("Get failed: %w", err)
	}
	return true, nil
}

// Delete removes a given key from Redis.
func (r *RedisClient) Delete(key string) error {
	fmt.Printf("DELETE key=%s\n", key)

	err := r.client.Del(key).Err()
	if err != nil {
		return fmt.Errorf("Delete failed: %w", err)
	}
	return nil
}

type OpType int

const (
	OpPut OpType = iota
	OpDelete
)

type BatchOperation struct {
	Type  OpType
	Key   string
	Value []byte // only used if Type == OpPut
}

func (r *RedisClient) Batch(ctx context.Context, operations []BatchOperation) error {
	// Create pipeline
	pipe := r.client.Pipeline()

	for _, op := range operations {
		switch op.Type {
		case OpPut:
			fmt.Printf("BATCH PUT key=%s\n", op.Key)
			pipe.Set(op.Key, op.Value, 0)
		case OpDelete:
			fmt.Printf("BATCH DELETE key=%s\n", op.Key)
			pipe.Del(op.Key)
		}
	}

	// Execute pipeline
	_, err := pipe.Exec()
	if err != nil {
		return fmt.Errorf("Batch failed: %w", err)
	}

	return nil
}

// DeleteBlock removes the block by slot.
func (r *RedisClient) DeleteBlock(ctx context.Context, slot types.TimeSlot) error {
	key := fmt.Sprintf("block:%d", slot)
	return r.Delete(key)
}

// SAdd inserts one or more members into a Redis set.
func (r *RedisClient) SAdd(key string, members ...[]byte) error {
	// For logging/tracing:
	for _, m := range members {
		log.Printf("SADD key=%s member(hex)=%s", key, hex.EncodeToString(m))
	}

	// Convert []byte members to interface{}
	interfaceVals := make([]interface{}, len(members))
	for i, mb := range members {
		interfaceVals[i] = mb
	}

	err := r.client.SAdd(key, interfaceVals...).Err()
	if err != nil {
		return fmt.Errorf("SAdd failed: %w", err)
	}
	return nil
}

// SRem removes one or more members from a Redis set.
func (r *RedisClient) SRem(key string, members ...[]byte) error {
	for _, m := range members {
		log.Printf("SREM key=%s member(hex)=%s", key, hex.EncodeToString(m))
	}

	interfaceVals := make([]interface{}, len(members))
	for i, mb := range members {
		interfaceVals[i] = mb
	}

	err := r.client.SRem(key, interfaceVals...).Err()
	if err != nil {
		return fmt.Errorf("SRem failed: %w", err)
	}
	return nil
}

// SMembers retrieves all members of a Redis set, returning them as [][]byte.
func (r *RedisClient) SMembers(key string) ([][]byte, error) {
	log.Printf("SMEMBERS key=%s", key)

	strVals, err := r.client.SMembers(key).Result()
	if err != nil {
		return nil, fmt.Errorf("SMembers failed: %w", err)
	}

	// Convert from string -> []byte
	byteVals := make([][]byte, 0, len(strVals))
	for _, s := range strVals {
		byteVals = append(byteVals, []byte(s))
	}

	return byteVals, nil
}

// SIsMember checks if a given member is in the set at key.
func (r *RedisClient) SIsMember(key string, member []byte) (bool, error) {
	log.Printf("SISMEMBER key=%s member(hex)=%s", key, hex.EncodeToString(member))

	ok, err := r.client.SIsMember(key, member).Result()
	if err != nil {
		return false, fmt.Errorf("SIsMember failed: %w", err)
	}
	return ok, nil
}
