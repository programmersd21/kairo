package help

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbletea"
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

	// Helper to extract keys from binding
	getK := func(b key.Binding) string {
		return strings.Join(b.Keys(), ", ")
	}

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
				{getK(m.km.ViewInbox) + ", " + getK(m.km.ViewToday) + ", " + getK(m.km.ViewUpcoming) + ", " + getK(m.km.ViewTag) + ", " + getK(m.km.ViewPriority), "Switch views"},
				{getK(m.km.OpenTask), "Open task details"},
				{getK(m.km.Back), "Back / Close"},
			},
		},
		{
			"Tasks",
			[]struct{ key, desc string }{
				{getK(m.km.NewTask), "New task"},
				{getK(m.km.EditTask), "Edit task"},
				{getK(m.km.DeleteTask), "Delete task"},
			},
		},
		{
			"App",
			[]struct{ key, desc string }{
				{getK(m.km.Palette), "Command palette"},
				{getK(m.km.TaskSearch), "Search tasks"},
				{getK(m.km.CycleTheme), "Theme menu"},
				{getK(m.km.OpenPluginDir), "Open plugins folder"},
				{getK(m.km.ManagePlugins), "Manage plugins"},
				{getK(m.km.Help), "Show help"},
				{getK(m.km.Quit), "Quit"},
			},
		},
	}

	var content []string
	content = append(content, lipgloss.NewStyle().Padding(0, 1).Render(header), "")

	for _, s := range sections {
		content = append(content, lipgloss.NewStyle().Bold(true).Foreground(m.styles.Theme.Accent).Padding(0, 1).Render(s.title))
		for _, k := range s.keys {
			keyStr := lipgloss.NewStyle().Foreground(m.styles.Theme.Good).Width(15).Render(k.key)
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
