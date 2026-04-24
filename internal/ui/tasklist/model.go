package tasklist

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/programmersd21/kairo/internal/core"
	"github.com/programmersd21/kairo/internal/ui/keymap"
	"github.com/programmersd21/kairo/internal/ui/render"
	"github.com/programmersd21/kairo/internal/ui/styles"
)

type Model struct {
	styles  styles.Styles
	vimMode bool
	km      keymap.Keymap

	width  int
	height int

	tasks []core.Task
	sel   int

	// Animation state — set by the app model, read-only during render.
	animatingTaskID  string
	animationStart   time.Time
	animationDur     time.Duration
	animationReverse bool

	creatingTaskID string
	creationStart  time.Time
	creationDur    time.Duration

	ViewTransitioning      bool
	ViewTransitionProgress float64

	DeletingTaskID string
	DeleteProgress float64
}

func New(s styles.Styles, vimMode bool, km keymap.Keymap) Model {
	return Model{styles: s, vimMode: vimMode, km: km}
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

func (m *Model) SetCreationAnimation(taskID string, start time.Time, duration time.Duration) {
	m.creatingTaskID = taskID
	m.creationStart = start
	m.creationDur = duration
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

		// Cascading reveal: wait until view transition reaches a threshold for this row
		if m.ViewTransitioning && m.ViewTransitionProgress < 1.0 {
			idx := i - start
			startThresh := float64(idx) * 0.05
			if m.ViewTransitionProgress < startThresh {
				// Return background-filled empty line
				emptyLine := lipgloss.NewStyle().
					Width(m.width).
					Background(m.styles.Theme.Bg).
					Render(strings.Repeat(" ", m.width))
				lines = append(lines, emptyLine)
				continue
			}
		}

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
	boxWidth := min(60, m.width-4)

	icon := lipgloss.NewStyle().
		Foreground(m.styles.Theme.Accent).
		Background(m.styles.Theme.Bg).
		Bold(true).
		Width(boxWidth).
		Align(lipgloss.Center).
		Render("\u2728 " + styles.IconTask)

	title := lipgloss.NewStyle().
		Foreground(m.styles.Theme.Fg).
		Background(m.styles.Theme.Bg).
		Bold(true).
		Margin(1, 0, 0, 0).
		Width(boxWidth).
		Align(lipgloss.Center).
		Render("No tasks here yet")

	subtitle := lipgloss.NewStyle().
		Foreground(m.styles.Theme.Muted).
		Background(m.styles.Theme.Bg).
		Margin(1, 0, 0, 0).
		Width(boxWidth).
		Align(lipgloss.Center).
		Render("Press 'n' to create a new task and start your journey")

	paletteKeys := strings.Join(m.km.Palette.Keys(), ", ")
	hint := lipgloss.NewStyle().
		Foreground(m.styles.Theme.Muted).
		Background(m.styles.Theme.Bg).
		Italic(true).
		Margin(2, 0, 0, 0).
		Width(boxWidth).
		Align(lipgloss.Center).
		Render(fmt.Sprintf("Tip: Use the command palette (%s) to access all features", paletteKeys))

	content := lipgloss.JoinVertical(lipgloss.Left,
		icon,
		title,
		subtitle,
		hint,
	)

	card := m.styles.Card.
		Width(boxWidth).
		Padding(2, 0). // Horizontal padding removed to respect explicit line widths
		Render(content)

	// Place centered; FillViewport at the top level will handle ANSI fixup.
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, card,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceBackground(m.styles.Theme.Bg),
	)
}

