package store

import (
	"errors"
	"log"
	"sync"

	"github.com/New-JAMneration/JAM-Protocol/config"
)

var (
	redisInitOnce     sync.Once
	globalRedisClient *RedisClient
)

func GetRedisClient() (*RedisClient, error) {
	redisInitOnce.Do(func() {
		// Initialize the Redis client
		redisConfig := config.Config.Redis
		globalRedisClient = NewRedisClient(redisConfig.Address, redisConfig.Password, redisConfig.Port)
		if err := globalRedisClient.Ping(); err != nil {
			log.Printf("failed to connect to Redis: %v", err)
			globalRedisClient = nil
		} else {
			log.Println("Redis client initialized successfully")
		}

		// TODO: Add the genesis block to the Redis store
	})

	if globalRedisClient == nil {
		return nil, errors.New("redis client is not initialized")
	}
	return globalRedisClient, nil
}
