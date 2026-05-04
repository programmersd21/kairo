package core

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// NextOccurrence computes the next occurrence of a task after the given time.
func (t Task) NextOccurrence(after time.Time) *time.Time {
	if t.Recurrence == RecurrenceNone {
		return nil
	}

	switch t.Recurrence {
	case RecurrenceWeekly:
		return t.nextWeekly(after)
	case RecurrenceMonthly:
		return t.nextMonthly(after)
	default:
		return nil
	}
}

var dayMap = map[string]string{
	"mon": "mon", "monday": "mon",
	"tue": "tue", "tuesday": "tue",
	"wed": "wed", "wednesday": "wed",
	"thu": "thu", "thursday": "thu",
	"fri": "fri", "friday": "fri",
	"sat": "sat", "saturday": "sat",
	"sun": "sun", "sunday": "sun",
}

func (t Task) nextWeekly(after time.Time) *time.Time {
	if len(t.RecurrenceWeekly) == 0 {
		return nil
	}

	days := make(map[time.Weekday]bool)
	for _, d := range t.RecurrenceWeekly {
		d = strings.ToLower(strings.TrimSpace(d))
		if short, ok := dayMap[d]; ok {
			switch short {
			case "sun":
				days[time.Sunday] = true
			case "mon":
				days[time.Monday] = true
			case "tue":
				days[time.Tuesday] = true
			case "wed":
				days[time.Wednesday] = true
			case "thu":
				days[time.Thursday] = true
			case "fri":
				days[time.Friday] = true
			case "sat":
				days[time.Saturday] = true
			}
		}
	}

	if len(days) == 0 {
		return nil
	}

	// Search up to 7 days ahead
	curr := after.AddDate(0, 0, 1)
	for i := 0; i < 7; i++ {
		if days[curr.Weekday()] {
			// Maintain same time of day as original deadline if it exists, else use 00:00
			res := t.withSameTime(curr)
			return &res
		}
		curr = curr.AddDate(0, 0, 1)
	}

	return nil
}

func (t Task) nextMonthly(after time.Time) *time.Time {
	if t.RecurrenceMonthly < 1 || t.RecurrenceMonthly > 31 {
		return nil
	}

	// Start looking from the next month
	year, month, _ := after.Date()

	// Try next months until we find a valid date
	for i := 1; i <= 12; i++ {
		targetMonth := month + time.Month(i)
		targetYear := year
		for targetMonth > 12 {
			targetMonth -= 12
			targetYear++
		}

		// Handle month length (e.g. Feb 30 -> Feb 28/29 or Mar 1?)
		// Requirement says: "handle safely (skip or clamp using simplest approach)"
		// Clamping is usually preferred: if you want something on the 31st,
		// and the month has only 30 days, it happens on the 30th.

		lastDay := lastDayOfMonth(targetYear, targetMonth)
		day := t.RecurrenceMonthly
		if day > lastDay {
			day = lastDay
		}

		resDate := time.Date(targetYear, targetMonth, day, 0, 0, 0, 0, after.Location())

		// Ensure it's actually after 'after'
		if resDate.After(after) {
			res := t.withSameTime(resDate)
			return &res
		}
	}

	return nil
}

func (t Task) withSameTime(d time.Time) time.Time {
	if t.Deadline == nil {
		return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, d.Location())
	}
	dl := t.Deadline.Local()
	return time.Date(d.Year(), d.Month(), d.Day(), dl.Hour(), dl.Minute(), dl.Second(), dl.Nanosecond(), dl.Location()).UTC()
}

func lastDayOfMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// ParseRecurrence parses a raw string into recurrence configuration.
// weekly: "mon,wed,fri" or "monday, wednesday"
// monthly: "15"
func ParseRecurrence(raw string) (RecurrenceType, []string, int, error) {
	raw = strings.TrimSpace(strings.ToLower(raw))
	if raw == "" || raw == "none" {
		return RecurrenceNone, nil, 0, nil
	}

	// Try to parse as integer (monthly)
	if day, err := strconv.Atoi(raw); err == nil {
		if day < 1 || day > 31 {
			return RecurrenceNone, nil, 0, errors.New("monthly recurrence must be 1-31")
		}
		return RecurrenceMonthly, nil, day, nil
	}

	// Else assume weekly
	parts := strings.Split(raw, ",")

	var weekly []string
	for _, p := range parts {
		p = strings.TrimSpace(strings.ToLower(p))
		if p == "" {
			continue
		}
		if short, ok := dayMap[p]; ok {
			weekly = append(weekly, short)
		} else {
			return RecurrenceNone, nil, 0, fmt.Errorf("invalid weekday: %s", p)
		}
	}

	if len(weekly) == 0 {
		return RecurrenceNone, nil, 0, nil
	}

	return RecurrenceWeekly, weekly, 0, nil
}
