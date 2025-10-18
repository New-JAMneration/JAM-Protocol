package memory

import (
	"errors"

	"github.com/New-JAMneration/JAM-Protocol/internal/database"
)

type batch struct {
	db       *memoryDB
	writeOps []writeOp
}

type writeOp struct {
	key      []byte
	value    []byte
	isDelete bool

	rangeFrom []byte
	rangeTo   []byte
}

func (db *memoryDB) NewBatch() database.Batch {
	return &batch{
		db: db,
	}
}

func (b *batch) Put(key, value []byte) error {
	copied := make([]byte, len(value))
	copy(copied, value)

	b.writeOps = append(b.writeOps, writeOp{key: key, value: copied})
	return nil
}

func (b *batch) Delete(key []byte) error {
	b.writeOps = append(b.writeOps, writeOp{key: key, isDelete: true})
	return nil
}

func (b *batch) DeleteRange(start, end []byte) error {
	b.writeOps = append(b.writeOps, writeOp{rangeFrom: start, rangeTo: end, isDelete: true})
	return nil
}

func (b *batch) Commit() error {
	b.db.mu.Lock()
	defer b.db.mu.Unlock()

	if b.db.data == nil {
		return errors.New("database is closed")
	}

	for _, op := range b.writeOps {
		if op.isDelete {
			if len(op.key) != 0 {
				delete(b.db.data, string(op.key))
			} else {
				// Range deletion [start, end)
				for key := range b.db.data {
					if op.rangeFrom != nil && key < string(op.rangeFrom) {
						continue
					}
					if op.rangeTo != nil && key >= string(op.rangeTo) {
						continue
					}
					delete(b.db.data, key)
				}
			}
		} else {
			b.db.data[string(op.key)] = op.value
		}
	}

	return nil
}
