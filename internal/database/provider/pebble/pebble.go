package pebble

import (
	"runtime"

	"github.com/cockroachdb/pebble"
)

type Database struct {
	datadir   string
	inner     *pebble.DB
	writeOpts *pebble.WriteOptions
}

func New(datadir string, readOnly bool) (*Database, error) {
	opt := &pebble.Options{
		// Default compaction concurrency is single-threaded, use all available logical cores
		// for speeding up compactions.
		MaxConcurrentCompactions: runtime.NumCPU,
		ReadOnly:                 readOnly,
	}
	pebbleDB, err := pebble.Open(datadir, opt)
	if err != nil {
		return nil, err
	}
	engine := &Database{
		datadir:   datadir,
		inner:     pebbleDB,
		writeOpts: pebble.NoSync,
	}
	return engine, nil
}

func (db *Database) Has(key []byte) (bool, error) {
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

func (db *Database) Get(key []byte) ([]byte, bool, error) {
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

func (db *Database) Put(key, value []byte) error {
	return db.inner.Set(key, value, db.writeOpts)
}

func (db *Database) Delete(key []byte) error {
	return db.inner.Delete(key, db.writeOpts)
}

func (db *Database) DeleteRange(start, end []byte) error {
	return db.inner.DeleteRange(start, end, db.writeOpts)
}

func (db *Database) Close() error {
	return db.inner.Close()
}
