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

type TaskItem struct {
	core.Task
	Depth       int
	HasChildren bool
}

type Model struct {
	styles     styles.Styles
	vimMode    bool
	Animations bool
	km         keymap.Keymap
	rightOrder []string

	width  int
	height int

	items    []TaskItem
	allTasks []core.Task // All tasks for stats calculation
	sel      int

	// Animation state - set by the app model, read-only during render.
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

	DueMinimal bool

	TagsHighlight map[string]string // Added for tag highlighting
	selectedIDs   map[string]bool   // Added for multi-selection
	lastKey       string            // For tracking key sequences like 'gg'
}

func New(s styles.Styles, vimMode bool, animations bool, km keymap.Keymap, dueMinimal bool) Model {
	return Model{
		styles:        s,
		vimMode:       vimMode,
		Animations:    animations,
		km:            km,
		rightOrder:    []string{"tags", "due", "priority"},
		DueMinimal:    dueMinimal,
		selectedIDs:   make(map[string]bool),
		TagsHighlight: make(map[string]string),
	}
}

func (m *Model) SetTagsConfig(cfg map[string]string) {
	m.TagsHighlight = cfg
}

func (m *Model) SetRightOrder(order []string) {
	if len(order) == 0 {
		m.rightOrder = []string{"tags", "due", "priority"}
		return
	}
	m.rightOrder = append([]string(nil), order...)
}

func (m Model) Selected() (TaskItem, bool) {
	if m.sel < 0 || m.sel >= len(m.items) {
		return TaskItem{}, false
	}
	return m.items[m.sel], true
}

func (m *Model) SetSize(w, h int) {
	m.width, m.height = w, h
}

func (m *Model) SetTasks(ts []core.Task) {
	ts = filterWaitUntil(ts, time.Now())
	m.items = buildVisibleTree(ts)
	if m.sel >= len(m.items) {
		m.sel = len(m.items) - 1
	}
	if m.sel < 0 {
		m.sel = 0
	}
}

func filterWaitUntil(ts []core.Task, now time.Time) []core.Task {
	if len(ts) == 0 {
		return ts
	}

	byID := make(map[string]core.Task, len(ts))
	visible := make(map[string]bool, len(ts))
	for _, t := range ts {
		byID[t.ID] = t
		visible[t.ID] = t.WaitUntil == nil || !now.Before(*t.WaitUntil)
	}

	// If a task is hidden, also hide its descendants to avoid orphaning.
	memo := map[string]bool{}
	var isVisible func(id string) bool
	isVisible = func(id string) bool {
		if v, ok := memo[id]; ok {
			return v
		}
		t, ok := byID[id]
		if !ok {
			memo[id] = true
			return true
		}
		if !visible[id] {
			memo[id] = false
			return false
		}
		if t.ParentID == "" {
			memo[id] = true
			return true
		}
		// If parent doesn't exist, treat this as a root.
		if _, ok := byID[t.ParentID]; !ok {
			memo[id] = true
			return true
		}
		v := isVisible(t.ParentID)
		memo[id] = v
		return v
	}

	out := make([]core.Task, 0, len(ts))
	for _, t := range ts {
		if isVisible(t.ID) {
			out = append(out, t)
		}
	}
	return out
}

func buildVisibleTree(ts []core.Task) []TaskItem {
	taskMap := make(map[string]core.Task)
	for _, t := range ts {
		taskMap[t.ID] = t
	}

	children := make(map[string][]core.Task)
	for _, t := range ts {
		if t.ParentID != "" {
			children[t.ParentID] = append(children[t.ParentID], t)
		}
	}

	var roots []core.Task
	for _, t := range ts {
		if t.ParentID == "" {
			roots = append(roots, t)
		} else if _, ok := taskMap[t.ParentID]; !ok {
			roots = append(roots, t)
		}
	}

	var out []TaskItem
	var walk func(t core.Task, depth int)
	walk = func(t core.Task, depth int) {
		kids := children[t.ID]
		out = append(out, TaskItem{
			Task:        t,
			Depth:       depth,
			HasChildren: len(kids) > 0,
		})

		if !t.Collapsed {
			for _, kid := range kids {
				walk(kid, depth+1)
			}
		}
	}

	for _, r := range roots {
		walk(r, 0)
	}
	return out
}

