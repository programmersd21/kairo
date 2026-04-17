package styles

import (
	"fmt"
	"strings"

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
	IconPriority0 = "0 "
	IconPriority1 = "1 "
	IconPriority2 = "2 "
	IconPriority3 = "3 "
	IconDeadline  = "⏲ "
	IconTag       = "# "
	IconSync      = "↻ "
	IconError     = "✖ "
	IconInfo      = "ℹ "
	IconHelp      = "? "
	IconTask      = "❖ "
	IconPlugin    = "🧩 "
)

// Design System Constants
const (
	// Spacing (compact grid)
	Spacing0 = 0
	Spacing1 = 1
	Spacing2 = 2
)

type Styles struct {
	Theme theme.Theme

	// Base
	App    lipgloss.Style
	Header lipgloss.Style
	Footer lipgloss.Style
	Panel  lipgloss.Style

	// Typography
	Title    lipgloss.Style
	Subtitle lipgloss.Style
	Muted    lipgloss.Style
	Text     lipgloss.Style
	Accent   lipgloss.Style

	// Tabs & Navigation
	TabActive   lipgloss.Style
	TabInactive lipgloss.Style
	Separator   lipgloss.Style

	// Rows & List Items
	RowSelected lipgloss.Style
	RowNormal   lipgloss.Style
	RowHovered  lipgloss.Style
	RowDimmed   lipgloss.Style
	RowFocused  lipgloss.Style

	// Badges & Status
	Badge            lipgloss.Style
	BadgeGood        lipgloss.Style
	BadgeWarn        lipgloss.Style
	BadgeBad         lipgloss.Style
	BadgeMuted       lipgloss.Style
	BadgeOutlineGood lipgloss.Style
	BadgeOutlineWarn lipgloss.Style
	BadgeOutlineBad  lipgloss.Style

	// Detail & Form
	DetailKey   lipgloss.Style
	DetailValue lipgloss.Style
	DetailLabel lipgloss.Style
	FormLabel   lipgloss.Style

	// Components
	Card             lipgloss.Style
	CardHeader       lipgloss.Style
	CardContent      lipgloss.Style
	CardFooter       lipgloss.Style
	Overlay          lipgloss.Style
	Input            lipgloss.Style
	InputFocused     lipgloss.Style
	InputPlaceholder lipgloss.Style
	Button           lipgloss.Style
	ButtonPrimary    lipgloss.Style
	ButtonSecondary  lipgloss.Style
	ButtonActive     lipgloss.Style
	Divider          lipgloss.Style
	Border           lipgloss.Style
	SoftBorder       lipgloss.Style

	// States
	Empty   lipgloss.Style
	Loading lipgloss.Style
	Error   lipgloss.Style
	Success lipgloss.Style
}

func New(t theme.Theme) Styles {
	base := lipgloss.NewStyle().Foreground(t.Fg).Background(t.Bg)
	accentStyle := lipgloss.NewStyle().Foreground(t.Accent)
	mutedStyle := lipgloss.NewStyle().Foreground(t.Muted)

	selection := lipgloss.NewStyle().
		Foreground(t.Bg).
		Background(t.Accent).
		Bold(true)

	return Styles{
		Theme: t,

		// Base
		App:    base,
		Header: base.Padding(0, 1).Height(1),
		Footer: base.Padding(0, 1).Height(1),
		Panel:  base,

		// Typography
		Title:    base.Bold(true).Foreground(t.Accent),
		Subtitle: base.Bold(true).Foreground(t.Muted),
		Muted:    mutedStyle,
		Text:     base,
		Accent:   accentStyle,

		// Tabs & Navigation
		TabActive: lipgloss.NewStyle().
			Foreground(t.Accent).
			Bold(true).
			Padding(0, 1),
		TabInactive: lipgloss.NewStyle().
			Foreground(t.Muted).
			Padding(0, 1),
		Separator: mutedStyle.SetString("│"),

		// Rows & List Items
		RowSelected: selection,
		RowNormal:   base,
		RowHovered:  lipgloss.NewStyle().Foreground(t.Bg).Background(t.Accent),
		RowDimmed:   mutedStyle,
		RowFocused:  selection,

		// Badges - Compact
		Badge: lipgloss.NewStyle().
			Foreground(t.Muted),
		BadgeGood: lipgloss.NewStyle().
			Foreground(t.Good),
		BadgeWarn: lipgloss.NewStyle().
			Foreground(t.Warn),
		BadgeBad: lipgloss.NewStyle().
			Foreground(t.Bad),
		BadgeMuted: lipgloss.NewStyle().
			Foreground(t.Muted),

		// Detail & Form
		DetailKey: mutedStyle.
			Bold(true).
			Width(12).
			MarginRight(1),
		DetailValue: base,
		DetailLabel: mutedStyle.Bold(true),
		FormLabel:   mutedStyle.Bold(true),

		// Components
		Card: lipgloss.NewStyle().
			Background(t.Bg).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(t.Border).
			Padding(0, 1),
		CardHeader:  accentStyle.Bold(true),
		CardContent: base,
		CardFooter:  mutedStyle,
		Overlay: lipgloss.NewStyle().
			Background(t.Bg).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(t.Accent).
			Padding(0, 1),
		Input: lipgloss.NewStyle().
			Background(t.Overlay).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(t.Border),
		InputFocused: lipgloss.NewStyle().
			Background(t.Overlay).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(t.Accent),
		InputPlaceholder: mutedStyle,
		Button:           lipgloss.NewStyle().Padding(0, 1).Foreground(t.Muted),
		ButtonPrimary:    lipgloss.NewStyle().Padding(0, 1).Foreground(t.Accent).Bold(true),
		ButtonSecondary:  lipgloss.NewStyle().Padding(0, 1).Foreground(t.Muted),
		ButtonActive:     lipgloss.NewStyle().Padding(0, 1).Foreground(t.Accent).Bold(true),
		Divider:          mutedStyle.SetString(strings.Repeat("─", 80)),
		Border:           lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderForeground(t.Border),
		SoftBorder:       lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderForeground(t.Border),

		// States
		Empty:   mutedStyle.Italic(true),
		Loading: accentStyle,
		Error:   lipgloss.NewStyle().Foreground(t.Bad),
		Success: lipgloss.NewStyle().Foreground(t.Good),
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
