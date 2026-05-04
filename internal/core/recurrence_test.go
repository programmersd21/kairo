package core

import (
	"testing"
	"time"
)

func TestNextOccurrence_Weekly(t *testing.T) {
	deadline := time.Date(2026, 5, 4, 10, 0, 0, 0, time.UTC) // Monday, May 4, 2026
	task := Task{
		Recurrence:       RecurrenceWeekly,
		RecurrenceWeekly: []string{"mon", "wed", "fri"},
		Deadline:         &deadline,
	}

	// Completing on Monday, should give Wednesday
	next := task.NextOccurrence(deadline)
	if next == nil {
		t.Fatal("expected next occurrence, got nil")
	}
	expected := time.Date(2026, 5, 6, 10, 0, 0, 0, time.UTC) // Wednesday, May 6
	if !next.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, *next)
	}

	// Completing on Wednesday, should give Friday
	next = task.NextOccurrence(expected)
	expected = time.Date(2026, 5, 8, 10, 0, 0, 0, time.UTC) // Friday, May 8
	if !next.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, *next)
	}

	// Completing on Friday, should give next Monday
	next = task.NextOccurrence(expected)
	expected = time.Date(2026, 5, 11, 10, 0, 0, 0, time.UTC) // Monday, May 11
	if !next.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, *next)
	}
}

func TestNextOccurrence_Monthly(t *testing.T) {
	deadline := time.Date(2026, 5, 15, 10, 0, 0, 0, time.UTC)
	task := Task{
		Recurrence:        RecurrenceMonthly,
		RecurrenceMonthly: 15,
		Deadline:          &deadline,
	}

	next := task.NextOccurrence(deadline)
	if next == nil {
		t.Fatal("expected next occurrence, got nil")
	}
	expected := time.Date(2026, 6, 15, 10, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, *next)
	}
}

func TestNextOccurrence_Monthly_Clamp(t *testing.T) {
	deadline := time.Date(2026, 1, 31, 10, 0, 0, 0, time.UTC)
	task := Task{
		Recurrence:        RecurrenceMonthly,
		RecurrenceMonthly: 31,
		Deadline:          &deadline,
	}

	// Next is Feb 28 (non-leap year)
	next := task.NextOccurrence(deadline)
	if next == nil {
		t.Fatal("expected next occurrence, got nil")
	}
	expected := time.Date(2026, 2, 28, 10, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, *next)
	}
}

func TestParseRecurrence(t *testing.T) {
	tests := []struct {
		input     string
		wantType  RecurrenceType
		wantWeek  []string
		wantMonth int
		wantErr   bool
	}{
		{"", RecurrenceNone, nil, 0, false},
		{"none", RecurrenceNone, nil, 0, false},
		{"15", RecurrenceMonthly, nil, 15, false},
		{"mon,wed", RecurrenceWeekly, []string{"mon", "wed"}, 0, false},
		{"MON, FRI", RecurrenceWeekly, []string{"mon", "fri"}, 0, false},
		{"32", RecurrenceNone, nil, 0, true},
		{"foo", RecurrenceNone, nil, 0, true},
	}

	for _, tt := range tests {
		gotType, gotWeek, gotMonth, err := ParseRecurrence(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("ParseRecurrence(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if gotType != tt.wantType {
			t.Errorf("ParseRecurrence(%q) gotType = %v, want %v", tt.input, gotType, tt.wantType)
		}
		if gotMonth != tt.wantMonth {
			t.Errorf("ParseRecurrence(%q) gotMonth = %v, want %v", tt.input, gotMonth, tt.wantMonth)
		}
		if len(gotWeek) != len(tt.wantWeek) {
			t.Errorf("ParseRecurrence(%q) gotWeek len = %v, want %v", tt.input, len(gotWeek), len(tt.wantWeek))
		}
	}
}
