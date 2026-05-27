package persistence

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dxe/adb/model"
	"github.com/jmoiron/sqlx"
)

func (r DBActivistRepository) DebugActivistQuery(options model.QueryActivistOptions, username string) (int64, error) {
	query, _, _, err := buildActivistsQueryFromShape(options.Shape)
	if err != nil {
		return 0, err
	}
	sqlStr, args := query.ToSQL()

	resolved := resolveSQLPlaceholders(sqlStr, args)

	explainOutput, err := runExplainAnalyze(r.db, sqlStr, args)
	if err != nil {
		return 0, fmt.Errorf("running EXPLAIN ANALYZE: %w", err)
	}

	res, err := r.db.Exec(
		`INSERT INTO debug_sql_queries (username, sql_query, explain_analyze_result) VALUES (?, ?, ?)`,
		username, resolved, explainOutput,
	)
	if err != nil {
		return 0, fmt.Errorf("inserting debug_sql_queries row: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("reading debug_sql_queries last insert id: %w", err)
	}
	return id, nil
}

// runExplainAnalyze runs EXPLAIN ANALYZE for the given query and returns the
// concatenated text output. MySQL 8.0+ returns the analyzed plan as one or
// more rows of a single TEXT column.
func runExplainAnalyze(db *sqlx.DB, sqlStr string, args []any) (string, error) {
	rows, err := db.Query("EXPLAIN ANALYZE "+sqlStr, args...)
	if err != nil {
		return "", err
	}
	defer func() { _ = rows.Close() }()

	var lines []string
	for rows.Next() {
		var line string
		if err := rows.Scan(&line); err != nil {
			return "", fmt.Errorf("scanning EXPLAIN ANALYZE row: %w", err)
		}
		lines = append(lines, line)
	}
	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("iterating EXPLAIN ANALYZE rows: %w", err)
	}
	return strings.Join(lines, "\n"), nil
}

// resolveSQLPlaceholders does a best-effort substitution of each `?`
// placeholder with its corresponding arg formatted as a SQL literal. The
// result is intended for human inspection only and must not be re-executed.
// Placeholders inside single- or double-quoted strings are left as-is, which
// matches how the query builder produces parameterized SQL (it never embeds
// `?` inside quoted literals).
func resolveSQLPlaceholders(sqlStr string, args []any) string {
	var b strings.Builder
	b.Grow(len(sqlStr))
	var quote byte // 0 when not inside a string, otherwise '\'' or '"'
	argIdx := 0
	for i := 0; i < len(sqlStr); i++ {
		c := sqlStr[i]
		if quote != 0 {
			b.WriteByte(c)
			if c == '\\' && i+1 < len(sqlStr) {
				b.WriteByte(sqlStr[i+1])
				i++
				continue
			}
			if c == quote {
				quote = 0
			}
			continue
		}
		if c == '\'' || c == '"' {
			quote = c
			b.WriteByte(c)
			continue
		}
		if c == '?' && argIdx < len(args) {
			b.WriteString(formatSQLLiteral(args[argIdx]))
			argIdx++
			continue
		}
		b.WriteByte(c)
	}
	return b.String()
}

func formatSQLLiteral(v any) string {
	if v == nil {
		return "NULL"
	}
	switch x := v.(type) {
	case string:
		return quoteString(x)
	case []byte:
		return quoteString(string(x))
	case bool:
		if x {
			return "1"
		}
		return "0"
	case int:
		return strconv.FormatInt(int64(x), 10)
	case int8:
		return strconv.FormatInt(int64(x), 10)
	case int16:
		return strconv.FormatInt(int64(x), 10)
	case int32:
		return strconv.FormatInt(int64(x), 10)
	case int64:
		return strconv.FormatInt(x, 10)
	case uint:
		return strconv.FormatUint(uint64(x), 10)
	case uint8:
		return strconv.FormatUint(uint64(x), 10)
	case uint16:
		return strconv.FormatUint(uint64(x), 10)
	case uint32:
		return strconv.FormatUint(uint64(x), 10)
	case uint64:
		return strconv.FormatUint(x, 10)
	case float32:
		return strconv.FormatFloat(float64(x), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	case time.Time:
		return quoteString(x.Format("2006-01-02 15:04:05"))
	case sql.NullString:
		if !x.Valid {
			return "NULL"
		}
		return quoteString(x.String)
	case sql.NullInt64:
		if !x.Valid {
			return "NULL"
		}
		return strconv.FormatInt(x.Int64, 10)
	case sql.NullInt32:
		if !x.Valid {
			return "NULL"
		}
		return strconv.FormatInt(int64(x.Int32), 10)
	case sql.NullFloat64:
		if !x.Valid {
			return "NULL"
		}
		return strconv.FormatFloat(x.Float64, 'f', -1, 64)
	case sql.NullBool:
		if !x.Valid {
			return "NULL"
		}
		if x.Bool {
			return "1"
		}
		return "0"
	case sql.NullTime:
		if !x.Valid {
			return "NULL"
		}
		return quoteString(x.Time.Format("2006-01-02 15:04:05"))
	default:
		return quoteString(fmt.Sprintf("%v", v))
	}
}

func quoteString(s string) string {
	var b strings.Builder
	b.Grow(len(s) + 2)
	b.WriteByte('\'')
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '\'':
			b.WriteString(`\'`)
		case '\\':
			b.WriteString(`\\`)
		case 0:
			b.WriteString(`\0`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case 0x1a:
			b.WriteString(`\Z`)
		default:
			b.WriteByte(c)
		}
	}
	b.WriteByte('\'')
	return b.String()
}
