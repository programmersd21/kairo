package palette

import (
	"strings"

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
	cardW := min(72, w-4)
	if cardW < 44 {
		cardW = w - 2
	}
	m.input.Width = cardW - 6

	header := m.styles.Title.Render("Search")
	input := m.styles.Input.Width(cardW - 4).Render(m.input.View())

	lines := []string{
		lipgloss.NewStyle().Padding(0, 2).Width(cardW - 4).Render(header),
		lipgloss.NewStyle().Padding(0, 1).Render(input),
	}

	if len(m.results) == 0 {
		lines = append(lines, m.styles.Muted.Padding(1, 2).Width(cardW-4).Render("No results found"))
	} else {
		for i, r := range m.results {
			lines = append(lines, m.renderResult(cardW-2, i, r))
		}
	}

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

func (m Model) renderResult(w, idx int, r search.Result) string {
	kind := strings.ToUpper(string(r.Item.Kind))
	kindStyle := m.styles.Muted
	switch r.Item.Kind {
	case search.KindCommand:
		kindStyle = lipgloss.NewStyle().Foreground(m.styles.Theme.Accent).Bold(true)
	case search.KindTask:
		kindStyle = m.styles.Muted
	case search.KindTag:
		kindStyle = lipgloss.NewStyle().Foreground(m.styles.Theme.Warn)
	}

	left := kindStyle.Width(10).Render(kind)
	title := r.Item.Title
	if r.Item.Hint != "" {
		title = title + m.styles.Muted.Render("  "+r.Item.Hint)
	}

	indicator := "  "
	st := lipgloss.NewStyle().Width(w).Padding(0, 2).Background(m.styles.Theme.Bg)
	if idx == m.sel {
		indicator = lipgloss.NewStyle().Foreground(m.styles.Theme.Accent).Render("→ ")
		st = st.Background(m.styles.Theme.Overlay).Foreground(m.styles.Theme.Accent).Bold(true)
	}

	line := indicator + left + title
	return st.Render(truncate(line, w-4))
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
