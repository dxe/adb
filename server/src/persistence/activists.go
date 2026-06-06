package persistence

import (
	"fmt"

	"github.com/dxe/adb/model"
	"github.com/dxe/adb/pkg/activists"

	"github.com/jmoiron/sqlx"
)

// DBActivistRepository implements model.ActivistRepository. The read query
// path (QueryActivists, StreamActivists, CountActivists, DebugActivistQuery)
// is provided by the shared *activists.Repository (embedded). The write path
// (PatchActivist) lives here in the server.
type DBActivistRepository struct {
	*activists.Repository
	db *sqlx.DB
}

func NewActivistRepository(db *sqlx.DB) *DBActivistRepository {
	return &DBActivistRepository{
		Repository: activists.NewRepository(db),
		db:         db,
	}
}

func (r DBActivistRepository) PatchActivist(id int, patch model.ActivistPatchData) error {
	sqlStr, args, err := BuildActivistPatchSQL(id, patch)
	if err != nil {
		return fmt.Errorf("building patch SQL: %w", err)
	}
	result, err := r.db.Exec(sqlStr, args...)
	if err != nil {
		return fmt.Errorf("executing activist patch: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("reading patch affected rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("%w: activist with id %d not found", model.ErrNotFound, id)
	}
	return nil
}
