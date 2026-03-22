package persistence

import (
	"fmt"
	"strings"

	"github.com/dxe/adb/model"
)

// timestampGroup defines a set of data columns whose changes should update a shared timestamp column.
// Mirrors the IF(old <> new, NOW(), old_timestamp) pattern in userUpdateActivistQuery.
type timestampGroup struct {
	timestampCol string
	dataFields   []model.ActivistColumnName
	// nullSafeFields are fields within dataFields that require null-safe comparison (NOT col <=> ?).
	nullSafeFields map[model.ActivistColumnName]bool
}

var activistTimestampGroups = []timestampGroup{
	{timestampCol: "name_updated", dataFields: []model.ActivistColumnName{model.ColName, model.ColPreferredName}},
	{timestampCol: "email_updated", dataFields: []model.ActivistColumnName{model.ColEmail}},
	{timestampCol: "phone_updated", dataFields: []model.ActivistColumnName{model.ColPhone}},
	{timestampCol: "address_updated", dataFields: []model.ActivistColumnName{model.ColStreetAddress, model.ColCity, model.ColState}},
	{timestampCol: "location_updated", dataFields: []model.ActivistColumnName{model.ColStreetAddress, model.ColCity, model.ColState, model.ColLocation}, nullSafeFields: map[model.ActivistColumnName]bool{model.ColLocation: true}},
}

// getWriteColumn returns the bare SQL column name for UPDATE, or empty string if the field is not writable.
func getWriteColumn(fieldName model.ActivistColumnName) string {
	col, ok := simpleColumns[fieldName]
	if !ok || col.writeCol == "" {
		return ""
	}
	return col.writeCol
}

// BuildActivistPatchSQL generates an UPDATE statement for the given patch data.
// Each field name must be in the writable column allowlist.
//
// Timestamp SET clauses are emitted before data SET clauses so that MySQL's left-to-right
// evaluation sees the old data values in the IF() comparisons.
func BuildActivistPatchSQL(id int, patch model.ActivistPatchData) (string, []any, error) {
	if len(patch.Fields) == 0 {
		return "", nil, fmt.Errorf("no fields to update")
	}

	// Validate all fields are writable and build lookup structures.
	type resolvedField struct {
		writeCol string
		value    any
	}
	resolved := make(map[model.ActivistColumnName]resolvedField, len(patch.Fields))
	for _, f := range patch.Fields {
		col := getWriteColumn(f.Name)
		if col == "" {
			return "", nil, fmt.Errorf("field %q is not a writable activist column", f.Name)
		}
		resolved[f.Name] = resolvedField{writeCol: col, value: f.Value}
	}

	var setClauses []string
	var args []any

	// Emit timestamp SET clauses first (must precede data columns; see userUpdateActivistQuery).
	for _, tg := range activistTimestampGroups {
		var comparisons []string
		var compArgs []any
		for _, df := range tg.dataFields {
			rf, ok := resolved[df]
			if !ok {
				continue
			}
			if tg.nullSafeFields[df] {
				comparisons = append(comparisons, fmt.Sprintf("NOT %s <=> ?", rf.writeCol))
			} else {
				// Standard comparison works because these columns are NOT NULL in the schema.
				comparisons = append(comparisons, fmt.Sprintf("%s <> ?", rf.writeCol))
			}
			compArgs = append(compArgs, rf.value)
		}
		if len(comparisons) > 0 {
			setClauses = append(setClauses, fmt.Sprintf("%s = IF(%s, NOW(), %s)",
				tg.timestampCol,
				strings.Join(comparisons, " OR "),
				tg.timestampCol,
			))
			args = append(args, compArgs...)
		}
	}

	// Emit data SET clauses in the same order as the input.
	for _, f := range patch.Fields {
		rf := resolved[f.Name]
		setClauses = append(setClauses, fmt.Sprintf("%s = ?", rf.writeCol))
		args = append(args, rf.value)
	}

	sql := "UPDATE activists SET " + strings.Join(setClauses, ", ") + " WHERE id = ?"
	args = append(args, id)

	return sql, args, nil
}
