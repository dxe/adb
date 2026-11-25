package persistence

import (
	"testing"
	"time"

	"github.com/dxe/adb/model"
	"github.com/dxe/adb/testdb"
	"github.com/stretchr/testify/require"
)

func TestQueryActivists_Empty(t *testing.T) {
	db := testdb.NewDB()
	defer func() { _ = db.Close() }()
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
	defer func() { _ = db.Close() }()
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
	defer func() { _ = db.Close() }()
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

func TestQueryActivists_IdsFilter(t *testing.T) {
	db := testdb.NewDB()
	defer func() { _ = db.Close() }()
	repo := NewActivistRepository(db)

	alice, err := model.GetOrCreateActivist(db, "Alice", model.SFBayChapterIdDevTest)
	require.NoError(t, err)
	_, err = model.GetOrCreateActivist(db, "Bob", model.SFBayChapterIdDevTest)
	require.NoError(t, err)

	result, err := repo.QueryActivists(model.QueryActivistOptions{
		Shape: model.QueryActivistShape{
			Columns: []model.ActivistColumnName{model.ColName, model.ColID},
			Filters: model.QueryActivistFilters{Ids: []int{alice.ID}},
		},
	})
	require.NoError(t, err)
	require.Len(t, result.Activists, 1)
	require.Equal(t, "Alice", result.Activists[0].Name)
}

// The ids filter narrows the other filters rather than overriding them.
func TestQueryActivists_IdsFilterCombinedWithOtherFilters(t *testing.T) {
	db := testdb.NewDB()
	defer func() { _ = db.Close() }()
	repo := NewActivistRepository(db)

	alice, err := model.GetOrCreateActivist(db, "Alice", model.SFBayChapterIdDevTest)
	require.NoError(t, err)

	result, err := repo.QueryActivists(model.QueryActivistOptions{
		Shape: model.QueryActivistShape{
			Columns: []model.ActivistColumnName{model.ColName, model.ColID},
			Filters: model.QueryActivistFilters{
				Ids:  []int{alice.ID},
				Name: model.NameFilter{NameContains: "Bob"},
			},
		},
	})
	require.NoError(t, err)
	require.Empty(t, result.Activists)
}

func TestQueryActivists_InvalidColumn(t *testing.T) {
	db := testdb.NewDB()
	defer func() { _ = db.Close() }()
	repo := NewActivistRepository(db)

	_, err := repo.QueryActivists(model.QueryActivistOptions{
		Shape: model.QueryActivistShape{
			Columns: []model.ActivistColumnName{"not_a_real_column"},
		},
	})
	require.Error(t, err)
}

// Sets up activists whose first events differ in date and type:
//
//	Alice: Outreach on 2025-01-10, then Meeting
//	Bob:   Meeting on 2025-01-15, then Outreach
//	Carol: no events
//	Dan:   Outreach on 2024-06-01
func setUpFirstEventActivists(t *testing.T) *DBActivistRepository {
	t.Helper()
	db := testdb.NewDB()
	t.Cleanup(func() { _ = db.Close() })

	newActivist := func(name string) model.Activist {
		a, err := model.GetOrCreateActivist(db, name, model.SFBayChapterIdDevTest)
		require.NoError(t, err)
		return a
	}
	alice, bob, dan := newActivist("Alice"), newActivist("Bob"), newActivist("Dan")
	newActivist("Carol")

	newEvent := func(date, eventType string, attendees ...model.Activist) {
		d, err := time.Parse("2006-01-02", date)
		require.NoError(t, err)
		_, err = model.InsertUpdateEvent(db, model.Event{
			EventName:      eventType + " " + date,
			EventDate:      d,
			EventType:      model.EventType(eventType),
			ChapterID:      model.SFBayChapterIdDevTest,
			AddedAttendees: attendees,
		})
		require.NoError(t, err)
	}
	newEvent("2024-06-01", "Outreach", dan)
	newEvent("2025-01-10", "Outreach", alice)
	newEvent("2025-01-15", "Meeting", bob)
	newEvent("2025-02-01", "Meeting", alice)
	newEvent("2025-02-01", "Outreach", bob)

	return NewActivistRepository(db)
}

func queryActivistNames(t *testing.T, repo *DBActivistRepository, filters model.QueryActivistFilters) []string {
	t.Helper()
	require.NoError(t, filters.Validate())
	result, err := repo.QueryActivists(model.QueryActivistOptions{
		Shape: model.QueryActivistShape{
			Columns: []model.ActivistColumnName{model.ColName, model.ColID},
			Filters: filters,
		},
	})
	require.NoError(t, err)
	names := make([]string, len(result.Activists))
	for i, a := range result.Activists {
		names[i] = a.Name
	}
	return names
}

func mustDateOnly(t *testing.T, s string) model.DateOnly {
	t.Helper()
	d, err := time.Parse("2006-01-02", s)
	require.NoError(t, err)
	return model.DateOnly{Time: d}
}

func TestQueryActivists_FirstEventFilter(t *testing.T) {
	repo := setUpFirstEventActivists(t)

	tests := []struct {
		name   string
		filter model.FirstEventFilter
		want   []string
	}{
		{
			name: "date only",
			filter: model.FirstEventFilter{
				Date: model.DateRange{Gte: mustDateOnly(t, "2025-01-01")},
			},
			want: []string{"Alice", "Bob"},
		},
		{
			name: "date only, or null",
			filter: model.FirstEventFilter{
				Date:   model.DateRange{Gte: mustDateOnly(t, "2025-01-01")},
				OrNull: true,
			},
			want: []string{"Alice", "Bob", "Carol"},
		},
		{
			name: "include types",
			filter: model.FirstEventFilter{
				Types: model.EventTypeFilter{Mode: "include", Values: []string{"Outreach"}},
			},
			want: []string{"Alice", "Dan"},
		},
		{
			name: "exclude types",
			filter: model.FirstEventFilter{
				Types: model.EventTypeFilter{Mode: "exclude", Values: []string{"Outreach"}},
			},
			want: []string{"Bob"},
		},
		{
			name: "types and date match the same event",
			filter: model.FirstEventFilter{
				Date:  model.DateRange{Gte: mustDateOnly(t, "2025-01-01")},
				Types: model.EventTypeFilter{Mode: "include", Values: []string{"Outreach"}},
			},
			want: []string{"Alice"},
		},
		{
			name: "types and date, or null",
			filter: model.FirstEventFilter{
				Date:   model.DateRange{Gte: mustDateOnly(t, "2025-01-01")},
				OrNull: true,
				Types:  model.EventTypeFilter{Mode: "include", Values: []string{"Outreach"}},
			},
			want: []string{"Alice", "Carol"},
		},
		{
			name:   "or null only",
			filter: model.FirstEventFilter{OrNull: true},
			want:   []string{"Carol"},
		},
		{
			name: "types, or null with no date bounds",
			filter: model.FirstEventFilter{
				OrNull: true,
				Types:  model.EventTypeFilter{Mode: "exclude", Values: []string{"Outreach"}},
			},
			want: []string{"Bob", "Carol"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := queryActivistNames(t, repo, model.QueryActivistFilters{FirstEvent: tt.filter})
			require.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestQueryActivistFilters_ValidateFirstEventTypes(t *testing.T) {
	invalid := []model.EventTypeFilter{
		{Mode: "", Values: []string{"Outreach"}},
		{Mode: "both", Values: []string{"Outreach"}},
		{Mode: "include", Values: []string{" "}},
	}
	for _, types := range invalid {
		f := model.QueryActivistFilters{FirstEvent: model.FirstEventFilter{Types: types}}
		require.Error(t, f.Validate(), "%+v", types)
	}
}

func TestQueryActivistFilters_ValidateOrNullNeedsOpenRange(t *testing.T) {
	closed := model.DateRange{
		Gte: mustDateOnly(t, "2025-01-01"),
		Lt:  mustDateOnly(t, "2025-02-01"),
	}
	open := model.DateRange{Gte: mustDateOnly(t, "2025-01-01")}

	for _, f := range []model.QueryActivistFilters{
		{LastEvent: model.DateRangeFilter{Date: closed, OrNull: true}},
		{FirstEvent: model.FirstEventFilter{Date: closed, OrNull: true}},
	} {
		require.Error(t, f.Validate())
	}
	for _, f := range []model.QueryActivistFilters{
		{LastEvent: model.DateRangeFilter{Date: closed}},
		{LastEvent: model.DateRangeFilter{Date: open, OrNull: true}},
		{FirstEvent: model.FirstEventFilter{Date: open, OrNull: true}},
	} {
		require.NoError(t, f.Validate())
	}
}

func TestQueryActivists_FirstEventTypeColumn(t *testing.T) {
	repo := setUpFirstEventActivists(t)

	result, err := repo.QueryActivists(model.QueryActivistOptions{
		Shape: model.QueryActivistShape{
			Columns: []model.ActivistColumnName{model.ColName, model.ColID, model.ColFirstEventType},
		},
	})
	require.NoError(t, err)
	got := map[string]string{}
	for _, a := range result.Activists {
		got[a.Name] = a.FirstEventType
	}
	require.Equal(t, map[string]string{
		"Alice": "Outreach",
		"Bob":   "Meeting",
		"Carol": "",
		"Dan":   "Outreach",
	}, got)
}
