package store

import (
	"context"
	"encoding/hex"
	"errors"
	"log"
	"os"
	"sync"

	"github.com/New-JAMneration/JAM-Protocol/config"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/alicebob/miniredis/v2"
)

var (
	redisInitOnce      sync.Once
	globalRedisClient  *RedisClient
	globalRedisBackend *RedisBackend
	globalMiniRedis    *miniredis.Miniredis
)

func GetRedisBackend() (*RedisBackend, error) {
	redisInitOnce.Do(func() {
		redisConfig := config.Config.Redis
		client := NewRedisClient(redisConfig.Address, redisConfig.Password, redisConfig.Port)
		if err := client.Ping(); err != nil {
			log.Printf("failed to connect to Redis: %v", err)

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

		// If we reach here, we have a valid Redis client (either real or miniredis)
		globalRedisClient = client
		globalRedisBackend = NewRedisBackend(client)

		// Set the genesis block in Redis
		genesisBlock := genGenesisBlock()
		err := globalRedisBackend.SetGenesisBlock(context.Background(), genesisBlock)
		if err != nil {
			log.Printf("failed to set genesis block in Redis: %v", err)
			return
		}
	})
	if globalRedisBackend == nil {
		return nil, errors.New("redis backend is not initialized")
	}
	return globalRedisBackend, nil
}

func genGenesisBlock() *types.Block {
	hash := "5c743dbc514284b2ea57798787c5a155ef9d7ac1e9499ec65910a7a3d65897b7"
	byteArray, _ := hex.DecodeString(hash)
	genesisBlock := types.Block{
		Header: types.Header{
			// hash string to jamTypes.HeaderHash
			Parent:          types.HeaderHash(byteArray),
			ParentStateRoot: types.StateRoot{},
			ExtrinsicHash:   types.OpaqueHash{},
			Slot:            0,
			EpochMark:       nil,
			TicketsMark:     nil,
			OffendersMark:   types.OffendersMark{},
			AuthorIndex:     0,
			EntropySource:   types.BandersnatchVrfSignature{},
			Seal:            types.BandersnatchVrfSignature{},
		},
		Extrinsic: types.Extrinsic{
			Tickets:    types.TicketsExtrinsic{},
			Preimages:  types.PreimagesExtrinsic{},
			Guarantees: types.GuaranteesExtrinsic{},
			Assurances: types.AssurancesExtrinsic{},
			Disputes:   types.DisputesExtrinsic{},
		},
	}

	return &genesisBlock
}
func CloseMiniRedis() {
	if globalMiniRedis != nil {
		globalMiniRedis.Close()
		globalMiniRedis = nil
	}
}
