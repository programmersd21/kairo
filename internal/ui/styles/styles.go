package styles

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/programmersd21/kairo/internal/core"
	"github.com/programmersd21/kairo/internal/ui/theme"
)

// Icons used throughout the app.
// Using standard Unicode symbols and emojis for maximum compatibility.
const (
	IconTodo      = "○ "
	IconDoing     = "◐ "
	IconDone      = "● "
	IconPriority0 = "⓪ "
	IconPriority1 = "① "
	IconPriority2 = "② "
	IconPriority3 = "③ "
	IconDeadline  = "⏲ "
	IconTag       = "# "
	IconSync      = "↻ "
	IconError     = "✖ "
	IconInfo      = "ℹ "
	IconHelp      = "? "
	IconTask      = "❖ "
	IconPlugin    = "🧩 "
)

type Styles struct {
	Theme theme.Theme

	App    lipgloss.Style
	Header lipgloss.Style
	Footer lipgloss.Style
	Panel  lipgloss.Style
	Title  lipgloss.Style
	Muted  lipgloss.Style

	// Tabs
	TabActive   lipgloss.Style
	TabInactive lipgloss.Style

	// Rows
	RowSelected lipgloss.Style
	RowNormal   lipgloss.Style
	RowDimmed   lipgloss.Style

	// Badges
	Badge      lipgloss.Style
	BadgeGood  lipgloss.Style
	BadgeWarn  lipgloss.Style
	BadgeBad   lipgloss.Style
	BadgeMuted lipgloss.Style

	// Detail
	DetailKey   lipgloss.Style
	DetailValue lipgloss.Style

	// Overlays
	Overlay lipgloss.Style
	Input   lipgloss.Style
}

func New(t theme.Theme) Styles {
	base := lipgloss.NewStyle().Foreground(t.Fg).Background(t.Bg)

	return Styles{
		Theme: t,
		App:   base,
		Header: lipgloss.NewStyle().
			Background(t.Bg).
			Padding(0, 1).
			Height(1),
		Footer: lipgloss.NewStyle().
			Background(t.Bg).
			Padding(0, 1).
			Height(1),
		Panel: base.Padding(0, 2),
		Title: base.Bold(true).Foreground(t.Accent),
		Muted: base.Foreground(t.Muted),

		TabActive: lipgloss.NewStyle().
			Foreground(t.Accent).
			Background(t.Overlay).
			Bold(true).
			Padding(0, 2),
		TabInactive: lipgloss.NewStyle().
			Foreground(t.Muted).
			Background(t.Bg).
			Padding(0, 2),

		RowSelected: lipgloss.NewStyle().
			Foreground(t.Accent).
			Background(t.Overlay).
			Bold(true),
		RowNormal: lipgloss.NewStyle().
			Foreground(t.Fg).
			Background(t.Bg),
		RowDimmed: lipgloss.NewStyle().
			Foreground(t.Muted).
			Background(t.Bg),

		Badge:      lipgloss.NewStyle().Padding(0, 1).Background(t.Overlay).Foreground(t.Fg),
		BadgeGood:  lipgloss.NewStyle().Padding(0, 1).Background(t.Good).Foreground(t.Bg).Bold(true),
		BadgeWarn:  lipgloss.NewStyle().Padding(0, 1).Background(t.Warn).Foreground(t.Bg).Bold(true),
		BadgeBad:   lipgloss.NewStyle().Padding(0, 1).Background(t.Bad).Foreground(t.Bg).Bold(true),
		BadgeMuted: lipgloss.NewStyle().Padding(0, 1).Background(t.Muted).Foreground(t.Bg),

		DetailKey: lipgloss.NewStyle().
			Foreground(t.Muted).
			Bold(true).
			Width(12),
		DetailValue: lipgloss.NewStyle().
			Foreground(t.Fg),

		Overlay: lipgloss.NewStyle().
			Background(t.Overlay).
			Foreground(t.Fg).
			Padding(1, 2).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(t.Accent),
		Input: lipgloss.NewStyle().
			Foreground(t.Fg).
			Background(t.Overlay).
			Padding(0, 1),
	}
}

func (s Styles) StatusBadge(st core.Status) string {
	switch st {
	case core.StatusTodo:
		return s.BadgeMuted.Render(IconTodo + "TODO")
	case core.StatusDoing:
		return s.BadgeWarn.Render(IconDoing + "DOING")
	case core.StatusDone:
		return s.BadgeGood.Render(IconDone + "DONE")
	default:
		return s.BadgeMuted.Render(string(st))
	}
}

func (s Styles) PriorityBadge(p core.Priority) string {
	switch p.Clamp() {
	case core.P0:
		return s.BadgeMuted.Render(IconPriority0 + "P0")
	case core.P1:
		return s.Badge.Render(IconPriority1 + "P1")
	case core.P2:
		return s.BadgeWarn.Render(IconPriority2 + "P2")
	case core.P3:
		return s.BadgeBad.Render(IconPriority3 + "P3")
	default:
		return s.BadgeMuted.Render(fmt.Sprintf("P%d", int(p)))
	}
}
