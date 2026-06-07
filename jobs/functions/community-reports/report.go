package main

import (
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/dxe/adb/pkg/activists"
)

// reportSection is one query and its rendering metadata. Each section becomes a
// titled table in the email; columns drives both the SELECT and the table
// columns so the two stay in sync.
type reportSection struct {
	title   string
	columns []activists.ActivistColumnName
	options activists.QueryActivistOptions
}

func sortByFirstEventAsc() activists.ActivistSortOptions {
	return activists.ActivistSortOptions{
		SortColumns: []activists.ActivistSortColumn{
			{ColumnName: activists.ColFirstEvent, Desc: false},
		},
	}
}

// runSection executes a section's query, returning all matching rows (not
// paginated).
func runSection(repo *activists.Repository, section reportSection) ([]activists.ActivistExtra, error) {
	var rows []activists.ActivistExtra
	if err := repo.StreamActivists(section.options, func(a activists.ActivistExtra) error {
		rows = append(rows, a)
		return nil
	}); err != nil {
		return nil, err
	}
	return rows, nil
}

// renderSection renders one section as an HTML heading plus a table of its rows.
func renderSection(section reportSection, rows []activists.ActivistExtra) string {
	var b strings.Builder
	fmt.Fprintf(&b, "<h2>%s</h2>", html.EscapeString(section.title))
	fmt.Fprintf(&b, "<p>%d activist(s).</p>", len(rows))

	if len(rows) == 0 {
		return b.String()
	}

	b.WriteString(`<table border="1" cellpadding="4" cellspacing="0" style="border-collapse:collapse">`)
	b.WriteString("<tr>")
	for _, col := range section.columns {
		// ID is only used to link names to the ADB website.
		if col == activists.ColID {
			continue
		}
		fmt.Fprintf(&b, "<th>%s</th>", html.EscapeString(string(col)))
	}
	b.WriteString("</tr>")

	for _, a := range rows {
		b.WriteString("<tr>")
		for _, col := range section.columns {
			if col == activists.ColID {
				continue
			}
			fmt.Fprintf(&b, "<td>%s</td>", renderCell(a, col))
		}
		b.WriteString("</tr>")
	}
	b.WriteString("</table>")

	return b.String()
}

// activistProfileURL is the ADB profile page for an activist, used to link names.
const activistProfileURL = "https://adb.dxe.io/v2/activists/"

// renderCell returns the HTML for a single table cell. The name column links to
// the activist's profile page; all other columns are plain escaped text.
func renderCell(a activists.ActivistExtra, col activists.ActivistColumnName) string {
	if col == activists.ColName {
		return fmt.Sprintf(`<a href="%s%d">%s</a>`,
			activistProfileURL, a.ID, html.EscapeString(a.Name))
	}
	return html.EscapeString(formatColumn(a, col))
}

// formatColumn renders a single activist column value as a string.
func formatColumn(a activists.ActivistExtra, col activists.ActivistColumnName) string {
	switch col {
	case activists.ColName:
		return a.Name
	case activists.ColEmail:
		return a.Email
	case activists.ColPhone:
		return a.Phone
	case activists.ColActivistLevel:
		return a.ActivistLevel
	case activists.ColTotalEvents:
		return fmt.Sprintf("%d", a.TotalEvents)
	case activists.ColFirstEvent:
		if a.FirstEvent.Valid {
			return a.FirstEvent.Time.Format("2006-01-02")
		}
		return ""
	case activists.ColFirstEventName:
		return a.FirstEventName
	case activists.ColLastEvent:
		if a.LastEvent.Valid {
			return a.LastEvent.Time.Format("2006-01-02")
		}
		return ""
	case activists.ColID:
		return fmt.Sprintf("%d", a.ID)
	default:
		return ""
	}
}

// dateOnly truncates t to a YYYY-MM-DD value in UTC, matching the wire format
// the frontend sends for date filters.
func dateOnly(t time.Time) activists.DateOnly {
	t = t.UTC()
	return activists.DateOnly{
		Time: time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC),
	}
}

func intPtr(v int) *int {
	return &v
}
