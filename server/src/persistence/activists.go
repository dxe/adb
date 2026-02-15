package persistence

import (
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

func (r DBActivistRepository) QueryActivists(options model.QueryActivistOptions) (model.QueryActivistResult, error) {
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
		query.SelectColumn(colSpec.selectExpr())
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

	// Normalize sort columns: default to name ASC, and append ID as tiebreaker
	// for deterministic ordering (required for cursor pagination).
	sortColumns := options.Sort.SortColumns
	if len(sortColumns) == 0 {
		sortColumns = []model.ActivistSortColumn{{ColumnName: "name"}}
	}
	if sortColumns[len(sortColumns)-1].ColumnName != "id" {
		sortColumns = append(sortColumns, model.ActivistSortColumn{ColumnName: "id"})
	}

	// Register joins needed by sort columns, add to SELECT if missing
	// (sorted columns must be selected for cursor pagination), and build sortSpecs.
	sortSpecs := make([]sortSpec, len(sortColumns))
	for i, sc := range sortColumns {
		colSpec := getColumnSpec(sc.ColumnName)
		if colSpec == nil {
			return model.QueryActivistResult{}, fmt.Errorf("invalid sort column: '%v'", sc.ColumnName)
		}
		for _, joinSpec := range colSpec.joins {
			registry.registerJoin(joinSpec)
		}
		if !slices.Contains(columns, sc.ColumnName) {
			query.SelectColumn(colSpec.selectExpr())
		}
		sortSpecs[i] = sortSpec{
			expr: colSpec.expr,
			desc: sc.Desc,
		}
	}

	for _, joinSQL := range registry.getJoins() {
		query.Join(joinSQL)
	}

	// Apply cursor seek condition if paginating.
	if len(options.After) > 0 {
		cursor, err := parsePaginationCursor(options.After)
		if err != nil {
			return model.QueryActivistResult{}, fmt.Errorf("parsing pagination cursor: %w", err)
		}

		numExpectedValues := len(sortColumns) - 1 // all sort columns except the id tiebreaker
		if len(cursor.SortOffsetValues) != numExpectedValues {
			return model.QueryActivistResult{}, fmt.Errorf("invalid pagination cursor: expected %d sort values, got %d", numExpectedValues, len(cursor.SortOffsetValues))
		}
		cursorWhere := buildCursorWhere(sortSpecs, cursor)
		query.Where(cursorWhere.sql, cursorWhere.args...)
	}

	for _, si := range sortSpecs {
		dir := "ASC"
		if si.desc {
			dir = "DESC"
		}
		query.OrderBy(fmt.Sprintf("%s %s", si.expr, dir))
	}

	// Fetch one extra row to detect if there are more results.
	limit := 50
	query.Limit(limit + 1)

	sqlStr, args := query.ToSQL()

	activists := []model.ActivistExtra{}
	if err := r.db.Select(&activists, sqlStr, args...); err != nil {
		return model.QueryActivistResult{}, fmt.Errorf("querying activists: %w", err)
	}

	// Determine if there are more pages and trim the extra row.
	hasMore := len(activists) > limit
	if hasMore {
		activists = activists[:limit]
	}

	nextCursor := ""
	if hasMore {
		var err error
		nextCursor, err = buildPaginationCursor(sortColumns, activists[len(activists)-1])
		if err != nil {
			return model.QueryActivistResult{}, fmt.Errorf("generating pagination cursor: %w", err)
		}
	}

	return model.QueryActivistResult{
		Activists: activists,
		Pagination: model.QueryActivistResultPagination{
			NextCursor: nextCursor,
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
