package core

import (
	"time"
)

type ViewID string

const (
	ViewInbox     ViewID = "inbox"
	ViewToday     ViewID = "today"
	ViewUpcoming  ViewID = "upcoming"
	ViewCompleted ViewID = "completed"
	ViewTag       ViewID = "tag"
	ViewPriority  ViewID = "priority"
)

type View struct {
	ID       ViewID
	Title    string
	Filter   Filter
	MetaHint string // e.g. active tag
}

type SortMode string

const (
	SortDefault  SortMode = ""
	SortDeadline SortMode = "deadline"
	SortPriority SortMode = "priority"
	SortUpdated  SortMode = "updated"
	SortCreated  SortMode = "created"
)

type Filter struct {
	Statuses           []Status
	Tags               []string
	Priority           *Priority
	From               *time.Time
	To                 *time.Time
	IncludeNilDeadline bool
	Sort               SortMode
}

func (f Filter) ApplyToTask(t *Task) {
	if len(f.Statuses) > 0 {
		t.Status = f.Statuses[0]
	}
	if len(f.Tags) > 0 {
		for _, required := range f.Tags {
			exists := false
			for _, existing := range t.Tags {
				if existing == required {
					exists = true
					break
				}
			}
			if !exists {
				t.Tags = append(t.Tags, required)
			}
		}
	}
	if f.Priority != nil {
		t.Priority = *f.Priority
	}
	// For views with a specific "To" date like "Today", set the deadline
	if f.To != nil && t.Deadline == nil {
		d := f.To.Add(-1 * time.Second) // Just before the boundary
		t.Deadline = &d
	}
}

func DefaultViews(now time.Time) []View {
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	dayEnd := dayStart.Add(24 * time.Hour)
	return []View{
		{
			ID:    ViewInbox,
			Title: "Inbox",
			Filter: Filter{
				Statuses:           []Status{StatusTodo, StatusDoing},
				IncludeNilDeadline: true,
				Sort:               SortUpdated,
			},
		},
		{
			ID:    ViewToday,
			Title: "Today",
			Filter: Filter{
				Statuses:           []Status{StatusTodo, StatusDoing},
				From:               &dayStart,
				To:                 &dayEnd,
				IncludeNilDeadline: false,
				Sort:               SortDeadline,
			},
		},
		{
			ID:    ViewUpcoming,
			Title: "Upcoming",
			Filter: Filter{
				Statuses:           []Status{StatusTodo, StatusDoing},
				From:               &dayEnd,
				IncludeNilDeadline: false,
				Sort:               SortDeadline,
			},
		},
		{
			ID:    ViewCompleted,
			Title: "Completed",
			Filter: Filter{
				Statuses:           []Status{StatusDone},
				IncludeNilDeadline: true,
				Sort:               SortUpdated,
			},
		},
		{
			ID:    ViewTag,
			Title: "By Tag",
			Filter: Filter{
				Statuses:           []Status{StatusTodo, StatusDoing, StatusDone},
				IncludeNilDeadline: true,
				Sort:               SortUpdated,
			},
		},
		{
			ID:    ViewPriority,
			Title: "By Priority",
			Filter: Filter{
				Statuses:           []Status{StatusTodo, StatusDoing, StatusDone},
				IncludeNilDeadline: true,
				Sort:               SortPriority,
			},
		},
	}
}
