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
	Muted:   lipgloss.Color("#81A1C1"),
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

// --- 2026 DARK THEMES ---

var ObsidianBloom = Theme{
	Name:    "obsidian_bloom",
	IsLight: false,
	Bg:      lipgloss.Color("#161616"), // Deep charcoal
	Fg:      lipgloss.Color("#E8E8E8"),
	Muted:   lipgloss.Color("#696969"),
	Border:  lipgloss.Color("#2D2D2D"),
	Accent:  lipgloss.Color("#FF1493"), // Neon magenta
	Good:    lipgloss.Color("#00CED1"), // Neon teal
	Warn:    lipgloss.Color("#FF6EC7"),
	Bad:     lipgloss.Color("#FF00FF"),
	Overlay: lipgloss.Color("#252525"),
}

var NeonReef = Theme{
	Name:    "neon_reef",
	IsLight: false,
	Bg:      lipgloss.Color("#000814"), // Ocean-black
	Fg:      lipgloss.Color("#E0FFFF"), // Cyan-tinted white
	Muted:   lipgloss.Color("#5A7A7A"),
	Border:  lipgloss.Color("#1A2A2A"),
	Accent:  lipgloss.Color("#00FFFF"), // Cyan
	Good:    lipgloss.Color("#00FF7F"), // Aqua green
	Warn:    lipgloss.Color("#00D4FF"), // Electric aqua
	Bad:     lipgloss.Color("#C71585"), // Electric purple
	Overlay: lipgloss.Color("#0F1F2A"),
}

var CarbonSunset = Theme{
	Name:    "carbon_sunset",
	IsLight: false,
	Bg:      lipgloss.Color("#0A0A0A"), // Near-black
	Fg:      lipgloss.Color("#F5E6D3"),
	Muted:   lipgloss.Color("#7A6F63"),
	Border:  lipgloss.Color("#1F1F1F"),
	Accent:  lipgloss.Color("#FF6B35"), // Burnt orange
	Good:    lipgloss.Color("#F4A261"), // Terracotta
	Warn:    lipgloss.Color("#E76F51"), // Rust
	Bad:     lipgloss.Color("#D62828"),
	Overlay: lipgloss.Color("#1A1A1A"),
}

var VantaAurora = Theme{
	Name:    "vanta_aurora",
	IsLight: false,
	Bg:      lipgloss.Color("#050810"), // Ultra-dark
	Fg:      lipgloss.Color("#D4E6FF"),
	Muted:   lipgloss.Color("#6B7D8C"),
	Border:  lipgloss.Color("#1A1F2E"),
	Accent:  lipgloss.Color("#00D4AA"), // Aurora green
	Good:    lipgloss.Color("#7B68EE"), // Aurora purple
	Warn:    lipgloss.Color("#00FF88"),
	Bad:     lipgloss.Color("#FF006E"),
	Overlay: lipgloss.Color("#0F1620"),
}

var PlasmaGrape = Theme{
	Name:    "plasma_grape",
	IsLight: false,
	Bg:      lipgloss.Color("#1A0033"), // Dark violet
	Fg:      lipgloss.Color("#F0E6FF"),
	Muted:   lipgloss.Color("#8B6BA8"),
	Border:  lipgloss.Color("#330066"),
	Accent:  lipgloss.Color("#FF1493"), // Red-magenta
	Good:    lipgloss.Color("#E91E63"),
	Warn:    lipgloss.Color("#FF0055"),
	Bad:     lipgloss.Color("#C41E3A"),
	Overlay: lipgloss.Color("#2D0052"),
}

var MidnightJade = Theme{
	Name:    "midnight_jade",
	IsLight: false,
	Bg:      lipgloss.Color("#0D2B2B"), // Deep teal
	Fg:      lipgloss.Color("#D4F1F4"),
	Muted:   lipgloss.Color("#5A8888"),
	Border:  lipgloss.Color("#1A4D4D"),
	Accent:  lipgloss.Color("#7FDBCA"), // Jade
	Good:    lipgloss.Color("#20B2AA"), // Subtle jade
	Warn:    lipgloss.Color("#DAA520"), // Subtle gold
	Bad:     lipgloss.Color("#CD5C5C"),
	Overlay: lipgloss.Color("#153D3D"),
}

