package tasklist

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/programmersd21/kairo/internal/core"
	"github.com/programmersd21/kairo/internal/ui/styles"
)

type Model struct {
	styles  styles.Styles
	vimMode bool

	width  int
	height int

	tasks []core.Task
	sel   int
}

func New(s styles.Styles, vimMode bool) Model {
	return Model{styles: s, vimMode: vimMode}
}

func (m Model) Selected() (core.Task, bool) {
	if m.sel < 0 || m.sel >= len(m.tasks) {
		return core.Task{}, false
	}
	return m.tasks[m.sel], true
}

func (m *Model) SetSize(w, h int) {
	m.width, m.height = w, h
}

func (m *Model) SetTasks(ts []core.Task) {
	m.tasks = append([]core.Task(nil), ts...)
	if m.sel >= len(m.tasks) {
		m.sel = len(m.tasks) - 1
	}
	if m.sel < 0 {
		m.sel = 0
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch x := msg.(type) {
	case tea.KeyMsg:
		switch x.String() {
		case "up":
			if m.sel > 0 {
				m.sel--
			}
		case "down":
			if m.sel < len(m.tasks)-1 {
				m.sel++
			}
		case "pgup":
			m.sel -= max(1, m.height-4)
			if m.sel < 0 {
				m.sel = 0
			}
		case "pgdown":
			m.sel += max(1, m.height-4)
			if m.sel > len(m.tasks)-1 {
				m.sel = len(m.tasks) - 1
			}
		case "home", "g":
			if !m.vimMode && x.String() == "g" {
				break
			}
			m.sel = 0
		case "end", "G":
			if x.String() == "G" && !m.vimMode {
				break
			}
			if len(m.tasks) > 0 {
				m.sel = len(m.tasks) - 1
			}
		case "j":
			if m.vimMode && m.sel < len(m.tasks)-1 {
				m.sel++
			}
		case "k":
			if m.vimMode && m.sel > 0 {
				m.sel--
			}
		}
	}
	return m, nil
}

func (m Model) View() string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}
	if len(m.tasks) == 0 {
		msg := lipgloss.NewStyle().
			Foreground(m.styles.Theme.Muted).
			Italic(true).
			Render("No tasks found.")
		hint := lipgloss.NewStyle().
			Foreground(m.styles.Theme.Accent).
			Render("Press 'n' to create your first task.")

		content := lipgloss.JoinVertical(lipgloss.Center, msg, hint)
		filled := lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content, lipgloss.WithWhitespaceChars(" "), lipgloss.WithWhitespaceBackground(m.styles.Theme.Bg))
		return filled
	}

	visible := max(1, m.height)
	start := clamp(m.sel-visible/2, 0, max(0, len(m.tasks)-visible))
	end := min(len(m.tasks), start+visible)

	lines := make([]string, 0, visible)
	for i := start; i < end; i++ {
		t := m.tasks[i]
		line := m.renderRow(t, i == m.sel)
		lines = append(lines, line)
	}
	// Fill remaining space with background
	for len(lines) < visible {
		emptyLine := lipgloss.NewStyle().Background(m.styles.Theme.Bg).Render(strings.Repeat(" ", m.width))
		lines = append(lines, emptyLine)
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderRow(t core.Task, selected bool) string {
	status := m.styles.StatusBadge(t.Status)
	pri := m.styles.PriorityBadge(t.Priority)

	dead := ""
	if t.Deadline != nil {
		dead = humanDeadline(*t.Deadline, time.Now())
	}
	tags := ""
	if len(t.Tags) > 0 {
		tags = "#" + strings.Join(t.Tags, " #")
	}

	// Selection indicator
	indicator := "  "
	if selected {
		indicator = lipgloss.NewStyle().Foreground(m.styles.Theme.Accent).Render("┃ ")
	}

	titleStyle := lipgloss.NewStyle().Foreground(m.styles.Theme.Fg).Background(m.styles.Theme.Bg)
	if selected {
		titleStyle = titleStyle.Foreground(m.styles.Theme.Accent).Background(m.styles.Theme.Overlay).Bold(true)
	} else if t.Status == core.StatusDone {
		titleStyle = m.styles.Muted.Strikethrough(true)
	}

	title := titleStyle.Render(truncate(t.Title, max(16, m.width/2)))

	left := lipgloss.JoinHorizontal(lipgloss.Left, indicator, status, " ", pri, " ", title)

	rightParts := []string{}
	if dead != "" {
		deadStyle := m.styles.Muted
		if t.Deadline.Before(time.Now()) && t.Status != core.StatusDone {
			deadStyle = lipgloss.NewStyle().Foreground(m.styles.Theme.Bad)
		}
		rightParts = append(rightParts, deadStyle.Render(dead))
	}
	if tags != "" {
		rightParts = append(rightParts, m.styles.Muted.Render(truncate(tags, max(10, m.width/4))))
	}
	right := strings.Join(rightParts, "  ")

	space := m.width - lipgloss.Width(left) - lipgloss.Width(right) - 1
	if space < 1 {
		space = 1
	}

	line := left + strings.Repeat(" ", space) + right
	rowStyle := lipgloss.NewStyle().Background(m.styles.Theme.Bg).Width(m.width)
	if selected {
		rowStyle = rowStyle.Background(m.styles.Theme.Overlay)
	}
	return rowStyle.Render(line)
}

func truncate(s string, w int) string {
	if w <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= w {
		return s
	}
	// keep it simple; assume mostly ascii titles.
	if w <= 1 {
		return "…"
	}
	r := []rune(s)
	if len(r) <= w-1 {
		return string(r)
	}
	return string(r[:w-1]) + "…"
}

func humanDeadline(t time.Time, now time.Time) string {
	d := t.Sub(now)
	if d < 0 {
		d = -d
		if d < 24*time.Hour {
			return "overdue"
		}
		return fmt.Sprintf("%dd overdue", int(d.Hours()/24))
	}
	if d < 2*time.Hour {
		return fmt.Sprintf("in %dm", int(d.Minutes()))
	}
	if d < 36*time.Hour {
		return fmt.Sprintf("in %dh", int(d.Hours()))
	}
	return fmt.Sprintf("in %dd", int(d.Hours()/24))
}

func clamp(x, lo, hi int) int {
	if x < lo {
		return lo
	}
	if x > hi {
		return hi
	}
	return x
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
