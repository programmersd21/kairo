package styles

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/programmersd21/kairo/internal/core"
	"github.com/programmersd21/kairo/internal/ui/theme"
)

type Styles struct {
	Theme theme.Theme

	App      lipgloss.Style
	Header   lipgloss.Style
	Footer   lipgloss.Style
	Panel    lipgloss.Style
	Title    lipgloss.Style
	Muted    lipgloss.Style
	Selected lipgloss.Style

	Badge      lipgloss.Style
	BadgeGood  lipgloss.Style
	BadgeWarn  lipgloss.Style
	BadgeBad   lipgloss.Style
	BadgeMuted lipgloss.Style

	Overlay lipgloss.Style
	Input   lipgloss.Style
}

func New(t theme.Theme) Styles {
	base := lipgloss.NewStyle().Foreground(t.Fg).Background(t.Bg)
	border := lipgloss.RoundedBorder()
	return Styles{
		Theme:    t,
		App:      base,
		Header:   base.Padding(0, 1).BorderBottom(true).BorderStyle(border).BorderForeground(t.Border),
		Footer:   base.Padding(0, 1).BorderTop(true).BorderStyle(border).BorderForeground(t.Border),
		Panel:    base.Padding(0, 1),
		Title:    base.Bold(true).Foreground(t.Accent),
		Muted:    base.Foreground(t.Muted),
		Selected: base.Background(t.Overlay).Foreground(t.Accent).Bold(true),

		Badge:      lipgloss.NewStyle().Padding(0, 1).Background(t.Overlay).Foreground(t.Fg),
		BadgeGood:  lipgloss.NewStyle().Padding(0, 1).Background(t.Good).Foreground(t.Bg).Bold(true),
		BadgeWarn:  lipgloss.NewStyle().Padding(0, 1).Background(t.Warn).Foreground(t.Bg).Bold(true),
		BadgeBad:   lipgloss.NewStyle().Padding(0, 1).Background(t.Bad).Foreground(t.Bg).Bold(true),
		BadgeMuted: lipgloss.NewStyle().Padding(0, 1).Background(t.Muted).Foreground(t.Bg),

		Overlay: lipgloss.NewStyle().Background(t.Overlay).Foreground(t.Fg).Padding(0, 1),
		Input:   lipgloss.NewStyle().Foreground(t.Fg).Background(t.Overlay).Padding(0, 1),
	}
}

func (s Styles) StatusBadge(st core.Status) string {
	switch st {
	case core.StatusTodo:
		return s.BadgeMuted.Render("TODO")
	case core.StatusDoing:
		return s.BadgeWarn.Render("DOING")
	case core.StatusDone:
		return s.BadgeGood.Render("DONE")
	default:
		return s.BadgeMuted.Render(string(st))
	}
}

func (s Styles) PriorityBadge(p core.Priority) string {
	switch p.Clamp() {
	case core.P0:
		return s.BadgeMuted.Render("P0")
	case core.P1:
		return s.Badge.Render("P1")
	case core.P2:
		return s.BadgeWarn.Render("P2")
	case core.P3:
		return s.BadgeBad.Render("P3")
	default:
		return s.BadgeMuted.Render(fmt.Sprintf("P%d", int(p)))
	}
}