var SynthwaveMinimal = Theme{
	Name:    "synthwave_minimal",
	IsLight: false,
	Bg:      lipgloss.Color("#1A1828"), // Muted dark
	Fg:      lipgloss.Color("#E8E0E0"),
	Muted:   lipgloss.Color("#6F5F6F"),
	Border:  lipgloss.Color("#2D2532"),
	Accent:  lipgloss.Color("#00F5FF"), // Controlled neon
	Good:    lipgloss.Color("#0FFF50"),
	Warn:    lipgloss.Color("#FF006E"),
	Bad:     lipgloss.Color("#FF0040"),
	Overlay: lipgloss.Color("#252030"),
}

var GraphiteMatcha = Theme{
	Name:    "graphite_matcha",
	IsLight: false,
	Bg:      lipgloss.Color("#2A2A2A"), // Soft graphite
	Fg:      lipgloss.Color("#D8D8D8"),
	Muted:   lipgloss.Color("#6F6F6F"),
	Border:  lipgloss.Color("#3D3D3D"),
	Accent:  lipgloss.Color("#7CB342"), // Muted green
	Good:    lipgloss.Color("#9CCC65"),
	Warn:    lipgloss.Color("#D4A574"),
	Bad:     lipgloss.Color("#E57373"),
	Overlay: lipgloss.Color("#353535"),
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
	Accent:  lipgloss.Color("#3B82F6"), // Vibrant Blue
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
	Accent:  lipgloss.Color("#D97706"), // Rich Amber
	Good:    lipgloss.Color("#4A5D23"),
	Warn:    lipgloss.Color("#A0522D"),
	Bad:     lipgloss.Color("#8B0000"),
	Overlay: lipgloss.Color("#EADCB8"),
}

// --- 2026 LIGHT THEMES ---

var CloudDancer = Theme{
	Name:    "cloud_dancer",
	IsLight: true,
	Bg:      lipgloss.Color("#F9F7F4"), // Warm off-white
	Fg:      lipgloss.Color("#4A4A4A"),
	Muted:   lipgloss.Color("#8C8C8C"),
	Border:  lipgloss.Color("#E8E8E8"),
	Accent:  lipgloss.Color("#B8860B"), // Soft neutral
	Good:    lipgloss.Color("#5C7D3F"),
	Warn:    lipgloss.Color("#C67C4E"),
	Bad:     lipgloss.Color("#C53030"),
	Overlay: lipgloss.Color("#F0EEE9"),
}

var SakuraSand = Theme{
	Name:    "sakura_sand",
	IsLight: true,
	Bg:      lipgloss.Color("#FCF5F3"), // Soft pink
	Fg:      lipgloss.Color("#52464B"),
	Muted:   lipgloss.Color("#8B7D87"),
	Border:  lipgloss.Color("#EDDBDC"),
	Accent:  lipgloss.Color("#D9825B"), // Sand/coral
	Good:    lipgloss.Color("#8B7355"), // Warm sand
	Warn:    lipgloss.Color("#D47D6E"),
	Bad:     lipgloss.Color("#A85A6A"),
	Overlay: lipgloss.Color("#F5E8E3"),
}

var OliveMist = Theme{
	Name:    "olive_mist",
	IsLight: true,
	Bg:      lipgloss.Color("#F7F9F4"), // Light sage base
	Fg:      lipgloss.Color("#5A6D52"),
	Muted:   lipgloss.Color("#8A9984"),
	Border:  lipgloss.Color("#E6EBE1"),
	Accent:  lipgloss.Color("#7A9E6B"), // Sage/olive
	Good:    lipgloss.Color("#6B8E54"),
	Warn:    lipgloss.Color("#C9A876"),
	Bad:     lipgloss.Color("#9B5C5C"),
	Overlay: lipgloss.Color("#EFF2E9"),
}

