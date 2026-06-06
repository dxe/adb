package activists

import (
	"fmt"
	"slices"

	"github.com/dxe/adb/pkg/shared"

	"github.com/jmoiron/sqlx"
)

// Repository runs activist read queries directly against the database. It is
// shared between the ADB server and standalone jobs (e.g. AWS Lambdas) so that
// callers can run activist queries natively without going through the HTTP API.
type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

const activistTableAlias = "a"

func (r Repository) QueryActivists(options QueryActivistOptions) (QueryActivistResult, error) {
	query, sortColumns, sortSpecs, err := buildActivistsQueryFromShape(options.Shape)
	if err != nil {
		return QueryActivistResult{}, err
	}

	// Apply cursor seek condition if paginating.
	if len(options.After) > 0 {
		cursor, err := parsePaginationCursor(options.After)
		if err != nil {
			return QueryActivistResult{}, fmt.Errorf("parsing pagination cursor: %w", err)
		}

		numExpectedValues := len(sortColumns) - 1 // all sort columns except the id tiebreaker
		if len(cursor.SortOffsetValues) != numExpectedValues {
			return QueryActivistResult{}, shared.ValidationErrorf("invalid pagination cursor: expected %d sort values, got %d", numExpectedValues, len(cursor.SortOffsetValues))
		}
		cursorWhere := buildCursorWhere(sortSpecs, cursor)
		query.Where(cursorWhere.sql, cursorWhere.args...)
	}

	limit := 50
	// Fetch one extra row to detect if there are more results.
	query.Limit(limit + 1)

	sqlStr, args := query.ToSQL()

	activists := []ActivistExtra{}
	if err := r.db.Select(&activists, sqlStr, args...); err != nil {
		return QueryActivistResult{}, fmt.Errorf("querying activists: %w", err)
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
			return QueryActivistResult{}, fmt.Errorf("generating pagination cursor: %w", err)
		}
	}

	return QueryActivistResult{
		Activists: activists,
		Pagination: QueryActivistResultPagination{
			NextCursor: nextCursor,
		},
	}, nil
}

