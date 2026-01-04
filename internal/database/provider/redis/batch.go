package redis

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/database"
	"github.com/go-redis/redis"
)

type batch struct {
	pipeline redis.Pipeliner
}

func (db *redisDB) NewBatch() database.Batch {
	return &batch{
		pipeline: db.client.Pipeline(),
	}
}

func (b *batch) Put(key, value []byte) error {
	return b.pipeline.Set(string(key), value, 0).Err()
}

func (b *batch) Delete(key []byte) error {
	return b.pipeline.Del(string(key)).Err()
}

func (b *batch) Commit() error {
	_, err := b.pipeline.Exec()
	return err
}

func (b *batch) Close() error {
	if b.pipeline != nil {
		return b.pipeline.Close()
	}
	return nil
}
