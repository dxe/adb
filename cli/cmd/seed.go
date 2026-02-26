package cmd

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dxe/adb/cli/internal/config"
	"github.com/dxe/adb/pkg/shared"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"

	_ "github.com/go-sql-driver/mysql"
)

// Seeded activists are generated from random names and share a 36-month pool of events (1-2 events per month) to create
// realistic overlap in attendance. Each activist is assigned a last event, then may attend prior events with
// probability based on a per-activist activeness value.
//
// Email/phone are generated with independent 10% blank rates, and activist level depends on whether the most recent
// event is within the last year.
//
// This is suitable for generating activists having varying:
// * "last event" dates,
// * total numbers of events
// * email/phone set or not set
// * MPI (i.e. months in the last year with event attendance, event type = "action") although this is only calculated
//   in production as of Feb 2026.

var seedChapterID int

const (
	seedActivistCount       = 200
	seedMonthsBack          = 36
	seedMinEventsPerMonth   = 1
	seedMaxEventsPerMonth   = 2
	seedBlankRate           = 0.10
	seedMinActivenessRate   = 0.10
	seedMaxActivenessRate   = 0.95
	seedRecentEventExponent = 2.0
)

func init() {
	rootCmd.AddCommand(seedCmd)
	seedCmd.AddCommand(seedActivistsCmd)
	seedActivistsCmd.Flags().IntVar(&seedChapterID, "chapter-id", shared.SFBayChapterIdDevTest, "Chapter ID to assign to seeded activists")
}

var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Commands for seeding the database with test data (non-production only)",
}

var seedActivistsCmd = &cobra.Command{
	Use:   "activists",
	Short: "Seed 200 activists with event history spread across the past 36 months",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireNotProd(); err != nil {
			return err
		}

		conn, err := sqlx.Connect("mysql", config.DBDataSource())
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		defer conn.Close()

		names := randomNames(seedActivistCount)
		now := time.Now()
		months := seedMonths(now, seedMonthsBack)
		sharedEvents, err := seedSharedEvents(conn, seedChapterID, months)
		if err != nil {
			return err
		}
		profiles := buildSeedProfiles(names, sharedEvents.All, seedChapterID, now)

		inserted := 0
		attendanceInserted := 0
		for _, profile := range profiles {
			res, err := conn.Exec(
				`INSERT IGNORE INTO activists (name, email, phone, chapter_id, activist_level) VALUES (?, ?, ?, ?, ?)`,
				profile.Name, profile.Email, profile.Phone, seedChapterID, profile.ActivistLevel,
			)
			if err != nil {
				return fmt.Errorf("failed to insert activist %q: %w", profile.Name, err)
			}
			if n, _ := res.RowsAffected(); n == 0 {
				continue // duplicate name in this chapter — skip
			}
			activistID, _ := res.LastInsertId()
			attendance := selectAttendanceEvents(profile, sharedEvents.ByMonth)

			for _, event := range attendance {
				if _, err := conn.Exec(
					`INSERT INTO event_attendance (activist_id, event_id) VALUES (?, ?)`,
					activistID, event.ID,
				); err != nil {
					return fmt.Errorf("failed to insert attendance for activist %q: %w", profile.Name, err)
				}
			}
			attendanceInserted += len(attendance)
			inserted++
		}

		fmt.Printf(
			"Seeded %d activists, %d events, and %d attendance rows in chapter %d\n",
			inserted, len(sharedEvents.All), attendanceInserted, seedChapterID,
		)
		return nil
	},
}

type seededEvent struct {
	ID         int64
	MonthIndex int
	Date       time.Time
}

// seededEvents groups shared events in both grouped and flattened forms.
type seededEvents struct {
	ByMonth [][]seededEvent
	All     []seededEvent
}

// seedProfile holds all randomized per-activist values before persistence.
type seedProfile struct {
	Name          string
	Email         string
	Phone         string
	LastEvent     seededEvent
	Activeness    float64
	ActivistLevel string
}

