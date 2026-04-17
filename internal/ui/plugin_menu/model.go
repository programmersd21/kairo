package plugin_menu

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/programmersd21/kairo/internal/plugins"
	"github.com/programmersd21/kairo/internal/ui/styles"
)

type CloseMsg struct{}
type UninstallMsg struct{ ID string }
type OpenFolderMsg struct{}
type ReloadMsg struct{}

type UninstallConfirmMsg struct{ ID string }

type Model struct {
	styles  styles.Styles
	width   int
	height  int
	plugins []plugins.PluginInfo
	sel     int
}

func New(s styles.Styles) Model {
	return Model{styles: s}
}

func (m *Model) SetSize(w, h int) {
	m.width, m.height = w, h
}

func (m *Model) SetPlugins(ps []plugins.PluginInfo) {
	m.plugins = ps
	if m.sel >= len(m.plugins) {
		m.sel = len(m.plugins) - 1
	}
	if m.sel < 0 {
		m.sel = 0
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch x := msg.(type) {
	case tea.KeyMsg:
		switch x.String() {
		case "esc", "q", "p":
			return m, func() tea.Msg { return CloseMsg{} }
		case "up", "k":
			if m.sel > 0 {
				m.sel--
			}
		case "down", "j":
			if m.sel < len(m.plugins)-1 {
				m.sel++
			}
		case "o": // Open folder
			return m, func() tea.Msg { return OpenFolderMsg{} }
		case "r": // Reload
			return m, func() tea.Msg { return ReloadMsg{} }
		case "x": // Uninstall key
			if m.sel >= 0 && m.sel < len(m.plugins) {
				id := m.plugins[m.sel].ID
				return m, func() tea.Msg { return UninstallConfirmMsg{ID: id} }
			}
		}
	}
	return m, nil
}

func (m Model) View() string {
	w := m.width
	if w <= 0 {
		w = 80
	}
	cardW := min(80, w-4)

	var rows []string
	if len(m.plugins) == 0 {
		rows = append(rows, m.styles.Muted.Padding(0, 1).Render("No plugins."))
	} else {
		for i, p := range m.plugins {
			rows = append(rows, m.renderPluginRow(cardW-2, i, p))
		}
	}

	return lipgloss.Place(w, m.height, lipgloss.Center, lipgloss.Center,
		m.styles.Overlay.Width(cardW).Render(lipgloss.JoinVertical(lipgloss.Left, rows...)),
		lipgloss.WithWhitespaceBackground(m.styles.Theme.Bg),
	)
}

func (m Model) renderPluginRow(w, idx int, p plugins.PluginInfo) string {
	prefix := "  "
	style := m.styles.Muted.Padding(0, 1)
	if idx == m.sel {
		prefix = "> "
		style = style.Foreground(m.styles.Theme.Accent).Bold(true)
	}

	return style.Render(prefix + p.Name + " " + p.Version)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
