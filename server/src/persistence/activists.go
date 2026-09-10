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

// GetActivistAssignInfo fetches the chapter and hidden flag for each of the
// given activist ids in one query. Ids matching no activist are simply absent
// from the result.
func (r DBActivistRepository) GetActivistAssignInfo(activistIDs []int) ([]model.ActivistAssignInfo, error) {
	if len(activistIDs) == 0 {
		return nil, nil
	}
	query, args, err := sqlx.In(`SELECT id, chapter_id, hidden FROM activists WHERE id IN (?)`, activistIDs)
	if err != nil {
		return nil, fmt.Errorf("building assign info query for %d activists: %w", len(activistIDs), err)
	}
	var infos []model.ActivistAssignInfo
	if err := r.db.Select(&infos, r.db.Rebind(query), args...); err != nil {
		return nil, fmt.Errorf("fetching activist assign info: %w", err)
	}
	return infos, nil
}

// AssignActivists sets assigned_to on the given activists in a single UPDATE,
// so either all of them are reassigned or none are.
//
// The caller passes a distinct set of ids it has already checked to exist and
// to be visible, so the UPDATE must match every one of them. If it matches
// fewer — an activist was deleted or hidden since that check — the transaction
// is rolled back and nothing is reassigned.
func (r DBActivistRepository) AssignActivists(activistIDs []int, userID int) error {
	if len(activistIDs) == 0 {
		return nil
	}
	query, args, err := sqlx.In(`UPDATE activists SET assigned_to = ? WHERE id IN (?) AND hidden = 0`,
		userID, activistIDs)
	if err != nil {
		return fmt.Errorf("building bulk assign query for %d activists: %w", len(activistIDs), err)
	}

	tx, err := r.db.Beginx()
	if err != nil {
		return fmt.Errorf("starting bulk assign transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	result, err := tx.Exec(r.db.Rebind(query), args...)
	if err != nil {
		return fmt.Errorf("executing bulk assign: %w", err)
	}
	// The DSN sets clientFoundRows=true, so this counts rows matched rather
	// than rows changed: activists already assigned to userID still count.
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("reading bulk assign affected rows: %w", err)
	}
	if rows != int64(len(activistIDs)) {
		return fmt.Errorf("%w: bulk assign matched %d of %d activists", model.ErrNotFound, rows, len(activistIDs))
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing bulk assign: %w", err)
	}
	return nil
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
