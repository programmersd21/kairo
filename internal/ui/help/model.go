package help

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/programmersd21/kairo/internal/ui/keymap"
	"github.com/programmersd21/kairo/internal/ui/styles"
)

type CloseMsg struct{}

type Model struct {
	styles styles.Styles
	km     keymap.Keymap
	width  int
	height int
}

func New(s styles.Styles, km keymap.Keymap) Model {
	return Model{styles: s, km: km}
}

func (m *Model) SetSize(w, h int) {
	m.width, m.height = w, h
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch x := msg.(type) {
	case tea.KeyMsg:
		switch x.String() {
		case "esc", "q", "?":
			return m, func() tea.Msg { return CloseMsg{} }
		}
	}
	return m, nil
}

func (m Model) View() string {
	w := m.width
	if w <= 0 {
		w = 80
	}
	cardW := min(72, w-4)
	if cardW < 44 {
		cardW = w - 2
	}

	header := m.styles.Title.Render(" Help & Keybindings ")

	sections := []struct {
		title string
		keys  []struct {
			key  string
			desc string
		}
	}{
		{
			"Navigation",
			[]struct{ key, desc string }{
				{"1-5", "Switch views (Inbox, Today, etc.)"},
				{"j/k, ↑/↓", "Move selection"},
				{"enter", "Open task details"},
				{"esc", "Back / Close"},
			},
		},
		{
			"Tasks",
			[]struct{ key, desc string }{
				{"n", "New task"},
				{"e", "Edit task"},
				{"d", "Delete task"},
			},
		},
		{
			"App",
			[]struct{ key, desc string }{
				{"ctrl+p", "Command palette"},
				{"t", "Theme menu"},
				{"?", "Show help"},
				{"q", "Quit"},
			},
		},
	}

	var content []string
	content = append(content, lipgloss.NewStyle().Padding(0, 1).Render(header), "")

	for _, s := range sections {
		content = append(content, lipgloss.NewStyle().Bold(true).Foreground(m.styles.Theme.Accent).Padding(0, 1).Render(s.title))
		for _, k := range s.keys {
			keyStr := lipgloss.NewStyle().Foreground(m.styles.Theme.Good).Width(10).Render(k.key)
			descStr := m.styles.Muted.Render(k.desc)
			content = append(content, lipgloss.NewStyle().Padding(0, 2).Render(keyStr+" "+descStr))
		}
		content = append(content, "")
	}

	card := lipgloss.NewStyle().
		Width(cardW).
		Background(m.styles.Theme.Bg).
		Border(lipgloss.ThickBorder()).
		BorderForeground(m.styles.Theme.Accent).
		Padding(1, 1).
		Render(strings.Join(content, "\n"))

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
