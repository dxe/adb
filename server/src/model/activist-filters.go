package model

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// DateOnly represents a date without time information (YYYY-MM-DD format).
// The time component is always 00:00:00 UTC.
type DateOnly struct {
	time.Time
}

// Compile-time checks that DateOnly implements json.Marshaler/json.Unmarshaler
var _ json.Unmarshaler = (*DateOnly)(nil)
var _ json.Marshaler = DateOnly{}

// UnmarshalJSON parses a date string in YYYY-MM-DD format as UTC midnight
func (d *DateOnly) UnmarshalJSON(data []byte) error {
	// Null means "no bound" for optional range filters.
	if string(data) == "null" {
		d.Time = time.Time{}
		return nil
	}

	var dateStr string
	if err := json.Unmarshal(data, &dateStr); err != nil {
		return fmt.Errorf("invalid date value: %w", err)
	}

	if strings.TrimSpace(dateStr) == "" {
		return fmt.Errorf("empty date string is not allowed")
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
type DateRangeFilter struct {
	Gte    DateOnly `json:"gte"` // Greater than or equal to (inclusive)
	Lt     DateOnly `json:"lt"`  // Less than (exclusive)
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
type IntRangeFilter struct {
	Gte *int `json:"gte,omitempty"` // Greater than or equal to (inclusive)
	Lt  *int `json:"lt,omitempty"`  // Less than (exclusive)
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

func assertNonNegative(f IntRangeFilter) error {
	if (f.Gte != nil && *f.Gte < 0) || (f.Lt != nil && *f.Lt < 0) {
		return fmt.Errorf("negative bounds")
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
type SourceFilter struct {
	ContainsAny    []string `json:"contains_any,omitempty"`
	NotContainsAny []string `json:"not_contains_any,omitempty"`
}

func (f *SourceFilter) IsEmpty() bool {
	return len(f.ContainsAny) == 0 && len(f.NotContainsAny) == 0
}

func (f *SourceFilter) Validate() error {
	containsSet := make(map[string]struct{}, len(f.ContainsAny))
	for i, v := range f.ContainsAny {
		v = strings.TrimSpace(v)
		if v == "" {
			return fmt.Errorf("contains_any[%d] cannot be empty", i)
		}
		containsSet[v] = struct{}{}
	}
	for i, v := range f.NotContainsAny {
		v = strings.TrimSpace(v)
		if v == "" {
			return fmt.Errorf("not_contains_any[%d] cannot be empty", i)
		}
		if _, exists := containsSet[v]; exists {
			return fmt.Errorf("source token %q cannot be both contains_any and not_contains_any", v)
		}
	}
	return nil
}

// TrainingFilter filters by training column completion status.
type TrainingFilter struct {
	Completed    []string `json:"completed,omitempty"`
	NotCompleted []string `json:"not_completed,omitempty"`
}

func (f *TrainingFilter) IsEmpty() bool {
	return len(f.Completed) == 0 && len(f.NotCompleted) == 0
}

var ValidTrainingColumns = map[string]bool{
	"training0":        true,
	"training1":        true,
	"training4":        true,
	"training5":        true,
	"training6":        true,
	"consent_quiz":     true,
	"training_protest": true,
	"dev_quiz":         true,
}

const (
	ProspectFilterChapterMember = "chapter_member"
	ProspectFilterOrganizer     = "organizer"
)

func (f *TrainingFilter) Validate() error {
	completedSet := make(map[string]struct{}, len(f.Completed))
	for i, v := range f.Completed {
		if strings.TrimSpace(v) == "" {
			return fmt.Errorf("completed[%d] cannot be empty", i)
		}
		if !ValidTrainingColumns[v] {
			return fmt.Errorf("invalid training column: %q", v)
		}
		completedSet[v] = struct{}{}
	}
	for i, v := range f.NotCompleted {
		if strings.TrimSpace(v) == "" {
			return fmt.Errorf("not_completed[%d] cannot be empty", i)
		}
		if !ValidTrainingColumns[v] {
			return fmt.Errorf("invalid training column: %q", v)
		}
		if _, exists := completedSet[v]; exists {
			return fmt.Errorf("training column %q cannot be both completed and not_completed", v)
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
	AssignedTo int    `json:"assigned_to"`
	Followups  string `json:"followups"` // "", "all", "due", "upcoming"
	Prospect   string `json:"prospect"`  // "", ProspectFilterChapterMember, ProspectFilterOrganizer
}

func (f *QueryActivistFilters) Validate() error {
	if f.ChapterId < 0 {
		return fmt.Errorf("invalid chapter_id value: %d", f.ChapterId)
	}
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
	if err := assertNonNegative(f.TotalEvents); err != nil {
		return fmt.Errorf("invalid total events filter: %w", err)
	}
	if err := f.TotalEvents.Validate(); err != nil {
		return fmt.Errorf("invalid total events filter: %w", err)
	}
	if err := assertNonNegative(f.TotalInteractions); err != nil {
		return fmt.Errorf("invalid total interactions filter: %w", err)
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
	if f.AssignedTo < -1 {
		return fmt.Errorf("invalid assigned_to value: %d", f.AssignedTo)
	}
	if f.Followups != "" && f.Followups != "all" && f.Followups != "due" && f.Followups != "upcoming" {
		return fmt.Errorf("invalid followups value: %q", f.Followups)
	}
	if f.Prospect != "" &&
		f.Prospect != ProspectFilterChapterMember &&
		f.Prospect != ProspectFilterOrganizer {
		return fmt.Errorf("invalid prospect value: %q", f.Prospect)
	}
	return nil
}
