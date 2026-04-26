package import_export_menu

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/programmersd21/kairo/internal/ui/styles"
)

type Action int

const (
	ActionExportJSON Action = iota
	ActionExportMD
	ActionExportCSV
	ActionExportTXT
	ActionImportJSON
	ActionImportMD
	ActionImportCSV
	ActionImportTXT
)

func (a Action) Label() string {
	switch a {
	case ActionExportJSON:
		return "Export to JSON"
	case ActionExportMD:
		return "Export to Markdown"
	case ActionExportCSV:
		return "Export to CSV"
	case ActionExportTXT:
		return "Export to Text"
	case ActionImportJSON:
		return "Import from JSON"
	case ActionImportMD:
		return "Import from Markdown"
	case ActionImportCSV:
		return "Import from CSV"
	case ActionImportTXT:
		return "Import from Text"
	default:
		return "Unknown"
	}
}

func (a Action) IsExport() bool {
	return a <= ActionExportTXT
}

func (a Action) Extension() string {
	switch a {
	case ActionExportJSON, ActionImportJSON:
		return ".json"
	case ActionExportMD, ActionImportMD:
		return ".md"
	case ActionExportCSV, ActionImportCSV:
		return ".csv"
	case ActionExportTXT, ActionImportTXT:
		return ".txt"
	default:
		return ".data"
	}
}

func (a Action) Format() string {
	switch a {
	case ActionExportJSON, ActionImportJSON:
		return "json"
	case ActionExportMD, ActionImportMD:
		return "md"
	case ActionExportCSV, ActionImportCSV:
		return "csv"
	case ActionExportTXT, ActionImportTXT:
		return "txt"
	default:
		return "json"
	}
}

type SelectMsg struct {
	Action Action
	Path   string
}

type CloseMsg struct{}

type Model struct {
	styles styles.Styles
	width  int
	height int
	sel    int
	items  []Action
	input  textinput.Model
	isPath bool
}

func New(s styles.Styles) Model {
	items := []Action{
		ActionExportJSON, ActionExportMD, ActionExportCSV, ActionExportTXT,
		ActionImportJSON, ActionImportMD, ActionImportCSV, ActionImportTXT,
	}
	ti := textinput.New()
	ti.Placeholder = "Enter file path..."
	ti.Width = 40
	return Model{
		styles: s,
		sel:    0,
		items:  items,
		input:  ti,
		isPath: false,
	}
}

func (m *Model) SetSize(w, h int) {
	m.width, m.height = w, h
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if m.isPath {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		if km, ok := msg.(tea.KeyMsg); ok {
			switch km.String() {
			case "esc":
				m.isPath = false
				return m, nil
			case "enter":
				path := m.input.Value()
				if path != "" {
					return m, func() tea.Msg { return SelectMsg{Action: m.items[m.sel], Path: path} }
				}
			}
		}
		return m, cmd
	}

	switch x := msg.(type) {
	case tea.KeyMsg:
		switch x.String() {
		case "esc", "q":
			return m, func() tea.Msg { return CloseMsg{} }
		case "up", "k":
			if m.sel > 0 {
				m.sel--
			}
		case "down", "j":
			if m.sel < len(m.items)-1 {
				m.sel++
			}
		case "enter":
			action := m.items[m.sel]
			if action.IsExport() {
				// Set default path for export
				fname := fmt.Sprintf("kairo_export_%s%s", time.Now().Format("2006-01-02_150405"), action.Extension())
				m.input.SetValue(fname)
			} else {
				m.input.SetValue("")
			}
			m.isPath = true
			m.input.Focus()
			return m, nil
		}
	}
	return m, nil
}

func (m Model) View() string {
	w := m.width
	if w <= 0 {
		w = 80
	}
	cardW := 50

	header := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Width(cardW).
		Render(m.styles.Title.Render(" Import / Export Tasks "))

	var lines []string
	lines = append(lines, header, "")

	if m.isPath {
		action := m.items[m.sel]
		label := "Save to:"
		if !action.IsExport() {
			label = "Import from:"
		}
		lines = append(lines, m.styles.Muted.Render(action.Label()))
		lines = append(lines, "")
		lines = append(lines, label)
		lines = append(lines, m.input.View())
		lines = append(lines, "")
		lines = append(lines, m.styles.Muted.Render("Press Enter to confirm, Esc to cancel"))
	} else {
		for i, action := range m.items {
			// Add a separator between Export and Import
			if i == 4 {
				lines = append(lines, "")
			}

			style := m.styles.RowNormal
			prefix := "  "
			if i == m.sel {
				style = m.styles.RowSelected
				prefix = "> "
			}
			lines = append(lines, style.Render(prefix+action.Label()))
		}
	}

	return lipgloss.Place(w, m.height, lipgloss.Center, lipgloss.Center,
		m.styles.Overlay.Width(cardW).Padding(1, 2).Render(lipgloss.JoinVertical(lipgloss.Left, lines...)),
		lipgloss.WithWhitespaceBackground(m.styles.Theme.Bg),
	)
}
