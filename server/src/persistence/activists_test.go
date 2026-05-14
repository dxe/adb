package persistence

import (
	"testing"

	"github.com/dxe/adb/model"
	"github.com/dxe/adb/testdb"
	"github.com/stretchr/testify/require"
)

func TestQueryActivists_Empty(t *testing.T) {
	db := testdb.NewDB()
	defer db.Close()
	repo := NewActivistRepository(db)

	result, err := repo.QueryActivists(model.QueryActivistOptions{
		Shape: model.QueryActivistShape{
			Columns: []model.ActivistColumnName{model.ColName, model.ColID},
		},
	})
	require.NoError(t, err)
	require.Empty(t, result.Activists)
	require.Empty(t, result.Pagination.NextCursor)
}

func TestQueryActivists_ReturnsActivists(t *testing.T) {
	db := testdb.NewDB()
	defer db.Close()
	repo := NewActivistRepository(db)

	_, err := model.GetOrCreateActivist(db, "Alice", model.SFBayChapterIdDevTest)
	require.NoError(t, err)
	_, err = model.GetOrCreateActivist(db, "Bob", model.SFBayChapterIdDevTest)
	require.NoError(t, err)

	result, err := repo.QueryActivists(model.QueryActivistOptions{
		Shape: model.QueryActivistShape{
			Columns: []model.ActivistColumnName{model.ColName, model.ColID},
		},
	})
	require.NoError(t, err)
	require.Len(t, result.Activists, 2)
}

func TestQueryActivists_NameFilter(t *testing.T) {
	db := testdb.NewDB()
	defer db.Close()
	repo := NewActivistRepository(db)

	_, err := model.GetOrCreateActivist(db, "Alice", model.SFBayChapterIdDevTest)
	require.NoError(t, err)
	_, err = model.GetOrCreateActivist(db, "Bob", model.SFBayChapterIdDevTest)
	require.NoError(t, err)

	result, err := repo.QueryActivists(model.QueryActivistOptions{
		Shape: model.QueryActivistShape{
			Columns: []model.ActivistColumnName{model.ColName, model.ColID},
			Filters: model.QueryActivistFilters{
				Name: model.NameFilter{NameContains: "Ali"},
			},
		},
	})
	require.NoError(t, err)
	require.Len(t, result.Activists, 1)
	require.Equal(t, "Alice", result.Activists[0].Name)
}

func TestQueryActivists_InvalidColumn(t *testing.T) {
	db := testdb.NewDB()
	defer db.Close()
	repo := NewActivistRepository(db)

	_, err := repo.QueryActivists(model.QueryActivistOptions{
		Shape: model.QueryActivistShape{
			Columns: []model.ActivistColumnName{"not_a_real_column"},
		},
	})
	require.Error(t, err)
}