var seedFirstNames = []string{
	"Alex", "Blake", "Cameron", "Dana", "Ellis", "Finley", "Gray", "Harper",
	"Indigo", "Jordan", "Kendall", "Lee", "Morgan", "Nova", "Oakley", "Parker",
	"Quinn", "Reese", "Sage", "Taylor", "Uma", "Val", "Wren", "Xan", "Yael",
}

var seedLastNames = []string{
	"Abbott", "Barnes", "Chen", "Davis", "Evans", "Foster", "Garcia", "Hill",
	"Iyer", "Jones", "Kim", "Lopez", "Miller", "Nguyen", "Okafor", "Patel",
	"Quinn", "Reyes", "Smith", "Torres", "Ueda", "Vargas", "Walker", "Xu", "Young",
}

// randomNames returns n unique "First Last" names drawn from the seed pools.
// n must be ≤ len(seedFirstNames) * len(seedLastNames).
func randomNames(n int) []string {
	nPossibleNames := len(seedFirstNames) * len(seedLastNames)
	if n > nPossibleNames {
		panic("randomNames: n is too large; limit is " + strconv.Itoa(nPossibleNames))
	}
	all := make([]string, 0, nPossibleNames)
	for _, f := range seedFirstNames {
		for _, l := range seedLastNames {
			all = append(all, f+" "+l)
		}
	}
	rand.Shuffle(len(all), func(i, j int) { all[i], all[j] = all[j], all[i] })
	return all[:n]
}

// seedMonths returns `count` month-start timestamps ordered oldest to newest.
// The final element is the first day of `now`'s month in `now`'s location.
func seedMonths(now time.Time, count int) []time.Time {
	months := make([]time.Time, 0, count)
	currentMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	for i := count - 1; i >= 0; i-- {
		months = append(months, currentMonth.AddDate(0, -i, 0))
	}
	return months
}

// seedSharedEvents inserts 1-2 "Action" events per month and returns:
//  1. ByMonth: events grouped by month index from `months`
//  2. All: a flattened oldest-to-newest event list used for last-event sampling
//  3. error: non-nil if any INSERT fails
func seedSharedEvents(conn *sqlx.DB, chapterID int, months []time.Time) (seededEvents, error) {
	eventsByMonth := make([][]seededEvent, len(months))
	allEvents := make([]seededEvent, 0, len(months)*2)

	for monthIdx, monthStart := range months {
		eventCount := seedMinEventsPerMonth + rand.Intn(seedMaxEventsPerMonth-seedMinEventsPerMonth+1)
		days := randomDistinctDays(daysInMonth(monthStart), eventCount)
		sort.Ints(days)

		for eventIdx, day := range days {
			date := time.Date(monthStart.Year(), monthStart.Month(), day, 0, 0, 0, 0, monthStart.Location())
			dateString := date.Format("2006-01-02")
			name := fmt.Sprintf("Seed Event %s #%d", monthStart.Format("2006-01"), eventIdx+1)

			res, err := conn.Exec(
				`INSERT INTO events (name, date, event_type, chapter_id) VALUES (?, ?, 'Action', ?)`,
				name, dateString, chapterID,
			)
			if err != nil {
				return seededEvents{}, fmt.Errorf("failed to insert shared event %q: %w", name, err)
			}

			eventID, _ := res.LastInsertId()
			event := seededEvent{
				ID:         eventID,
				MonthIndex: monthIdx,
				Date:       date,
			}
			eventsByMonth[monthIdx] = append(eventsByMonth[monthIdx], event)
			allEvents = append(allEvents, event)
		}
	}

	return seededEvents{
		ByMonth: eventsByMonth,
		All:     allEvents,
	}, nil
}

