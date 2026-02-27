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
	var result []filter
	f := options.Filters

	if f.ChapterId != 0 {
		result = append(result, &chapterFilter{ChapterId: f.ChapterId})
	}

	if f.Name.NameContains != "" {
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
			Include: f.ActivistLevel.Include,
			Exclude: f.ActivistLevel.Exclude,
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
