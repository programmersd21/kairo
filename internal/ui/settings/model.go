package settings

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/programmersd21/kairo/internal/config"
	"github.com/programmersd21/kairo/internal/ui/styles"
)

type ConfigChangedMsg struct {
	Config config.Config
}

type CloseMsg struct{}

type settingItem struct {
	label string
	key   string
	kind  string // "bool", "string"
	val   interface{}
}

type Model struct {
	styles styles.Styles
	cfg    config.Config
	width  int
	height int
	sel    int
	items  []settingItem
}

func New(s styles.Styles, cfg config.Config) Model {
	m := Model{
		styles: s,
		cfg:    cfg,
	}
	m.rebuildItems()
	return m
}

func (m *Model) rebuildItems() {
	m.items = []settingItem{
		{"Vim Mode", "vim_mode", "bool", m.cfg.App.VimMode},
		{"Show Help Footer", "show_help", "bool", m.cfg.App.ShowHelp},
		{"Rainbow Logo", "rainbow", "bool", m.cfg.App.Rainbow},
		{"Git Sync Enabled", "sync_enabled", "bool", m.cfg.Sync.Enabled},
		{"Auto Push (Git)", "auto_push", "bool", m.cfg.Sync.AutoPush},
	}
}

func (m *Model) SetSize(w, h int) {
	m.width, m.height = w, h
}

func (m *Model) SetConfig(cfg config.Config) {
	m.cfg = cfg
	m.rebuildItems()
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch x := msg.(type) {
	case tea.KeyMsg:
		switch x.String() {
		case "esc", "ctrl+s":
			return m, func() tea.Msg { return CloseMsg{} }
		case "g":
			path, _ := config.ConfigPath()
			return m, func() tea.Msg {
				var cmd *exec.Cmd
				switch runtime.GOOS {
				case "windows":
					cmd = exec.Command("cmd", "/c", "start", path)
				case "darwin":
					cmd = exec.Command("open", path)
				default:
					cmd = exec.Command("xdg-open", path)
				}
				_ = cmd.Run()
				return nil
			}
		case "up", "k":
			if m.sel > 0 {
				m.sel--
			}
		case "down", "j":
			if m.sel < len(m.items)-1 {
				m.sel++
			}
		case "enter", " ":
			item := m.items[m.sel]
			if item.kind == "bool" {
				val := !item.val.(bool)
				switch item.key {
				case "vim_mode":
					m.cfg.App.VimMode = val
				case "show_help":
					m.cfg.App.ShowHelp = val
				case "rainbow":
					m.cfg.App.Rainbow = val
				case "sync_enabled":
					m.cfg.Sync.Enabled = val
				case "auto_push":
					m.cfg.Sync.AutoPush = val
				}
				m.rebuildItems()
				// Save config immediately and notify app
				_ = m.cfg.Save()
				return m, func() tea.Msg { return ConfigChangedMsg{Config: m.cfg} }
			}
		}
	}
	return m, nil
}

func (m Model) View() string {
	w := m.width
	if w <= 0 {
		w = 80
	}
	cardW := min(60, w-4)

	title := m.styles.Title.Render("SETTINGS")
	var lines []string
	lines = append(lines, title, "")

	for i, item := range m.items {
		style := m.styles.RowNormal
		if i == m.sel {
			style = m.styles.RowSelected
		}

		status := ""
		if item.kind == "bool" {
			if item.val.(bool) {
				status = m.styles.BadgeGood.Render(" ON  ")
			} else {
				status = m.styles.BadgeMuted.Render(" OFF ")
			}
		}

		label := item.label
		padding := cardW - lipgloss.Width(label) - lipgloss.Width(status) - 4
		if padding < 0 {
			padding = 0
		}

		line := style.Render(fmt.Sprintf(" %s%s%s ", label, strings.Repeat(" ", padding), status))
		lines = append(lines, line)
	}

	hint := m.styles.Muted.Render("\n Tip: You can edit 'config.toml' for advanced options.")
	footer := m.styles.Muted.Render(" esc/ctrl+s close • 'g' open config • enter toggle")
	lines = append(lines, hint, footer)

	return lipgloss.Place(w, m.height, lipgloss.Center, lipgloss.Center,
		m.styles.Overlay.Width(cardW).Padding(1, 2).Render(lipgloss.JoinVertical(lipgloss.Left, lines...)),
		lipgloss.WithWhitespaceBackground(m.styles.Theme.Bg),
	)
}
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