var TerracottaAir = Theme{
	Name:    "terracotta_air",
	IsLight: true,
	Bg:      lipgloss.Color("#FBF8F5"), // Light background
	Fg:      lipgloss.Color("#5D4E47"),
	Muted:   lipgloss.Color("#8C8077"),
	Border:  lipgloss.Color("#E8DFD6"),
	Accent:  lipgloss.Color("#C85A3A"), // Terracotta
	Good:    lipgloss.Color("#8B7355"),
	Warn:    lipgloss.Color("#D47D5F"),
	Bad:     lipgloss.Color("#A0453D"),
	Overlay: lipgloss.Color("#F3EDEB"),
}

var VanillaSky = Theme{
	Name:    "vanilla_sky",
	IsLight: true,
	Bg:      lipgloss.Color("#FFFBF7"), // Warm white
	Fg:      lipgloss.Color("#4A4A4A"),
	Muted:   lipgloss.Color("#8C8C8C"),
	Border:  lipgloss.Color("#E8E8E8"),
	Accent:  lipgloss.Color("#5B9FBD"), // Sky blue
	Good:    lipgloss.Color("#5C7D3F"),
	Warn:    lipgloss.Color("#D9A574"),
	Bad:     lipgloss.Color("#C53030"),
	Overlay: lipgloss.Color("#F5F0ED"),
}

var PeachFuzzNeo = Theme{
	Name:    "peach_fuzz_neo",
	IsLight: true,
	Bg:      lipgloss.Color("#FEF5F1"), // Soft peach
	Fg:      lipgloss.Color("#5A4A47"),
	Muted:   lipgloss.Color("#8B7B78"),
	Border:  lipgloss.Color("#E8D9D6"),
	Accent:  lipgloss.Color("#E8B4A8"), // Peach
	Good:    lipgloss.Color("#B49FA3"), // Lavender accent
	Warn:    lipgloss.Color("#D4A574"),
	Bad:     lipgloss.Color("#C85A3A"),
	Overlay: lipgloss.Color("#F5E8E3"),
}

var CoastalDrift = Theme{
	Name:    "coastal_drift",
	IsLight: true,
	Bg:      lipgloss.Color("#F5F9F8"), // Soft white
	Fg:      lipgloss.Color("#4A5D6D"),
	Muted:   lipgloss.Color("#7A8E9E"),
	Border:  lipgloss.Color("#D9E8E8"),
	Accent:  lipgloss.Color("#6BBBB8"), // Seafoam
	Good:    lipgloss.Color("#5C8F7D"),
	Warn:    lipgloss.Color("#B8A574"),
	Bad:     lipgloss.Color("#A0453D"),
	Overlay: lipgloss.Color("#EBF2F1"),
}

var MatchaLatte = Theme{
	Name:    "matcha_latte",
	IsLight: true,
	Bg:      lipgloss.Color("#F9F7F5"), // Creamy base
	Fg:      lipgloss.Color("#52483E"),
	Muted:   lipgloss.Color("#8B7D72"),
	Border:  lipgloss.Color("#E8DBCE"),
	Accent:  lipgloss.Color("#8BA85C"), // Matcha green
	Good:    lipgloss.Color("#7B9E45"),
	Warn:    lipgloss.Color("#C9A876"), // Light brown
	Bad:     lipgloss.Color("#A85A3C"),
	Overlay: lipgloss.Color("#F0E8E0"),
}

// --- HYBRID / SIGNATURE THEMES ---

