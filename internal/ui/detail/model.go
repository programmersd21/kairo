package detail

import (
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"

	"github.com/programmersd21/kairo/internal/core"
	"github.com/programmersd21/kairo/internal/ui/styles"
)

type Model struct {
	styles styles.Styles
	width  int
	height int

	task   core.Task
	ShowID bool

	renderer *glamour.TermRenderer
	mdCache  string
	mdSrc    string
}

func New(s styles.Styles) Model {
	return Model{styles: s}
}

func (m *Model) SetSize(w, h int) {
	m.width, m.height = w, h
	m.resetRenderer()
}

func (m *Model) SetTask(t core.Task) {
	m.task = t
	if m.mdSrc != t.Description {
		m.mdSrc = t.Description
		m.mdCache = ""
	}
}

func (m Model) Task() core.Task {
	return m.task
}

func (m *Model) resetRenderer() {
	if m.width <= 0 {
		return
	}
	// Glamour styles
	style := "dark"
	if m.styles.Theme.IsLight {
		style = "light"
	}
	r, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle(style),
		glamour.WithWordWrap(m.width-8), // More padding
	)
	if err == nil {
		m.renderer = r
		m.mdCache = ""
	}
}

func (m *Model) View() string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}

	// Header
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.styles.Theme.Accent).
		Padding(1, 2).
		Render(styles.IconTask + m.task.Title)

	// Divider
	divider := lipgloss.NewStyle().
		Foreground(m.styles.Theme.Border).
		Padding(0, 2).
		Render(strings.Repeat("─", m.width-4))

	// Metadata
	meta := m.renderMeta()

	// Description
	body := m.renderMarkdown(m.task.Description)
	if strings.TrimSpace(body) == "" {
		body = lipgloss.NewStyle().
			Foreground(m.styles.Theme.Muted).
			Italic(true).
			Padding(1, 4).
			Render("No description provided.")
	} else {
		body = lipgloss.NewStyle().Padding(0, 2).Render(body)
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		title,
		divider,
		meta,
		lipgloss.NewStyle().Height(1).Render(""),
		lipgloss.NewStyle().Foreground(m.styles.Theme.Accent).Bold(true).Padding(0, 2).Render("Description"),
		body,
	)

	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Background(m.styles.Theme.Bg).
		Render(content)
}

func (m Model) renderMeta() string {
	type field struct {
		label string
		value string
	}
	fields := []field{}

	if m.ShowID {
		fields = append(fields, field{"ID", m.task.ID})
	}
	fields = append(fields, field{"Status", m.styles.StatusBadge(m.task.Status)})
	fields = append(fields, field{"Priority", m.styles.PriorityBadge(m.task.Priority)})
	if m.task.Deadline != nil {
		fields = append(fields, field{"Due", styles.IconDeadline + m.task.Deadline.Local().Format("Mon, Jan 02 15:04")})
	}
	if m.task.Project != "" {
		fields = append(fields, field{"Project", m.task.Project})
	}
	if len(m.task.Tags) > 0 {
		tagStr := ""
		for i, t := range m.task.Tags {
			if i > 0 {
				tagStr += " "
			}
			tagStr += "#" + t
		}
		fields = append(fields, field{"Tags", tagStr})
	}
	if m.task.OpenIssueID != "" {
		fields = append(fields, field{"Issue ID", m.task.OpenIssueID})
	}
	if m.task.Responsible != "" {
		fields = append(fields, field{"Resp", m.task.Responsible})
	}
	if m.task.Result != "" {
		fields = append(fields, field{"Result", m.task.Result})
	}

	maxLabel := 0
	for _, f := range fields {
		if n := lipgloss.Width(f.label); n > maxLabel {
			maxLabel = n
		}
	}
	labelStyle := lipgloss.NewStyle().Width(maxLabel + 1)

	var meta []string
	for i, f := range fields {
		line := lipgloss.JoinHorizontal(lipgloss.Left,
			labelStyle.Render(m.styles.Muted.Render(f.label+":")),
			m.styles.DetailValue.Render(" "+f.value),
		)
		pad := 0
		if i < 2 && m.ShowID {
			pad = 0
		} else if i < 1 && !m.ShowID {
			pad = 0
		}
		_ = pad
		meta = append(meta, lipgloss.NewStyle().Padding(0, 2).Render(line))
	}

	return lipgloss.JoinVertical(lipgloss.Left, meta...)
}

func (m *Model) renderMarkdown(src string) string {
	src = strings.TrimSpace(src)
	if src == "" {
		return ""
	}
	if m.mdCache != "" {
		return m.mdCache
	}
	if m.renderer == nil {
		m.resetRenderer()
	}
	if m.renderer == nil {
		m.mdCache = src
		return m.mdCache
	}
	out, err := m.renderer.Render(src)
	if err != nil {
		m.mdCache = src
		return m.mdCache
	}
	m.mdCache = strings.TrimRight(out, "\n")
	return m.mdCache
}
