package changelog

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/programmersd21/kairo/internal/ui/styles"
)

type CloseMsg struct{}

type Model struct {
	styles   styles.Styles
	viewport viewport.Model
	width    int
	height   int
	ready    bool
}

func New(s styles.Styles) Model {
	return Model{styles: s}
}

func (m *Model) SetSize(w, h int) {
	m.width, m.height = w, h
	cardW := min(72, w-4)
	if cardW < 44 {
		cardW = w - 2
	}
	cardH := h - 4
	if !m.ready {
		m.viewport = viewport.New(cardW, cardH)
		m.ready = true
	} else {
		m.viewport.Width = cardW
		m.viewport.Height = cardH
	}
}

func (m *Model) SetContent(content string) {
	m.viewport.SetContent(content)
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch x := msg.(type) {
	case tea.KeyMsg:
		switch x.String() {
		case "esc", "q", "c":
			return m, func() tea.Msg { return CloseMsg{} }
		}
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	if !m.ready {
		return ""
	}

	header := m.styles.Title.Render(" Changelog ")
	footer := m.styles.Muted.Render(" q/esc/c close • ↑/↓ scroll ")

	content := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Padding(0, 1).Render(header),
		"",
		m.viewport.View(),
		"",
		lipgloss.NewStyle().Padding(0, 1).Render(footer),
	)

	card := lipgloss.NewStyle().
		Width(m.viewport.Width+2).
		Background(m.styles.Theme.Bg).
		Border(lipgloss.ThickBorder()).
		BorderForeground(m.styles.Theme.Accent).
		Padding(1, 1).
		Render(content)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, card,
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
