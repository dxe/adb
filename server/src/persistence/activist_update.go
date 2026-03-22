package persistence

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/dxe/adb/model"
)

// BuildActivistPatchSQL generates an UPDATE statement for the given patch data.
func BuildActivistPatchSQL(id int, patch model.ActivistPatchData) (string, []any, error) {
	if len(patch.Fields) == 0 {
		return "", nil, fmt.Errorf("no fields to update")
	}

	type resolvedField struct {
		info  model.ActivistColumnInfo
		value any
	}
	resolved := make(map[model.ActivistColumnName]resolvedField, len(patch.Fields))
	for _, f := range patch.Fields {
		info, ok := model.ActivistColumns[f.Name]
		if !ok || info.DbCol == "" {
			return "", nil, fmt.Errorf("field %q is not a writable activist column", f.Name)
		}
		if _, exists := resolved[f.Name]; exists {
			return "", nil, fmt.Errorf("duplicate field %q in patch", f.Name)
		}
		resolved[f.Name] = resolvedField{info: info, value: f.Value}
	}

	// Collect every distinct timestamp that any resolved field bumps, sorted
	// for deterministic SQL output.
	tsSet := map[string]struct{}{}
	for _, rf := range resolved {
		for _, ts := range rf.info.BumpTimestamps {
			tsSet[ts] = struct{}{}
		}
	}
	timestampCols := make([]string, 0, len(tsSet))
	for ts := range tsSet {
		timestampCols = append(timestampCols, ts)
	}
	sort.Strings(timestampCols)

	var setClauses []string
	var args []any

	// Timestamp SET clauses are emitted before data SET clauses so that MySQL's
	// left-to-right evaluation sees the old data values in the IF() comparisons.

	// Emit timestamp SET clauses first. For each timestamp, walk patch.Fields
	// in input order so the OR-ed comparisons and their args are stable.
	for _, ts := range timestampCols {
		var comparisons []string
		var compArgs []any
		for _, f := range patch.Fields {
			rf := resolved[f.Name]
			if !slices.Contains(rf.info.BumpTimestamps, ts) {
				continue
			}
			comparisons = append(comparisons, fmt.Sprintf("NOT %s <=> ?", rf.info.DbCol))
			compArgs = append(compArgs, rf.value)
		}
		setClauses = append(setClauses, fmt.Sprintf("%s = IF(%s, NOW(), %s)",
			ts,
			strings.Join(comparisons, " OR "),
			ts,
		))
		args = append(args, compArgs...)
	}

	// Emit data SET clauses in input order.
	for _, f := range patch.Fields {
		rf := resolved[f.Name]
		setClauses = append(setClauses, fmt.Sprintf("%s = ?", rf.info.DbCol))
		args = append(args, rf.value)
	}

	sql := "UPDATE activists SET " + strings.Join(setClauses, ", ") + " WHERE id = ?"
	args = append(args, id)

	return sql, args, nil
}