var DigitalLavender = Theme{
	Name:    "digital_lavender",
	IsLight: false,
	Bg:      lipgloss.Color("#2A1F3D"), // Soft purple base
	Fg:      lipgloss.Color("#E8D5F2"),
	Muted:   lipgloss.Color("#8B7BA3"),
	Border:  lipgloss.Color("#3F2E5C"),
	Accent:  lipgloss.Color("#A78BFA"), // Lavender accent
	Good:    lipgloss.Color("#C4B5FD"),
	Warn:    lipgloss.Color("#F0ABFC"),
	Bad:     lipgloss.Color("#E879F9"),
	Overlay: lipgloss.Color("#3D2D55"),
}

var NeoMintSystem = Theme{
	Name:    "neo_mint_system",
	IsLight: true,
	Bg:      lipgloss.Color("#F0FDF8"), // Fresh mint light
	Fg:      lipgloss.Color("#3D5555"),
	Muted:   lipgloss.Color("#7A9E94"),
	Border:  lipgloss.Color("#D1E8E3"),
	Accent:  lipgloss.Color("#10B981"), // Fresh mint green
	Good:    lipgloss.Color("#059669"), // Vibrant accent
	Warn:    lipgloss.Color("#F59E0B"),
	Bad:     lipgloss.Color("#DC2626"),
	Overlay: lipgloss.Color("#D5EEEA"),
}

var SunsetGradientPro = Theme{
	Name:    "sunset_gradient_pro",
	IsLight: false,
	Bg:      lipgloss.Color("#1A0F0A"), // Warm gradient base
	Fg:      lipgloss.Color("#F5E8D8"),
	Muted:   lipgloss.Color("#9B7D6B"),
	Border:  lipgloss.Color("#3D2416"),
	Accent:  lipgloss.Color("#FF8C42"), // Orange
	Good:    lipgloss.Color("#FF6B9D"), // Pink transition
	Warn:    lipgloss.Color("#C239B3"), // Purple transition
	Bad:     lipgloss.Color("#FF0055"),
	Overlay: lipgloss.Color("#2D1810"),
}

var ForestSanctuary = Theme{
	Name:    "forest_sanctuary",
	IsLight: false,
	Bg:      lipgloss.Color("#1B3D2C"), // Deep forest green
	Fg:      lipgloss.Color("#D4E6D4"),
	Muted:   lipgloss.Color("#7A9F8B"),
	Border:  lipgloss.Color("#2D5A42"),
	Accent:  lipgloss.Color("#6BBB99"), // Sage accent
	Good:    lipgloss.Color("#52B788"),
	Warn:    lipgloss.Color("#D4A574"),
	Bad:     lipgloss.Color("#CD5C5C"),
	Overlay: lipgloss.Color("#2D4A39"),
}

var Steel = Theme{
	Name:    "steel",
	IsLight: false,
	Bg:      lipgloss.Color("#1e2a35"),
	Fg:      lipgloss.Color("#dce8f0"),
	Muted:   lipgloss.Color("#4a6478"),
	Border:  lipgloss.Color("#3a4f62"),
	Accent:  lipgloss.Color("#5dade2"),
	Good:    lipgloss.Color("#a6e3a1"),
	Warn:    lipgloss.Color("#f9e2af"),
	Bad:     lipgloss.Color("#f38ba8"),
	Overlay: lipgloss.Color("#253342"),
}

func Builtins() []Theme {
	return []Theme{
		Steel,
		Catppuccin, Midnight, Aurora, Cyberpunk, Dracula, Nord,
		ObsidianBloom, NeonReef, CarbonSunset, VantaAurora, PlasmaGrape, MidnightJade, SynthwaveMinimal, GraphiteMatcha,
		Vanilla, Solarized, Rose, Matcha, Cloud, Sepia,
		CloudDancer, SakuraSand, OliveMist, TerracottaAir, VanillaSky, PeachFuzzNeo, CoastalDrift, MatchaLatte,
		DigitalLavender, NeoMintSystem, SunsetGradientPro, ForestSanctuary,
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

func FindBuiltin(name string) Theme {
	return ByName(name)
}
