package detail

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"

	"github.com/programmersd21/kairo/internal/core"
	"github.com/programmersd21/kairo/internal/ui/styles"
)

type Model struct {
	styles styles.Styles
	width  int
	height int

	task core.Task

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

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.styles.Theme.Accent).
		Padding(0, 1).
		Render(styles.IconTask + m.task.Title)

	meta := m.renderMeta()

	descriptionHeader := lipgloss.NewStyle().
		Foreground(m.styles.Theme.Muted).
		Bold(true).
		Padding(1, 0, 0, 1).
		Render("DESCRIPTION")

	body := m.renderMarkdown(m.task.Description)
	if strings.TrimSpace(body) == "" {
		body = "  " + m.styles.Muted.Render("No description provided.")
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		title,
		"\n",
		meta,
		"\n",
		descriptionHeader,
		body,
	)

	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Padding(1, 2).
		Render(content)
}

func (m Model) renderMeta() string {
	rows := []string{
		lipgloss.JoinHorizontal(lipgloss.Left, m.styles.DetailKey.Render("Status"), m.styles.StatusBadge(m.task.Status)),
		lipgloss.JoinHorizontal(lipgloss.Left, m.styles.DetailKey.Render("Priority"), m.styles.PriorityBadge(m.task.Priority)),
	}

	if m.task.Deadline != nil {
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Left,
			m.styles.DetailKey.Render("Deadline"),
			m.styles.DetailValue.Render(styles.IconDeadline+m.task.Deadline.Local().Format("Mon, Jan 02 15:04"))))
	}

	if len(m.task.Tags) > 0 {
		tagStr := ""
		for _, t := range m.task.Tags {
			tagStr += styles.IconTag + t + " "
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Left,
			m.styles.DetailKey.Render("Tags"),
			m.styles.DetailValue.Render(tagStr)))
	}

	rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Left,
		m.styles.DetailKey.Render("Updated"),
		m.styles.DetailValue.Render(humanTime(m.task.UpdatedAt, time.Now()))))

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
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

func humanTime(t time.Time, now time.Time) string {
	d := now.Sub(t)
	if d < 0 {
		d = -d
	}
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}
