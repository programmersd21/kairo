package editor

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/cursor"
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
type SelectParentMsg struct{}
type SelectProjectMsg struct{}

type Model struct {
	styles styles.Styles
	mode   Mode

	width  int
	height int

	orig core.Task

	title       textinput.Model
	tags        textinput.Model
	priority    textinput.Model
	deadline    textinput.Model
	waitUntil   textinput.Model
	status      textinput.Model
	recur       textinput.Model
	until       textinput.Model
	project     textinput.Model
	parentID    textinput.Model
	responsible textinput.Model
	result      textinput.Model
	desc        textarea.Model

	focus int

	deadlinePreview string
	deadlineValue   *time.Time
	deadlineErr     string

	waitUntilPreview string
	waitUntilValue   *time.Time
	waitUntilErr     string

	recurPreview string
	recurErr     string

	untilPreview string
	untilValue   *time.Time
	untilErr     string

	showPreview bool
	renderer    *glamour.TermRenderer
}

func New(s styles.Styles, mode Mode, t core.Task, preview bool) Model {
	applyFieldStyles := func(in *textinput.Model) {
		// Keep labels fully highlighted even when the field is not focused.
		in.PromptStyle = s.Accent.Bold(true)
		in.TextStyle = s.Text

		// Show a blinking cursor for better cursor position visibility, combined with
		// background highlighting for improved focus indication.
		in.Cursor.SetMode(cursor.CursorBlink)
	}
	ti := textinput.New()
	ti.Prompt = ""
	ti.CharLimit = 200
	ti.SetValue(strings.TrimSpace(t.Title))
	ti.Focus()
	applyFieldStyles(&ti)

	tags := textinput.New()
	tags.Prompt = ""
	tags.CharLimit = 200
	if len(t.Tags) > 0 {
		tags.SetValue("#" + strings.Join(t.Tags, " #"))
	}
	applyFieldStyles(&tags)

	pr := textinput.New()
	pr.Prompt = ""
	pr.CharLimit = 2
	pr.SetValue(fmt.Sprintf("%d", int(t.Priority.Clamp())))
	applyFieldStyles(&pr)

	dl := textinput.New()
	dl.Prompt = ""
	dl.CharLimit = 64
	if t.Deadline != nil {
		dl.SetValue(t.Deadline.Local().Format("2006-01-02 15:04"))
	}
	applyFieldStyles(&dl)

	wu := textinput.New()
	wu.Prompt = ""
	wu.CharLimit = 64
	if t.WaitUntil != nil {
		wu.SetValue(t.WaitUntil.Local().Format("2006-01-02 15:04"))
	}
	applyFieldStyles(&wu)

	st := textinput.New()
	st.Prompt = ""
	st.CharLimit = 16
	if t.Status == "" {
		st.SetValue(string(core.StatusTodo))
	} else {
		st.SetValue(string(t.Status))
	}
	applyFieldStyles(&st)

	re := textinput.New()
	re.Prompt = ""
	re.CharLimit = 64
	switch t.Recurrence {
	case core.RecurrenceWeekly:
		re.SetValue(strings.Join(t.RecurrenceWeekly, ","))
	case core.RecurrenceMonthly:
		re.SetValue(fmt.Sprintf("%d", t.RecurrenceMonthly))
	}
	applyFieldStyles(&re)

	un := textinput.New()
	un.Prompt = ""
	un.CharLimit = 64
	if t.Until != nil {
		un.SetValue(t.Until.Local().Format("2006-01-02 15:04"))
	}
	applyFieldStyles(&un)

	pj := textinput.New()
	pj.Prompt = ""
	pj.CharLimit = 64
	pj.SetValue(t.Project)
	applyFieldStyles(&pj)

	pid := textinput.New()
	pid.Prompt = ""
	pid.CharLimit = 64
	pid.SetValue(t.ParentID)
	applyFieldStyles(&pid)

	resp := textinput.New()
	resp.Prompt = ""
	resp.CharLimit = 64
	resp.SetValue(t.Responsible)
	applyFieldStyles(&resp)

	res := textinput.New()
	res.Prompt = ""
	res.CharLimit = 200
	res.SetValue(t.Result)
	applyFieldStyles(&res)

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
		waitUntil:   wu,
		status:      st,
		recur:       re,
		until:       un,
		project:     pj,
		parentID:    pid,
		responsible: resp,
		result:      res,
		desc:        d,
		focus:       0,
		showPreview: preview,
	}
	m.recomputeDeadline()
	m.recomputeWaitUntil()
	m.recomputeRecurrence()
	m.recomputeUntil()
	return m
}

