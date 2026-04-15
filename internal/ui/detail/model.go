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
	style := "dark"
	if m.styles.Theme.Name == "paper" {
		style = "light"
	}
	r, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle(style),
		glamour.WithWordWrap(m.width-4),
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
		Background(m.styles.Theme.Bg).
		BorderBottom(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(m.styles.Theme.Border).
		Width(m.width - 4).
		Render(m.task.Title)

	meta := m.renderMeta()

	body := m.renderMarkdown(m.task.Description)
	if strings.TrimSpace(body) == "" {
		body = m.styles.Muted.Render("No description.")
	}

	content := lipgloss.JoinVertical(lipgloss.Left, title, "", meta, "", body)

	box := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Background(m.styles.Theme.Bg).
		Padding(1, 2).
		Render(content)
	return box
}

func (m Model) renderMeta() string {
	parts := []string{
		m.styles.StatusBadge(m.task.Status),
		" ",
		m.styles.PriorityBadge(m.task.Priority),
	}

	metaStyle := m.styles.Muted.PaddingLeft(2)

	if m.task.Deadline != nil {
		parts = append(parts, metaStyle.Render("Due "+m.task.Deadline.Local().Format("Mon Jan 2 15:04")))
	}

	if len(m.task.Tags) > 0 {
		parts = append(parts, metaStyle.Render("#"+strings.Join(m.task.Tags, " #")))
	}

	parts = append(parts, metaStyle.Render("Updated "+humanTime(m.task.UpdatedAt, time.Now())))
	return lipgloss.JoinHorizontal(lipgloss.Center, parts...)
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
