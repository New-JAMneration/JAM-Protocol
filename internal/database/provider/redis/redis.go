package redis

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/database"
	"github.com/go-redis/redis"
)

type redisDB struct {
	client *redis.Client
}

// NewDatabase creates a new Redis-backed database.
func NewDatabase(addr string, password string, db int) database.Database {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return &redisDB{
		client: client,
	}
}

func (db *redisDB) Has(key []byte) (bool, error) {
	result, err := db.client.Exists(string(key)).Result()
	if err != nil {
		return false, err
	}

	return result > 0, nil
}

func (db *redisDB) Get(key []byte) ([]byte, bool, error) {
	value, err := db.client.Get(string(key)).Bytes()
	if err == redis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	return value, true, nil
}

func (db *redisDB) Put(key, value []byte) error {
	return db.client.Set(string(key), value, 0).Err()
}

func (db *redisDB) Delete(key []byte) error {
	return db.client.Del(string(key)).Err()
}

func (db *redisDB) Close() error {
	return db.client.Close()
}
