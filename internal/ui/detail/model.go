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
		Render(content)
}

func (m Model) renderMeta() string {
	var meta []string

	// Status & Priority in one line
	status := lipgloss.JoinHorizontal(lipgloss.Left,
		m.styles.Muted.Render("Status:   "),
		m.styles.StatusBadge(m.task.Status),
	)
	priority := lipgloss.JoinHorizontal(lipgloss.Left,
		m.styles.Muted.Render("Priority: "),
		m.styles.PriorityBadge(m.task.Priority),
	)
	meta = append(meta, lipgloss.JoinHorizontal(lipgloss.Left,
		lipgloss.NewStyle().Padding(1, 2).Render(status),
		lipgloss.NewStyle().Padding(1, 4).Render(priority),
	))

	// Deadline & Tags
	if m.task.Deadline != nil {
		meta = append(meta, lipgloss.NewStyle().Padding(0, 2).Render(
			lipgloss.JoinHorizontal(lipgloss.Left,
				m.styles.Muted.Render("Due:      "),
				m.styles.DetailValue.Render(styles.IconDeadline+m.task.Deadline.Local().Format("Mon, Jan 02 15:04")),
			)))
	}

	if len(m.task.Tags) > 0 {
		tagStr := ""
		for i, t := range m.task.Tags {
			if i > 0 {
				tagStr += " "
			}
			tagStr += "#" + t
		}
		meta = append(meta, lipgloss.NewStyle().Padding(0, 2).Render(
			lipgloss.JoinHorizontal(lipgloss.Left,
				m.styles.Muted.Render("Tags:     "),
				m.styles.DetailValue.Render(tagStr),
			)))
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
