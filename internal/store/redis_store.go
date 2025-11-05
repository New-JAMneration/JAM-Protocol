package store

import (
	"context"
	"errors"
	"log"
	"os"
	"sync"

	"github.com/New-JAMneration/JAM-Protocol/config"
	"github.com/alicebob/miniredis/v2"
)

var (
	redisInitOnce      sync.Once
	globalRedisBackend *RedisBackend
	globalMiniRedis    *miniredis.Miniredis
)

func GetRedisBackend() (*RedisBackend, error) {
	redisInitOnce.Do(func() {
		redisConfig := config.Config.Redis
		client := NewRedisClient(redisConfig.Address, redisConfig.Password, redisConfig.Port)
		if err := client.Ping(); err != nil {
			// If Redis connection fails, check if we should use miniredis for testing
			useMini := os.Getenv("USE_MINI_REDIS")
			if useMini == "true" {
				log.Printf("USE_MINI_REDIS is set to true, starting miniredis for testing")
				mr, miniErr := miniredis.Run()
				if miniErr != nil {
					log.Printf("failed to start miniredis: %v", miniErr)
					return
				}
				globalMiniRedis = mr
				client = NewRedisClient(mr.Addr(), "", 0)
			} else {
				log.Printf("failed to connect to Redis: %v", err)
				return
			}
		}

		globalRedisBackend = NewRedisBackend(client)

		hashSegmentMap := genInitHashSegmentMap()
		err := globalRedisBackend.SetHashSegmentMap(context.Background(), hashSegmentMap)
		if err != nil {
			log.Printf("failed to set hash segment map in Redis: %v", err)
			return
		}
	})
	if globalRedisBackend == nil {
		return nil, errors.New("redis backend is not initialized")
	}
	return globalRedisBackend, nil
}

func genInitHashSegmentMap() map[string]string {
	return make(map[string]string)
}

func CloseMiniRedis() {
	if globalMiniRedis != nil {
		globalMiniRedis.Close()
		globalMiniRedis = nil
	}
}
