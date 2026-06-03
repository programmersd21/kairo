package codec

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strings"
	"time"

	"github.com/programmersd21/kairo/internal/core"
)

// MarshalCSV encodes tasks into CSV format.
func MarshalCSV(tasks []core.Task) ([]byte, error) {
	var b bytes.Buffer
	w := csv.NewWriter(&b)

	header := []string{
		"ID", "Title", "Description", "Tags", "Priority", "Status",
		"Deadline", "Recurrence", "RecurrenceWeekly", "RecurrenceMonthly",
		"ParentID", "Collapsed", "CreatedAt", "UpdatedAt",
		"OpenIssueID", "Result", "Responsible",
	}
	if err := w.Write(header); err != nil {
		return nil, err
	}

	for _, t := range tasks {
		deadline := ""
		if t.Deadline != nil {
			deadline = t.Deadline.Format(time.RFC3339)
		}
		collapsed := "false"
		if t.Collapsed {
			collapsed = "true"
		}
		row := []string{
			t.ID,
			t.Title,
			t.Description,
			strings.Join(t.Tags, ","),
			fmt.Sprintf("%d", t.Priority),
			string(t.Status),
			deadline,
			string(t.Recurrence),
			strings.Join(t.RecurrenceWeekly, ";"),
			fmt.Sprintf("%d", t.RecurrenceMonthly),
			t.ParentID,
			collapsed,
			t.CreatedAt.Format(time.RFC3339),
			t.UpdatedAt.Format(time.RFC3339),
			t.OpenIssueID,
			t.Result,
			t.Responsible,
		}
		if err := w.Write(row); err != nil {
			return nil, err
		}
	}

	w.Flush()
	return b.Bytes(), w.Error()
}

// UnmarshalCSV decodes tasks from CSV format.
func UnmarshalCSV(b []byte) ([]core.Task, error) {
	r := csv.NewReader(bytes.NewReader(b))
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) < 1 {
		return nil, nil
	}

	// Map headers to indices for robustness
	headerMap := make(map[string]int)
	for i, h := range records[0] {
		headerMap[h] = i
	}

	var tasks []core.Task
	for i, row := range records {
		if i == 0 {
			continue // skip header
		}

		get := func(key string) string {
			if idx, ok := headerMap[key]; ok && idx < len(row) {
				return row[idx]
			}
			return ""
		}

		priority := core.P1
		_, _ = fmt.Sscanf(get("Priority"), "%d", &priority)

		recurrenceMonthly := 0
		_, _ = fmt.Sscanf(get("RecurrenceMonthly"), "%d", &recurrenceMonthly)

		t := core.Task{
			ID:                get("ID"),
			Title:             get("Title"),
			Description:       get("Description"),
			Tags:              strings.Split(get("Tags"), ","),
			Priority:          priority,
			Status:            core.Status(get("Status")),
			Recurrence:        core.RecurrenceType(get("Recurrence")),
			RecurrenceWeekly:  strings.Split(get("RecurrenceWeekly"), ";"),
			RecurrenceMonthly: recurrenceMonthly,
			ParentID:          get("ParentID"),
			Collapsed:         get("Collapsed") == "true",
		}

		// Clean up empty strings from split
		if len(t.Tags) == 1 && t.Tags[0] == "" {
			t.Tags = nil
		}
		if len(t.RecurrenceWeekly) == 1 && t.RecurrenceWeekly[0] == "" {
			t.RecurrenceWeekly = nil
		}

		if dl := get("Deadline"); dl != "" {
			if dt, err := time.Parse(time.RFC3339, dl); err == nil {
				t.Deadline = &dt
			}
		}

		if ca := get("CreatedAt"); ca != "" {
			if dt, err := time.Parse(time.RFC3339, ca); err == nil {
				t.CreatedAt = dt
			}
		}

		if ua := get("UpdatedAt"); ua != "" {
			if dt, err := time.Parse(time.RFC3339, ua); err == nil {
				t.UpdatedAt = dt
			}
		}

		tasks = append(tasks, t)
	}
	return tasks, nil
}
