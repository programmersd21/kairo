package palette

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/programmersd21/kairo/internal/search"
	"github.com/programmersd21/kairo/internal/ui/styles"
)

type SelectMsg struct {
	Item search.Item
}

type CloseMsg struct{}

type Model struct {
	styles styles.Styles

	width  int
	height int

	input textinput.Model

	index   *search.Index
	results []search.Result
	sel     int
}

func New(s styles.Styles) Model {
	in := textinput.New()
	in.Prompt = "› "
	in.Placeholder = "Search tasks, commands, tags…"
	in.CharLimit = 128
	in.Width = 48
	in.Focus()

	return Model{
		styles: s,
		input:  in,
		index:  search.NewIndex(nil),
	}
}

func (m *Model) SetSize(w, h int) {
	m.width, m.height = w, h
}

func (m *Model) SetIndex(idx *search.Index) {
	m.index = idx
	m.refresh()
}

func (m *Model) SetPlaceholder(p string) {
	m.input.Placeholder = p
}

func (m *Model) Open() tea.Cmd {
	m.input.SetValue("")
	m.input.CursorEnd()
	m.sel = 0
	m.refresh()
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch x := msg.(type) {
	case tea.KeyMsg:
		switch x.String() {
		case "esc":
			return m, func() tea.Msg { return CloseMsg{} }
		case "enter":
			if m.sel >= 0 && m.sel < len(m.results) {
				it := m.results[m.sel].Item
				return m, func() tea.Msg { return SelectMsg{Item: it} }
			}
			return m, nil
		case "up":
			if m.sel > 0 {
				m.sel--
			}
			return m, nil
		case "down":
			if m.sel < len(m.results)-1 {
				m.sel++
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	// If query changes, refresh results.
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.Type {
		case tea.KeyRunes, tea.KeyBackspace, tea.KeyDelete:
			m.refresh()
		}
	}
	return m, cmd
}

func (m *Model) refresh() {
	if m.index == nil {
		m.results = nil
		return
	}
	m.results = m.index.Search(m.input.Value(), 12)
	if m.sel >= len(m.results) {
		m.sel = len(m.results) - 1
	}
	if m.sel < 0 {
		m.sel = 0
	}
}

func (m Model) View() string {
	w := m.width
	if w <= 0 {
		w = 80
	}
	cardW := min(76, w-4)

	input := lipgloss.NewStyle().
		Padding(0, 1).
		Render(m.input.View())

	var results []string
	if len(m.results) == 0 {
		results = append(results, m.styles.Muted.Padding(0, 2).Render("No results."))
	} else {
		for i, r := range m.results {
			results = append(results, m.renderResult(cardW-2, i, r))
		}
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		input,
		"",
		lipgloss.JoinVertical(lipgloss.Left, results...),
	)

	return lipgloss.Place(w, m.height, lipgloss.Center, lipgloss.Center,
		m.styles.Overlay.Width(cardW).Render(content),
		lipgloss.WithWhitespaceBackground(m.styles.Theme.Bg),
	)
}

func (m Model) renderResult(w, idx int, r search.Result) string {
	indicator := "  "
	style := lipgloss.NewStyle().Width(w).Padding(0, 1)
	if idx == m.sel {
		indicator = "> "
		style = style.Foreground(m.styles.Theme.Accent)
	}

	title := r.Item.Title
	hint := ""
	if r.Item.Hint != "" {
		hint = " " + m.styles.Muted.Render(r.Item.Hint)
	}

	line := indicator + title + hint
	return style.Render(truncate(line, w-2))
}

func truncate(s string, w int) string {
	if w <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= w {
		return s
	}
	r := []rune(s)
	if w <= 1 {
		return "…"
	}
	if len(r) <= w-1 {
		return string(r)
	}
	return string(r[:w-1]) + "…"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
