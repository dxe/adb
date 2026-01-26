package persistence

import (
	"strconv"
	"strings"
)

type sqlQueryBuilder struct {
	base    string
	columns []string
	joins   []queryClause
	filters []queryClause
	orderBy []string
	limit   *int
}

type queryClause struct {
	sql  string
	args []any
}

func NewSqlQueryBuilder() *sqlQueryBuilder {
	return &sqlQueryBuilder{base: "FROM activists"}
}

func (b *sqlQueryBuilder) SelectColumn(column string) *sqlQueryBuilder {
	b.columns = append(b.columns, column)
	return b
}

func (b *sqlQueryBuilder) From(base string) *sqlQueryBuilder {
	if strings.TrimSpace(base) != "" {
		b.base = base
	}
	return b
}

func (b *sqlQueryBuilder) Join(clause string, args ...any) *sqlQueryBuilder {
	if strings.TrimSpace(clause) != "" {
		b.joins = append(b.joins, queryClause{sql: clause, args: args})
	}
	return b
}

func (b *sqlQueryBuilder) Where(clause string, args ...any) *sqlQueryBuilder {
	if strings.TrimSpace(clause) != "" {
		b.filters = append(b.filters, queryClause{sql: clause, args: args})
	}
	return b
}

func (b *sqlQueryBuilder) OrderBy(order ...string) *sqlQueryBuilder {
	b.orderBy = append(b.orderBy, order...)
	return b
}

func (b *sqlQueryBuilder) Limit(limit int) *sqlQueryBuilder {
	if limit < 0 {
		b.limit = nil
		return b
	}
	b.limit = &limit
	return b
}

func (b *sqlQueryBuilder) ToSQL() (string, []any) {
	columns := "*"
	if len(b.columns) > 0 {
		columns = strings.Join(b.columns, ", ")
	}

	var builder strings.Builder
	builder.WriteString("SELECT ")
	builder.WriteString(columns)
	builder.WriteString(" ")
	builder.WriteString(b.base)

	args := make([]any, 0)

	for _, join := range b.joins {
		builder.WriteString(" ")
		builder.WriteString(join.sql)
		args = append(args, join.args...)
	}

	if len(b.filters) > 0 {
		builder.WriteString(" WHERE ")
		parts := make([]string, 0, len(b.filters))
		for _, filter := range b.filters {
			parts = append(parts, filter.sql)
			args = append(args, filter.args...)
		}
		builder.WriteString(strings.Join(parts, " AND "))
	}

	if len(b.orderBy) > 0 {
		builder.WriteString(" ORDER BY ")
		builder.WriteString(strings.Join(b.orderBy, ", "))
	}

	if b.limit != nil {
		builder.WriteString(" LIMIT ")
		builder.WriteString(strconv.Itoa(*b.limit))
	}

	return builder.String(), args
}
