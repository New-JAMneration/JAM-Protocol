package pebble

import (
	"runtime"

	"github.com/New-JAMneration/JAM-Protocol/internal/database"
	"github.com/cockroachdb/pebble"
)

type pebbleDB struct {
	datadir   string
	inner     *pebble.DB
	writeOpts *pebble.WriteOptions
}

func NewDatabase(datadir string, readOnly bool) (database.Database, error) {
	opt := &pebble.Options{
		// Default compaction concurrency is 1, use all available logical cores
		// for speeding up compactions.
		MaxConcurrentCompactions: runtime.NumCPU,
		ReadOnly:                 readOnly,
	}
	db, err := pebble.Open(datadir, opt)
	if err != nil {
		return nil, err
	}
	engine := &pebbleDB{
		datadir:   datadir,
		inner:     db,
		writeOpts: pebble.NoSync,
	}
	return engine, nil
}

func (db *pebbleDB) Has(key []byte) (bool, error) {
	_, closer, err := db.inner.Get(key)
	if err == pebble.ErrNotFound {
		return false, nil
	} else if err != nil {
		return false, err
	}
	if err = closer.Close(); err != nil {
		return false, err
	}
	return true, nil
}

func (db *pebbleDB) Get(key []byte) ([]byte, bool, error) {
	value, closer, err := db.inner.Get(key)
	if err != nil {
		if err == pebble.ErrNotFound {
			return nil, false, nil
		}
		return nil, false, err
	}
	data := make([]byte, len(value))
	copy(data, value)
	if err = closer.Close(); err != nil {
		return nil, false, err
	}
	return data, true, nil
}

func (db *pebbleDB) Put(key, value []byte) error {
	return db.inner.Set(key, value, db.writeOpts)
}

func (db *pebbleDB) Delete(key []byte) error {
	return db.inner.Delete(key, db.writeOpts)
}

func (db *pebbleDB) Close() error {
	return db.inner.Close()
}
