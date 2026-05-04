package core

import (
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"
)

type Status string

const (
	StatusTodo  Status = "todo"
	StatusDoing Status = "doing"
	StatusDone  Status = "done"
)

func (s Status) Valid() bool {
	switch s {
	case StatusTodo, StatusDoing, StatusDone:
		return true
	default:
		return false
	}
}

type Priority int

const (
	P0 Priority = 0
	P1 Priority = 1
	P2 Priority = 2
	P3 Priority = 3
)

func (p Priority) Clamp() Priority {
	switch {
	case p < P0:
		return P0
	case p > P3:
		return P3
	default:
		return p
	}
}

type RecurrenceType string

const (
	RecurrenceNone    RecurrenceType = "none"
	RecurrenceWeekly  RecurrenceType = "weekly"
	RecurrenceMonthly RecurrenceType = "monthly"
)

func (r RecurrenceType) Valid() bool {
	switch r {
	case RecurrenceNone, RecurrenceWeekly, RecurrenceMonthly:
		return true
	default:
		return false
	}
}

type Task struct {
	ID                string
	Title             string
	Description       string
	Tags              []string
	Priority          Priority
	Deadline          *time.Time
	Status            Status
	Recurrence        RecurrenceType
	RecurrenceWeekly  []string
	RecurrenceMonthly int
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (t Task) NormalizedTags() []string {
	m := make(map[string]struct{}, len(t.Tags))
	for _, tag := range t.Tags {
		tag = NormalizeTag(tag)
		if tag == "" {
			continue
		}
		m[tag] = struct{}{}
	}
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func NormalizeTag(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	s = strings.TrimPrefix(s, "#")
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, " ", "-")
	return s
}

func ParseTags(s string) []string {
	parts := strings.FieldsFunc(s, func(r rune) bool { return r == ',' || r == ' ' || r == '\t' || r == '\n' || r == '\r' })
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = NormalizeTag(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func (t Task) Validate() error {
	if strings.TrimSpace(t.Title) == "" {
		return errors.New("title required")
	}
	if !t.Status.Valid() {
		return errors.New("invalid status")
	}
	if !t.Recurrence.Valid() {
		return errors.New("invalid recurrence type")
	}
	if t.Recurrence == RecurrenceMonthly && (t.RecurrenceMonthly < 1 || t.RecurrenceMonthly > 31) {
		return errors.New("invalid monthly recurrence day (must be 1-31)")
	}
	return nil
}

type TaskPatch struct {
	Title             *string
	Description       *string
	Tags              *[]string
	Priority          *Priority
	Deadline          **time.Time
	Status            *Status
	Recurrence        *RecurrenceType
	RecurrenceWeekly  *[]string
	RecurrenceMonthly *int
}

func (p TaskPatch) ApplyTo(t Task) Task {
	if p.Title != nil {
		t.Title = *p.Title
	}
	if p.Description != nil {
		t.Description = *p.Description
	}
	if p.Tags != nil {
		t.Tags = append([]string(nil), (*p.Tags)...)
	}
	if p.Priority != nil {
		t.Priority = (*p.Priority).Clamp()
	}
	if p.Deadline != nil {
		t.Deadline = *p.Deadline
	}
	if p.Status != nil {
		t.Status = *p.Status
	}
	if p.Recurrence != nil {
		t.Recurrence = *p.Recurrence
	}
	if p.RecurrenceWeekly != nil {
		t.RecurrenceWeekly = append([]string(nil), (*p.RecurrenceWeekly)...)
	}
	if p.RecurrenceMonthly != nil {
		t.RecurrenceMonthly = *p.RecurrenceMonthly
	}
	return t
}

func (t Task) MarshalJSON() ([]byte, error) {
	type wire struct {
		ID                string    `json:"id"`
		Title             string    `json:"title"`
		Description       string    `json:"description,omitempty"`
		Tags              []string  `json:"tags,omitempty"`
		Priority          int       `json:"priority"`
		Deadline          *string   `json:"deadline,omitempty"`
		Status            Status    `json:"status"`
		Recurrence        string    `json:"recurrence,omitempty"`
		RecurrenceWeekly  []string  `json:"recurrence_weekly,omitempty"`
		RecurrenceMonthly int       `json:"recurrence_monthly,omitempty"`
		CreatedAt         time.Time `json:"created_at"`
		UpdatedAt         time.Time `json:"updated_at"`
	}
	var d *string
	if t.Deadline != nil {
		s := t.Deadline.UTC().Format(time.RFC3339Nano)
		d = &s
	}
	return json.Marshal(wire{
		ID:                t.ID,
		Title:             t.Title,
		Description:       t.Description,
		Tags:              t.NormalizedTags(),
		Priority:          int(t.Priority.Clamp()),
		Deadline:          d,
		Status:            t.Status,
		Recurrence:        string(t.Recurrence),
		RecurrenceWeekly:  t.RecurrenceWeekly,
		RecurrenceMonthly: t.RecurrenceMonthly,
		CreatedAt:         t.CreatedAt.UTC(),
		UpdatedAt:         t.UpdatedAt.UTC(),
	})
}
