package core

import (
	"time"
)

type ViewID string

const (
	ViewInbox    ViewID = "inbox"
	ViewToday    ViewID = "today"
	ViewUpcoming ViewID = "upcoming"
	ViewTag      ViewID = "tag"
	ViewPriority ViewID = "priority"
)

type View struct {
	ID       ViewID
	Title    string
	Filter   Filter
	MetaHint string // e.g. active tag
}

type Filter struct {
	Statuses           []Status
	Tag                string
	Priority           *Priority
	From               *time.Time
	To                 *time.Time
	IncludeNilDeadline bool
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
			},
		},
		{
			ID:    ViewUpcoming,
			Title: "Upcoming",
			Filter: Filter{
				Statuses:           []Status{StatusTodo, StatusDoing},
				From:               &dayEnd,
				IncludeNilDeadline: false,
			},
		},
		{
			ID:    ViewTag,
			Title: "By Tag",
			Filter: Filter{
				Statuses:           []Status{StatusTodo, StatusDoing, StatusDone},
				IncludeNilDeadline: true,
			},
		},
		{
			ID:    ViewPriority,
			Title: "By Priority",
			Filter: Filter{
				Statuses:           []Status{StatusTodo, StatusDoing, StatusDone},
				IncludeNilDeadline: true,
			},
		},
	}
}
