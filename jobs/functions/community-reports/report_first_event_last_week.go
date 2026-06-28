package main

import (
	"time"

	"github.com/dxe/adb/pkg/activists"
)

// newFirstEventLastWeekSection: activists whose first event was within the last
// week, sorted by first event ascending.
func newFirstEventLastWeekSection(now time.Time) reportSection {
	columns := []activists.ActivistColumnName{
		activists.ColName,
		activists.ColEmail,
		activists.ColPhone,
		activists.ColActivistLevel,
		activists.ColTotalEvents,
		activists.ColFirstEvent,
		activists.ColFirstEventName,
		activists.ColLastEvent,
		activists.ColID,
	}

	return reportSection{
		title:   "First event in the last week",
		columns: columns,
		options: activists.QueryActivistOptions{
			Shape: activists.QueryActivistShape{
				Columns: columns,
				Filters: activists.QueryActivistFilters{
					ChapterId:     reportChapterID,
					IncludeHidden: false,
					FirstEvent: activists.DateRangeFilter{
						Gte: dateOnly(now.AddDate(0, 0, -7)),
					},
				},
				Sort: sortByFirstEventAsc(),
			},
		},
	}
}
