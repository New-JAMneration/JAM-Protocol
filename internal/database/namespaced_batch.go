package database

// NewBatch creates a new batch from the root database and wraps it with the namespace.
func (db *namespacedDatabase) NewBatch() Batch {
	return db.BindBatch(db.inner.NewBatch())
}

// BindBatch wraps an existing batch with the namespace prefix.
// This allows multiple namespaced databases to share the same batch for atomic commits.
func (db *namespacedDatabase) BindBatch(batch Batch) Batch {
	return &namespacedBatch{
		inner: batch,
		ns:    db.ns,
	}
}

type namespacedBatch struct {
	inner Batch
	ns    []byte
}

func (b *namespacedBatch) Put(key, value []byte) error {
	return b.inner.Put(append(b.ns, key...), value)
}

func (b *namespacedBatch) Delete(key []byte) error {
	return b.inner.Delete(append(b.ns, key...))
}

func (b *namespacedBatch) DeleteRange(start, end []byte) error {
	return b.inner.DeleteRange(append(b.ns, start...), append(b.ns, end...))
}

func (b *namespacedBatch) Commit() error {
	return b.inner.Commit()
}
