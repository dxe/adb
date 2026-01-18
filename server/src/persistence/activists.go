package persistence

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/dxe/adb/model"
	"github.com/jmoiron/sqlx"
)

type DBActivistRepository struct {
	db *sqlx.DB
}

func NewActivistRepository(db *sqlx.DB) *DBActivistRepository {
	return &DBActivistRepository{db: db}
}

type activistPaginationCursor struct {
	// values of the last row of the previous page corresponding to the sort columns.
	// Required for this cursor pagination implementation.
	SortOffsetValues []any `json:"sort_values"`

	// ID of the activist in the last row of the previous page.
	IdOffset int `json:"activist_id"`
}

func (r DBActivistRepository) QueryActivists(options model.QueryActivistOptions) (model.QueryActivistResult, error) {
	var cursor activistPaginationCursor
	if len(options.After) > 0 {
		decoded, err := base64.StdEncoding.DecodeString(options.After)
		if err != nil {
			return model.QueryActivistResult{}, fmt.Errorf("invalid pagination cursor: %w", err)
		}
		if err := json.Unmarshal(decoded, &cursor); err != nil {
			return model.QueryActivistResult{}, fmt.Errorf("invalid pagination cursor: %w", err)
		}
	}
	_ = cursor

	// TODO: implement this function
	return model.QueryActivistResult{}, nil
}