// buildSeedProfiles returns one profile per name containing all randomized fields used to insert activists and
// determine attendance.
func buildSeedProfiles(names []string, allEvents []seededEvent, chapterID int, now time.Time) []seedProfile {
	profiles := make([]seedProfile, 0, len(names))
	for _, name := range names {
		lastEvent := allEvents[pickLastEventIndex(len(allEvents))]
		profiles = append(profiles, seedProfile{
			Name:          name,
			Email:         maybeBlank(seedEmail(name, chapterID)),
			Phone:         maybeBlank(seedPhone()),
			LastEvent:     lastEvent,
			Activeness:    randomActivenessRate(),
			ActivistLevel: seedActivistLevel(lastEvent.Date, now),
		})
	}
	return profiles
}

// selectAttendanceEvents returns attendance rows for a profile. It always includes the profile's last event, then
// samples prior months using one per-activist activeness value.
func selectAttendanceEvents(profile seedProfile, eventsByMonth [][]seededEvent) []seededEvent {
	attendance := make([]seededEvent, 0, profile.LastEvent.MonthIndex+1)
	attendance = append(attendance, profile.LastEvent)

	for monthIdx := 0; monthIdx < profile.LastEvent.MonthIndex; monthIdx++ {
		for _, event := range eventsByMonth[monthIdx] {
			if rand.Float64() >= profile.Activeness {
				continue
			}
			attendance = append(attendance, event)
		}
	}
	return attendance
}

// randomDistinctDays returns up to `count` unique day numbers in [1, daysInMonth].
// It returns an empty slice when count <= 0 and clamps count to daysInMonth.
func randomDistinctDays(daysInMonth, count int) []int {
	if count <= 0 {
		return []int{}
	}
	if count > daysInMonth {
		count = daysInMonth
	}

	seen := make(map[int]struct{}, count)
	days := make([]int, 0, count)
	for len(days) < count {
		day := 1 + rand.Intn(daysInMonth)
		if _, exists := seen[day]; exists {
			continue
		}
		seen[day] = struct{}{}
		days = append(days, day)
	}
	return days
}

// daysInMonth returns the number of calendar days in monthStart's month.
func daysInMonth(monthStart time.Time) int {
	return monthStart.AddDate(0, 1, -1).Day()
}

// pickLastEventIndex returns an index into `allEvents` and biases toward newer events.
// The returned value is always in [0, totalEvents-1], or 0 when totalEvents <= 1.
func pickLastEventIndex(totalEvents int) int {
	if totalEvents <= 1 {
		return 0
	}

	// Skew toward recent events by selecting smaller offsets more often.
	offset := int(math.Pow(rand.Float64(), seedRecentEventExponent) * float64(totalEvents))
	if offset >= totalEvents {
		offset = totalEvents - 1
	}
	return totalEvents - 1 - offset
}

// randomActivenessRate returns one per-activist attendance probability in
// [seedMinActivenessRate, seedMaxActivenessRate].
func randomActivenessRate() float64 {
	return seedMinActivenessRate + rand.Float64()*(seedMaxActivenessRate-seedMinActivenessRate)
}

// seedActivistLevel returns an activist level based on recency:
// - always "Supporter" if lastEventDate is older than 1 year
// - otherwise randomly "Organizer" (10%), "Chapter Member" (40%), or "Supporter" (50%)
func seedActivistLevel(lastEventDate time.Time, now time.Time) string {
	if lastEventDate.Before(now.AddDate(-1, 0, 0)) {
		return "Supporter"
	}

	roll := rand.Intn(10)
	switch {
	case roll == 0:
		return "Organizer"
	case roll < 5:
		return "Chapter Member"
	default:
		return "Supporter"
	}
}

// seedEmail returns a deterministic synthetic email for name/chapter pairs.
func seedEmail(name string, chapterID int) string {
	localPart := strings.ToLower(strings.ReplaceAll(name, " ", "."))
	return fmt.Sprintf("%s+ch%d@seed.example", localPart, chapterID)
}

// seedPhone returns a synthetic +1-415-555-XXXX phone number.
func seedPhone() string {
	return fmt.Sprintf("+1-415-555-%04d", rand.Intn(10000))
}

// maybeBlank returns an empty string ~10% of the time; otherwise it returns v.
func maybeBlank(v string) string {
	if rand.Float64() < seedBlankRate {
		return ""
	}
	return v
}
