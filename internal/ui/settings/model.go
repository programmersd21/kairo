package settings

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
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

	editing bool
	input   textinput.Model
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
		{"Show Task IDs", "show_id", "bool", m.cfg.App.ShowID},
		{"Rainbow Logo", "rainbow", "bool", m.cfg.App.Rainbow},
		{"Git Sync Enabled", "sync_enabled", "bool", m.cfg.Sync.Enabled},
		{"Auto Push (Git)", "auto_push", "bool", m.cfg.Sync.AutoPush},
		{"MCP Server Enabled", "mcp_enabled", "bool", m.cfg.App.MCPEnabled},
		{"Animations", "animations", "bool", m.cfg.App.Animations},
		{"AI Model (←/→)", "ai_model", "enum", m.cfg.App.AIModel},
		{"Gemini API Key", "gemini_api_key", "string", m.cfg.App.GeminiAPIKey},
		{"AI Assistant Shortcut", "ai_toggle", "string", m.cfg.Keymap.AIPanelToggle},
	}
}

func (m *Model) SetSize(w, h int) {
	m.width, m.height = w, h
}

func (m *Model) SetConfig(cfg config.Config) {
	m.cfg = cfg
	m.rebuildItems()
}

func (m *Model) SetStyles(s styles.Styles) {
	m.styles = s
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if m.editing {
		switch x := msg.(type) {
		case tea.KeyMsg:
			switch x.String() {
			case "enter":
				item := m.items[m.sel]
				if item.key == "gemini_api_key" {
					m.cfg.App.GeminiAPIKey = m.input.Value()
				}
				if item.key == "ai_toggle" {
					m.cfg.Keymap.AIPanelToggle = m.input.Value()
				}
				m.editing = false
				m.rebuildItems()
				_ = m.cfg.Save()
				return m, func() tea.Msg { return ConfigChangedMsg{Config: m.cfg} }
			case "esc":
				m.editing = false
				return m, nil
			}
		}
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}

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
		case "r":
			m.cfg = config.Default()
			m.rebuildItems()
			_ = m.cfg.Save()
			return m, func() tea.Msg { return ConfigChangedMsg{Config: m.cfg} }
		case "up", "k":
			if m.sel > 0 {
				m.sel--
			}
		case "down", "j":
			if m.sel < len(m.items)-1 {
				m.sel++
			}
		case "left", "right", "h", "l":
			item := m.items[m.sel]
			if item.key == "ai_model" {
				models := []string{"gemini-3.1-flash-lite-preview", "gemini-2.0-flash-lite", "gemini-2.5-flash-lite"}
				curr := m.cfg.App.AIModel
				if curr == "" {
					curr = "gemini-3.1-flash-lite-preview"
				}
				idx := 0
				for i, mod := range models {
					if mod == curr {
						idx = i
						break
					}
				}
				if x.String() == "left" || x.String() == "h" {
					idx--
					if idx < 0 {
						idx = len(models) - 1
					}
				} else {
					idx = (idx + 1) % len(models)
				}
				m.cfg.App.AIModel = models[idx]
				m.rebuildItems()
				_ = m.cfg.Save()
				return m, func() tea.Msg { return ConfigChangedMsg{Config: m.cfg} }
			}
		case "enter", " ":
			item := m.items[m.sel]
			switch item.kind {
			case "bool":
				val := !item.val.(bool)
				switch item.key {
				case "vim_mode":
					m.cfg.App.VimMode = val
				case "show_help":
					m.cfg.App.ShowHelp = val
				case "show_id":
					m.cfg.App.ShowID = val
				case "rainbow":
					m.cfg.App.Rainbow = val
				case "sync_enabled":
					m.cfg.Sync.Enabled = val
				case "auto_push":
					m.cfg.Sync.AutoPush = val
				case "mcp_enabled":
					m.cfg.App.MCPEnabled = val
				case "animations":
					m.cfg.App.Animations = val
				}
				m.rebuildItems()
				// Save config immediately and notify app
				_ = m.cfg.Save()
				return m, func() tea.Msg { return ConfigChangedMsg{Config: m.cfg} }
			case "string":
				m.editing = true
				m.input = textinput.New()
				m.input.SetValue(item.val.(string))
				m.input.Focus()
				m.input.Width = 30
				return m, nil
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
		switch item.kind {
		case "bool":
			if item.val.(bool) {
				status = m.styles.BadgeGood.Render(" ON  ")
			} else {
				status = m.styles.BadgeMuted.Render(" OFF ")
			}
		case "string":
			if i == m.sel && m.editing {
				status = m.input.View()
			} else {
				v := item.val.(string)
				if v == "" {
					status = m.styles.Muted.Render("None")
				} else {
					if len(v) > 8 {
						status = m.styles.Muted.Render(v[:8] + "...")
					} else {
						status = m.styles.Muted.Render(v)
					}
				}
			}
		case "enum":
			v := item.val.(string)
			status = m.styles.BadgeWarn.Render(" " + v + " ")
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
	footer := m.styles.Muted.Render(" esc/ctrl+s close • 'g' open config • 'r' reset • enter toggle • 'j' move down • 'k' move up")
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