// StreamActivists executes a query for the given options and invokes fn for
// each matching row. Unlike QueryActivists, results are not paginated — all
// matching rows are streamed via the callback. If fn returns an error,
// iteration stops and that error is returned.
func (r Repository) StreamActivists(options QueryActivistOptions, fn func(ActivistExtra) error) error {
	query, _, _, err := buildActivistsQueryFromShape(options.Shape)
	if err != nil {
		return err
	}

	sqlStr, args := query.ToSQL()

	rows, err := r.db.Queryx(sqlStr, args...)
	if err != nil {
		return fmt.Errorf("querying activists: %w", err)
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var activist ActivistExtra
		if err := rows.StructScan(&activist); err != nil {
			return fmt.Errorf("scanning activist row: %w", err)
		}
		if err := fn(activist); err != nil {
			return err
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterating activist rows: %w", err)
	}
	return nil
}

func (r Repository) CountActivists(filters QueryActivistFilters) (int, error) {
	query := NewSqlQueryBuilder()
	query.From(fmt.Sprintf("FROM activists %s", activistTableAlias))
	query.SelectColumn("COUNT(*)")

	filterList := buildFiltersFromShape(QueryActivistShape{Filters: filters})
	registry := newJoinRegistry()
	for _, f := range filterList {
		for _, clause := range f.buildWhere() {
			query.Where(clause.sql, clause.args...)
		}
		for _, joinSpec := range f.getJoins() {
			registry.registerJoin(joinSpec)
		}
	}
	for _, joinSQL := range registry.getJoins() {
		query.Join(joinSQL)
	}

	sqlStr, args := query.ToSQL()
	var count int
	if err := r.db.Get(&count, sqlStr, args...); err != nil {
		return 0, fmt.Errorf("counting activists: %w", err)
	}
	return count, nil
}

// buildActivistsQueryFromShape applies the columns, filters, joins, and
// ORDER BY (with id tiebreaker) for an activist query.
func buildActivistsQueryFromShape(shape QueryActivistShape) (*sqlQueryBuilder, []ActivistSortColumn, []sortSpec, error) {
	query := NewSqlQueryBuilder()
	query.From(fmt.Sprintf("FROM activists %s", activistTableAlias))

	filters := buildFiltersFromShape(shape)

	// Ensure chapter_id is in columns if not filtering by chapter.
	columns := shape.Columns
	if shape.Filters.ChapterId == 0 && !slices.Contains(columns, ColChapterID) {
		columns = append(columns, ColChapterID)
	}

	registry := newJoinRegistry()

	for _, colName := range columns {
		colSpec := getColumnSpec(colName)
		if colSpec == nil {
			return nil, nil, nil, shared.ValidationErrorf("invalid column name: '%v'", colName)
		}
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
	sortColumns := shape.Sort.SortColumns
	if len(sortColumns) == 0 {
		sortColumns = []ActivistSortColumn{{ColumnName: ColName}}
	}
	// buildPaginationCursor requires ID to be the trailing sort column.
	sortColumns = ensureTrailingIdTiebreaker(sortColumns)

	// Register joins needed by sort columns, add to SELECT if missing
	// (sorted columns must be selected for cursor pagination), and build sortSpecs.
	sortSpecs := make([]sortSpec, len(sortColumns))
	for i, sc := range sortColumns {
		colSpec := getColumnSpec(sc.ColumnName)
		if colSpec == nil {
			return nil, nil, nil, shared.ValidationErrorf("invalid sort column: '%v'", sc.ColumnName)
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

	for _, si := range sortSpecs {
		dir := "ASC"
		if si.desc {
			dir = "DESC"
		}
		query.OrderBy(fmt.Sprintf("%s %s", si.expr, dir))
	}

	return query, sortColumns, sortSpecs, nil
}

// ensureTrailingIdTiebreaker ensures sort columns end with ColID,
// removing any columns after it if found (which has no actual effect on
// sorting since ID is unique), or adding it to the end otherwise.
func ensureTrailingIdTiebreaker(sortColumns []ActivistSortColumn) []ActivistSortColumn {
	for i, sc := range sortColumns {
		if sc.ColumnName == ColID {
			return sortColumns[:i+1]
		}
	}
	return append(sortColumns, ActivistSortColumn{ColumnName: ColID})
}

func buildFiltersFromShape(shape QueryActivistShape) []filter {
	var result []filter
	f := shape.Filters

	if f.ChapterId != 0 {
		result = append(result, &chapterFilter{ChapterId: f.ChapterId})
	}

	if !f.Name.IsEmpty() {
		result = append(result, &nameFilter{NameContains: f.Name.NameContains})
	}

	if !f.LastEvent.IsEmpty() {
		lastEventJoin := lastEventSubqueryJoin
		result = append(result, &dateRangeFilter{
			filter: f.LastEvent,
			join:   &lastEventJoin,
			expr:   fmt.Sprintf("%s.last_event_date", lastEventSubqueryJoin.Key),
		})
	}

	if !f.IncludeHidden {
		result = append(result, &hiddenFilter{})
	}

	if !f.ActivistLevel.IsEmpty() {
		result = append(result, &activistLevelFilter{
			Mode:   f.ActivistLevel.Mode,
			Values: f.ActivistLevel.Values,
		})
	}

	if !f.InterestDate.IsEmpty() {
		result = append(result, &dateRangeFilter{
			filter: f.InterestDate,
			expr:   fmt.Sprintf("%s.interest_date", activistTableAlias),
		})
	}

	if !f.FirstEvent.IsEmpty() {
		firstEventJoin := firstEventSubqueryJoin
		result = append(result, &dateRangeFilter{
			filter: f.FirstEvent,
			join:   &firstEventJoin,
			expr:   fmt.Sprintf("%s.first_event_date", firstEventSubqueryJoin.Key),
		})
	}

	if !f.TotalEvents.IsEmpty() {
		totalEventsJoin := totalEventsSubqueryJoin
		result = append(result, &intRangeFilter{
			filter: f.TotalEvents,
			join:   &totalEventsJoin,
			expr:   fmt.Sprintf("%s.event_count", totalEventsSubqueryJoin.Key),
		})
	}

	if !f.TotalInteractions.IsEmpty() {
		interactionsJoin := interactionsSubqueryJoin
		result = append(result, &intRangeFilter{
			filter: f.TotalInteractions,
			join:   &interactionsJoin,
			expr:   fmt.Sprintf("%s.interaction_count", interactionsSubqueryJoin.Key),
		})
	}

	if !f.Source.IsEmpty() {
		result = append(result, &sourceFilter{
			ContainsAny:    f.Source.ContainsAny,
			NotContainsAny: f.Source.NotContainsAny,
		})
	}

	if !f.Training.IsEmpty() {
		result = append(result, &trainingFilter{
			Completed:    f.Training.Completed,
			NotCompleted: f.Training.NotCompleted,
		})
	}

	if f.AssignedTo != 0 {
		result = append(result, &assignedToFilter{AssignedTo: f.AssignedTo})
	}

	if f.Followups != "" {
		result = append(result, &followupsFilter{Followups: f.Followups})
	}

	if f.Prospect != "" {
		result = append(result, &prospectFilter{Prospect: f.Prospect})
	}

	return result
}

// GetActivistEventData loads first/last/total event aggregates for a single
// activist by ID.
func (a Activist) GetActivistEventData(db *sqlx.DB) (ActivistEventData, error) {
	query := `
SELECT
  MIN(e.date) AS first_event,
  MAX(e.date) AS last_event,
  COUNT(*) as total_events
FROM events e
JOIN event_attendance
  ON event_attendance.event_id = e.id
WHERE
  event_attendance.activist_id = ?
`
	var data ActivistEventData
	if err := db.Get(&data, query, a.ID); err != nil {
		return ActivistEventData{}, fmt.Errorf("failed to get activist event data: %w", err)
	}
	return data, nil
}
