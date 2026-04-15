package codec

import (
	"encoding/json"
	"time"

	"github.com/programmersd21/kairo/internal/core"
)

type JSONExport struct {
	ExportedAt time.Time   `json:"exported_at"`
	Tasks      []core.Task `json:"tasks"`
}

func MarshalJSON(tasks []core.Task) ([]byte, error) {
	w := JSONExport{
		ExportedAt: time.Now().UTC(),
		Tasks:      tasks,
	}
	return json.MarshalIndent(w, "", "  ")
}

func UnmarshalJSON(b []byte) ([]core.Task, error) {
	var w JSONExport
	if err := json.Unmarshal(b, &w); err != nil {
		// Also accept a bare list for interoperability.
		var tasks []core.Task
		if err2 := json.Unmarshal(b, &tasks); err2 != nil {
			return nil, err
		}
		return tasks, nil
	}
	return w.Tasks, nil
}
