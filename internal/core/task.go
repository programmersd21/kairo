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
	Project           string
	Tags              []string
	Priority          Priority
	Deadline          *time.Time
	WaitUntil         *time.Time
	Until             *time.Time
	Status            Status
	Recurrence        RecurrenceType
	RecurrenceWeekly  []string
	RecurrenceMonthly int
	ParentID          string
	Collapsed         bool
	CreatedAt         time.Time
	UpdatedAt         time.Time
	CompletedAt       *time.Time
	Result            string
	OpenIssueID       string
	Responsible       string
}

func (t Task) DeepCopy() Task {
	clone := t
	if t.Deadline != nil {
		d := *t.Deadline
		clone.Deadline = &d
	}
	if t.WaitUntil != nil {
		w := *t.WaitUntil
		clone.WaitUntil = &w
	}
	if t.Until != nil {
		u := *t.Until
		clone.Until = &u
	}
	if t.CompletedAt != nil {
		c := *t.CompletedAt
		clone.CompletedAt = &c
	}
	if t.Tags != nil {
		clone.Tags = append([]string(nil), t.Tags...)
	}
	if t.RecurrenceWeekly != nil {
		clone.RecurrenceWeekly = append([]string(nil), t.RecurrenceWeekly...)
	}
	return clone
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
	Project           *string
	Tags              *[]string
	Priority          *Priority
	Deadline          **time.Time
	WaitUntil         **time.Time
	Until             **time.Time
	Status            *Status
	Recurrence        *RecurrenceType
	RecurrenceWeekly  *[]string
	RecurrenceMonthly *int
	ParentID          *string
	Collapsed         *bool
	CompletedAt       **time.Time
	Result            *string
	OpenIssueID       *string
	Responsible       *string
}

