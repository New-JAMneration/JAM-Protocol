package store

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/database"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type Repository struct {
	db      database.Database
	encoder *types.Encoder
	decoder *types.Decoder
}

func NewRepository(db database.Database) *Repository {
	return &Repository{
		db:      db,
		encoder: types.NewEncoder(),
		decoder: types.NewDecoder(),
	}
}

func (repo *Repository) Database() database.Database {
	return repo.db
}

// NewBatch creates a new batch for batched writes.
// User of this method is responsible for committing and closing the batch.
func (repo *Repository) NewBatch() database.Batch {
	return repo.db.NewBatch()
}

// WithBatch executes the given function within a batch.
// It creates a new batch, passes it to the function, and commits the batch if the function returns no error.
// The batch is closed after the function execution.
func (repo *Repository) WithBatch(fn func(batch database.Batch) error) error {
	batch := repo.db.NewBatch()
	defer batch.Close()

	if err := fn(batch); err != nil {
		return err
	}
	return batch.Commit()
}
