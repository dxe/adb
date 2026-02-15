package persistence

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/dxe/adb/model"
	"github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
)

// sortSpec combines SQL expression with sort direction for ORDER BY and cursor WHERE.
type sortSpec struct {
	expr string
	desc bool
}

type activistPaginationCursor struct {
	// values of the last row of the previous page corresponding to the sort columns.
	// Required for this cursor pagination implementation.
	// Note: time.Time values serialize to RFC3339 strings, and numeric values become float64 after
	// JSON unmarshal. MySQL handles these conversions correctly in prepared statement comparisons.
	SortOffsetValues []any `json:"sort_values"`

	// ID of the activist in the last row of the previous page.
	IdOffset int `json:"activist_id"`
}

func buildPaginationCursor(sortColumns []model.ActivistSortColumn, lastRow model.ActivistExtra) (string, error) {
	if len(sortColumns) == 0 || sortColumns[len(sortColumns)-1].ColumnName != "id" {
		return "", errors.New("last sort column must be ID")
	}

	newCursor := activistPaginationCursor{
		IdOffset: lastRow.ID,
	}

	for _, sc := range sortColumns[:len(sortColumns)-1] { // exclude id tiebreaker
		sortValue, err := getSortValue(lastRow, sc.ColumnName)
		if err != nil {
			return "", fmt.Errorf("could not get sort value: %w", err)
		}
		newCursor.SortOffsetValues = append(newCursor.SortOffsetValues, sortValue)
	}

	cursorJSON, err := json.Marshal(newCursor)
	if err != nil {
		return "", fmt.Errorf("encoding pagination cursor: %w", err)
	}

	return base64.StdEncoding.EncodeToString(cursorJSON), nil
}

func parsePaginationCursor(str string) (activistPaginationCursor, error) {
	var cursor activistPaginationCursor
	decoded, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return activistPaginationCursor{}, fmt.Errorf("invalid pagination cursor: %w", err)
	}
	if err := json.Unmarshal(decoded, &cursor); err != nil {
		return activistPaginationCursor{}, fmt.Errorf("invalid pagination cursor: %w", err)
	}
	return cursor, nil
}

// buildCursorWhere constructs the seek-method WHERE clause for cursor pagination.
// For sort columns (A ASC, B DESC, id ASC) with cursor values (a, b, id), it generates:
//
//	(A > a) OR (A = a AND B < b) OR (A = a AND B = b AND id > id)
//
// NULL handling follows MySQL's default ordering (NULLs first for ASC, last for DESC).
func buildCursorWhere(sortSpecs []sortSpec, cursor activistPaginationCursor) queryClause {
	// Build cursor values list: sort offset values + id offset at the end.
	cursorValues := make([]any, len(cursor.SortOffsetValues)+1)
	copy(cursorValues, cursor.SortOffsetValues)
	cursorValues[len(cursorValues)-1] = cursor.IdOffset

	var disjuncts []string
	var allArgs []any

	for i := range sortSpecs {
		// Each disjunct: equality on all preceding columns, then "past" on column i.
		var parts []string
		var args []any

		// Equality prefix for columns 0..i-1
		for j := 0; j < i; j++ {
			colExpr := sortSpecs[j].expr
			val := cursorValues[j]
			if val == nil {
				parts = append(parts, fmt.Sprintf("%s IS NULL", colExpr))
			} else {
				parts = append(parts, fmt.Sprintf("%s = ?", colExpr))
				args = append(args, val)
			}
		}

		// "Past" condition for column i
		colExpr := sortSpecs[i].expr
		val := cursorValues[i]
		desc := sortSpecs[i].desc

		pastSQL, pastArgs := buildSeekPast(colExpr, val, desc)
		parts = append(parts, pastSQL)
		args = append(args, pastArgs...)

		disjuncts = append(disjuncts, "("+strings.Join(parts, " AND ")+")")
		allArgs = append(allArgs, args...)
	}

	return queryClause{
		sql:  "(" + strings.Join(disjuncts, " OR ") + ")",
		args: allArgs,
	}
}

// buildSeekPast returns the SQL condition and args for seeking past a cursor value on a single column.
// MySQL NULL ordering: NULLs first for ASC, NULLs last for DESC.
// colExpr is the raw SQL expression (not an alias).
func buildSeekPast(colExpr string, cursorVal any, desc bool) (string, []any) {
	if cursorVal == nil {
		if desc {
			// DESC + NULL cursor: NULLs are last in DESC, so no rows sort after this column's value.
			// Return FALSE so only deeper disjuncts (tiebreaker) can advance past the cursor.
			return "FALSE", nil
		}
		// ASC + NULL cursor: NULLs are first in ASC. Everything non-NULL comes after.
		return fmt.Sprintf("%s IS NOT NULL", colExpr), nil
	}

	if desc {
		// DESC + non-NULL cursor: rows with smaller value come next, plus NULLs (last in DESC).
		return fmt.Sprintf("(%s < ? OR %s IS NULL)", colExpr, colExpr), []any{cursorVal}
	}
	// ASC + non-NULL cursor: rows with larger value come next. (NULLs are first, already passed.)
	return fmt.Sprintf("%s > ?", colExpr), []any{cursorVal}
}

// getSortValue extracts the value for the given database column tag from an ActivistExtra using
// reflection on `db:` struct tags. Unwraps sql.Null* and mysql.NullTime types, returning
// nil for NULL values.
func getSortValue(activist model.ActivistExtra, colTag model.ActivistColumnName) (any, error) {
	target := string(colTag)
	v := reflect.ValueOf(activist)
	result := findDBTagValue(v, target)
	if result == notFound {
		return nil, fmt.Errorf("could not find field with tag: %v", colTag)
	}
	return result, nil
}

func findDBTagValue(v reflect.Value, tag string) any {
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldVal := v.Field(i)

		// Recurse into embedded structs.
		if field.Anonymous {
			if result := findDBTagValue(fieldVal, tag); result != notFound {
				return result
			}
			continue
		}

		dbTag := field.Tag.Get("db")
		if dbTag != tag {
			continue
		}

		return unwrapNullable(fieldVal)
	}
	return notFound
}

// sentinel value to distinguish "field not found" from a real nil.
var notFound = &struct{}{}

// unwrapNullable converts sql.Null* and mysql.NullTime to their underlying value or nil.
func unwrapNullable(v reflect.Value) any {
	iface := v.Interface()
	switch val := iface.(type) {
	case sql.NullString:
		if val.Valid {
			return val.String
		}
		return nil
	case sql.NullInt64:
		if val.Valid {
			return val.Int64
		}
		return nil
	case sql.NullFloat64:
		if val.Valid {
			return val.Float64
		}
		return nil
	case sql.NullBool:
		if val.Valid {
			return val.Bool
		}
		return nil
	case sql.NullTime:
		if val.Valid {
			return val.Time
		}
		return nil
	case mysql.NullTime:
		if val.Valid {
			return val.Time
		}
		return nil
	default:
		return iface
	}
}
