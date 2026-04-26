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

	header := []string{"ID", "Title", "Description", "Tags", "Priority", "Status", "Deadline", "CreatedAt", "UpdatedAt"}
	if err := w.Write(header); err != nil {
		return nil, err
	}

	for _, t := range tasks {
		deadline := ""
		if t.Deadline != nil {
			deadline = t.Deadline.Format(time.RFC3339)
		}
		row := []string{
			t.ID,
			t.Title,
			t.Description,
			strings.Join(t.Tags, ","),
			fmt.Sprintf("%d", t.Priority),
			string(t.Status),
			deadline,
			t.CreatedAt.Format(time.RFC3339),
			t.UpdatedAt.Format(time.RFC3339),
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

	var tasks []core.Task
	for i, row := range records {
		if i == 0 {
			continue // skip header
		}
		if len(row) < 6 {
			continue
		}

		priority := core.P1
		_, _ = fmt.Sscanf(row[4], "%d", &priority)

		t := core.Task{
			ID:          row[0],
			Title:       row[1],
			Description: row[2],
			Tags:        strings.Split(row[3], ","),
			Priority:    priority,
			Status:      core.Status(row[5]),
		}

		if len(row) > 6 && row[6] != "" {
			if dt, err := time.Parse(time.RFC3339, row[6]); err == nil {
				t.Deadline = &dt
			}
		}

		tasks = append(tasks, t)
	}
	return tasks, nil
}
