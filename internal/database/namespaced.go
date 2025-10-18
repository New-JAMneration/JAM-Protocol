package database

type namespacedDatabase struct {
	inner Database
	ns    []byte
}

// NewNamespaced creates a new namespaced database that prefixes all keys with the given namespace.
func NewNamespaced(db Database, ns []byte) *namespacedDatabase {
	return &namespacedDatabase{
		inner: db,
		ns:    ns,
	}
}

func (db *namespacedDatabase) Has(key []byte) (bool, error) {
	return db.inner.Has(append(db.ns, key...))
}

func (db *namespacedDatabase) Get(key []byte) ([]byte, bool, error) {
	return db.inner.Get(append(db.ns, key...))
}

func (db *namespacedDatabase) Put(key, value []byte) error {
	return db.inner.Put(append(db.ns, key...), value)
}

func (db *namespacedDatabase) Delete(key []byte) error {
	return db.inner.Delete(append(db.ns, key...))
}

func (db *namespacedDatabase) DeleteRange(start, end []byte) error {
	return db.inner.DeleteRange(append(db.ns, start...), append(db.ns, end...))
}

func (db *namespacedDatabase) Close() error {
	// DB should be only closed by the root database.
	return nil
}
