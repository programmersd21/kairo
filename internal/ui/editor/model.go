package editor

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
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

	showPreview bool
	renderer    *glamour.TermRenderer
}

func New(s styles.Styles, mode Mode, t core.Task) Model {
	ti := textinput.New()
	ti.Prompt = styles.IconTask + "Title: "
	ti.CharLimit = 200
	ti.SetValue(strings.TrimSpace(t.Title))
	ti.Focus()

	tags := textinput.New()
	tags.Prompt = styles.IconTag + "Tags:  "
	tags.CharLimit = 200
	if len(t.Tags) > 0 {
		tags.SetValue("#" + strings.Join(t.Tags, " #"))
	}

	pr := textinput.New()
	pr.Prompt = styles.IconPriority1 + "Pri:   "
	pr.CharLimit = 2
	pr.SetValue(fmt.Sprintf("%d", int(t.Priority.Clamp())))

	dl := textinput.New()
	dl.Prompt = styles.IconDeadline + "Due:   "
	dl.CharLimit = 64
	if t.Deadline != nil {
		dl.SetValue(t.Deadline.Local().Format("2006-01-02 15:04"))
	}

	st := textinput.New()
	st.Prompt = styles.IconDoing + "Status:"
	st.CharLimit = 16
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
		styles:      s,
		mode:        mode,
		orig:        t,
		title:       ti,
		tags:        tags,
		priority:    pr,
		deadline:    dl,
		status:      st,
		desc:        d,
		focus:       0,
		showPreview: true,
	}
	m.recomputeDeadline()
	return m
}

func (m *Model) SetSize(w, h int) {
	m.width, m.height = w, h

	// Base sizes
	editorW := max(20, min(80, w-10))
	if w > 120 && m.showPreview {
		editorW = w / 2
	}

	m.desc.SetWidth(max(20, editorW-10))
	m.desc.SetHeight(max(4, h-16))
	m.title.Width = max(20, editorW-20)
	m.tags.Width = max(20, editorW-20)
	m.priority.Width = 6
	m.deadline.Width = max(20, editorW-20)
	m.status.Width = 10

	// Recreate renderer with new width
	style := "dark"
	if m.styles.Theme.IsLight {
		style = "light"
	}

	previewW := w - editorW - 10
	r, _ := glamour.NewTermRenderer(
		glamour.WithStandardStyle(style),
		glamour.WithWordWrap(max(20, previewW-4)),
	)
	m.renderer = r
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
		case "ctrl+p":
			m.showPreview = !m.showPreview
			m.SetSize(m.width, m.height)
			return m, nil
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

	useSplit := m.width > 120 && m.showPreview
	cardW := min(84, w-6)
	if useSplit {
		cardW = m.width - 4
	}

	titleText := "NEW TASK"
	if m.mode == ModeEdit {
		titleText = "EDIT TASK"
	}
	header := m.styles.Title.Padding(0, 1).MarginBottom(1).Render(titleText)

	fields := []string{
		header,
		m.renderField("Title", m.title.View(), m.focus == 0),
		m.renderField("Tags", m.tags.View(), m.focus == 1),
		lipgloss.JoinHorizontal(lipgloss.Left,
			m.renderField("Pri", m.priority.View(), m.focus == 2),
			m.renderField("Status", m.status.View(), m.focus == 4),
		),
		m.renderField("Due", m.deadline.View(), m.focus == 3),
	}

	if m.deadlineErr != "" {
		fields = append(fields, m.styles.Error.Padding(0, 2).Render(m.deadlineErr))
	} else if m.deadlinePreview != "" {
		fields = append(fields, m.styles.Muted.Padding(0, 2).Render(m.deadlinePreview))
	}

	descView := m.desc.View()
	fields = append(fields, "", lipgloss.NewStyle().Padding(0, 2).Render(descView))

	editorContent := lipgloss.JoinVertical(lipgloss.Left, fields...)

	var finalContent string
	if useSplit {
		previewContent := "No description"
		if strings.TrimSpace(m.desc.Value()) != "" {
			var err error
			previewContent, err = m.renderer.Render(m.desc.Value())
			if err != nil {
				previewContent = m.styles.Error.Render("Preview error: " + err.Error())
			}
		}

		previewBox := lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(m.styles.Theme.Border).
			Padding(0, 2).
			Height(m.height - 10).
			Width(m.width - (m.width / 2) - 10).
			Render(previewContent)

		finalContent = lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().Width(m.width/2).Render(editorContent),
			previewBox,
		)
	} else {
		finalContent = editorContent
	}

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		m.styles.Overlay.Width(cardW).Render(finalContent),
		lipgloss.WithWhitespaceBackground(m.styles.Theme.Bg),
	)
}

func (m Model) renderField(label, input string, focused bool) string {
	labelStyle := m.styles.Muted.Width(8)
	if focused {
		labelStyle = labelStyle.Foreground(m.styles.Theme.Accent)
	}

	return lipgloss.NewStyle().Padding(0, 1).Render(
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
