package status_menu

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/programmersd21/kairo/internal/core"
	"github.com/programmersd21/kairo/internal/ui/styles"
)

type SelectMsg struct {
	Status core.Status
}

type CloseMsg struct{}

type Model struct {
	styles styles.Styles
	width  int
	height int
	sel    int
}

func New(s styles.Styles) Model {
	return Model{
		styles: s,
	}
}

func (m *Model) SetSize(w, h int) {
	m.width, m.height = w, h
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch x := msg.(type) {
	case tea.KeyMsg:
		switch x.String() {
		case "esc", "q":
			return m, func() tea.Msg { return CloseMsg{} }
		case "up":
			if m.sel > 0 {
				m.sel--
			}
		case "down":
			if m.sel < len(statuses)-1 {
				m.sel++
			}
		case "enter":
			return m, func() tea.Msg { return SelectMsg{Status: statuses[m.sel]} }
		}
	}
	return m, nil
}

var statuses = []core.Status{
	core.StatusTodo,
	core.StatusDoing,
	core.StatusDone,
}

var statusLabels = map[core.Status]string{
	core.StatusTodo:  "Todo",
	core.StatusDoing: "Doing",
	core.StatusDone:  "Done",
}

var statusIcons = map[core.Status]string{
	core.StatusTodo:  "○",
	core.StatusDoing: "◉",
	core.StatusDone:  "✓",
}

func (m Model) View() string {
	w := m.width
	if w <= 0 {
		w = 80
	}
	cardW := min(36, w-4)

	header := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Width(cardW).
		Render(m.styles.Title.Render(" Set Status "))

	var lines []string
	lines = append(lines, header, "")
	for i, st := range statuses {
		style := m.styles.RowNormal
		prefix := "  "
		if i == m.sel {
			style = m.styles.RowSelected
			prefix = "> "
		}

		icon := statusIcons[st]
		label := statusLabels[st]

		itemStyle := lipgloss.NewStyle().Width(cardW - 2)
		line := prefix + icon + " " + label
		lines = append(lines, style.Render(itemStyle.Render(line)))
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
