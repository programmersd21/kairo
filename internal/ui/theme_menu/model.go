package theme_menu

import (
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
	cardW := min(40, w-4)

	var lines []string
	for i, t := range m.themes {
		style := m.styles.RowNormal
		prefix := "  "
		if i == m.sel {
			style = m.styles.RowSelected
			prefix = "> "
		}
		lines = append(lines, style.Render(prefix+t.Name))
	}

	return lipgloss.Place(w, m.height, lipgloss.Center, lipgloss.Center,
		m.styles.Overlay.Width(cardW).Render(lipgloss.JoinVertical(lipgloss.Left, lines...)),
		lipgloss.WithWhitespaceBackground(m.styles.Theme.Bg),
	)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
