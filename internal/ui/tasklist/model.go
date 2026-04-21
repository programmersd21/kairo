package tasklist

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/programmersd21/kairo/internal/core"
	"github.com/programmersd21/kairo/internal/ui/render"
	"github.com/programmersd21/kairo/internal/ui/styles"
)

type Model struct {
	styles  styles.Styles
	vimMode bool

	width  int
	height int

	tasks []core.Task
	sel   int

	// Animation state
	animatingTaskID  string
	animationStart   time.Time
	animationDur     time.Duration
	animationReverse bool
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

func (m *Model) SetAnimation(taskID string, start time.Time, duration time.Duration, reverse bool) {
	m.animatingTaskID = taskID
	m.animationStart = start
	m.animationDur = duration
	m.animationReverse = reverse
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch x := msg.(type) {
	case tea.KeyMsg:
		switch x.String() {
		case "up", "k":
			if x.String() == "k" && !m.vimMode {
				break
			}
			if m.sel > 0 {
				m.sel--
			}
		case "down", "j":
			if x.String() == "j" && !m.vimMode {
				break
			}
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
		case "home":
			m.sel = 0
		case "end", "G":
			if x.String() == "G" && !m.vimMode {
				break
			}
			if len(m.tasks) > 0 {
				m.sel = len(m.tasks) - 1
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
		return m.renderEmpty()
	}

	visible := m.height
	start := clamp(m.sel-visible/2, 0, max(0, len(m.tasks)-visible))
	end := min(len(m.tasks), start+visible)

	lines := make([]string, 0, visible)
	for i := start; i < end; i++ {
		t := m.tasks[i]
		line := m.renderRow(t, i == m.sel)
		lines = append(lines, line)
	}

	// Pad remaining rows with background-filled empty lines.
	// The outer FillViewport also handles this, but doing it here
	// ensures the tasklist always returns a consistent height.
	emptyLine := lipgloss.NewStyle().
		Width(m.width).
		Background(m.styles.Theme.Bg).
		Render(strings.Repeat(" ", m.width))
	for len(lines) < visible {
		lines = append(lines, emptyLine)
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m Model) renderEmpty() string {
	icon := lipgloss.NewStyle().
		Foreground(m.styles.Theme.Accent).
		Background(m.styles.Theme.Bg).
		Bold(true).
		Render("\u2728 " + styles.IconTask)

	title := lipgloss.NewStyle().
		Foreground(m.styles.Theme.Fg).
		Background(m.styles.Theme.Bg).
		Bold(true).
		Margin(1, 0, 0, 0).
		Render("No tasks here yet")

	subtitle := lipgloss.NewStyle().
		Foreground(m.styles.Theme.Muted).
		Background(m.styles.Theme.Bg).
		Margin(1, 0, 0, 0).
		Render("Press 'n' to create a new task and start your journey")

	hint := lipgloss.NewStyle().
		Foreground(m.styles.Theme.Muted).
		Background(m.styles.Theme.Bg).
		Italic(true).
		Margin(2, 0, 0, 0).
		Render("Tip: Use the command palette (Ctrl+K) to access all features")

	content := lipgloss.JoinVertical(lipgloss.Center,
		icon,
		title,
		subtitle,
		hint,
	)

	// Place centered; FillViewport at the top level will handle ANSI fixup.
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceBackground(m.styles.Theme.Bg),
	)
}

func (m Model) renderRow(t core.Task, selected bool) string {
	// Check if this task is being animated
	isAnimating := m.animatingTaskID == t.ID && m.animatingTaskID != ""
	animProgress := 0.0
	if isAnimating {
		elapsed := time.Since(m.animationStart)
		if elapsed < m.animationDur {
			animProgress = float64(elapsed) / float64(m.animationDur)
		} else {
			animProgress = 1.0
		}
	}

	rowBg := m.styles.Theme.Bg
	if selected {
		rowBg = m.styles.Theme.Overlay
	}

	// Status icon
	statusIcon := styles.IconTodo
	statusStyle := lipgloss.NewStyle().Foreground(m.styles.Theme.Muted).Background(rowBg)
	switch t.Status {
	case core.StatusDoing:
		statusIcon = styles.IconDoing
		statusStyle = lipgloss.NewStyle().Foreground(m.styles.Theme.Warn).Background(rowBg)
	case core.StatusDone:
		statusIcon = styles.IconDone
		statusStyle = lipgloss.NewStyle().Foreground(m.styles.Theme.Good).Background(rowBg)
	}

	// Selection indicator
	indicator := lipgloss.NewStyle().Background(rowBg).Render("  ")
	if selected {
		indicator = lipgloss.NewStyle().Foreground(m.styles.Theme.Accent).Background(rowBg).Render("\u2503 ")
	}

	titleStyle := m.styles.RowNormal
	if selected {
		titleStyle = m.styles.RowSelected
	} else if t.Status == core.StatusDone {
		titleStyle = m.styles.RowDimmed.Strikethrough(true)
	}

	titleText := t.Title
	var title string
	if isAnimating {
		displayTitle := m.renderProgressiveStrikethrough(titleText, animProgress)
		title = m.styles.RowDimmed.Render(truncate(displayTitle, max(20, m.width-40)))
	} else {
		title = titleStyle.Render(truncate(titleText, max(20, m.width-40)))
	}

	// Build left side
	spaceBg := lipgloss.NewStyle().Background(rowBg).Render(" ")
	left := indicator + statusStyle.Render(statusIcon) + spaceBg + title

	rightParts := []string{}

	// Priority badge
	pri := m.styles.PriorityBadge(t.Priority)
	rightParts = append(rightParts, pri)

	// Deadline
	if t.Deadline != nil {
		deadText := humanDeadline(*t.Deadline, time.Now())
		deadStyle := m.styles.Muted
		if t.Deadline.Before(time.Now()) && t.Status != core.StatusDone {
			deadStyle = lipgloss.NewStyle().Foreground(m.styles.Theme.Bad).Background(rowBg)
		}
		rightParts = append(rightParts, deadStyle.Render(styles.IconDeadline+deadText))
	}

	// Tags
	if len(t.Tags) > 0 {
		tagStr := ""
		for i, tag := range t.Tags {
			if i > 0 {
				tagStr += " "
			}
			tagStr += "#" + tag
		}
		rightParts = append(rightParts, m.styles.Muted.Render(truncate(tagStr, max(10, m.width/6))))
	}

	right := strings.Join(rightParts, lipgloss.NewStyle().Background(rowBg).Render("  "))

	// Use render.BarLine: fills the gap between left and right with bg-styled spaces.
	// Subtract 2 for the Padding(0,1) applied by rowStyle below.
	innerWidth := m.width - 2
	if innerWidth < 0 {
		innerWidth = m.width
	}
	line := render.BarLine(left, right, innerWidth, rowBg)

	rowStyle := lipgloss.NewStyle().Width(m.width).Padding(0, 1).Background(rowBg)
	return rowStyle.Render(line)
}

func (m Model) renderProgressiveStrikethrough(text string, progress float64) string {
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}

	runes := []rune(text)

	if m.animationReverse {
		// Reverse animation: start fully struckthrough, remove from left to right
		normalCount := int(float64(len(runes)) * progress)

		if normalCount == 0 {
			// All struckthrough
			return m.styles.RowDimmed.Strikethrough(true).Render(text)
		}
		if normalCount >= len(runes) {
			// All normal
			return text
		}

		// Split: normal part on left, struckthrough on right
		normal := m.styles.RowDimmed.Render(string(runes[:normalCount]))
		strikethrough := m.styles.RowDimmed.Strikethrough(true).Render(string(runes[normalCount:]))
		return normal + strikethrough
	} else {
		// Forward animation: start normal, apply strikethrough from left to right
		strikeCount := int(float64(len(runes)) * progress)

		if strikeCount == 0 {
			return text
		}
		if strikeCount >= len(runes) {
			// All struckthrough
			return m.styles.RowDimmed.Strikethrough(true).Render(text)
		}

		// Split: struckthrough on left, normal on right
		strikethrough := m.styles.RowDimmed.Strikethrough(true).Render(string(runes[:strikeCount]))
		normal := m.styles.RowDimmed.Render(string(runes[strikeCount:]))
		return strikethrough + normal
	}
}

func truncate(s string, w int) string {
	if w <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= w {
		return s
	}
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
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 36*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return fmt.Sprintf("%dd", int(d.Hours()/24))
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
