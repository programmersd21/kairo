package codec

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"

	"github.com/programmersd21/kairo/internal/core"
)

// MarshalText encodes tasks into a simple plain text format.
func MarshalText(tasks []core.Task) []byte {
	var b bytes.Buffer
	for _, t := range tasks {
		status := " "
		if t.Status == core.StatusDone {
			status = "x"
		}
		tags := ""
		if len(t.Tags) > 0 {
			tags = " " + tagsInline(t.Tags)
		}
		fmt.Fprintf(&b, "[%s] %s%s\n", status, t.Title, tags)
		if t.Description != "" {
			for _, line := range strings.Split(strings.TrimRight(t.Description, "\n"), "\n") {
				fmt.Fprintf(&b, "  %s\n", line)
			}
		}
		fmt.Fprintln(&b)
	}
	return b.Bytes()
}

// UnmarshalText decodes tasks from a simple plain text format.
func UnmarshalText(b []byte) ([]core.Task, error) {
	s := bufio.NewScanner(bytes.NewReader(b))
	var tasks []core.Task
	var cur *core.Task
	for s.Scan() {
		line := s.Text()
		if strings.HasPrefix(line, "[") && len(line) >= 4 && line[2] == ']' {
			// Handle "[ ] Title" or "[x] Title"
			if cur != nil {
				tasks = append(tasks, *cur)
			}
			st := core.StatusTodo
			if line[1] == 'x' || line[1] == 'X' {
				st = core.StatusDone
			}
			titleAndTags := strings.TrimSpace(line[4:])
			tags := extractTags(titleAndTags)
			title := stripTags(titleAndTags)
			cur = &core.Task{
				Title:    title,
				Tags:     tags,
				Status:   st,
				Priority: core.P1,
			}
		} else if cur != nil && strings.HasPrefix(line, "  ") {
			cur.Description += strings.TrimPrefix(line, "  ") + "\n"
		} else if line == "" && cur != nil {
			// Separator
			tasks = append(tasks, *cur)
			cur = nil
		}
	}
	if cur != nil {
		tasks = append(tasks, *cur)
	}
	for i := range tasks {
		tasks[i].Description = strings.TrimRight(tasks[i].Description, "\n")
	}
	return tasks, nil
}