func (m Model) renderRow(t core.Task, selected bool) string {
	// Compute animation progress for strike (completion toggle).
	// Progress is always clamped to [0, 1] — no overshoot.
	isAnimating := m.animatingTaskID == t.ID && m.animatingTaskID != ""
	animProgress := 0.0
	if isAnimating {
		elapsed := time.Since(m.animationStart)
		if elapsed < m.animationDur {
			raw := float64(elapsed) / float64(m.animationDur)
			animProgress = render.EaseOutCubic(raw)
		} else {
			animProgress = 1.0
		}
	}

	// Compute animation progress for bloom (new task creation).
	isCreating := m.creatingTaskID == t.ID && m.creatingTaskID != ""
	creationProgress := 0.0
	if isCreating {
		elapsed := time.Since(m.creationStart)
		if elapsed < m.creationDur {
			raw := float64(elapsed) / float64(m.creationDur)
			creationProgress = render.EaseOutQuad(raw)
		} else {
			creationProgress = 1.0
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

	// Selection indicator — stays in place, no spatial shifting
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

	// Bombastic "Glitch & Vaporize" Deletion Animation
	isDeleting := m.DeletingTaskID == t.ID
	if isDeleting && m.DeleteProgress > 0 {
		titleStyle = m.styles.RowDimmed.Foreground(m.styles.Theme.Bad) // Turn text red
		statusIcon = "✖"
		statusStyle = lipgloss.NewStyle().Foreground(m.styles.Theme.Bad).Background(rowBg)

		runes := []rune(titleText)
		particles := []rune{'*', 'x', '.', ' ', '·', 'º'}
		glitchStart := int(float64(len(runes)) * m.DeleteProgress)
		for i := glitchStart; i < len(runes); i++ {
			// Scramble characters based on position and progress
			if (i*7+int(m.DeleteProgress*100))%3 == 0 {
				runes[i] = particles[(i+int(m.DeleteProgress*10))%len(particles)]
			}
		}
		titleText = string(runes)
		// Truncate length progressively to "shrink" the task into nothing
		shrinkLen := len(runes) - int(float64(len(runes))*(m.DeleteProgress*m.DeleteProgress))
		if shrinkLen < 0 {
			shrinkLen = 0
		}
		if shrinkLen < len(runes) {
			titleText = string(runes[:shrinkLen])
		}
	}

	// Bloom: progressive character reveal with smooth easing.
	// Characters appear left-to-right. No spatial shifting of the row.
	if isCreating && creationProgress < 1.0 {
		runes := []rune(titleText)
		showCount := int(float64(len(runes)) * creationProgress)
		if showCount < 0 {
			showCount = 0
		}
		if showCount > len(runes) {
			showCount = len(runes)
		}
		titleText = string(runes[:showCount])
	}

	var title string
	if isAnimating {
		// Clean left-to-right strikethrough wipe
		title = m.renderStrikeWipe(titleText, animProgress, rowBg)
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
		tagParts := []string{}
		for _, tag := range t.Tags {
			pill := lipgloss.JoinHorizontal(lipgloss.Left,
				m.styles.TagLeft.Render(),
				m.styles.Tag.Render(tag),
				m.styles.TagRight.Render(),
			)
			tagParts = append(tagParts, pill)
		}
		rightParts = append(rightParts, strings.Join(tagParts, " "))
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

// renderStrikeWipe renders a clean left-to-right strikethrough animation.
// Progress [0, 1] controls how much of the text is struck through.
//
// Forward (Todo → Done): characters progressively gain strikethrough + dim.
// Reverse (Done → Todo): characters progressively lose strikethrough from left.
func (m Model) renderStrikeWipe(text string, progress float64, rowBg lipgloss.Color) string {
	progress = render.Clamp01(progress)

	runes := []rune(text)
	if len(runes) == 0 {
		return ""
	}

	maxWidth := max(20, m.width-40)
	text = truncate(text, maxWidth)
	runes = []rune(text)

	splitIdx := int(float64(len(runes)) * progress)
	if splitIdx > len(runes) {
		splitIdx = len(runes)
	}

	struckStyle := m.styles.RowDimmed.Strikethrough(true).Background(rowBg)
	normalStyle := m.styles.RowNormal.Background(rowBg)

	if m.animationReverse {
		// Reverse: left portion clears strikethrough, right stays struck
		if splitIdx >= len(runes) {
			return normalStyle.Render(text)
		}
		cleared := normalStyle.Render(string(runes[:splitIdx]))
		remaining := struckStyle.Render(string(runes[splitIdx:]))
		return cleared + remaining
	}

	// Forward: left portion gets struck, right stays normal
	if splitIdx >= len(runes) {
		return struckStyle.Render(text)
	}
	struck := struckStyle.Render(string(runes[:splitIdx]))
	remaining := normalStyle.Render(string(runes[splitIdx:]))
	return struck + remaining
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