func (m *Model) SetSize(w, h int) {
	m.width, m.height = w, h

	// Base sizes
	editorW := max(20, min(80, w-10))
	useSplit := m.width >= 100 && m.showPreview
	if useSplit {
		editorW = m.width / 2
		if editorW < 50 {
			editorW = 50
		}
	}

	m.desc.SetWidth(max(20, editorW-10))
	m.desc.SetHeight(max(4, h-18)) // Reduced height to fit more fields
	m.title.Width = max(20, editorW-20)
	m.tags.Width = max(20, editorW-20)
	m.priority.Width = 6
	m.deadline.Width = max(20, editorW-20)
	m.waitUntil.Width = max(20, editorW-20)
	m.status.Width = 10
	m.recur.Width = max(20, editorW-20)
	m.until.Width = max(20, editorW-20)
	m.project.Width = max(20, editorW-20)
	m.parentID.Width = max(20, editorW-20)
	m.responsible.Width = max(20, editorW-20)
	m.result.Width = max(20, editorW-20)
	style := "dark"
	if m.styles.Theme.IsLight {
		style = "light"
	}

	previewW := w - editorW - 6
	if !useSplit {
		previewW = w - 10
	}

	r, _ := glamour.NewTermRenderer(
		glamour.WithStandardStyle(style),
		glamour.WithWordWrap(max(20, previewW-4)),
	)
	m.renderer = r
}

func (m *Model) SetParentID(id string) {
	m.parentID.SetValue(id)
}

func (m *Model) SetProject(p string) {
	m.project.SetValue(p)
}

func (m Model) Init() tea.Cmd { return nil }