func (m *Model) SetAllTasks(ts []core.Task) {
	m.allTasks = append([]core.Task(nil), ts...)
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

func (m Model) GetSelectedTasks() []core.Task {
	var selected []core.Task
	for _, item := range m.items {
		if m.selectedIDs[item.ID] {
			selected = append(selected, item.Task)
		}
	}
	// If nothing is selected, fall back to current selection (as per original functionality)
	if len(selected) == 0 {
		if item, ok := m.Selected(); ok {
			selected = append(selected, item.Task)
		}
	}
	return selected
}

func (m Model) GetVisibleTasks() []TaskItem {
	return m.items
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch x := msg.(type) {
	case tea.KeyMsg:
		switch x.String() {
		case "up", "k":
			m.lastKey = ""
			if x.String() == "k" && !m.vimMode {
				break
			}
			if m.sel > 0 {
				m.sel--
			}
		case "down", "j":
			m.lastKey = ""
			if x.String() == "j" && !m.vimMode {
				break
			}
			if m.sel < len(m.items)-1 {
				m.sel++
			}
		case "pgup":
			m.lastKey = ""
			m.sel -= max(1, m.height-4)
			if m.sel < 0 {
				m.sel = 0
			}
		case "pgdown":
			m.lastKey = ""
			m.sel += max(1, m.height-4)
			if m.sel > len(m.items)-1 {
				m.sel = len(m.items) - 1
			}
		case "home":
			m.lastKey = ""
			m.sel = 0
		case "end", "G":
			m.lastKey = ""
			if x.String() == "G" && !m.vimMode {
				break
			}
			if len(m.items) > 0 {
				m.sel = len(m.items) - 1
			}
		case "g":
			if !m.vimMode {
				break
			}
			if m.lastKey == "g" {
				m.sel = 0
				m.lastKey = ""
			} else {
				m.lastKey = "g"
				return m, nil // Don't reset lastKey yet
			}
		case " ":
			m.lastKey = ""
			if item, ok := m.Selected(); ok {
				if m.selectedIDs[item.ID] {
					delete(m.selectedIDs, item.ID)
				} else {
					m.selectedIDs[item.ID] = true
				}
			}
		case "esc":
			m.lastKey = ""
			m.selectedIDs = make(map[string]bool)
		default:
			m.lastKey = ""
		}
	}
	return m, nil
}

func (m Model) View() string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}
	if len(m.items) == 0 {
		return m.renderEmpty()
	}

	visible := m.height
	start := clamp(m.sel-visible/2, 0, max(0, len(m.items)-visible))
	end := min(len(m.items), start+visible)

	maxDueWidth := 0
	if m.DueMinimal {
		for i := start; i < end; i++ {
			item := m.items[i]
			if item.Deadline != nil {
				d := humanDeadline(*item.Deadline, time.Now())
				d = formatDue(d, true)
				w := lipgloss.Width(styles.IconDeadline + d)
				if w > maxDueWidth {
					maxDueWidth = w
				}
			}
		}
	}

	lines := make([]string, 0, visible)
	for i := start; i < end; i++ {
		item := m.items[i]

		// Cascading reveal: wait until view transition reaches a threshold for this row
		if m.Animations && m.ViewTransitioning && m.ViewTransitionProgress < 1.0 {
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

		line := m.renderRow(item, i == m.sel, maxDueWidth)
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
	boxWidth := min(64, m.width-4)
	accent := m.styles.Theme.Accent

	// 1. Productivity Velocity
	completedCount := 0
	for _, t := range m.allTasks {
		if t.Status == core.StatusDone {
			completedCount++
		}
	}

	velocity := ""
	if completedCount > 0 {
		velocity = lipgloss.NewStyle().
			Foreground(m.styles.Theme.Good).
			Background(m.styles.Theme.Bg).
			Padding(1, 0).
			Width(boxWidth).
			Align(lipgloss.Center).
			Render(fmt.Sprintf("󰄬 %d Tasks Completed", completedCount))
	} else {
		velocity = lipgloss.NewStyle().
			Foreground(m.styles.Theme.Muted).
			Padding(1, 0).
			Width(boxWidth).
			Align(lipgloss.Center).
			Render("No recent velocity data.")
	}

	// 2. Greeting & Motivation
	hour := time.Now().Hour()
	greeting := "Good evening"
	if hour < 12 {
		greeting = "Good morning"
	} else if hour < 18 {
		greeting = "Good afternoon"
	}

	header := lipgloss.NewStyle().
		Foreground(accent).
		Bold(true).
		Width(boxWidth).
		Align(lipgloss.Center).
		Render(greeting)

	body := lipgloss.NewStyle().
		Foreground(m.styles.Theme.Fg).
		Width(boxWidth).
		Align(lipgloss.Center).
		Render("Nothing on the horizon. Your schedule is clear.")

	// 3. Action Hint
	paletteKeys := strings.Join(m.km.Palette.Keys(), ", ")
	hint := lipgloss.NewStyle().
		Foreground(m.styles.Theme.Muted).
		Italic(true).
		MarginTop(2).
		Width(boxWidth).
		Align(lipgloss.Center).
		Render(fmt.Sprintf("Press 'n' to create or %s for tools", paletteKeys))

	// Composite inside a minimal border
	dashboard := lipgloss.JoinVertical(lipgloss.Center,
		header,
		body,
		velocity,
		hint,
	)

	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(m.styles.Overlay.Width(boxWidth).Padding(2).Render(dashboard))
}
func (m Model) renderRow(item TaskItem, selected bool, maxDueWidth int) string {
	t := item.Task
	// Compute animation progress for strike (completion toggle).
	// Progress is always clamped to [0, 1] - no overshoot.
	isAnimating := m.Animations && m.animatingTaskID == t.ID && m.animatingTaskID != ""
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
	isCreating := m.Animations && m.creatingTaskID == t.ID && m.creatingTaskID != ""
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

	// Selection indicator - render only for selected task
	indicator := " "
	if selected || m.selectedIDs[t.ID] {
		indicatorStyle := m.styles.Theme.Muted
		if m.selectedIDs[t.ID] {
			indicatorStyle = m.styles.Theme.Accent
		}
		indicator = lipgloss.NewStyle().
			Foreground(indicatorStyle).
			Background(rowBg).
			Bold(true).
			Render("│")
	}

	// Hierarchy indentation and icons
	indent := strings.Repeat("  ", item.Depth)
	expandIcon := "  "
	if item.HasChildren {
		if t.Collapsed {
			expandIcon = "▶ "
		} else {
			expandIcon = "▼ "
		}
	}

	titleStyle := m.styles.RowNormal
	if selected {
		titleStyle = m.styles.RowSelected
	} else if t.Status == core.StatusDone {
		titleStyle = m.styles.RowDimmed.Strikethrough(true)
	}

	titleText := t.Title

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

	order := m.rightOrder
	if len(order) == 0 {
		order = []string{"tags", "due", "priority"}
	}

	rightParts := []string{}
	for _, f := range order {
		switch f {
		case "priority":
			pri := m.styles.PriorityBadge(t.Priority)
			rightParts = append(rightParts, lipgloss.NewStyle().Width(8).Align(lipgloss.Right).Render(pri))
		case "due":
			if t.Deadline != nil {
				now := time.Now()
				deadText := humanDeadline(*t.Deadline, now)
				deadStyleColor := m.styles.Theme.Muted
				if t.Deadline.Before(now) && t.Status != core.StatusDone {
					deadStyleColor = m.styles.Theme.Bad
				}

				dueContent := styles.IconDeadline + formatDue(deadText, m.DueMinimal)

				badge := m.styles.BadgeMuted.
					Background(m.styles.Theme.Muted).
					Foreground(m.styles.Theme.Bg).
					Padding(0, 0)

				pill := lipgloss.JoinHorizontal(lipgloss.Left,
					m.styles.TagLeft.Foreground(deadStyleColor).Render(),
					badge.Background(deadStyleColor).Render(dueContent),
					m.styles.TagRight.Foreground(deadStyleColor).Render(),
				)

				styledPill := lipgloss.NewStyle().Width(16).Align(lipgloss.Center).Render(pill)
				rightParts = append(rightParts, styledPill)
			} else {
				dash := lipgloss.NewStyle().
					Foreground(m.styles.Theme.Muted).
					Width(16).
					Align(lipgloss.Center).
					Render("-")
				rightParts = append(rightParts, dash)
			}
		case "tags":
			tagStr := ""
			if len(t.Tags) > 0 {
				tagParts := []string{}
				for _, tag := range t.Tags {
					tagStyle := m.styles.Tag
					if highlightStr, ok := m.TagsHighlight[tag]; ok {
						highlight := styles.ParseTagHighlightString(highlightStr)
						tagStyle = styles.ApplyTagHighlight(m.styles.Tag, highlight, m.styles.Theme)
					}

					tagContent := tagStyle.
						Padding(0, 0).
						Render(tag)
					pill := lipgloss.JoinHorizontal(lipgloss.Left,
						m.styles.TagLeft.Foreground(tagStyle.GetBackground()).Render(),
						tagContent,
						m.styles.TagRight.Foreground(tagStyle.GetBackground()).Render(),
					)
					tagParts = append(tagParts, pill)
				}
				tagStr = strings.Join(tagParts, " ")
			}
			rightParts = append(rightParts, tagStr)
		case "open_issue_id":
			if t.OpenIssueID != "" {
				pill := lipgloss.NewStyle().
					Width(12).
					Align(lipgloss.Right).
					Foreground(m.styles.Theme.Muted).
					Render(t.OpenIssueID)
				rightParts = append(rightParts, pill)
			} else {
				pill := lipgloss.NewStyle().
					Width(12).
					Render("")
				rightParts = append(rightParts, pill)
			}
		case "project":
			var projPill string
			if t.Project != "" {
				projPill = lipgloss.JoinHorizontal(lipgloss.Left,
					m.styles.TagLeft.Foreground(m.styles.Theme.Muted).Render(),
					m.styles.Tag.
						Background(m.styles.Theme.Muted).
						Foreground(m.styles.Theme.Bg).
						Padding(0, 0).
						Render(t.Project),
					m.styles.TagRight.Foreground(m.styles.Theme.Muted).Render(),
				)
			}
			rightParts = append(rightParts, lipgloss.NewStyle().Width(14).Align(lipgloss.Right).Render(projPill))
		}
	}

	var right string
	if len(rightParts) == 0 {
		right = ""
	} else {
		right = rightParts[0]
		for i := 1; i < len(rightParts); i++ {
			right = lipgloss.JoinHorizontal(lipgloss.Right, right, lipgloss.NewStyle().Width(1).Render(""), rightParts[i])
		}
	}

	rightWidth := lipgloss.Width(right)
	innerWidth := m.width - 2
	if innerWidth < 0 {
		innerWidth = m.width
	}

	var title string
	if isAnimating {
		title = m.renderStrikeWipe(titleText, animProgress, rowBg)
	} else {
		leftPrefixWidth := lipgloss.Width(indicator) + lipgloss.Width(indent) + lipgloss.Width(expandIcon) + lipgloss.Width(statusIcon) + 1
		titleWidth := innerWidth - leftPrefixWidth - rightWidth
		if titleWidth < 10 {
			titleWidth = 10
		}
		title = titleStyle.Render(truncate(titleText, titleWidth))
	}

	left := indicator +
		lipgloss.NewStyle().Background(rowBg).Render(indent) +
		lipgloss.NewStyle().Foreground(m.styles.Theme.Accent).Background(rowBg).Render(expandIcon) +
		statusStyle.Render(statusIcon) +
		lipgloss.NewStyle().Background(rowBg).Render(" ") +
		title

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

func formatDue(due string, minimal bool) string {
	if !minimal {
		return due
	}
	return strings.ReplaceAll(due, "overdue", "OD")
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
