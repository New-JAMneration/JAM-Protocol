package pebble

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/database"
	"github.com/cockroachdb/pebble"
)

type batch struct {
	pb *pebble.Batch
	db *Database
}

func (db *Database) NewBatch() database.Batch {
	return &batch{
		pb: db.inner.NewBatch(),
		db: db,
	}
}

func (b *batch) Put(key, value []byte) error {
	copied := make([]byte, len(value))
	copy(copied, value)

	err := b.pb.Set(key, copied, b.db.writeOpts)
	if err != nil {
		return err
	}
	return nil
}

func (b *batch) Delete(key []byte) error {
	err := b.pb.Delete(key, b.db.writeOpts)
	if err != nil {
		return err
	}
	return nil
}

func (b *batch) DeleteRange(start, end []byte) error {
	err := b.pb.DeleteRange(start, end, b.db.writeOpts)
	if err != nil {
		return err
	}
	return nil
}

func (b *batch) Commit() error {
	return b.pb.Commit(b.db.writeOpts)
}
