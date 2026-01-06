package keystore

import (
	"errors"
	"fmt"
	"sync"

	"github.com/New-JAMneration/JAM-Protocol/internal/database"
)

// DatabaseStorage implements Storage interface using database.Database
type DatabaseStorage struct {
	db database.Database
}

// Compile-time interface check
var _ Storage = (*DatabaseStorage)(nil)
var _ Transaction = (*DatabaseTransaction)(nil)

// NewDatabaseStorage creates a new DatabaseStorage instance
func NewDatabaseStorage(db database.Database) *DatabaseStorage {
	return &DatabaseStorage{db: db}
}

func (d *DatabaseStorage) Put(key string, value []byte) error {
	return d.db.Put([]byte(key), value)
}

func (d *DatabaseStorage) Get(key string) ([]byte, error) {
	value, exists, err := d.db.Get([]byte(key))
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrKeyNotFound
	}
	return value, nil
}

func (d *DatabaseStorage) Delete(key string) error {
	return d.db.Delete([]byte(key))
}

// Set operations using prefix-based storage for efficiency
// Format: {setKey}:member:{memberValue} => "1"
func (d *DatabaseStorage) SetAdd(key string, members ...[]byte) error {
	if len(members) == 0 {
		return nil
	}

	batch := d.db.NewBatch()
	defer batch.Close()

	for _, member := range members {
		memberKey := fmt.Sprintf("%s:member:%s", key, string(member))
		if err := batch.Put([]byte(memberKey), []byte("1")); err != nil {
			return fmt.Errorf("failed to add set member: %w", err)
		}
	}

	return batch.Commit()
}

func (d *DatabaseStorage) SetRemove(key string, members ...[]byte) error {
	if len(members) == 0 {
		return nil
	}

	batch := d.db.NewBatch()
	defer batch.Close()

	for _, member := range members {
		memberKey := fmt.Sprintf("%s:member:%s", key, string(member))
		if err := batch.Delete([]byte(memberKey)); err != nil {
			return fmt.Errorf("failed to remove set member: %w", err)
		}
	}

	return batch.Commit()
}

func (d *DatabaseStorage) SetMembers(key string) ([][]byte, error) {
	prefix := []byte(fmt.Sprintf("%s:member:", key))
	iter, err := d.db.NewIterator(prefix, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create iterator: %w", err)
	}
	defer iter.Close()

	var members [][]byte
	for iter.Next() {
		iterKey := iter.Key()
		// Extract member value from key: {setKey}:member:{memberValue}
		if len(iterKey) > len(prefix) {
			member := iterKey[len(prefix):]
			members = append(members, member)
		}
	}

	if err := iter.Error(); err != nil {
		return nil, fmt.Errorf("iterator error: %w", err)
	}

	return members, nil
}

func (d *DatabaseStorage) SetIsMember(key string, member []byte) (bool, error) {
	memberKey := fmt.Sprintf("%s:member:%s", key, string(member))
	exists, err := d.db.Has([]byte(memberKey))
	if err != nil {
		return false, fmt.Errorf("failed to check set membership: %w", err)
	}
	return exists, nil
}

// GetMultiple retrieves multiple keys efficiently
func (d *DatabaseStorage) GetMultiple(keys []string) (map[string][]byte, error) {
	result := make(map[string][]byte, len(keys))
	for _, key := range keys {
		value, exists, err := d.db.Get([]byte(key))
		if err != nil {
			return nil, fmt.Errorf("failed to get key %s: %w", key, err)
		}
		if exists {
			result[key] = value
		}
	}
	return result, nil
}

// Begin starts a transaction
func (d *DatabaseStorage) Begin() (Transaction, error) {
	return &DatabaseTransaction{
		db:  d.db,
		ops: make([]operation, 0),
	}, nil
}

type operation struct {
	opType string // "put", "delete", "setadd", "setremove"
	key    string
	value  []byte
	member []byte
}

type DatabaseTransaction struct {
	db   database.Database
	ops  []operation
	mu   sync.Mutex
	done bool
}

func (t *DatabaseTransaction) Put(key string, value []byte) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.done {
		return errors.New("transaction already committed or rolled back")
	}
	t.ops = append(t.ops, operation{opType: "put", key: key, value: value})
	return nil
}

func (t *DatabaseTransaction) Delete(key string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.done {
		return errors.New("transaction already committed or rolled back")
	}
	t.ops = append(t.ops, operation{opType: "delete", key: key})
	return nil
}

func (t *DatabaseTransaction) SetAdd(key string, members ...[]byte) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.done {
		return errors.New("transaction already committed or rolled back")
	}
	for _, member := range members {
		t.ops = append(t.ops, operation{opType: "setadd", key: key, member: member})
	}
	return nil
}

func (t *DatabaseTransaction) SetRemove(key string, members ...[]byte) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.done {
		return errors.New("transaction already committed or rolled back")
	}
	for _, member := range members {
		t.ops = append(t.ops, operation{opType: "setremove", key: key, member: member})
	}
	return nil
}

func (t *DatabaseTransaction) Commit() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.done {
		return errors.New("transaction already committed or rolled back")
	}

	batch := t.db.NewBatch()
	defer batch.Close()

	for _, op := range t.ops {
		switch op.opType {
		case "put":
			if err := batch.Put([]byte(op.key), op.value); err != nil {
				return fmt.Errorf("failed to put in transaction: %w", err)
			}
		case "delete":
			if err := batch.Delete([]byte(op.key)); err != nil {
				return fmt.Errorf("failed to delete in transaction: %w", err)
			}
		case "setadd":
			memberKey := fmt.Sprintf("%s:member:%s", op.key, string(op.member))
			if err := batch.Put([]byte(memberKey), []byte("1")); err != nil {
				return fmt.Errorf("failed to setadd in transaction: %w", err)
			}
		case "setremove":
			memberKey := fmt.Sprintf("%s:member:%s", op.key, string(op.member))
			if err := batch.Delete([]byte(memberKey)); err != nil {
				return fmt.Errorf("failed to setremove in transaction: %w", err)
			}
		}
	}

	t.done = true
	return batch.Commit()
}

func (t *DatabaseTransaction) Rollback() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.done {
		return errors.New("transaction already committed or rolled back")
	}
	t.done = true
	t.ops = nil
	return nil
}
