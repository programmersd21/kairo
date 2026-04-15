package codec

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/programmersd21/kairo/internal/core"
)

func MarshalMarkdown(tasks []core.Task) []byte {
	byStatus := map[core.Status][]core.Task{
		core.StatusTodo:  {},
		core.StatusDoing: {},
		core.StatusDone:  {},
	}
	for _, t := range tasks {
		byStatus[t.Status] = append(byStatus[t.Status], t)
	}
	for st := range byStatus {
		sort.Slice(byStatus[st], func(i, j int) bool {
			return byStatus[st][i].UpdatedAt.After(byStatus[st][j].UpdatedAt)
		})
	}

	var b bytes.Buffer
	fmt.Fprintf(&b, "# Kairo Export\n\n")
	fmt.Fprintf(&b, "_Exported: %s_\n\n", time.Now().UTC().Format(time.RFC3339))

	writeSection := func(title string, st core.Status) {
		fmt.Fprintf(&b, "## %s\n\n", title)
		for _, t := range byStatus[st] {
			box := " "
			if st == core.StatusDone {
				box = "x"
			}
			line := fmt.Sprintf("- [%s] %s", box, escapeMDInline(strings.TrimSpace(t.Title)))
			if t.Deadline != nil {
				line += "  _(due " + t.Deadline.Local().Format("2006-01-02") + ")_"
			}
			if len(t.Tags) > 0 {
				line += "  " + tagsInline(t.Tags)
			}
			fmt.Fprintln(&b, line)
			if strings.TrimSpace(t.Description) != "" {
				for _, ln := range strings.Split(strings.TrimRight(t.Description, "\n"), "\n") {
					fmt.Fprintln(&b, "  "+ln)
				}
			}
		}
		fmt.Fprintln(&b)
	}

	writeSection("Todo", core.StatusTodo)
	writeSection("Doing", core.StatusDoing)
	writeSection("Done", core.StatusDone)
	return b.Bytes()
}

var mdTaskRe = regexp.MustCompile(`^\s*-\s*\[( |x|X)\]\s+(.*)$`)

func UnmarshalMarkdown(b []byte) ([]core.Task, error) {
	s := bufio.NewScanner(bytes.NewReader(b))
	var tasks []core.Task
	var cur *core.Task
	for s.Scan() {
		line := s.Text()
		if m := mdTaskRe.FindStringSubmatch(line); m != nil {
			if cur != nil {
				tasks = append(tasks, *cur)
			}
			title := strings.TrimSpace(m[2])
			st := core.StatusTodo
			if m[1] == "x" || m[1] == "X" {
				st = core.StatusDone
			}
			tags := extractTags(title)
			title = stripTags(title)
			cur = &core.Task{
				Title:    title,
				Tags:     tags,
				Status:   st,
				Priority: core.P1,
			}
			continue
		}
		if cur != nil {
			if strings.HasPrefix(line, "  ") {
				cur.Description += strings.TrimPrefix(line, "  ") + "\n"
			}
		}
	}
	if cur != nil {
		tasks = append(tasks, *cur)
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	for i := range tasks {
		tasks[i].Description = strings.TrimRight(tasks[i].Description, "\n")
	}
	return tasks, nil
}

func escapeMDInline(s string) string {
	repl := strings.NewReplacer("|", "\\|", "*", "\\*", "_", "\\_", "`", "\\`")
	return repl.Replace(s)
}

func tagsInline(tags []string) string {
	out := make([]string, 0, len(tags))
	for _, t := range tags {
		t = core.NormalizeTag(t)
		if t != "" {
			out = append(out, "#"+t)
		}
	}
	sort.Strings(out)
	return strings.Join(out, " ")
}

func extractTags(s string) []string {
	parts := strings.Fields(s)
	var tags []string
	for _, p := range parts {
		if strings.HasPrefix(p, "#") && len(p) > 1 {
			tags = append(tags, core.NormalizeTag(p))
		}
	}
	return tags
}

func stripTags(s string) string {
	parts := strings.Fields(s)
	keep := parts[:0]
	for _, p := range parts {
		if strings.HasPrefix(p, "#") && len(p) > 1 {
			continue
		}
		keep = append(keep, p)
	}
	return strings.Join(keep, " ")
}
