package theme_menu

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/programmersd21/kairo/internal/ui/styles"
	"github.com/programmersd21/kairo/internal/ui/theme"
)

type SelectMsg struct {
	Theme theme.Theme
}

type CloseMsg struct{}

type Model struct {
	styles styles.Styles
	width  int
	height int
	themes []theme.Theme
	sel    int
}

func New(s styles.Styles) Model {
	th := theme.Builtins()
	sel := 0
	for i, t := range th {
		if t.Name == s.Theme.Name {
			sel = i
			break
		}
	}
	return Model{
		styles: s,
		themes: th,
		sel:    sel,
	}
}

func (m *Model) SetSize(w, h int) {
	m.width, m.height = w, h
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch x := msg.(type) {
	case tea.KeyMsg:
		switch x.String() {
		case "esc", "q", "t":
			return m, func() tea.Msg { return CloseMsg{} }
		case "up", "k":
			if m.sel > 0 {
				m.sel--
			}
		case "down", "j":
			if m.sel < len(m.themes)-1 {
				m.sel++
			}
		case "enter":
			return m, func() tea.Msg { return SelectMsg{Theme: m.themes[m.sel]} }
		}
	}
	return m, nil
}

func (m Model) View() string {
	w := m.width
	if w <= 0 {
		w = 80
	}
	cardW := min(48, w-4)
	if cardW < 32 {
		cardW = w - 2
	}

	header := m.styles.Title.Render(" Select Theme ")

	var lines []string
	lines = append(lines, lipgloss.NewStyle().Padding(0, 1).Render(header), "")

	for i, t := range m.themes {
		indicator := "  "
		style := lipgloss.NewStyle().Padding(0, 2).Width(cardW - 4).Background(m.styles.Theme.Bg)
		if i == m.sel {
			indicator = lipgloss.NewStyle().Foreground(m.styles.Theme.Accent).Render("→ ")
			style = style.Background(m.styles.Theme.Overlay).Foreground(m.styles.Theme.Accent).Bold(true)
		}

		name := t.Name
		if t.Name == m.styles.Theme.Name {
			name += " (current)"
		}

		lines = append(lines, style.Render(indicator+name))
	}

	lines = append(lines, "", m.styles.Muted.Padding(0, 2).Render("enter to select • esc to close"))

	card := lipgloss.NewStyle().
		Width(cardW).
		Background(m.styles.Theme.Bg).
		Border(lipgloss.ThickBorder()).
		BorderForeground(m.styles.Theme.Accent).
		Padding(1, 0).
		Render(strings.Join(lines, "\n"))

	return lipgloss.Place(w, m.height, lipgloss.Center, lipgloss.Center, card,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceBackground(m.styles.Theme.Bg),
	)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
