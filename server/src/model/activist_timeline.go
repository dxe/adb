package model

import (
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/jmoiron/sqlx"
)

// ActivistTimelineMaxItems caps how many timeline items a single request can
// return. Long-standing activists have attended many thousands of events, and
// the timeline is only ever read as "what happened recently", so the newest
// items are kept and the rest are dropped.
const ActivistTimelineMaxItems = 200

// Timeline item types. These are the discriminator sent to the client.
const (
	TimelineItemTypeEvent       = "event"
	TimelineItemTypeInteraction = "interaction"
)

// activistTimelineLocation is the timezone the timeline is ordered in. Event
// attendance is stored as a bare calendar date while an interaction is stored
// as an instant, so the two are only comparable once a zone is picked. This
// matches the zone the activist UI formats dates in (see
// ACTIVISTS_TIME_ZONE in frontend-v2), which keeps the order the server sends
// consistent with the dates the reader sees.
var activistTimelineLocation = func() *time.Location {
	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		// Only possible if the tzdata is missing from the image, in which case
		// ordering degrades to UTC rather than the whole endpoint failing.
		return time.UTC
	}
	return loc
}()

// ActivistTimelineEvent is the payload of a TimelineItemTypeEvent item.
type ActivistTimelineEvent struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// ActivistTimelineInteraction is the payload of a TimelineItemTypeInteraction
// item.
type ActivistTimelineInteraction struct {
	Method   string `json:"method"`
	Outcome  string `json:"outcome"`
	Notes    string `json:"notes"`
	UserName string `json:"user_name"`
}

// ActivistTimelineItem is one entry in an activist's engagement timeline. The
// fields above the payloads are common to every item; exactly one payload is
// non-nil, and Type says which. The payloads are nested rather than flattened
// so a variant's fields cannot be half-populated or collide with another
// variant's, and so each payload's own fields are always emitted (an
// interaction with no notes sends "notes": "", not a missing key).
type ActivistTimelineItem struct {
	Type string `json:"type"`
	ID   int    `json:"id"`
	// Date is the calendar date (YYYY-MM-DD) the item is filed under. It is
	// what the reader sees, so the client should not re-derive it from
	// Timestamp.
	Date string `json:"date"`
	// Timestamp is the instant the item is ordered by, RFC3339 in UTC. It is
	// set on every item so the client has one field to format regardless of
	// variant.
	Timestamp string `json:"timestamp"`
	// HasTime reports whether Timestamp's time of day is real. An attended
	// event with no recorded start time is given a placeholder instant purely
	// so it sorts sensibly (see activistTimelineEventInstant); HasTime is
	// false there, and the client must not present that time of day as the
	// event's own.
	HasTime bool `json:"has_time"`

	Event       *ActivistTimelineEvent       `json:"event,omitempty"`
	Interaction *ActivistTimelineInteraction `json:"interaction,omitempty"`

	// sortKey is the instant the item is ordered by. Not serialized; Timestamp
	// is its wire form.
	sortKey time.Time
}

// ActivistTimeline is an activist's engagement history, newest first.
type ActivistTimeline struct {
	Items []ActivistTimelineItem `json:"items"`
	// Truncated reports that older items exist beyond the ones returned, so
	// the client can say the list is capped rather than imply it is complete.
	Truncated bool `json:"truncated"`
}

// GetActivistTimeline returns the activist's event attendance and interactions
// merged into a single newest-first timeline, capped at
// ActivistTimelineMaxItems.
//
// The caller must have organizer access and must either be an admin or belong
// to the activist's chapter.
func GetActivistTimeline(db *sqlx.DB, authedUser ADBUser, activistID int) (ActivistTimeline, error) {
	if activistID <= 0 {
		return ActivistTimeline{}, ValidationErrorf("activist id must be supplied")
	}
	if !UserHasOrganizerAccess(authedUser) {
		return ActivistTimeline{}, ValidationErrorf("lacking permission to view activist engagement")
	}

	var existing struct {
		ChapterID int `db:"chapter_id"`
	}
	if err := db.Get(&existing, `SELECT chapter_id FROM activists WHERE id = ?`, activistID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ActivistTimeline{}, fmt.Errorf("%w: activist with id %d not found", ErrNotFound, activistID)
		}
		return ActivistTimeline{}, fmt.Errorf("failed to look up activist %d: %w", activistID, err)
	}
	if err := CheckChapterAccess(authedUser, existing.ChapterID); err != nil {
		return ActivistTimeline{}, err
	}

	// Each side is capped before merging: taking the newest N of each
	// guarantees the newest N of the union is among them, so the merged result
	// is exact without reading either table in full. One extra row per side
	// distinguishes "exactly at the cap" from "there is more".
	fetchLimit := ActivistTimelineMaxItems + 1
	events, err := getActivistTimelineEvents(db, activistID, fetchLimit)
	if err != nil {
		return ActivistTimeline{}, err
	}
	interactions, err := getActivistTimelineInteractions(db, activistID, fetchLimit)
	if err != nil {
		return ActivistTimeline{}, err
	}

	items := make([]ActivistTimelineItem, 0, len(events)+len(interactions))
	items = append(items, events...)
	items = append(items, interactions...)

	sort.SliceStable(items, func(i, j int) bool {
		if !items[i].sortKey.Equal(items[j].sortKey) {
			return items[i].sortKey.After(items[j].sortKey)
		}
		// Stable tiebreak so pagination-free reloads keep the same order. Ties
		// are common: every event lacking a start time lands on the same noon.
		if items[i].Type != items[j].Type {
			return items[i].Type < items[j].Type
		}
		return items[i].ID > items[j].ID
	})

	truncated := len(items) > ActivistTimelineMaxItems
	if truncated {
		items = items[:ActivistTimelineMaxItems]
	}
	return ActivistTimeline{Items: items, Truncated: truncated}, nil
}

