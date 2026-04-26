package ai_panel

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/programmersd21/kairo/internal/ui/styles"
)

type Styles struct {
	Panel       lipgloss.Style
	Header      lipgloss.Style
	User        lipgloss.Style
	AI          lipgloss.Style
	Tool        lipgloss.Style
	Footer      lipgloss.Style
	InputBorder lipgloss.Style
}

func DefaultStyles(s styles.Styles) Styles {
	return Styles{
		Panel: lipgloss.NewStyle().
			Border(lipgloss.ThickBorder(), false, false, false, true).
			BorderForeground(s.Theme.Accent).
			Background(s.Theme.Bg).
			Padding(0, 1),
		Header: lipgloss.NewStyle().
			Bold(true).
			Foreground(s.Theme.Bg).
			Background(s.Theme.Accent).
			Padding(0, 1).
			Align(lipgloss.Center),
		User: lipgloss.NewStyle().
			Bold(true).
			Foreground(s.Theme.Good),
		AI: lipgloss.NewStyle().
			Foreground(s.Theme.Fg),
		Tool: lipgloss.NewStyle().
			Italic(true).
			Foreground(s.Theme.Warn),
		Footer: lipgloss.NewStyle().
			Foreground(s.Theme.Muted).
			Italic(true),
		InputBorder: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(s.Theme.Accent).
			Padding(0, 1),
	}
}
