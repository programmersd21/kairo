package theme

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	Name    string
	IsLight bool

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

// --- DARK THEMES ---

var Catppuccin = Theme{
	Name:    "catppuccin",
	IsLight: false,
	Bg:      lipgloss.Color("#1E1E2E"),
	Fg:      lipgloss.Color("#CDD6F4"),
	Muted:   lipgloss.Color("#6C7086"),
	Border:  lipgloss.Color("#45475A"),
	Accent:  lipgloss.Color("#89B4FA"),
	Good:    lipgloss.Color("#A6E3A1"),
	Warn:    lipgloss.Color("#F9E2AF"),
	Bad:     lipgloss.Color("#F38BA8"),
	Overlay: lipgloss.Color("#313244"),
}

var Midnight = Theme{
	Name:    "midnight",
	IsLight: false,
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
	IsLight: false,
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
	IsLight: false,
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

var Aurora = Theme{
	Name:    "aurora",
	IsLight: false,
	Bg:      lipgloss.Color("#0F172A"),
	Fg:      lipgloss.Color("#E2E8F0"),
	Muted:   lipgloss.Color("#64748B"),
	Border:  lipgloss.Color("#1E293B"),
	Accent:  lipgloss.Color("#06B6D4"),
	Good:    lipgloss.Color("#10B981"),
	Warn:    lipgloss.Color("#F59E0B"),
	Bad:     lipgloss.Color("#EF4444"),
	Overlay: lipgloss.Color("#1E293B"),
}

var Cyberpunk = Theme{
	Name:    "cyberpunk",
	IsLight: false,
	Bg:      lipgloss.Color("#0D0221"),
	Fg:      lipgloss.Color("#E0F4FF"),
	Muted:   lipgloss.Color("#8F0053"),
	Border:  lipgloss.Color("#5A189A"),
	Accent:  lipgloss.Color("#00F5FF"),
	Good:    lipgloss.Color("#00FF00"),
	Warn:    lipgloss.Color("#FF006E"),
	Bad:     lipgloss.Color("#FF0040"),
	Overlay: lipgloss.Color("#3A0CA3"),
}

// --- LIGHT THEMES (EYE ERGONOMIC) ---

var Vanilla = Theme{
	Name:    "vanilla",
	IsLight: true,
	Bg:      lipgloss.Color("#FFF9E5"), // Soft Cream
	Fg:      lipgloss.Color("#433422"), // Dark Coffee
	Muted:   lipgloss.Color("#A69076"), // Muted Sand
	Border:  lipgloss.Color("#E6D5BC"), // Light Latte
	Accent:  lipgloss.Color("#D97706"), // Amber/Honey
	Good:    lipgloss.Color("#166534"), // Deep Forest Green
	Warn:    lipgloss.Color("#9A3412"), // Burnt Orange
	Bad:     lipgloss.Color("#991B1B"), // Soft Crimson
	Overlay: lipgloss.Color("#F3EAD3"), // Toasted Cream
}

var Solarized = Theme{
	Name:    "solarized",
	IsLight: true,
	Bg:      lipgloss.Color("#FDF6E3"),
	Fg:      lipgloss.Color("#657B83"),
	Muted:   lipgloss.Color("#93A1A1"),
	Border:  lipgloss.Color("#EEE8D5"),
	Accent:  lipgloss.Color("#268BD2"),
	Good:    lipgloss.Color("#859900"),
	Warn:    lipgloss.Color("#B58900"),
	Bad:     lipgloss.Color("#DC322F"),
	Overlay: lipgloss.Color("#EEE8D5"),
}

var Rose = Theme{
	Name:    "rose",
	IsLight: true,
	Bg:      lipgloss.Color("#FFF7F3"), // Soft Rose
	Fg:      lipgloss.Color("#575279"), // Deep Indigo
	Muted:   lipgloss.Color("#797593"),
	Border:  lipgloss.Color("#F2E9E1"),
	Accent:  lipgloss.Color("#D7827E"),
	Good:    lipgloss.Color("#286983"),
	Warn:    lipgloss.Color("#EA9D34"),
	Bad:     lipgloss.Color("#B4637A"),
	Overlay: lipgloss.Color("#F2E9E1"),
}

var Matcha = Theme{
	Name:    "matcha",
	IsLight: true,
	Bg:      lipgloss.Color("#F0F4F0"), // Soft Matcha Green
	Fg:      lipgloss.Color("#2D3436"),
	Muted:   lipgloss.Color("#636E72"),
	Border:  lipgloss.Color("#D1D8D1"),
	Accent:  lipgloss.Color("#2D8A4E"),
	Good:    lipgloss.Color("#417505"),
	Warn:    lipgloss.Color("#D97706"),
	Bad:     lipgloss.Color("#A91E2C"),
	Overlay: lipgloss.Color("#E3EAE3"),
}

var Cloud = Theme{
	Name:    "cloud",
	IsLight: true,
	Bg:      lipgloss.Color("#F1F5F9"), // Soft Blue Gray
	Fg:      lipgloss.Color("#334155"),
	Muted:   lipgloss.Color("#64748B"),
	Border:  lipgloss.Color("#E2E8F0"),
	Accent:  lipgloss.Color("#0F172A"),
	Good:    lipgloss.Color("#10B981"),
	Warn:    lipgloss.Color("#F59E0B"),
	Bad:     lipgloss.Color("#EF4444"),
	Overlay: lipgloss.Color("#E2E8F0"),
}

var Sepia = Theme{
	Name:    "sepia",
	IsLight: true,
	Bg:      lipgloss.Color("#F4ECD8"), // Old Paper
	Fg:      lipgloss.Color("#5B4636"),
	Muted:   lipgloss.Color("#8C7A6B"),
	Border:  lipgloss.Color("#DED0B6"),
	Accent:  lipgloss.Color("#8B4513"),
	Good:    lipgloss.Color("#4A5D23"),
	Warn:    lipgloss.Color("#A0522D"),
	Bad:     lipgloss.Color("#8B0000"),
	Overlay: lipgloss.Color("#EADCB8"),
}

func Builtins() []Theme {
	return []Theme{
		Catppuccin, Midnight, Aurora, Cyberpunk, Dracula, Nord,
		Vanilla, Solarized, Rose, Matcha, Cloud, Sepia,
	}
}

func ByName(name string) Theme {
	for _, t := range Builtins() {
		if t.Name == name {
			return t
		}
	}
	return Catppuccin
}
