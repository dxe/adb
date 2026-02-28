package model

import (
	"encoding/json"
	"fmt"
	"time"
)

// DateOnly represents a date without time information (YYYY-MM-DD format).
// The time component is always 00:00:00 UTC.
type DateOnly struct {
	time.Time
}

// Compile-time check that DateOnly implements json.Unmarshaler
var _ json.Unmarshaler = (*DateOnly)(nil)

// UnmarshalJSON parses a date string in YYYY-MM-DD format as UTC midnight
func (d *DateOnly) UnmarshalJSON(data []byte) error {
	// Remove quotes from JSON string
	dateStr := string(data)
	if len(dateStr) < 2 {
		return nil
	}
	dateStr = dateStr[1 : len(dateStr)-1]

	if dateStr == "" {
		return nil
	}

	parsed, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return fmt.Errorf("invalid date format (expected YYYY-MM-DD): %w", err)
	}

	d.Time = time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, time.UTC)
	return nil
}

// MarshalJSON formats the date as YYYY-MM-DD
func (d DateOnly) MarshalJSON() ([]byte, error) {
	if d.Time.IsZero() {
		return []byte("null"), nil
	}
	return []byte(`"` + d.Time.Format("2006-01-02") + `"`), nil
}

// DateRangeFilter filters by a date column with optional NULL inclusion.
// JSON: {"gte": "2025-01-01", "lt": "2025-06-01", "or_null": true}
type DateRangeFilter struct {
	Gte    DateOnly `json:"gte"`
	Lt     DateOnly `json:"lt"`
	OrNull bool     `json:"or_null,omitempty"`
}

func (f *DateRangeFilter) IsEmpty() bool {
	return f.Gte.IsZero() && f.Lt.IsZero() && !f.OrNull
}

func (f *DateRangeFilter) Validate() error {
	if !f.Gte.IsZero() && !f.Lt.IsZero() {
		if !f.Gte.Time.Before(f.Lt.Time) {
			return fmt.Errorf("invalid date range")
		}
		if f.OrNull {
			return fmt.Errorf("or_null is only valid for open-ended ranges (one bound must be missing)")
		}
	}
	return nil
}

// IntRangeFilter filters by an integer column.
// JSON: {"gte": 1, "lt": 4}
type IntRangeFilter struct {
	Gte *int `json:"gte,omitempty"`
	Lt  *int `json:"lt,omitempty"`
}

func (f *IntRangeFilter) IsEmpty() bool {
	return f.Gte == nil && f.Lt == nil
}

func (f *IntRangeFilter) Validate() error {
	if f.Gte != nil && f.Lt != nil && *f.Gte >= *f.Lt {
		return fmt.Errorf("invalid integer range")
	}
	return nil
}

// NameFilter filters activists by name using LIKE.
type NameFilter struct {
	NameContains string `json:"name_contains"`
}

// ActivistLevelFilter filters by activist_level values using one mode.
// mode="include" means match any listed values.
// mode="exclude" means exclude any listed values.
// JSON: {"mode": "include", "values": ["Supporter"]}
type ActivistLevelFilter struct {
	Mode   string   `json:"mode,omitempty"`
	Values []string `json:"values,omitempty"`
}

func (f *ActivistLevelFilter) IsEmpty() bool {
	return len(f.Values) == 0
}

var ValidActivistLevels = map[string]bool{
	"Supporter":             true,
	"Chapter Member":        true,
	"Organizer":             true,
	"Non-Local":             true,
	"Global Network Member": true,
}

func (f *ActivistLevelFilter) Validate() error {
	if len(f.Values) == 0 {
		return nil
	}
	if f.Mode != "include" && f.Mode != "exclude" {
		return fmt.Errorf("invalid activist level mode: %q", f.Mode)
	}
	for _, v := range f.Values {
		if !ValidActivistLevels[v] {
			return fmt.Errorf("invalid activist level: %q", v)
		}
	}
	return nil
}

// SourceFilter filters by the source column using LIKE patterns.
// JSON: {"contains_any": ["form", "petition"], "not_contains_any": ["application"]}
type SourceFilter struct {
	ContainsAny    []string `json:"contains_any,omitempty"`
	NotContainsAny []string `json:"not_contains_any,omitempty"`
}

