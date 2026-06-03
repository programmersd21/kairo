package core

import (
	"reflect"
	"testing"
)

func TestTaskPatch_ApplyTo(t *testing.T) {
	task := Task{
		ID:          "1",
		Title:       "Original",
		Description: "Desc",
		Project:     "Proj",
		Tags:        []string{"a"},
		Priority:    P1,
		Status:      StatusTodo,
	}

	newTitle := "Updated"
	newTags := []string{"b", "c"}
	newStatus := StatusDone
	patch := TaskPatch{
		Title:  &newTitle,
		Tags:   &newTags,
		Status: &newStatus,
	}

	updated := patch.ApplyTo(task)

	if updated.Title != "Updated" {
		t.Errorf("expected title Updated, got %s", updated.Title)
	}
	if !reflect.DeepEqual(updated.Tags, []string{"b", "c"}) {
		t.Errorf("expected tags [b c], got %v", updated.Tags)
	}
	if updated.Status != StatusDone {
		t.Errorf("expected status Done, got %s", updated.Status)
	}
	// Verify immutability of original
	if task.Title != "Original" {
		t.Error("original task title was mutated")
	}
}

func TestTaskPatch_ApplyTo_NewFields(t *testing.T) {
	task := Task{
		ID:      "1",
		Title:   "Test",
		Status:  StatusTodo,
		Project: "Work",
	}

	result := "Completed successfully"
	responsible := "alice"
	patch := TaskPatch{
		Result:      &result,
		Responsible: &responsible,
	}

	updated := patch.ApplyTo(task)

	if updated.Result != "Completed successfully" {
		t.Errorf("expected result 'Completed successfully', got %q", updated.Result)
	}
	if updated.Responsible != "alice" {
		t.Errorf("expected responsible 'alice', got %q", updated.Responsible)
	}
	if updated.OpenIssueID != "" {
		t.Errorf("expected empty OpenIssueID, got %q", updated.OpenIssueID)
	}
	// Verify immutability
	if task.Result != "" || task.Responsible != "" {
		t.Error("original task was mutated")
	}
}
