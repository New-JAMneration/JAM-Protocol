package memory

import (
	"errors"

	"github.com/New-JAMneration/JAM-Protocol/internal/database"
)

type batch struct {
	db       *Database
	writeOps []writeOp
}

type writeOp struct {
	isDelete bool

	key   []byte
	value []byte

	rangeStart []byte
	rangeEnd   []byte
}

func (db *Database) NewBatch() database.Batch {
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
	b.writeOps = append(b.writeOps, writeOp{isDelete: true, key: key})
	return nil
}

func (b *batch) DeleteRange(start, end []byte) error {
	b.writeOps = append(b.writeOps, writeOp{isDelete: true, rangeStart: start, rangeEnd: end})
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
					if op.rangeStart != nil && key < string(op.rangeStart) {
						continue
					}
					if op.rangeEnd != nil && key >= string(op.rangeEnd) {
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