func (m *Model) SetPreview(preview bool) {
	m.showPreview = preview
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch x := msg.(type) {
	case tea.KeyMsg:
		switch x.String() {
		case "esc":
			return m, func() tea.Msg { return CloseMsg{} }
		case "tab":
			m.blurAll()
			m.focus = (m.focus + 1) % 13
			m.focusField()
			return m, nil
		case "shift+tab":
			m.blurAll()
			m.focus--
			if m.focus < 0 {
				m.focus = 12
			}
			m.focusField()
			return m, nil
		case "ctrl+s":
			return m, m.saveCmd()
		case "ctrl+p":
			m.showPreview = !m.showPreview
			m.SetSize(m.width, m.height)
			return m, nil
		case "enter":
			if m.focus == 8 {
				return m, func() tea.Msg { return SelectProjectMsg{} }
			}
			if m.focus == 9 {
				return m, func() tea.Msg { return SelectParentMsg{} }
			}
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
		m.status, cmd = m.status.Update(msg)
	case 4:
		prev := m.deadline.Value()
		m.deadline, cmd = m.deadline.Update(msg)
		if m.deadline.Value() != prev {
			m.recomputeDeadline()
		}
	case 5:
		prev := m.waitUntil.Value()
		m.waitUntil, cmd = m.waitUntil.Update(msg)
		if m.waitUntil.Value() != prev {
			m.recomputeWaitUntil()
		}
	case 6:
		prev := m.recur.Value()
		m.recur, cmd = m.recur.Update(msg)
		if m.recur.Value() != prev {
			m.recomputeRecurrence()
		}
	case 7:
		prev := m.until.Value()
		m.until, cmd = m.until.Update(msg)
		if m.until.Value() != prev {
			m.recomputeUntil()
		}
	case 8:
		m.project, cmd = m.project.Update(msg)
	case 9:
		m.parentID, cmd = m.parentID.Update(msg)
	case 10:
		m.responsible, cmd = m.responsible.Update(msg)
	case 11:
		m.result, cmd = m.result.Update(msg)
	case 12:
		m.desc, cmd = m.desc.Update(msg)
	}
	return m, cmd
}

func (m Model) View() string {
	w := m.width
	if w <= 0 {
		w = 80
	}

	useSplit := m.width >= 100 && m.showPreview
	cardW := min(84, w-6)
	if useSplit {
		cardW = m.width - 4
	}

	titleText := "NEW TASK"
	if m.mode == ModeEdit {
		titleText = "EDIT TASK"
	}
	header := m.styles.Title.Padding(0, 1).MarginBottom(1).Render(titleText)

	// Helper for rendering structured fields
	renderField := func(icon, label string, input string, focused bool) string {
		s := lipgloss.NewStyle()

		// Style prompt icon and label based on focus
		promptStyle := m.styles.Muted
		if focused {
			promptStyle = promptStyle.Foreground(m.styles.Theme.Accent).Bold(true)
			// Use Overlay color for better focus visibility, with bold accent border for extra emphasis
			input = lipgloss.NewStyle().
				Background(m.styles.Theme.Overlay).
				Foreground(m.styles.Theme.Accent).
				Bold(true).
				Padding(0, 1).
				Render(input)
		}

		prompt := lipgloss.JoinHorizontal(lipgloss.Left, icon, label)
		return s.Render(lipgloss.JoinHorizontal(lipgloss.Left, promptStyle.Width(14).Render(prompt), input))
	}

	fields := []string{
		header,
		renderField(styles.IconTask, "Title:", m.title.View(), m.focus == 0),
		renderField(styles.IconTag, "Tags:", m.tags.View(), m.focus == 1),
		lipgloss.JoinHorizontal(lipgloss.Left,
			renderField(styles.IconPriority1, "Pri:", m.priority.View(), m.focus == 2),
			renderField(styles.IconDoing, "Status:", m.status.View(), m.focus == 3),
		),
		renderField(styles.IconDeadline, "Due:", m.deadline.View(), m.focus == 4),
	}

	if m.deadlineErr != "" {
		fields = append(fields, m.styles.Error.Padding(0, 2).Render(m.deadlineErr))
	} else if m.deadlinePreview != "" {
		fields = append(fields, m.styles.Muted.Padding(0, 2).Render(m.deadlinePreview))
	}

	fields = append(fields, renderField(styles.IconWaitUntil, "Wait Until:", m.waitUntil.View(), m.focus == 5))
	if m.waitUntilErr != "" {
		fields = append(fields, m.styles.Error.Padding(0, 2).Render(m.waitUntilErr))
	} else if m.waitUntilPreview != "" {
		fields = append(fields, m.styles.Muted.Padding(0, 2).Render(m.waitUntilPreview))
	}

	fields = append(fields, renderField("󰑖 ", "Recur:", m.recur.View(), m.focus == 6))
	if m.recurErr != "" {
		fields = append(fields, m.styles.Error.Padding(0, 2).Render(m.recurErr))
	} else if m.recurPreview != "" {
		fields = append(fields, m.styles.Muted.Padding(0, 2).Render(m.recurPreview))
	} else {
		fields = append(fields, m.styles.Muted.Padding(0, 2).Render("e.g. mon,wed or 15"))
	}

	fields = append(fields, renderField(styles.IconUntil, "Until:", m.until.View(), m.focus == 7))
	if m.untilErr != "" {
		fields = append(fields, m.styles.Error.Padding(0, 2).Render(m.untilErr))
	} else if m.untilPreview != "" {
		fields = append(fields, m.styles.Muted.Padding(0, 2).Render(m.untilPreview))
	}

	fields = append(fields, renderField("󱓡 ", "Project:", m.project.View(), m.focus == 8))
	fields = append(fields, renderField("󰓆 ", "Resp:", m.responsible.View(), m.focus == 10))
	fields = append(fields, renderField("󰄬 ", "Result:", m.result.View(), m.focus == 11))

	parentIDDisplay := m.parentID.Value()
	if parentIDDisplay == "" {
		parentIDDisplay = "None"
	}
	parentField := renderField("󱗼 ", "Parent:", parentIDDisplay, m.focus == 9)
	// Add an action hint for selecting
	if m.focus == 9 {
		parentField += m.styles.Muted.PaddingLeft(2).Render("press enter to select")
	}
	fields = append(fields, parentField)

	descView := m.desc.View()
	fields = append(fields, lipgloss.NewStyle().Padding(0, 2).Render(descView))

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

func (m *Model) blurAll() {
	m.title.Blur()
	m.tags.Blur()
	m.priority.Blur()
	m.deadline.Blur()
	m.waitUntil.Blur()
	m.status.Blur()
	m.recur.Blur()
	m.until.Blur()
	m.project.Blur()
	m.parentID.Blur()
	m.responsible.Blur()
	m.result.Blur()
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
		m.status.Focus()
	case 4:
		m.deadline.Focus()
	case 5:
		m.waitUntil.Focus()
	case 6:
		m.recur.Focus()
	case 7:
		m.until.Focus()
	case 8:
		m.project.Focus()
	case 9:
		m.parentID.Focus()
	case 10:
		m.responsible.Focus()
	case 11:
		m.result.Focus()
	case 12:
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
	// If deadline changes, recurrence preview might change too
	m.recomputeRecurrence()
}

func (m *Model) recomputeWaitUntil() {
	m.waitUntilErr = ""
	m.waitUntilPreview = ""
	m.waitUntilValue = nil
	raw := strings.TrimSpace(m.waitUntil.Value())
	if raw == "" {
		return
	}
	t, err := nlp.ParseDeadline(raw, time.Now())
	if err != nil {
		m.waitUntilErr = err.Error()
		return
	}
	if t == nil {
		return
	}
	m.waitUntilValue = t
	m.waitUntilPreview = t.Local().Format("Mon Jan 2 15:04")
	// If wait-until changes, recurrence preview might change too
	m.recomputeRecurrence()
}

func (m *Model) recomputeRecurrence() {
	m.recurErr = ""
	m.recurPreview = ""
	raw := strings.TrimSpace(m.recur.Value())
	if raw == "" {
		return
	}

	recType, weekly, monthly, err := core.ParseRecurrence(raw)
	if err != nil {
		m.recurErr = err.Error()
		return
	}

	if recType == core.RecurrenceNone {
		return
	}

	// Use deadline as reference if available, else now
	ref := time.Now()
	if m.deadlineValue != nil {
		ref = *m.deadlineValue
	}

	task := core.Task{
		Recurrence:        recType,
		RecurrenceWeekly:  weekly,
		RecurrenceMonthly: monthly,
		Deadline:          m.deadlineValue,
		WaitUntil:         m.waitUntilValue,
		Until:             m.untilValue,
	}

	next := task.NextOccurrence(ref)
	if next != nil {
		m.recurPreview = "Next: " + next.Local().Format("Mon Jan 2 15:04")
	}
}

func (m *Model) recomputeUntil() {
	m.untilErr = ""
	m.untilPreview = ""
	m.untilValue = nil
	raw := strings.TrimSpace(m.until.Value())
	if raw == "" {
		return
	}
	t, err := nlp.ParseDeadline(raw, time.Now())
	if err != nil {
		m.untilErr = err.Error()
		return
	}
	if t == nil {
		return
	}
	m.untilValue = t
	m.untilPreview = t.Local().Format("Mon Jan 2 15:04")
	// If until changes, recurrence preview might change too
	m.recomputeRecurrence()
}

func (m Model) saveCmd() tea.Cmd {
	title := strings.TrimSpace(m.title.Value())
	desc := strings.TrimSpace(m.desc.Value())
	prj := strings.TrimSpace(m.project.Value())
	res := strings.TrimSpace(m.result.Value())
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

	var waitUntil *time.Time
	if m.waitUntilValue != nil {
		d := (*m.waitUntilValue).UTC()
		waitUntil = &d
	}

	var until *time.Time
	if m.untilValue != nil {
		d := (*m.untilValue).UTC()
		until = &d
	}

	recType, weekly, monthly, err := core.ParseRecurrence(m.recur.Value())
	if err != nil {
		m.recurErr = err.Error()
		return nil
	}

	if m.mode == ModeNew {
		task := core.Task{
			Title:             title,
			Description:       desc,
			Project:           prj,
			Tags:              tags,
			Priority:          pri,
			Deadline:          deadline,
			WaitUntil:         waitUntil,
			Until:             until,
			Status:            st,
			Recurrence:        recType,
			RecurrenceWeekly:  weekly,
			RecurrenceMonthly: monthly,
			ParentID:          strings.TrimSpace(m.parentID.Value()),
			Responsible:       strings.TrimSpace(m.responsible.Value()),
			Result:            res,
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
	if (waitUntil == nil) != (m.orig.WaitUntil == nil) || (waitUntil != nil && m.orig.WaitUntil != nil && !waitUntil.Equal(*m.orig.WaitUntil)) {
		wu := waitUntil
		patch.WaitUntil = &wu
	}
	if (until == nil) != (m.orig.Until == nil) || (until != nil && m.orig.Until != nil && !until.Equal(*m.orig.Until)) {
		u := until
		patch.Until = &u
	}
	if st != m.orig.Status {
		patch.Status = &st
	}
	if prj != m.orig.Project {
		patch.Project = &prj
	}
	if recType != m.orig.Recurrence {
		patch.Recurrence = &recType
	}
	if strings.Join(weekly, ",") != strings.Join(m.orig.RecurrenceWeekly, ",") {
		patch.RecurrenceWeekly = &weekly
	}
	if monthly != m.orig.RecurrenceMonthly {
		patch.RecurrenceMonthly = &monthly
	}
	pid := strings.TrimSpace(m.parentID.Value())
	if pid != m.orig.ParentID {
		patch.ParentID = &pid
	}
	resp := strings.TrimSpace(m.responsible.Value())
	if resp != m.orig.Responsible {
		patch.Responsible = &resp
	}
	if res != m.orig.Result {
		patch.Result = &res
	}
	return func() tea.Msg { return SavePatchMsg{ID: m.orig.ID, Patch: patch} }
}
