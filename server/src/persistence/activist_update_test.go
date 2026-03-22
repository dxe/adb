package persistence

import (
	"testing"

	"github.com/dxe/adb/model"
	"github.com/stretchr/testify/require"
)

func TestBuildActivistPatchSQL(t *testing.T) {
	t.Run("no fields returns error", func(t *testing.T) {
		_, _, err := BuildActivistPatchSQL(1, model.ActivistPatchData{})
		require.Error(t, err)
	})

	t.Run("unknown field returns error", func(t *testing.T) {
		patch := model.ActivistPatchData{Fields: []model.ActivistPatchField{
			{Name: "not_a_real_column", Value: "x"},
		}}
		_, _, err := BuildActivistPatchSQL(1, patch)
		require.Error(t, err)
	})

	t.Run("duplicate field returns error", func(t *testing.T) {
		patch := model.ActivistPatchData{Fields: []model.ActivistPatchField{
			{Name: model.ColFacebook, Value: "fb1"},
			{Name: model.ColFacebook, Value: "fb2"},
		}}
		_, _, err := BuildActivistPatchSQL(1, patch)
		require.Error(t, err)
	})

	t.Run("single field with no timestamp", func(t *testing.T) {
		patch := model.ActivistPatchData{Fields: []model.ActivistPatchField{
			{Name: model.ColFacebook, Value: "myfb"},
		}}
		sql, args, err := BuildActivistPatchSQL(42, patch)
		require.NoError(t, err)
		require.Equal(t, "UPDATE activists SET facebook = ? WHERE id = ?", sql)
		require.Equal(t, []any{"myfb", 42}, args)
	})

	t.Run("multiple fields with no timestamps preserve input order", func(t *testing.T) {
		patch := model.ActivistPatchData{Fields: []model.ActivistPatchField{
			{Name: model.ColFacebook, Value: "myfb"},
			{Name: model.ColPronouns, Value: "they/them"},
		}}
		sql, args, err := BuildActivistPatchSQL(7, patch)
		require.NoError(t, err)
		require.Equal(t, "UPDATE activists SET facebook = ?, pronouns = ? WHERE id = ?", sql)
		require.Equal(t, []any{"myfb", "they/them", 7}, args)
	})

	t.Run("single field with one timestamp", func(t *testing.T) {
		// ColEmail bumps email_updated.
		patch := model.ActivistPatchData{Fields: []model.ActivistPatchField{
			{Name: model.ColEmail, Value: "a@b.com"},
		}}
		sql, args, err := BuildActivistPatchSQL(5, patch)
		require.NoError(t, err)
		require.Equal(t,
			"UPDATE activists SET email_updated = IF(NOT email <=> ?, NOW(), email_updated), email = ? WHERE id = ?",
			sql,
		)
		require.Equal(t, []any{"a@b.com", "a@b.com", 5}, args)
	})

	t.Run("timestamp clause precedes data clause", func(t *testing.T) {
		// ColName bumps name_updated; ColFacebook has no timestamp.
		// Timestamp SET clause must come before both data SET clauses for
		// correct SQL evaluation order.
		patch := model.ActivistPatchData{Fields: []model.ActivistPatchField{
			{Name: model.ColFacebook, Value: "myfb"},
			{Name: model.ColName, Value: "Alice"},
		}}
		sql, args, err := BuildActivistPatchSQL(3, patch)
		require.NoError(t, err)
		require.Equal(t,
			"UPDATE activists SET name_updated = IF(NOT name <=> ?, NOW(), name_updated), facebook = ?, name = ? WHERE id = ?",
			sql,
		)
		// arg order: ts comparison value, then data values in input order, then id
		require.Equal(t, []any{"Alice", "myfb", "Alice", 3}, args)
	})

	t.Run("multiple fields sharing one timestamp", func(t *testing.T) {
		// ColName and ColPreferredName both bump name_updated.
		// The IF condition should OR their comparisons in input order.
		patch := model.ActivistPatchData{Fields: []model.ActivistPatchField{
			{Name: model.ColName, Value: "Alice"},
			{Name: model.ColPreferredName, Value: "Ali"},
		}}
		sql, args, err := BuildActivistPatchSQL(9, patch)
		require.NoError(t, err)
		require.Equal(t,
			"UPDATE activists SET name_updated = IF(NOT name <=> ? OR NOT preferred_name <=> ?, NOW(), name_updated), name = ?, preferred_name = ? WHERE id = ?",
			sql,
		)
		require.Equal(t, []any{"Alice", "Ali", "Alice", "Ali", 9}, args)
	})

	t.Run("single field with multiple timestamps sorted alphabetically", func(t *testing.T) {
		// ColStreetAddress bumps address_updated and location_updated.
		// address_updated < location_updated alphabetically, so address_updated comes first.
		patch := model.ActivistPatchData{Fields: []model.ActivistPatchField{
			{Name: model.ColStreetAddress, Value: "123 Main St"},
		}}
		sql, args, err := BuildActivistPatchSQL(11, patch)
		require.NoError(t, err)
		require.Equal(t,
			"UPDATE activists SET "+
				"address_updated = IF(NOT street_address <=> ?, NOW(), address_updated), "+
				"location_updated = IF(NOT street_address <=> ?, NOW(), location_updated), "+
				"street_address = ? WHERE id = ?",
			sql,
		)
		require.Equal(t, []any{"123 Main St", "123 Main St", "123 Main St", 11}, args)
	})

	t.Run("shared timestamp from multiple fields with mixed other timestamps", func(t *testing.T) {
		// ColStreetAddress bumps address_updated + location_updated.
		// ColCity also bumps address_updated + location_updated.
		// Each timestamp's IF should OR both fields' comparisons in input order.
		patch := model.ActivistPatchData{Fields: []model.ActivistPatchField{
			{Name: model.ColStreetAddress, Value: "123 Main St"},
			{Name: model.ColCity, Value: "Springfield"},
		}}
		sql, args, err := BuildActivistPatchSQL(2, patch)
		require.NoError(t, err)
		require.Equal(t,
			"UPDATE activists SET "+
				"address_updated = IF(NOT street_address <=> ? OR NOT city <=> ?, NOW(), address_updated), "+
				"location_updated = IF(NOT street_address <=> ? OR NOT city <=> ?, NOW(), location_updated), "+
				"street_address = ?, city = ? WHERE id = ?",
			sql,
		)
		require.Equal(t, []any{"123 Main St", "Springfield", "123 Main St", "Springfield", "123 Main St", "Springfield", 2}, args)
	})
}
