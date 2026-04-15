package editor

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/programmersd21/kairo/internal/core"
	"github.com/programmersd21/kairo/internal/core/nlp"
	"github.com/programmersd21/kairo/internal/ui/styles"
)

type Mode int

const (
	ModeNew Mode = iota
	ModeEdit
)

type SaveNewMsg struct{ Task core.Task }
type SavePatchMsg struct {
	ID    string
	Patch core.TaskPatch
}
type CloseMsg struct{}

type Model struct {
	styles styles.Styles
	mode   Mode

	width  int
	height int

	orig core.Task

	title    textinput.Model
	tags     textinput.Model
	priority textinput.Model
	deadline textinput.Model
	status   textinput.Model
	desc     textarea.Model

	focus int

	deadlinePreview string
	deadlineValue   *time.Time
	deadlineErr     string
}

func New(s styles.Styles, mode Mode, t core.Task) Model {
	ti := textinput.New()
	ti.Prompt = "Title: "
	ti.CharLimit = 200
	ti.SetValue(strings.TrimSpace(t.Title))
	ti.Focus()

	tags := textinput.New()
	tags.Prompt = "Tags:  "
	tags.CharLimit = 200
	if len(t.Tags) > 0 {
		tags.SetValue("#" + strings.Join(t.Tags, " #"))
	}

	pr := textinput.New()
	pr.Prompt = "Pri:   "
	pr.CharLimit = 2
	pr.SetValue(fmt.Sprintf("%d", int(t.Priority.Clamp())))

	dl := textinput.New()
	dl.Prompt = "Due:   "
	dl.CharLimit = 64
	if t.Deadline != nil {
		dl.SetValue(t.Deadline.Local().Format("2006-01-02 15:04"))
	}

	st := textinput.New()
	st.Prompt = "Status:"
	st.CharLimit = 8
	if t.Status == "" {
		st.SetValue(string(core.StatusTodo))
	} else {
		st.SetValue(string(t.Status))
	}

	d := textarea.New()
	d.Placeholder = "Description (Markdown)…"
	d.SetValue(t.Description)
	d.Focus()
	d.Blur()
	d.ShowLineNumbers = false

	m := Model{
		styles:   s,
		mode:     mode,
		orig:     t,
		title:    ti,
		tags:     tags,
		priority: pr,
		deadline: dl,
		status:   st,
		desc:     d,
		focus:    0,
	}
	m.recomputeDeadline()
	return m
}

func (m *Model) SetSize(w, h int) {
	m.width, m.height = w, h
	m.desc.SetWidth(max(20, w-4))
	m.desc.SetHeight(max(4, h-12))
	m.title.Width = max(20, w-10)
	m.tags.Width = max(20, w-10)
	m.priority.Width = 6
	m.deadline.Width = max(20, w-10)
	m.status.Width = 10
}

func (m Model) Init() tea.Cmd { return textinput.Blink }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch x := msg.(type) {
	case tea.KeyMsg:
		switch x.String() {
		case "esc":
			return m, func() tea.Msg { return CloseMsg{} }
		case "tab":
			m.blurAll()
			m.focus = (m.focus + 1) % 6
			m.focusField()
			return m, nil
		case "shift+tab":
			m.blurAll()
			m.focus--
			if m.focus < 0 {
				m.focus = 5
			}
			m.focusField()
			return m, nil
		case "ctrl+s":
			return m, m.saveCmd()
		}
	}

	var cmd tea.Cmd
	switch m.focus {
	case 0:
		m.title, cmd = m.title.Update(msg)
	case 1:
		m.tags, cmd = m.tags.Update(msg)
	case 2:
		m.priority, cmd = m.priority.Update(msg)
	case 3:
		prev := m.deadline.Value()
		m.deadline, cmd = m.deadline.Update(msg)
		if m.deadline.Value() != prev {
			m.recomputeDeadline()
		}
	case 4:
		m.status, cmd = m.status.Update(msg)
	case 5:
		m.desc, cmd = m.desc.Update(msg)
	}
	return m, cmd
}

