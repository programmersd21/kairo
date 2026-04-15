package theme

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	Name string

	Bg lipgloss.Color
	Fg lipgloss.Color

	Muted   lipgloss.Color
	Border  lipgloss.Color
	Accent  lipgloss.Color
	Good    lipgloss.Color
	Warn    lipgloss.Color
	Bad     lipgloss.Color
	Overlay lipgloss.Color
}

var Midnight = Theme{
	Name:    "midnight",
	Bg:      lipgloss.Color("#1A1B26"),
	Fg:      lipgloss.Color("#A9B1D6"),
	Muted:   lipgloss.Color("#565F89"),
	Border:  lipgloss.Color("#414868"),
	Accent:  lipgloss.Color("#7AA2F7"),
	Good:    lipgloss.Color("#9ECE6A"),
	Warn:    lipgloss.Color("#E0AF68"),
	Bad:     lipgloss.Color("#F7768E"),
	Overlay: lipgloss.Color("#24283B"),
}

var Dracula = Theme{
	Name:    "dracula",
	Bg:      lipgloss.Color("#282A36"),
	Fg:      lipgloss.Color("#F8F8F2"),
	Muted:   lipgloss.Color("#6272A4"),
	Border:  lipgloss.Color("#44475A"),
	Accent:  lipgloss.Color("#BD93F9"),
	Good:    lipgloss.Color("#50FA7B"),
	Warn:    lipgloss.Color("#FFB86C"),
	Bad:     lipgloss.Color("#FF5555"),
	Overlay: lipgloss.Color("#343746"),
}

var Nord = Theme{
	Name:    "nord",
	Bg:      lipgloss.Color("#2E3440"),
	Fg:      lipgloss.Color("#D8DEE9"),
	Muted:   lipgloss.Color("#4C566A"),
	Border:  lipgloss.Color("#3B4252"),
	Accent:  lipgloss.Color("#88C0D0"),
	Good:    lipgloss.Color("#A3BE8C"),
	Warn:    lipgloss.Color("#EBCB8B"),
	Bad:     lipgloss.Color("#BF616A"),
	Overlay: lipgloss.Color("#3B4252"),
}

var Paper = Theme{
	Name:    "paper",
	Bg:      lipgloss.Color("#FAFAFA"),
	Fg:      lipgloss.Color("#111827"),
	Muted:   lipgloss.Color("#6B7280"),
	Border:  lipgloss.Color("#E5E7EB"),
	Accent:  lipgloss.Color("#2563EB"),
	Good:    lipgloss.Color("#059669"),
	Warn:    lipgloss.Color("#B45309"),
	Bad:     lipgloss.Color("#DC2626"),
	Overlay: lipgloss.Color("#F3F4F6"),
}

func Builtins() []Theme { return []Theme{Midnight, Dracula, Nord, Paper} }

func ByName(name string) Theme {
	for _, t := range Builtins() {
		if t.Name == name {
			return t
		}
	}
	return Midnight
}
