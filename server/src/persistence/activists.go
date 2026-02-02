package persistence

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/dxe/adb/model"
	"github.com/jmoiron/sqlx"
)

type DBActivistRepository struct {
	db *sqlx.DB
}

func NewActivistRepository(db *sqlx.DB) *DBActivistRepository {
	return &DBActivistRepository{db: db}
}

const activistTableAlias = "a"

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
	// TODO: use cursor value
	_ = cursor

	query := NewSqlQueryBuilder()
	query.From(fmt.Sprintf("FROM activists %s", activistTableAlias))

	// Convert options to filters and columns
	filters := buildFiltersFromOptions(options)

	// Ensure chapter_id is in columns if not filtering by chapter
	columns := options.Columns
	if options.Filters.ChapterId == 0 && !slices.Contains(columns, "chapter_id") {
		columns = append(columns, "chapter_id")
	}

	registry := newJoinRegistry()

	columnSpecs := []*activistColumn{}
	for _, colName := range columns {
		colSpec := getColumnSpec(colName)
		if colSpec == nil {
			return model.QueryActivistResult{}, fmt.Errorf("invalid column name: '%v'", colName)
		}
		columnSpecs = append(columnSpecs, colSpec)
		query.SelectColumn(colSpec.sql)
		for _, joinSpec := range colSpec.joins {
			registry.registerJoin(joinSpec)
		}
	}

	for _, filter := range filters {
		for _, whereClause := range filter.buildWhere() {
			query.Where(whereClause.sql, whereClause.args...)
		}
		for _, joinSpec := range filter.getJoins() {
			registry.registerJoin(joinSpec)
		}
	}

	for _, joinSQL := range registry.getJoins() {
		query.Join(joinSQL)
	}

	// TODO: Apply sort options from options.Sort
	// TODO: Increase pagination limit for prod
	limit := 20
	query.Limit(limit)

	sqlStr, args := query.ToSQL()

	activists := []model.ActivistExtra{}
	if err := r.db.Select(&activists, sqlStr, args...); err != nil {
		return model.QueryActivistResult{}, fmt.Errorf("querying activists: %w", err)
	}

	return model.QueryActivistResult{
		Activists: activists,
		Pagination: model.QueryActivistResultPagination{
			// TODO: set NextCursor if there are more results
			NextCursor: "",
		},
	}, nil
}

func buildFiltersFromOptions(options model.QueryActivistOptions) []filter {
	var filters []filter

	if options.Filters.ChapterId != 0 {
		filters = append(filters, &chapterFilter{ChapterId: options.Filters.ChapterId})
	}

	if options.Filters.Name.NameContains != "" {
		filters = append(filters, &nameFilter{NameContains: options.Filters.Name.NameContains})
	}

	if !options.Filters.LastEvent.LastEventLt.IsZero() || !options.Filters.LastEvent.LastEventGte.IsZero() {
		filters = append(filters, &lastEventFilter{
			LastEventGte: options.Filters.LastEvent.LastEventGte,
			LastEventLt:  options.Filters.LastEvent.LastEventLt,
		})
	}

	if !options.Filters.IncludeHidden {
		filters = append(filters, &hiddenFilter{})
	}

	return filters
}