func (f *SourceFilter) IsEmpty() bool {
	return len(f.ContainsAny) == 0 && len(f.NotContainsAny) == 0
}

// SourceFilter doesn't need complex validation beyond non-empty checks.
func (f *SourceFilter) Validate() error {
	return nil
}

// TrainingFilter filters by training column completion status.
// JSON: {"completed": ["training0"], "not_completed": ["training1"]}
type TrainingFilter struct {
	Completed    []string `json:"completed,omitempty"`
	NotCompleted []string `json:"not_completed,omitempty"`
}

func (f *TrainingFilter) IsEmpty() bool {
	return len(f.Completed) == 0 && len(f.NotCompleted) == 0
}

var validTrainingColumns = map[string]bool{
	"training0":        true,
	"training1":        true,
	"training4":        true,
	"training5":        true,
	"training6":        true,
	"consent_quiz":     true,
	"training_protest": true,
	"dev_quiz":         true,
}

func (f *TrainingFilter) Validate() error {
	for _, v := range f.Completed {
		if !validTrainingColumns[v] {
			return fmt.Errorf("invalid training column: %q", v)
		}
	}
	for _, v := range f.NotCompleted {
		if !validTrainingColumns[v] {
			return fmt.Errorf("invalid training column: %q", v)
		}
	}
	return nil
}

// QueryActivistFilters contains all filter parameters for querying activists.
type QueryActivistFilters struct {
	// 0 means search all chapters and requires that the "chapter" column be requested.
	// Must be set to ID of current chapter if user only has permission for current chapter.
	ChapterId     int             `json:"chapter_id"`
	Name          NameFilter      `json:"name"`
	LastEvent     DateRangeFilter `json:"last_event"`
	IncludeHidden bool            `json:"include_hidden"`

	ActivistLevel     ActivistLevelFilter `json:"activist_level"`
	InterestDate      DateRangeFilter     `json:"interest_date"`
	FirstEvent        DateRangeFilter     `json:"first_event"`
	TotalEvents       IntRangeFilter      `json:"total_events"`
	TotalInteractions IntRangeFilter      `json:"total_interactions"`
	Source            SourceFilter        `json:"source"`
	Training          TrainingFilter      `json:"training"`

	// 0 = no filter, -1 = any assignee (assigned_to <> 0), >0 = specific user ID.
	// Frontend should resolve "me" to the actual user ID before sending.
	AssignedTo int    `json:"assigned_to"`
	Followups  string `json:"followups"` // "", "all", "due", "upcoming"
	Prospect   string `json:"prospect"`  // "", "chapter_member", "organizer"
}

func (f *QueryActivistFilters) Validate() error {
	if err := f.LastEvent.Validate(); err != nil {
		return fmt.Errorf("invalid last event filter: %w", err)
	}
	if err := f.ActivistLevel.Validate(); err != nil {
		return fmt.Errorf("invalid activist level filter: %w", err)
	}
	if err := f.InterestDate.Validate(); err != nil {
		return fmt.Errorf("invalid interest date filter: %w", err)
	}
	if err := f.FirstEvent.Validate(); err != nil {
		return fmt.Errorf("invalid first event filter: %w", err)
	}
	if err := f.TotalEvents.Validate(); err != nil {
		return fmt.Errorf("invalid total events filter: %w", err)
	}
	if err := f.TotalInteractions.Validate(); err != nil {
		return fmt.Errorf("invalid total interactions filter: %w", err)
	}
	if err := f.Source.Validate(); err != nil {
		return fmt.Errorf("invalid source filter: %w", err)
	}
	if err := f.Training.Validate(); err != nil {
		return fmt.Errorf("invalid training filter: %w", err)
	}
	if f.Followups != "" && f.Followups != "all" && f.Followups != "due" && f.Followups != "upcoming" {
		return fmt.Errorf("invalid followups value: %q", f.Followups)
	}
	if f.Prospect != "" && f.Prospect != "chapter_member" && f.Prospect != "organizer" {
		return fmt.Errorf("invalid prospect value: %q", f.Prospect)
	}
	return nil
}
