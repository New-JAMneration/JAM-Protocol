package redis

import (
	"math"

	"github.com/go-redis/redis"
)

type KVStore struct {
	client *redis.Client
}

func New(addr, password string, db int) (*KVStore, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	if err := client.Ping().Err(); err != nil {
		return nil, err
	}

	return &KVStore{
		client: client,
	}, nil
}

func (kv *KVStore) Has(key []byte) (bool, error) {
	result := kv.client.Exists(string(key))
	if result.Err() != nil {
		return false, result.Err()
	}
	return result.Val() > 0, nil
}

func (kv *KVStore) Get(key []byte) ([]byte, error) {
	result := kv.client.Get(string(key))
	if result.Err() == redis.Nil {
		return nil, nil
	}
	if result.Err() != nil {
		return nil, result.Err()
	}
	return []byte(result.Val()), nil
}

func (kv *KVStore) Set(key, value []byte) error {
	return kv.client.Set(string(key), value, 0).Err()
}

func (kv *KVStore) Delete(key []byte) error {
	return kv.client.Del(string(key)).Err()
}

// DeleteRange deletes all keys in the range [start, end).
// Note: Redis does not support range deletion natively, so this implementation scans all keys.
// Should not be used in production for large datasets.
func (kv *KVStore) DeleteRange(start, end []byte) error {
	iter := kv.client.Scan(0, "*", math.MaxInt64).Iterator()
	keys := make([]string, 0)

	for iter.Next() {
		key := iter.Val()
		if key >= string(start) && key < string(end) {
			keys = append(keys, key)
		}
	}

	if err := iter.Err(); err != nil {
		return err
	}

	if len(keys) > 0 {
		return kv.client.Del(keys...).Err()
	}

	return nil
}

func (kv *KVStore) Close() error {
	return kv.client.Close()
}
