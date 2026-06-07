package main

import (
	"time"

	"github.com/dxe/adb/pkg/activists"
)

// newSingleEventSection: activists whose first (and only) event was 30–60 days
// ago — exactly one total event — to surface people who came once and haven't
// returned. Sorted by first event ascending.
func newSingleEventSection(now time.Time) reportSection {
	columns := []activists.ActivistColumnName{
		activists.ColName,
		activists.ColEmail,
		activists.ColPhone,
		activists.ColActivistLevel,
		activists.ColTotalEvents,
		activists.ColFirstEvent,
		activists.ColFirstEventName,
		activists.ColID,
	}

	return reportSection{
		title:   "Single event, 30-60 days ago",
		columns: columns,
		options: activists.QueryActivistOptions{
			Shape: activists.QueryActivistShape{
				Columns: columns,
				Filters: activists.QueryActivistFilters{
					ChapterId:     reportChapterID,
					IncludeHidden: false,
					FirstEvent: activists.DateRangeFilter{
						Gte: dateOnly(now.AddDate(0, 0, -60)),
						Lt:  dateOnly(now.AddDate(0, 0, -30)),
					},
					TotalEvents: activists.IntRangeFilter{
						Gte: intPtr(1),
						Lt:  intPtr(2),
					},
				},
				Sort: sortByFirstEventAsc(),
			},
		},
	}
}
