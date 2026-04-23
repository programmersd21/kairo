package core

import (
	"testing"
	"time"
)

func TestDefaultViews(t *testing.T) {
	views := DefaultViews(time.Now())
	if len(views) == 0 {
		t.Error("expected views, got none")
	}
}

func TestFilter_ApplyToTask(t *testing.T) {
	f := Filter{
		Tags:     []string{"test"},
		Priority: new(Priority),
	}
	*f.Priority = P0

	task := Task{Title: "Test Task"}
	f.ApplyToTask(&task)

	if task.Priority != P0 {
		t.Errorf("expected priority P0, got %v", task.Priority)
	}
	hasTag := false
	for _, tg := range task.Tags {
		if tg == "test" {
			hasTag = true
		}
	}
	if !hasTag {
		t.Error("expected tag 'test' not found")
	}
}