func (p TaskPatch) ApplyTo(t Task) Task {
	if p.Title != nil {
		t.Title = *p.Title
	}
	if p.Description != nil {
		t.Description = *p.Description
	}
	if p.Project != nil {
		t.Project = *p.Project
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
	if p.WaitUntil != nil {
		t.WaitUntil = *p.WaitUntil
	}
	if p.Until != nil {
		t.Until = *p.Until
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
	if p.ParentID != nil {
		t.ParentID = *p.ParentID
	}
	if p.Collapsed != nil {
		t.Collapsed = *p.Collapsed
	}
	if p.CompletedAt != nil {
		t.CompletedAt = *p.CompletedAt
	}
	if p.Result != nil {
		t.Result = *p.Result
	}
	if p.OpenIssueID != nil {
		t.OpenIssueID = *p.OpenIssueID
	}
	if p.Responsible != nil {
		t.Responsible = *p.Responsible
	}
	return t
}

func (t Task) MarshalJSON() ([]byte, error) {
	type wire struct {
		ID                string    `json:"id"`
		Title             string    `json:"title"`
		Description       string    `json:"description,omitempty"`
		Project           string    `json:"project,omitempty"`
		Tags              []string  `json:"tags,omitempty"`
		Priority          int       `json:"priority"`
		Deadline          *string   `json:"deadline,omitempty"`
		WaitUntil         *string   `json:"wait_until,omitempty"`
		Until             *string   `json:"until,omitempty"`
		Status            Status    `json:"status"`
		Recurrence        string    `json:"recurrence,omitempty"`
		RecurrenceWeekly  []string  `json:"recurrence_weekly,omitempty"`
		RecurrenceMonthly int       `json:"recurrence_monthly,omitempty"`
		ParentID          string    `json:"parent_id,omitempty"`
		Collapsed         bool      `json:"collapsed,omitempty"`
		CreatedAt         time.Time `json:"created_at"`
		UpdatedAt         time.Time `json:"updated_at"`
		CompletedAt       *string   `json:"completed_at,omitempty"`
		Result            string    `json:"result,omitempty"`
		OpenIssueID       string    `json:"open_issue_id,omitempty"`
		Responsible       string    `json:"responsible,omitempty"`
	}
	var d *string
	if t.Deadline != nil {
		s := t.Deadline.UTC().Format(time.RFC3339Nano)
		d = &s
	}
	var wu *string
	if t.WaitUntil != nil {
		s := t.WaitUntil.UTC().Format(time.RFC3339Nano)
		wu = &s
	}
	var u *string
	if t.Until != nil {
		s := t.Until.UTC().Format(time.RFC3339Nano)
		u = &s
	}
	var c *string
	if t.CompletedAt != nil {
		s := t.CompletedAt.UTC().Format(time.RFC3339Nano)
		c = &s
	}
	return json.Marshal(wire{
		ID:                t.ID,
		Title:             t.Title,
		Description:       t.Description,
		Project:           t.Project,
		Tags:              t.NormalizedTags(),
		Priority:          int(t.Priority.Clamp()),
		Deadline:          d,
		WaitUntil:         wu,
		Until:             u,
		Status:            t.Status,
		Recurrence:        string(t.Recurrence),
		RecurrenceWeekly:  t.RecurrenceWeekly,
		RecurrenceMonthly: t.RecurrenceMonthly,
		ParentID:          t.ParentID,
		Collapsed:         t.Collapsed,
		CreatedAt:         t.CreatedAt.UTC(),
		UpdatedAt:         t.UpdatedAt.UTC(),
		CompletedAt:       c,
		Result:            t.Result,
		OpenIssueID:       t.OpenIssueID,
		Responsible:       t.Responsible,
	})
}

func (t *Task) UnmarshalJSON(data []byte) error {
	type wire struct {
		ID                string    `json:"id"`
		Title             string    `json:"title"`
		Description       string    `json:"description,omitempty"`
		Project           string    `json:"project,omitempty"`
		Tags              []string  `json:"tags,omitempty"`
		Priority          int       `json:"priority"`
		Deadline          *string   `json:"deadline,omitempty"`
		WaitUntil         *string   `json:"wait_until,omitempty"`
		Until             *string   `json:"until,omitempty"`
		Status            Status    `json:"status"`
		Recurrence        string    `json:"recurrence,omitempty"`
		RecurrenceWeekly  []string  `json:"recurrence_weekly,omitempty"`
		RecurrenceMonthly int       `json:"recurrence_monthly,omitempty"`
		ParentID          string    `json:"parent_id,omitempty"`
		Collapsed         bool      `json:"collapsed,omitempty"`
		CreatedAt         time.Time `json:"created_at"`
		UpdatedAt         time.Time `json:"updated_at"`
		CompletedAt       *string   `json:"completed_at,omitempty"`
		Result            string    `json:"result,omitempty"`
		OpenIssueID       string    `json:"open_issue_id,omitempty"`
		Responsible       string    `json:"responsible,omitempty"`
	}
	var w wire
	if err := json.Unmarshal(data, &w); err != nil {
		return err
	}
	t.ID = w.ID
	t.Title = w.Title
	t.Description = w.Description
	t.Project = w.Project
	t.Tags = w.Tags
	t.Priority = Priority(w.Priority).Clamp()
	t.Status = w.Status
	t.Recurrence = RecurrenceType(w.Recurrence)
	t.RecurrenceWeekly = w.RecurrenceWeekly
	t.RecurrenceMonthly = w.RecurrenceMonthly
	t.ParentID = w.ParentID
	t.Collapsed = w.Collapsed
	t.CreatedAt = w.CreatedAt
	t.UpdatedAt = w.UpdatedAt
	t.Result = w.Result
	t.OpenIssueID = w.OpenIssueID
	t.Responsible = w.Responsible

	if w.Deadline != nil {
		dt, err := time.Parse(time.RFC3339Nano, *w.Deadline)
		if err != nil {
			return err
		}
		t.Deadline = &dt
	}
	if w.WaitUntil != nil {
		dt, err := time.Parse(time.RFC3339Nano, *w.WaitUntil)
		if err != nil {
			return err
		}
		t.WaitUntil = &dt
	}
	if w.Until != nil {
		dt, err := time.Parse(time.RFC3339Nano, *w.Until)
		if err != nil {
			return err
		}
		t.Until = &dt
	}
	if w.CompletedAt != nil {
		ct, err := time.Parse(time.RFC3339Nano, *w.CompletedAt)
		if err != nil {
			return err
		}
		t.CompletedAt = &ct
	}
	return nil
}
