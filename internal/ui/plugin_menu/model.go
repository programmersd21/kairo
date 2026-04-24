package plugin_menu

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/programmersd21/kairo/internal/plugins"
	"github.com/programmersd21/kairo/internal/ui/styles"
)

type CloseMsg struct{}
type TransitionMsg struct{}
type UninstallMsg struct{ ID string }
type OpenFolderMsg struct{}
type ReloadMsg struct{}

type Model struct {
	styles  styles.Styles
	width   int
	height  int
	plugins []plugins.PluginInfo
	sel     int
	detail  *plugins.PluginInfo // nil means in list mode
	confirm string              // non-empty means showing confirm for uninstalling this ID
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
		if m.confirm != "" {
			switch x.String() {
			case "y":
				id := m.confirm
				m.confirm = ""
				return m, func() tea.Msg { return UninstallMsg{ID: id} }
			case "n", "esc":
				m.confirm = ""
				return m, nil
			}
			return m, nil
		}

		if m.detail != nil {
			switch x.String() {
			case "esc", "q", "enter":
				m.detail = nil
				return m, func() tea.Msg { return TransitionMsg{} }
			}
			return m, nil
		}

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
		case "enter":
			if m.sel >= 0 && m.sel < len(m.plugins) {
				p := m.plugins[m.sel]
				m.detail = &p
				return m, func() tea.Msg { return TransitionMsg{} }
			}
		case "o":
			return m, func() tea.Msg { return OpenFolderMsg{} }
		case "r":
			return m, func() tea.Msg { return ReloadMsg{} }
		case "u":
			if m.sel >= 0 && m.sel < len(m.plugins) {
				m.confirm = m.plugins[m.sel].ID
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

	if m.confirm != "" {
		return lipgloss.Place(w, m.height, lipgloss.Center, lipgloss.Center,
			m.styles.Overlay.Width(cardW).Render(
				lipgloss.JoinVertical(lipgloss.Center,
					"Uninstall plugin?",
					"",
					m.styles.Accent.Render("[y] Yes  [n] No"),
				),
			),
		)
	}

	if m.detail != nil {
		p := m.detail
		content := lipgloss.JoinVertical(lipgloss.Left,
			m.styles.Title.Render(p.Name),
			"",
			"Version: "+p.Version,
			"Author: "+p.Author,
			"",
			p.Description,
			"",
			m.styles.Muted.Render("Press Enter/Esc to return"),
		)
		return lipgloss.Place(w, m.height, lipgloss.Center, lipgloss.Center,
			m.styles.Overlay.Width(cardW).Padding(2).Render(content),
		)
	}

	var rows []string
	if len(m.plugins) == 0 {
		rows = append(rows, m.styles.Muted.Padding(0, 1).Render("No plugins."))
	} else {
		for i, p := range m.plugins {
			rows = append(rows, m.renderPluginRow(cardW-2, i, p))
		}
	}

	return lipgloss.Place(w, m.height, lipgloss.Center, lipgloss.Center,
		m.styles.Overlay.Width(cardW).Render(
			lipgloss.JoinVertical(lipgloss.Left,
				lipgloss.JoinVertical(lipgloss.Left, rows...),
				"",
				m.styles.Muted.Padding(0, 1).Render("enter detail • u uninstall • o open folder • r reload • esc/p close • ↑/↓ navigate"),
			),
		),
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