type timelineEventRow struct {
	ID        int            `db:"id"`
	Name      string         `db:"name"`
	EventType string         `db:"event_type"`
	Date      time.Time      `db:"date"`
	StartTime sql.NullString `db:"start_time"`
	Timezone  string         `db:"timezone"`
}

// activistTimelineEventInstant returns the instant an attended event is
// ordered by, and whether that instant's time of day is the event's own.
//
// An event records a bare calendar date plus an optional wall-clock start time
// in its own zone. With no start time there is nothing to place it within the
// day, so it is treated as noon as an approximation. Noon is a sort key, not
// a fact about the event, hence the second return value.
func activistTimelineEventInstant(date time.Time, startTime sql.NullString, timezone string) (time.Time, bool) {
	if startTime.Valid {
		loc := activistTimelineLocation
		if timezone != "" {
			// An unloadable zone falls back rather than dropping the event.
			if eventLoc, err := time.LoadLocation(timezone); err == nil {
				loc = eventLoc
			}
		}
		// MySQL renders TIME as "15:04:05", but tolerate a seconds-less value.
		for _, layout := range []string{"15:04:05", "15:04"} {
			if t, err := time.Parse(layout, startTime.String); err == nil {
				return time.Date(date.Year(), date.Month(), date.Day(),
					t.Hour(), t.Minute(), 0, 0, loc), true
			}
		}
	}
	// The placeholder sits in the timeline's zone: with no start time there is
	// no event-local clock for a zone to interpret.
	return time.Date(date.Year(), date.Month(), date.Day(),
		12, 0, 0, 0, activistTimelineLocation), false
}

func getActivistTimelineEvents(db *sqlx.DB, activistID, limit int) ([]ActivistTimelineItem, error) {
	var rows []timelineEventRow
	// Ordered by date rather than by the instant the merged timeline sorts on.
	// That instant folds in start_time and the event's own zone, plus a noon
	// placeholder when there is no start time (see
	// activistTimelineEventInstant), none of which SQL reproduces cheaply. The
	// two orders pick the same newest events except at the cutoff, where events
	// sharing the boundary date are dropped by id rather than by time of day.
	query := `SELECT e.id, e.name, e.event_type, e.date, e.start_time, e.timezone
		FROM event_attendance ea
		JOIN events e ON e.id = ea.event_id
		WHERE ea.activist_id = ?
		ORDER BY e.date DESC, e.id DESC
		LIMIT ?`
	if err := db.Select(&rows, query, activistID, limit); err != nil {
		return nil, fmt.Errorf("failed to select attended events: %w", err)
	}

	items := make([]ActivistTimelineItem, 0, len(rows))
	for _, r := range rows {
		instant, hasTime := activistTimelineEventInstant(r.Date, r.StartTime, r.Timezone)
		items = append(items, ActivistTimelineItem{
			Type: TimelineItemTypeEvent,
			ID:   r.ID,
			// The event's own date rather than the timeline zone's projection
			// of instant, so the timeline agrees with the event's page. The two
			// disagree only for a chapter far enough east that the projection
			// crosses midnight, where the row's date and time then disagree
			// instead; projecting here would trade one for the other.
			Date:      r.Date.Format(EventDateLayout),
			Timestamp: instant.UTC().Format(time.RFC3339),
			HasTime:   hasTime,
			Event: &ActivistTimelineEvent{
				Name: r.Name,
				Type: r.EventType,
			},
			sortKey: instant,
		})
	}
	return items, nil
}

type timelineInteractionRow struct {
	ID int `db:"id"`
	// An interaction is stored as an instant, so its time of day is always
	// real.
	Timestamp time.Time `db:"timestamp"`
	Method    string    `db:"method"`
	Outcome   string    `db:"outcome"`
	Notes     string    `db:"notes"`
	UserName  string    `db:"user_name"`
}

func getActivistTimelineInteractions(db *sqlx.DB, activistID, limit int) ([]ActivistTimelineItem, error) {
	var rows []timelineInteractionRow
	query := `SELECT i.id, i.timestamp, IFNULL(i.method, '') AS method,
			IFNULL(i.outcome, '') AS outcome, IFNULL(i.notes, '') AS notes,
			IFNULL(u.name, '') AS user_name
		FROM interactions i
		LEFT JOIN adb_users u ON u.id = i.user_id
		WHERE i.activist_id = ?
		ORDER BY i.timestamp DESC, i.id DESC
		LIMIT ?`
	if err := db.Select(&rows, query, activistID, limit); err != nil {
		return nil, fmt.Errorf("failed to select interactions: %w", err)
	}

	items := make([]ActivistTimelineItem, 0, len(rows))
	for _, r := range rows {
		local := r.Timestamp.In(activistTimelineLocation)
		items = append(items, ActivistTimelineItem{
			Type:      TimelineItemTypeInteraction,
			ID:        r.ID,
			Date:      local.Format(EventDateLayout),
			Timestamp: r.Timestamp.UTC().Format(time.RFC3339),
			HasTime:   true,
			Interaction: &ActivistTimelineInteraction{
				Method:   r.Method,
				Outcome:  r.Outcome,
				Notes:    r.Notes,
				UserName: r.UserName,
			},
			sortKey: r.Timestamp,
		})
	}
	return items, nil
}
