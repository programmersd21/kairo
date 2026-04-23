package styles

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/programmersd21/kairo/internal/core"
	"github.com/programmersd21/kairo/internal/ui/theme"
)

// Icons used throughout the app.
// Designed with a "Premium & Sentimental" aesthetic for modern terminals.
const (
	IconTodo      = "󰄱 "
	IconDoing     = "󰔟 "
	IconDone      = "󰄲 "
	IconPriority0 = "󰼎 "
	IconPriority1 = "󰼏 "
	IconPriority2 = "󰼐 "
	IconPriority3 = "󰼑 "
	IconDeadline  = "󰃰 "
	IconTag       = "󰓹 "
	IconSync      = "󰑓 "
	IconError     = "󰅚 "
	IconInfo      = "󰋽 "
	IconHelp      = "󰋗 "
	IconTask      = "󰈈 "
	IconPlugin    = "󰡀 "
	// UI Affordances
	IconPalette   = "󰳟 "
	IconNew       = "󰐕 "
	IconDelete    = "󰆴 "
	IconView      = "󰈈 "
	IconStrike    = "󱐌 "
	IconIssues    = "󰋽 "
	IconChangelog = "󰠠 "
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
	Badge             lipgloss.Style
	BadgeGood         lipgloss.Style
	BadgeWarn         lipgloss.Style
	BadgeBad          lipgloss.Style
	BadgeMuted        lipgloss.Style
	BadgeDelete       lipgloss.Style
	BadgeQuit         lipgloss.Style
	BadgeOutlineGood  lipgloss.Style
	BadgeOutlineWarn  lipgloss.Style
	BadgeOutlineBad   lipgloss.Style
	BadgeOutlineMuted lipgloss.Style

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
	accentStyle := lipgloss.NewStyle().Foreground(t.Accent).Background(t.Bg)
	mutedStyle := lipgloss.NewStyle().Foreground(t.Muted).Background(t.Bg)

	selection := lipgloss.NewStyle().
		Foreground(t.Bg).
		Background(t.Accent).
		Bold(true)

	// Contrast color for badge text
	contrast := lipgloss.Color("#FFFFFF")
	if t.IsLight {
		contrast = t.Bg // Use theme background (usually light) for text on colored badges in light themes
	}

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
			Foreground(t.Bg).
			Background(t.Accent).
			Bold(true).
			Padding(0, 1),
		TabInactive: lipgloss.NewStyle().
			Foreground(t.Muted).
			Background(t.Bg).
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
			Foreground(t.Muted).
			Background(t.Bg),
		BadgeGood: lipgloss.NewStyle().
			Foreground(t.Good).
			Background(t.Bg),
		BadgeWarn: lipgloss.NewStyle().
			Foreground(t.Warn).
			Background(t.Bg),
		BadgeBad: lipgloss.NewStyle().
			Foreground(t.Bad).
			Background(t.Bg),
		BadgeMuted: lipgloss.NewStyle().
			Foreground(t.Muted).
			Background(t.Bg),
		BadgeDelete: lipgloss.NewStyle().
			Foreground(contrast).
			Background(t.Bad).
			Bold(true).
			Padding(0, 1),
		BadgeQuit: lipgloss.NewStyle().
			Foreground(contrast).
			Background(t.Warn).
			Bold(true).
			Padding(0, 1),
		BadgeOutlineGood: lipgloss.NewStyle().
			Foreground(contrast).
			Background(t.Good).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(t.Good).
			BorderTop(false).
			BorderBottom(false).
			Bold(true).
			Padding(0, 1),
		BadgeOutlineWarn: lipgloss.NewStyle().
			Foreground(contrast).
			Background(t.Warn).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(t.Warn).
			BorderTop(false).
			BorderBottom(false).
			Bold(true).
			Padding(0, 1),
		BadgeOutlineBad: lipgloss.NewStyle().
			Foreground(contrast).
			Background(t.Bad).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(t.Bad).
			BorderTop(false).
			BorderBottom(false).
			Bold(true).
			Padding(0, 1),
		BadgeOutlineMuted: lipgloss.NewStyle().
			Foreground(contrast).
			Background(t.Muted).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(t.Muted).
			BorderTop(false).
			BorderBottom(false).
			Bold(true).
			Padding(0, 1),

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
		return s.BadgeOutlineMuted.Render(IconPriority0 + "P0")
	case core.P1:
		return s.BadgeOutlineMuted.Render(IconPriority1 + "P1")
	case core.P2:
		return s.BadgeOutlineWarn.Render(IconPriority2 + "P2")
	case core.P3:
		return s.BadgeOutlineBad.Render(IconPriority3 + "P3")
	default:
		return s.BadgeOutlineMuted.Render(fmt.Sprintf("P%d", int(p)))
	}
}