func (m Model) View() string {
	w := m.width
	if w <= 0 {
		w = 80
	}
	cardW := min(80, w-6)

	title := "NEW TASK"
	if m.mode == ModeEdit {
		title = "EDIT TASK"
	}

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.styles.Theme.Accent).
		Padding(0, 1).
		Render(title)

	help := m.styles.Muted.Padding(0, 1).Render("ctrl+s save • esc cancel • tab navigate")

	// Input fields with labels
	fields := []string{
		m.renderField("TITLE", m.title.View(), 0 == m.focus),
		m.renderField("TAGS", m.tags.View(), 1 == m.focus),
		lipgloss.JoinHorizontal(lipgloss.Left,
			m.renderField("PRIORITY", m.priority.View(), 2 == m.focus),
			m.renderField("STATUS", m.status.View(), 4 == m.focus),
		),
		m.renderField("DUE", m.deadline.View(), 3 == m.focus),
	}

	if m.deadlineErr != "" {
		fields = append(fields, lipgloss.NewStyle().Foreground(m.styles.Theme.Bad).Padding(0, 2).Render("ERROR: "+m.deadlineErr))
	} else if m.deadlinePreview != "" {
		fields = append(fields, m.styles.Muted.Padding(0, 2).Render("→ "+m.deadlinePreview))
	}

	descView := m.desc.View()
	descBox := lipgloss.NewStyle().
		Padding(0, 1).
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderLeftForeground(m.styles.Theme.Border)

	if 5 == m.focus {
		descBox = descBox.BorderLeftForeground(m.styles.Theme.Accent)
	}

	fields = append(fields, "", lipgloss.NewStyle().Padding(0, 2).Render(descBox.Render(descView)))

	content := lipgloss.JoinVertical(lipgloss.Left, append([]string{header, help, ""}, fields...)...)

	return lipgloss.NewStyle().
		Width(cardW).
		Background(m.styles.Theme.Bg).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.styles.Theme.Accent).
		Padding(1, 2).
		Render(content)
}

func (m Model) renderField(label, input string, focused bool) string {
	labelStyle := lipgloss.NewStyle().
		Foreground(m.styles.Theme.Muted).
		Bold(true).
		Width(12)

	if focused {
		labelStyle = labelStyle.Foreground(m.styles.Theme.Accent)
	}

	return lipgloss.NewStyle().Padding(0, 2).Render(
		lipgloss.JoinHorizontal(lipgloss.Left, labelStyle.Render(label), input),
	)
}

func (m *Model) blurAll() {
	m.title.Blur()
	m.tags.Blur()
	m.priority.Blur()
	m.deadline.Blur()
	m.status.Blur()
	m.desc.Blur()
}

func (m *Model) focusField() {
	switch m.focus {
	case 0:
		m.title.Focus()
	case 1:
		m.tags.Focus()
	case 2:
		m.priority.Focus()
	case 3:
		m.deadline.Focus()
	case 4:
		m.status.Focus()
	case 5:
		m.desc.Focus()
	}
}

func (m Model) renderInput(s string) string {
	return lipgloss.NewStyle().Padding(0, 1).Render(s)
}

func (m *Model) recomputeDeadline() {
	m.deadlineErr = ""
	m.deadlinePreview = ""
	m.deadlineValue = nil
	raw := strings.TrimSpace(m.deadline.Value())
	if raw == "" {
		return
	}
	t, err := nlp.ParseDeadline(raw, time.Now())
	if err != nil {
		m.deadlineErr = err.Error()
		return
	}
	if t == nil {
		return
	}
	m.deadlineValue = t
	m.deadlinePreview = t.Local().Format("Mon Jan 2 15:04")
}

func (m Model) saveCmd() tea.Cmd {
	title := strings.TrimSpace(m.title.Value())
	desc := strings.TrimSpace(m.desc.Value())
	tags := core.ParseTags(m.tags.Value())
	priRaw := strings.TrimSpace(m.priority.Value())
	priInt, _ := strconv.Atoi(priRaw)
	pri := core.Priority(priInt).Clamp()
	st := core.Status(strings.ToLower(strings.TrimSpace(m.status.Value())))
	if st == "" {
		st = core.StatusTodo
	}

	var deadline *time.Time
	if m.deadlineValue != nil {
		d := (*m.deadlineValue).UTC()
		deadline = &d
	}

	if m.mode == ModeNew {
		task := core.Task{
			Title:       title,
			Description: desc,
			Tags:        tags,
			Priority:    pri,
			Deadline:    deadline,
			Status:      st,
		}
		return func() tea.Msg { return SaveNewMsg{Task: task} }
	}

	patch := core.TaskPatch{}
	if title != m.orig.Title {
		patch.Title = &title
	}
	if desc != m.orig.Description {
		patch.Description = &desc
	}
	nt := core.Task{Tags: tags}.NormalizedTags()
	ot := core.Task{Tags: m.orig.Tags}.NormalizedTags()
	if strings.Join(nt, ",") != strings.Join(ot, ",") {
		patch.Tags = &nt
	}
	if pri != m.orig.Priority {
		patch.Priority = &pri
	}
	if (deadline == nil) != (m.orig.Deadline == nil) || (deadline != nil && m.orig.Deadline != nil && !deadline.Equal(*m.orig.Deadline)) {
		d := deadline
		patch.Deadline = &d
	}
	if st != m.orig.Status {
		patch.Status = &st
	}
	return func() tea.Msg { return SavePatchMsg{ID: m.orig.ID, Patch: patch} }
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
