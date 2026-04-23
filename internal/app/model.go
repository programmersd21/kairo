package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/programmersd21/kairo/internal/buildinfo"
	"github.com/programmersd21/kairo/internal/config"
	"github.com/programmersd21/kairo/internal/core"
	"github.com/programmersd21/kairo/internal/plugins"
	"github.com/programmersd21/kairo/internal/search"
	"github.com/programmersd21/kairo/internal/service"
	ksync "github.com/programmersd21/kairo/internal/sync"
	"github.com/programmersd21/kairo/internal/ui/detail"
	"github.com/programmersd21/kairo/internal/ui/editor"
	"github.com/programmersd21/kairo/internal/ui/help"
	"github.com/programmersd21/kairo/internal/ui/keymap"
	"github.com/programmersd21/kairo/internal/ui/palette"
	"github.com/programmersd21/kairo/internal/ui/plugin_menu"
	"github.com/programmersd21/kairo/internal/ui/render"
	"github.com/programmersd21/kairo/internal/ui/styles"
	"github.com/programmersd21/kairo/internal/ui/tasklist"
	"github.com/programmersd21/kairo/internal/ui/theme"
	"github.com/programmersd21/kairo/internal/ui/theme_menu"
	"github.com/programmersd21/kairo/internal/updater"
	"github.com/programmersd21/kairo/internal/util"
)

// FilterState manages the tag filter with explicit lifecycle:
// - active: whether a filter is currently applied
// - value: the tag being filtered (only meaningful when active)
// This replaces the previous plain `tagParam` string to provide clear state management
// and enable proper reset/clear functionality.
type FilterState struct {
	active bool
	value  string
}

// Set activates the filter with a specific tag value
func (f *FilterState) Set(tag string) {
	f.active = true
	f.value = strings.TrimSpace(tag)
}

// Clear deactivates the filter and resets the value
func (f *FilterState) Clear() {
	f.active = false
	f.value = ""
}

// IsActive returns whether a filter is currently applied
func (f *FilterState) IsActive() bool {
	return f.active
}

// Value returns the current filter value (empty if not active)
func (f *FilterState) Value() string {
	if !f.active {
		return ""
	}
	return f.value
}

type Mode int

const (
	ModeList Mode = iota
	ModeDetail
	ModeEditor
	ModePalette
	ModeConfirmDelete
	ModeHelp
	ModeThemeMenu
	ModePluginMenu
	ModeTagFilter
	ModeConfirmQuit
)

type Model struct {
	ctx context.Context

	cfg   config.Config
	svc   service.TaskService
	km    keymap.Keymap
	theme theme.Theme
	s     styles.Styles

	width  int
	height int

	mode Mode

	views     []core.View
	activeIdx int
	tagFilter FilterState // Replaced plain tagParam with proper state management
	priParam  *core.Priority

	list tasklist.Model
	pal  palette.Model
	det  detail.Model
	edit *editor.Model
	hlp  help.Model
	tm   theme_menu.Model
	pm   plugin_menu.Model

	tagFilterInput textinput.Model // Input field for direct tag filtering in Tag View

	palFullIdx   *search.Index
	palTasksIdx  *search.Index
	palTasksOnly bool

	tasks []core.Task
	all   []core.Task
	tags  []string

	statusText string
	isErr      bool

	updateAvailable *updateAvailableMsg

	syncEngine *ksync.Engine

	plugHost *plugins.Host
	plugCh   chan struct{}

	RainbowAnimationOffset int
	animatingTaskID        string
	animationStarted       time.Time
	animationDuration      time.Duration
	animationReverse       bool
}

func (m *Model) rainbowTickCmd() tea.Cmd {
	return tea.Tick(150*time.Millisecond, func(time.Time) tea.Msg {
		return rainbowTickMsg{}
	})
}

func New(ctx context.Context, cfg config.Config, svc service.TaskService) (tea.Model, error) {
	th := applyThemeOverride(theme.ByName(cfg.App.Theme), cfg.Theme)
	s := styles.New(th)
	km := keymap.FromConfig(cfg.Keymap)

	// Initialize tag filter input
	tagInput := textinput.New()
	tagInput.Prompt = "#"
	tagInput.Placeholder = "Enter tag to filter…"
	tagInput.CharLimit = 64
	tagInput.Width = 40

	m := &Model{
		ctx:                    ctx,
		cfg:                    cfg,
		svc:                    svc,
		km:                     km,
		theme:                  th,
		s:                      s,
		mode:                   ModeList,
		tagFilterInput:         tagInput,
		RainbowAnimationOffset: 0,
	}
	m.list = tasklist.New(m.s, cfg.App.VimMode, m.km)
	m.pal = palette.New(m.s)
	m.det = detail.New(m.s)
	m.hlp = help.New(m.s, m.km)
	m.tm = theme_menu.New(m.s)
	m.pm = plugin_menu.New(m.s)

	// Sync.
	if cfg.Sync.Enabled && strings.TrimSpace(cfg.Sync.RepoPath) != "" {
		m.syncEngine = ksync.New(svc.Repo(), cfg.Sync.RepoPath, cfg.Sync.Remote, cfg.Sync.Branch, ksync.Strategy(cfg.Sync.Strategy), cfg.Sync.AutoPush)
	}

	// Plugins.
	if cfg.Plugins.Enabled {
		dir := strings.TrimSpace(cfg.Plugins.Dir)
		if dir == "" {
			stateDir, err := util.AppStateDir("kairo")
			if err == nil {
				dir = filepath.Join(stateDir, "plugins")
			}
		}
		if dir != "" {
			_ = os.MkdirAll(dir, 0o755)
			m.plugHost = plugins.New(svc, dir)
			m.plugHost.SetNotifyFunc(func(msg string, isErr bool) {
				m.statusText = msg
				m.isErr = isErr
			})
			_ = m.plugHost.LoadAll()
			m.plugCh = make(chan struct{}, 8)
			_ = m.plugHost.Watch(ctx, func() {
				select {
				case m.plugCh <- struct{}{}:
				default:
				}
			})
		}
	}

	m.rebuildViews()
	m.activeIdx = 0
	return m, nil
}

func applyThemeOverride(t theme.Theme, o config.ThemeConfig) theme.Theme {
	set := func(cur lipgloss.Color, v string) lipgloss.Color {
		v = strings.TrimSpace(v)
		if v == "" {
			return cur
		}
		return lipgloss.Color(v)
	}
	t.Bg = set(t.Bg, o.Bg)
	t.Fg = set(t.Fg, o.Fg)
	t.Muted = set(t.Muted, o.Muted)
	t.Border = set(t.Border, o.Border)
	t.Accent = set(t.Accent, o.Accent)
	t.Good = set(t.Good, o.Good)
	t.Warn = set(t.Warn, o.Warn)
	t.Bad = set(t.Bad, o.Bad)
	t.Overlay = set(t.Overlay, o.Overlay)
	return t
}

func (m *Model) Init() tea.Cmd {
	cmds := []tea.Cmd{m.loadTagsCmd(), m.loadTasksCmd(), m.loadAllTasksCmd(), m.checkUpdateCmd()}
	if m.cfg.App.Rainbow {
		cmds = append(cmds, m.rainbowTickCmd())
	}
	if m.plugCh != nil {
		cmds = append(cmds, m.listenPluginsCmd())
	}
	return tea.Batch(cmds...)
}

// isInputFocused returns true if the current mode has an active input field where
// the user is typing. This is used to prevent global keybindings from firing while
// input is focused, ensuring proper focus management and event routing.
func (m *Model) isInputFocused() bool {
	switch m.mode {
	case ModeEditor:
		// Editor has multiple text input fields that accept user input
		return true
	case ModePalette:
		// Palette has a search input field that's always focused when active
		return true
	case ModeTagFilter:
		// Tag filter input field is active when filtering by tag
		return true
	default:
		// All other modes don't have active text input fields, so keybindings can safely fire
		return false
	}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch x := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = x.Width, x.Height
		m.rebuildComponentSizes()
		return m, nil

	case errMsg:
		m.statusText = x.Err.Error()
		m.isErr = true
		return m, nil

	case tasksLoadedMsg:
		m.tasks = x.Tasks
		m.list.SetTasks(m.tasks)
		m.rebuildPaletteIndex()
		return m, nil

	case tagsLoadedMsg:
		m.tags = x.Tags
		m.rebuildPaletteIndex()
		return m, nil

	case allTasksLoadedMsg:
		m.all = x.Tasks
		m.rebuildPaletteIndex()
		return m, nil

	case palette.CloseMsg:
		if m.mode == ModePalette {
			m.mode = ModeList
		}
		return m, nil

	case help.CloseMsg:
		if m.mode == ModeHelp {
			m.mode = ModeList
		}
		return m, nil

	case theme_menu.CloseMsg:
		if m.mode == ModeThemeMenu {
			m.mode = ModeList
		}
		return m, nil

	case theme_menu.SelectMsg:
		m.theme = x.Theme
		m.cfg.App.Theme = x.Theme.Name
		_ = m.cfg.Save()
		m.refreshStyles()
		m.mode = ModeList
		return m, nil

	case plugin_menu.CloseMsg:
		if m.mode == ModePluginMenu {
			m.mode = ModeList
		}
		return m, nil

	case plugin_menu.UninstallMsg:
		if m.plugHost != nil {
			err := m.plugHost.DeletePlugin(x.ID)
			if err != nil {
				m.statusText = err.Error()
				m.isErr = true
			} else {
				m.statusText = "Plugin uninstalled"
				m.isErr = false
				m.pm.SetPlugins(m.plugHost.Plugins())
			}
			m.rebuildViews()
			m.rebuildPaletteIndex()
		}
		if m.mode == ModePluginMenu {
			// Stay in plugin menu, refresh handled by SetPlugins above
		} else {
			m.mode = ModeList
		}
		return m, nil

	case plugin_menu.OpenFolderMsg:
		if m.plugHost != nil {
			dir := m.cfg.Plugins.Dir
			if dir == "" {
				stateDir, _ := util.AppStateDir("kairo")
				dir = filepath.Join(stateDir, "plugins")
			}
			return m, openFolderCmd(dir)
		}
		return m, nil

	case plugin_menu.ReloadMsg:
		if m.plugHost != nil {
			_ = m.plugHost.LoadAll()
			m.pm.SetPlugins(m.plugHost.Plugins())
			m.rebuildViews()
			m.rebuildPaletteIndex()
			m.statusText = "Plugins reloaded"
			m.isErr = false
		}
		return m, nil

	case palette.SelectMsg:
		if m.mode != ModePalette {
			return m, nil
		}
		m.mode = ModeList
		switch x.Item.Kind {
		case search.KindTask:
			return m, m.fetchOpenTaskCmd(x.Item.ID)
		case search.KindTag:
			m.tagFilter.Set(x.Item.ID)
			m.setActiveView(core.ViewTag)
			m.rebuildComponentSizes() // Recalculate layout when filter changes
			return m, m.loadTasksCmd()
		case search.KindCommand:
			return m, m.runCommand(x.Item.ID)
		}
		return m, nil

	case editor.CloseMsg:
		if m.mode == ModeEditor {
			m.edit = nil
			m.mode = ModeList
		}
		return m, nil

	case editor.SaveNewMsg:
		return m, tea.Batch(m.createTaskCmd(x.Task), func() tea.Msg { return editor.CloseMsg{} })

	case editor.SavePatchMsg:
		return m, tea.Batch(m.updateTaskCmd(x.ID, x.Patch), func() tea.Msg { return editor.CloseMsg{} })

	case taskCreatedMsg:
		return m, tea.Batch(m.loadTagsCmd(), m.loadTasksCmd(), m.loadAllTasksCmd(), m.syncIfEnabledCmd())

	case taskUpdatedMsg:
		return m, tea.Batch(m.loadTagsCmd(), m.loadTasksCmd(), m.loadAllTasksCmd(), m.syncIfEnabledCmd())

	case taskDeletedMsg:
		return m, tea.Batch(m.loadTagsCmd(), m.loadTasksCmd(), m.loadAllTasksCmd(), m.syncIfEnabledCmd())

	case rainbowTickMsg:
		// Linear rainbow animation: increment offset each tick
		m.RainbowAnimationOffset = (m.RainbowAnimationOffset + 1) % 7 // 7 colors in rainbow
		return m, m.rainbowTickCmd()

	case strikeAnimationTickMsg:
		if m.animatingTaskID != x.TaskID {
			return m, nil
		}
		elapsed := time.Since(m.animationStarted)
		if elapsed >= m.animationDuration {
			// Animation complete, update the task
			m.animatingTaskID = ""
			var taskToUpdate core.Task
			for _, t := range m.all {
				if t.ID == x.TaskID {
					taskToUpdate = t
					break
				}
			}
			newStatus := core.StatusDone
			if taskToUpdate.Status == core.StatusDone {
				newStatus = core.StatusTodo
			}
			patch := core.TaskPatch{Status: &newStatus}
			return m, m.updateTaskCmd(x.TaskID, patch)
		}
		// Continue animation
		return m, m.strikeAnimationTickCmd(x.TaskID)

	case openTaskMsg:
		m.det.SetTask(x.Task)
		m.mode = ModeDetail
		return m, nil

	case openEditMsg:
		e := editor.New(m.s, editor.ModeEdit, x.Task)
		m.edit = &e
		m.rebuildComponentSizes()
		m.mode = ModeEditor
		return m, m.edit.Init()

	case pluginChangedMsg:
		if m.plugHost != nil {
			_ = m.plugHost.LoadAll()
			m.rebuildViews()
			m.rebuildPaletteIndex()
			return m, m.listenPluginsCmd()
		}
		return m, nil

	case syncDoneMsg:
		if x.Err != nil {
			m.statusText = x.Err.Error()
			m.isErr = true
		}
		return m, nil

	case updateAvailableMsg:
		m.updateAvailable = &x
		return m, nil
	}

	if km, ok := msg.(tea.KeyMsg); ok {
		if m.mode == ModeConfirmDelete {
			switch km.String() {
			case "y", "enter":
				if t, ok := m.list.Selected(); ok {
					m.mode = ModeList
					return m, m.deleteTaskCmd(t.ID)
				}
			case "n", "esc":
				m.mode = ModeList
				return m, nil
			}
		}

		if m.mode == ModeConfirmQuit {
			switch km.String() {
			case "y", "enter":
				return m, tea.Quit
			case "n", "esc":
				m.mode = ModeList
				return m, nil
			}
		}

		if m.mode == ModeTagFilter {
			// Handle critical global keys even in input mode
			if keymapMatch(m.km.Quit, km) {
				// Ask for confirmation even from tag filter
				m.tagFilterInput.Blur()
				m.mode = ModeConfirmQuit
				return m, nil
			}

			switch km.String() {
			case "enter":
				// Apply the filter and return to list view
				tagValue := strings.TrimSpace(m.tagFilterInput.Value())
				m.tagFilterInput.Blur()
				if tagValue != "" {
					m.tagFilter.Set(tagValue)
					m.setActiveView(core.ViewTag)
					m.rebuildComponentSizes() // Recalculate layout when filter changes
					m.mode = ModeList
					return m, m.loadTasksCmd()
				}
				// Empty input: just close without applying
				m.mode = ModeList
				return m, nil
			case "esc":
				// Cancel and clear input
				m.tagFilterInput.Blur()
				m.tagFilterInput.SetValue("")
				m.mode = ModeList
				return m, nil
			case "ctrl+u":
				// Clear the entire input
				m.tagFilterInput.SetValue("")
				return m, nil
			}
		}

		// Global key handling - only process keybindings if no input field is focused.
		// This ensures that text input has exclusive focus and keybindings are disabled
		// while typing in menus or editors.
		if !m.isInputFocused() {
			if keymapMatch(m.km.Quit, km) {
				m.mode = ModeConfirmQuit
				return m, nil
			}
			if keymapMatch(m.km.Palette, km) {
				m.palTasksOnly = false
				m.applyPaletteIndex()
				m.pal.SetPlaceholder("Search tasks, commands, tags…")
				m.mode = ModePalette
				return m, m.pal.Open()
			}
			if keymapMatch(m.km.TaskSearch, km) && (m.mode == ModeList || m.mode == ModeDetail) {
				m.palTasksOnly = true
				m.applyPaletteIndex()
				m.pal.SetPlaceholder("Search tasks…")
				m.mode = ModePalette
				return m, m.pal.Open()
			}
			if keymapMatch(m.km.CycleTheme, km) {
				m.mode = ModeThemeMenu
				return m, nil
			}
			if keymapMatch(m.km.ManagePlugins, km) {
				if m.mode == ModePluginMenu {
					m.mode = ModeList
					return m, nil
				}
				if m.plugHost != nil {
					m.pm.SetPlugins(m.plugHost.Plugins())
					m.mode = ModePluginMenu
				}
				return m, nil
			}
			if keymapMatch(m.km.OpenPluginDir, km) {
				if m.plugHost != nil {
					dir := m.cfg.Plugins.Dir
					if dir == "" {
						stateDir, err := util.AppStateDir("kairo")
						if err == nil {
							dir = filepath.Join(stateDir, "plugins")
						}
					}
					if dir != "" {
						return m, openFolderCmd(dir)
					}
				}
				return m, nil
			}
			if keymapMatch(m.km.Help, km) {
				m.mode = ModeHelp
				return m, nil
			}
			if keymapMatch(m.km.Issues, km) {
				return m, openURLCmd("https://github.com/programmersd21/kairo/issues")
			}
			if keymapMatch(m.km.Changelog, km) {
				return m, openURLCmd("https://github.com/programmersd21/kairo/blob/main/CHANGELOG.md")
			}

			// Plugin reload - single character keybinding only valid in ModeList
			if km.String() == "g" && m.mode == ModeList {
				if m.plugHost != nil {
					_ = m.plugHost.LoadAll()
					m.rebuildViews()
					m.rebuildPaletteIndex()
					m.statusText = "Plugins reloaded"
					m.isErr = false
					return m, m.loadTasksCmd()
				}
			}

			if m.mode == ModeList {
				// Dynamic view switching (1-9)
				if len(km.String()) == 1 && km.String() >= "1" && km.String() <= "9" {
					digit := int(km.String()[0] - '0')
					idx := digit - 1
					if idx >= 0 && idx < len(m.views) {
						m.activeIdx = idx
						m.tagFilter.Clear()
						m.rebuildComponentSizes()
						return m, m.loadTasksCmd()
					}
				}

				switch {
				case km.String() == "f":
					m.setActiveView(core.ViewTag)
					m.tagFilterInput.SetValue(m.tagFilter.Value())
					m.tagFilterInput.Focus()
					m.mode = ModeTagFilter
					return m, m.loadTasksCmd()
				case km.String() == "tab":
					m.activeIdx = (m.activeIdx + 1) % len(m.views)
					return m, m.loadTasksCmd()
				case km.String() == "shift+tab":
					m.activeIdx--
					if m.activeIdx < 0 {
						m.activeIdx = len(m.views) - 1
					}
					return m, m.loadTasksCmd()
				case keymapMatch(m.km.NewTask, km):
					task := core.Task{Status: core.StatusTodo, Priority: core.P1}
					m.activeFilter().ApplyToTask(&task)
					e := editor.New(m.s, editor.ModeNew, task)
					m.edit = &e
					m.rebuildComponentSizes()
					m.mode = ModeEditor
					return m, m.edit.Init()
				case keymapMatch(m.km.EditTask, km):
					if t, ok := m.list.Selected(); ok {
						return m, m.fetchOpenEditCmd(t.ID)
					}
				case keymapMatch(m.km.DeleteTask, km):
					if _, ok := m.list.Selected(); ok {
						m.mode = ModeConfirmDelete
						return m, nil
					}
				case keymapMatch(m.km.OpenTask, km):
					if t, ok := m.list.Selected(); ok {
						return m, m.fetchOpenTaskCmd(t.ID)
					}
				case keymapMatch(m.km.ToggleStrike, km):
					if t, ok := m.list.Selected(); ok {
						m.animatingTaskID = t.ID
						m.animationStarted = time.Now()
						m.animationDuration = 400 * time.Millisecond
						m.animationReverse = (t.Status == core.StatusDone)
						return m, m.strikeAnimationTickCmd(t.ID)
					}
				}
			}

			if m.mode == ModeDetail {
				if keymapMatch(m.km.Back, km) {
					m.mode = ModeList
					return m, nil
				}
				if keymapMatch(m.km.EditTask, km) {
					return m, m.fetchOpenEditCmd(m.det.Task().ID)
				}
				if keymapMatch(m.km.ToggleStrike, km) {
					t := m.det.Task()
					m.animatingTaskID = t.ID
					m.animationStarted = time.Now()
					m.animationDuration = 400 * time.Millisecond
					m.animationReverse = (t.Status == core.StatusDone)
					return m, m.strikeAnimationTickCmd(t.ID)
				}
			}
		}
	}

	// Delegate to active component.
	switch m.mode {
	case ModeTagFilter:
		// Handle text input for tag filtering
		var cmd tea.Cmd
		m.tagFilterInput, cmd = m.tagFilterInput.Update(msg)
		return m, cmd
	case ModeList, ModeConfirmDelete:
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	case ModePalette:
		var cmd tea.Cmd
		m.pal, cmd = m.pal.Update(msg)
		return m, cmd
	case ModeEditor:
		if m.edit == nil {
			m.mode = ModeList
			return m, nil
		}
		e, cmd := m.edit.Update(msg)
		*m.edit = e
		return m, cmd
	case ModeDetail:
		var cmd tea.Cmd
		return m, cmd
	case ModeHelp:
		var cmd tea.Cmd
		m.hlp, cmd = m.hlp.Update(msg)
		return m, cmd
	case ModeThemeMenu:
		var cmd tea.Cmd
		m.tm, cmd = m.tm.Update(msg)
		return m, cmd
	case ModePluginMenu:
		var cmd tea.Cmd
		m.pm, cmd = m.pm.Update(msg)
		return m, cmd
	default:
		return m, nil
	}
}

func (m *Model) View() string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}

	var content string

	switch m.mode {
	case ModePalette:
		content = m.pal.View()
	case ModeHelp:
		content = m.hlp.View()
	case ModeThemeMenu:
		content = m.tm.View()
	default:
		content = m.renderMainUI()
	}

	// Final rendering pipeline: FillViewport guarantees that every cell in
	// the width×height viewport has the background color applied.
	// It pads lines, fills missing rows, and—critically—re-applies the
	// background ANSI sequence after every SGR reset (\x1b[0m), which is
	// the root cause of terminal default background bleeding through.
	return render.FillViewport(content, m.width, m.height, m.s.Theme.Bg)
}

func (m *Model) renderMainUI() string {
	head := m.renderHeader()
	foot := m.renderFooter()

	hHeight := lipgloss.Height(head)
	fHeight := lipgloss.Height(foot)
	availableHeight := m.height - hHeight - fHeight
	if availableHeight < 0 {
		availableHeight = 0
	}

	// Sync animation state to tasklist
	if m.animatingTaskID != "" {
		m.list.SetAnimation(m.animatingTaskID, m.animationStarted, m.animationDuration, m.animationReverse)
	}

	var body string
	switch m.mode {
	case ModeList, ModeConfirmDelete:
		body = m.list.View()
	case ModeDetail:
		body = m.det.View()
	case ModePluginMenu:
		body = m.pm.View()
	case ModeTagFilter:
		body = m.renderTagFilterOverlay(availableHeight)
	case ModeEditor:
		if m.edit != nil {
			body = m.edit.View()
		} else {
			body = m.list.View()
		}
	default:
		body = m.list.View()
	}

	// Ensure body fills its allocated height.
	// The outer FillViewport handles width filling and ANSI reset fixup.
	body = lipgloss.NewStyle().
		Height(availableHeight).
		Width(m.width).
		Background(m.s.Theme.Bg).
		Render(body)

	return lipgloss.JoinVertical(lipgloss.Left, head, body, foot)
}

func (m *Model) rebuildComponentSizes() {
	// Calculate header height dynamically
	head := m.renderHeader()
	foot := m.renderFooter()
	hHeight := lipgloss.Height(head)
	fHeight := lipgloss.Height(foot)

	avail := m.height - hHeight - fHeight
	if avail < 0 {
		avail = 0
	}

	m.list.SetSize(m.width, avail)
	m.det.SetSize(m.width, avail)
	m.pal.SetSize(m.width, m.height)
	m.pm.SetSize(m.width, m.height)
	m.hlp.SetSize(m.width, m.height)
	m.tm.SetSize(m.width, m.height)
	if m.edit != nil {
		m.edit.SetSize(m.width, avail)
	}
}

// Add to Model struct:
// RainbowAnimationOffset int
// And inside New():
// m.RainbowAnimationOffset = 0

// Updated RenderHeader:
func (m *Model) renderHeader() string {
	// Logo with themed background container
	logoText := "KAIRO"
	var logo string
	if m.cfg.App.Rainbow {
		rainbowColors := []string{"#ff0000", "#ff7f00", "#ffff00", "#00ff00", "#0000ff", "#4b0082", "#9400d3"}
		var logoBuilder strings.Builder
		for i, char := range logoText {
			color := rainbowColors[(i+m.RainbowAnimationOffset)%len(rainbowColors)]
			logoBuilder.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(string(char)))
		}
		logo = lipgloss.NewStyle().Bold(true).Render(logoBuilder.String())
	} else {
		logo = lipgloss.NewStyle().Foreground(m.s.Theme.Accent).Bold(true).Render(logoText)
	}

	// Tabs
	tabs := []string{}
	for i, v := range m.views {
		style := m.s.TabInactive
		if i == m.activeIdx {
			style = m.s.TabActive
		}
		tabs = append(tabs, style.Render(v.Title))
	}
	tabRow := lipgloss.JoinHorizontal(lipgloss.Left, tabs...)

	sep := lipgloss.NewStyle().Background(m.s.Theme.Bg).Render("  ")
	topBarLeft := lipgloss.JoinHorizontal(lipgloss.Left, logo, sep, tabRow)

	count := fmt.Sprintf("%d tasks", len(m.tasks))
	topBarRight := m.s.Muted.Render(count)

	topBar := render.BarLine(topBarLeft, topBarRight, m.width, m.s.Theme.Bg)

	// Detail line (tag/priority info)
	detailLine := ""
	v := m.views[m.activeIdx]
	if v.ID == core.ViewTag && m.tagFilter.IsActive() {
		detailLine = "  " + m.s.Muted.Render(styles.IconTag) + m.s.Title.Render(m.tagFilter.Value()) + "  " + m.s.Muted.Render("[press f to edit]")
	} else if v.ID == core.ViewPriority && m.priParam != nil {
		detailLine = "  " + m.s.Muted.Render("Priority: ") + m.s.PriorityBadge(*m.priParam)
	}

	header := topBar
	if detailLine != "" {
		detailLine = render.BarLine(detailLine, "", m.width, m.s.Theme.Bg)
		header = lipgloss.JoinVertical(lipgloss.Left, topBar, detailLine)
	}

	return lipgloss.NewStyle().
		Width(m.width).
		Background(m.s.Theme.Bg).
		BorderBottom(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(m.s.Theme.Border).
		BorderBackground(m.s.Theme.Bg).
		Render(header)
}

// firstKey returns the first configured key for a binding, or "?" if unset.
func firstKey(b key.Binding) string {
	keys := b.Keys()
	if len(keys) == 0 {
		return "?"
	}
	return keys[0]
}

func (m *Model) renderFooter() string {
	fk := firstKey
	left := ""
	switch m.mode {
	case ModeConfirmDelete:
		left = m.s.BadgeDelete.Render(" DELETE? ") + " " + m.s.Muted.Render("y/enter confirm • n/esc cancel")
	case ModeConfirmQuit:
		left = m.s.BadgeQuit.Render(" QUIT? ") + " " + m.s.Muted.Render("y/enter confirm • n/esc cancel")
	case ModeTagFilter:
		left = " " + m.s.Muted.Render("enter apply • esc cancel • ctrl+u clear")
	case ModeDetail:
		left = " " + m.s.Muted.Render(
			fk(m.km.Back)+" 󰌍back • "+
				fk(m.km.EditTask)+" 󰏫edit • "+
				fk(m.km.Palette)+" "+styles.IconPalette+"palette • "+
				fk(m.km.Help)+" "+styles.IconHelp+"help • "+
				fk(m.km.Issues)+" "+styles.IconIssues+"issues • "+
				fk(m.km.Changelog)+" "+styles.IconChangelog+"changelog",
		)
	case ModeEditor:
		left = " " + m.s.Muted.Render("ctrl+s "+styles.IconDone+"save • esc "+styles.IconError+"cancel • tab navigate")
	case ModePalette:
		left = " " + m.s.Muted.Render("enter select • esc/p close • ↑/↓ navigate")
	case ModeHelp:
		left = " " + m.s.Muted.Render("esc/q/"+fk(m.km.Help)+" close")
	case ModeThemeMenu:
		left = " " + m.s.Muted.Render("enter select • esc/q/"+fk(m.km.CycleTheme)+" close • ↑/↓ navigate")
	case ModePluginMenu:
		left = " " + m.s.Muted.Render("enter detail • u uninstall • o open folder • r reload • p/"+fk(m.km.ManagePlugins)+" close • ↑/↓ navigate")
	default:
		left = " " + m.s.Muted.Render(
			fk(m.km.Palette)+" "+styles.IconPalette+"palette • "+
				fk(m.km.NewTask)+" "+styles.IconNew+"new • "+
				"f "+styles.IconTag+"tag • "+
				fk(m.km.ToggleStrike)+" "+styles.IconStrike+"strike • "+
				fk(m.km.DeleteTask)+" "+styles.IconDelete+"delete • "+
				fk(m.km.Help)+" "+styles.IconHelp+"help • "+
				fk(m.km.Issues)+" "+styles.IconIssues+"issues • "+
				fk(m.km.Changelog)+" "+styles.IconChangelog+"changelog • "+
				"1-9 "+styles.IconView+"view",
		)
	}

	right := ""
	if m.statusText != "" {
		icon := styles.IconInfo
		if m.isErr {
			icon = styles.IconError
			right = m.s.Muted.Foreground(m.s.Theme.Bad).Bold(true).Render(icon+" ") + m.s.Muted.Render(m.statusText+" ")
		} else {
			right = m.s.Muted.Foreground(m.s.Theme.Good).Bold(true).Render(icon+" ") + m.s.Muted.Render(m.statusText+" ")
		}
	} else {
		syncStatus := ""
		if m.syncEngine != nil && m.syncEngine.Enabled() {
			syncStatus = styles.IconSync + " "
		}

		versionText := buildinfo.VersionTag()
		if m.updateAvailable != nil {
			cur := m.updateAvailable.Current
			if !strings.HasPrefix(cur, "v") {
				cur = "v" + cur
			}
			lat := m.updateAvailable.Latest
			if !strings.HasPrefix(lat, "v") {
				lat = "v" + lat
			}
			versionText = fmt.Sprintf("Update: %s → %s (run `kairo update`)", cur, lat)
			right = m.s.Muted.Foreground(m.s.Theme.Accent).Bold(true).Render(syncStatus + versionText + " ")
		} else {
			right = m.s.Muted.Render(syncStatus + versionText + " ")
		}
	}

	line := render.BarLine(left, right, m.width, m.s.Theme.Bg)
	return lipgloss.NewStyle().
		Width(m.width).
		Background(m.s.Theme.Bg).
		BorderTop(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(m.s.Theme.Border).
		BorderBackground(m.s.Theme.Bg).
		Render(line)
}

// renderTagFilterOverlay renders the tag filter input modal
func (m *Model) renderTagFilterOverlay(h int) string {
	// Create filter input modal
	inputLabel := m.s.Title.Render("Filter by Tag")
	input := lipgloss.NewStyle().Padding(0, 1).Render(m.tagFilterInput.View())
	hint := m.s.Muted.Render("Current tags: " + strings.Join(m.tags, ", "))
	if len(m.tags) > 10 {
		hint = m.s.Muted.Render("(Showing available tags)")
	}

	modal := lipgloss.JoinVertical(lipgloss.Left,
		inputLabel,
		input,
		hint,
	)

	cardStyle := m.s.Overlay.Width(60)
	card := cardStyle.Render(modal)

	// Overlay the modal on top of the screen with proper background
	return lipgloss.Place(m.width, h, lipgloss.Center, lipgloss.Center, card,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceBackground(m.s.Theme.Bg),
	)
}

func (m *Model) setActiveView(id core.ViewID) {
	for i, v := range m.views {
		if v.ID == id {
			m.activeIdx = i
			return
		}
	}
}

func (m *Model) rebuildViews() {
	base := core.DefaultViews(time.Now())
	if m.plugHost != nil {
		for _, v := range m.plugHost.Views() {
			base = append(base, core.View{ID: core.ViewID(v.ID), Title: v.Title, Filter: v.Filter})
		}
	}
	m.views = base
	if m.activeIdx >= len(m.views) {
		m.activeIdx = 0
	}
}

func (m *Model) activeFilter() core.Filter {
	v := m.views[m.activeIdx]
	f := v.Filter

	// Apply dynamic parameters if it's a built-in view that supports them
	if v.ID == core.ViewTag {
		f.Tag = m.tagFilter.Value() // Use the new FilterState
	}
	if v.ID == core.ViewPriority && m.priParam != nil {
		f.Priority = m.priParam
	}

	// If it's a plugin-defined view, the filter is already set in rebuildViews
	return f
}

func (m *Model) loadTasksCmd() tea.Cmd {
	f := m.activeFilter()
	return func() tea.Msg {
		ts, err := m.svc.List(m.ctx, f)
		if err != nil {
			return errMsg{Err: err}
		}
		return tasksLoadedMsg{Tasks: ts}
	}
}

func (m *Model) loadAllTasksCmd() tea.Cmd {
	return func() tea.Msg {
		ts, err := m.svc.ListAll(m.ctx)
		if err != nil {
			return errMsg{Err: err}
		}
		return allTasksLoadedMsg{Tasks: ts}
	}
}

func (m *Model) loadTagsCmd() tea.Cmd {
	return func() tea.Msg {
		tags, err := m.svc.ListTags(m.ctx)
		if err != nil {
			return errMsg{Err: err}
		}
		return tagsLoadedMsg{Tags: tags}
	}
}

func (m *Model) createTaskCmd(t core.Task) tea.Cmd {
	return func() tea.Msg {
		created, err := m.svc.Create(m.ctx, t)
		if err != nil {
			return errMsg{Err: err}
		}
		return taskCreatedMsg{Task: created}
	}
}

func (m *Model) updateTaskCmd(id string, p core.TaskPatch) tea.Cmd {
	return func() tea.Msg {
		updated, err := m.svc.Update(m.ctx, id, p)
		if err != nil {
			return errMsg{Err: err}
		}
		return taskUpdatedMsg{Task: updated}
	}
}

func (m *Model) deleteTaskCmd(id string) tea.Cmd {
	return func() tea.Msg {
		if err := m.svc.Delete(m.ctx, id); err != nil {
			return errMsg{Err: err}
		}
		return taskDeletedMsg{ID: id}
	}
}

func (m *Model) strikeAnimationTickCmd(taskID string) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(16 * time.Millisecond) // ~60 FPS
		return strikeAnimationTickMsg{TaskID: taskID}
	}
}

func (m *Model) fetchOpenTaskCmd(id string) tea.Cmd {
	return func() tea.Msg {
		t, err := m.svc.GetByID(m.ctx, id)
		if err != nil {
			return errMsg{Err: err}
		}
		return openTaskMsg{Task: t}
	}
}

func (m *Model) fetchOpenEditCmd(id string) tea.Cmd {
	return func() tea.Msg {
		t, err := m.svc.GetByID(m.ctx, id)
		if err != nil {
			return errMsg{Err: err}
		}
		return openEditMsg{Task: t}
	}
}

func (m *Model) refreshStyles() {
	m.s = styles.New(m.theme)
	m.list = tasklist.New(m.s, m.cfg.App.VimMode, m.km)
	m.list.SetTasks(m.tasks)
	m.pal = palette.New(m.s)
	m.det = detail.New(m.s)
	m.hlp = help.New(m.s, m.km)
	m.tm = theme_menu.New(m.s)
	m.pm = plugin_menu.New(m.s)

	// Reinitialize tag filter input with new styles
	tagInput := textinput.New()
	tagInput.Prompt = "#"
	tagInput.Placeholder = "Enter tag to filter…"
	tagInput.CharLimit = 64
	tagInput.Width = 40
	m.tagFilterInput = tagInput

	m.rebuildComponentSizes()
	m.rebuildPaletteIndex()
}

func (m *Model) rebuildPaletteIndex() {
	items := make([]search.Item, 0, len(m.all)+len(m.tags)+32)

	items = append(items,
		search.Item{ID: "cmd:new", Kind: search.KindCommand, Title: "New task", Hint: "Create a task"},
		search.Item{ID: "cmd:sync", Kind: search.KindCommand, Title: "Sync now", Hint: "Git pull/push"},
		search.Item{ID: "cmd:theme", Kind: search.KindCommand, Title: "Theme menu", Hint: "Switch theme"},
		search.Item{ID: "cmd:view:inbox", Kind: search.KindCommand, Title: "View: Inbox", Hint: "1"},
		search.Item{ID: "cmd:view:today", Kind: search.KindCommand, Title: "View: Today", Hint: "2"},
		search.Item{ID: "cmd:view:upcoming", Kind: search.KindCommand, Title: "View: Upcoming", Hint: "3"},
		search.Item{ID: "cmd:view:completed", Kind: search.KindCommand, Title: "View: Completed", Hint: "4"},
		search.Item{ID: "cmd:view:tag", Kind: search.KindCommand, Title: "View: By Tag", Hint: "f"},
		search.Item{ID: "cmd:view:priority", Kind: search.KindCommand, Title: "View: By Priority", Hint: "5"},
	)

	for _, t := range m.tags {
		items = append(items, search.Item{ID: t, Kind: search.KindTag, Title: "#" + t, Hint: "tag"})
	}

	for p := core.P0; p <= core.P3; p++ {
		items = append(items, search.Item{ID: fmt.Sprintf("pri:%d", int(p)), Kind: search.KindCommand, Title: fmt.Sprintf("Priority: P%d", int(p)), Hint: "set priority view"})
	}

	if m.plugHost != nil {
		for _, c := range m.plugHost.Commands() {
			items = append(items, search.Item{ID: c.ID, Kind: search.KindCommand, Title: c.Title, Hint: "plugin • " + c.PluginID})
		}
		for _, v := range m.plugHost.Views() {
			items = append(items, search.Item{ID: "cmd:view:" + v.ID, Kind: search.KindCommand, Title: "View: " + v.Title, Hint: "plugin • " + v.PluginID})
		}
	}

	for _, t := range m.all {
		hint := string(t.Status)
		if t.Deadline != nil {
			hint += " • due " + t.Deadline.Local().Format("Jan 2")
		}
		items = append(items, search.Item{ID: t.ID, Kind: search.KindTask, Title: t.Title, Desc: t.Description, Hint: hint})
	}

	m.palFullIdx = search.NewIndex(items)

	taskItems := make([]search.Item, 0, len(m.all))
	for _, t := range m.all {
		hint := string(t.Status)
		if t.Deadline != nil {
			hint += " • due " + t.Deadline.Local().Format("Jan 2")
		}
		taskItems = append(taskItems, search.Item{ID: t.ID, Kind: search.KindTask, Title: t.Title, Desc: t.Description, Hint: hint})
	}
	m.palTasksIdx = search.NewIndex(taskItems)
	m.applyPaletteIndex()
}

func (m *Model) applyPaletteIndex() {
	if m.palTasksOnly {
		if m.palTasksIdx != nil {
			m.pal.SetIndex(m.palTasksIdx)
		}
		return
	}
	if m.palFullIdx != nil {
		m.pal.SetIndex(m.palFullIdx)
	}
}
func (m *Model) runCommand(id string) tea.Cmd {
	switch id {
	case "cmd:new":
		e := editor.New(m.s, editor.ModeNew, core.Task{Status: core.StatusTodo, Priority: core.P1})
		m.edit = &e
		m.rebuildComponentSizes()
		m.mode = ModeEditor
		return m.edit.Init()
	case "cmd:theme":
		m.mode = ModeThemeMenu
		return nil
	case "cmd:view:inbox":
		m.setActiveView(core.ViewInbox)
		return m.loadTasksCmd()
	case "cmd:view:today":
		m.setActiveView(core.ViewToday)
		return m.loadTasksCmd()
	case "cmd:view:upcoming":
		m.setActiveView(core.ViewUpcoming)
		return m.loadTasksCmd()
	case "cmd:view:completed":
		m.setActiveView(core.ViewCompleted)
		return m.loadTasksCmd()
	case "cmd:view:tag":
		m.setActiveView(core.ViewTag)
		return m.loadTasksCmd()
	case "cmd:view:priority":
		m.setActiveView(core.ViewPriority)
		return m.loadTasksCmd()
	case "cmd:sync":
		return m.syncNowCmd()
	}

	if strings.HasPrefix(id, "cmd:view:plugin:") {
		viewID := strings.TrimPrefix(id, "cmd:view:")
		m.setActiveView(core.ViewID(viewID))
		return m.loadTasksCmd()
	}

	if strings.HasPrefix(id, "plugin:") && m.plugHost != nil {
		return func() tea.Msg {
			if err := m.plugHost.RunCommand(m.ctx, id); err != nil {
				return errMsg{Err: err}
			}
			return taskUpdatedMsg{Task: core.Task{}}
		}
	}

	if strings.HasPrefix(id, "pri:") {
		raw := strings.TrimPrefix(id, "pri:")
		var p int
		_, _ = fmt.Sscanf(raw, "%d", &p)
		pp := core.Priority(p).Clamp()
		m.priParam = &pp
		m.setActiveView(core.ViewPriority)
		return m.loadTasksCmd()
	}

	return func() tea.Msg { return errMsg{Err: errors.New("unknown command")} }
}

func (m *Model) listenPluginsCmd() tea.Cmd {
	return func() tea.Msg {
		<-m.plugCh
		return pluginChangedMsg{}
	}
}

func (m *Model) syncIfEnabledCmd() tea.Cmd {
	if m.syncEngine == nil || !m.syncEngine.Enabled() {
		return nil
	}
	return m.syncNowCmd()
}

func (m *Model) syncNowCmd() tea.Cmd {
	if m.syncEngine == nil || !m.syncEngine.Enabled() {
		return func() tea.Msg { return errMsg{Err: errors.New("sync not configured")} }
	}
	return func() tea.Msg {
		err := m.syncEngine.SyncNow(m.ctx)
		return syncDoneMsg{Err: err}
	}
}

func (m *Model) checkUpdateCmd() tea.Cmd {
	return func() tea.Msg {
		v := buildinfo.EffectiveVersion()
		if v == "dev" {
			return nil
		}
		cfg := updater.DefaultConfig()
		res, _, err := cfg.Check(m.ctx, v)
		if err != nil {
			return nil // Silently fail update check
		}
		if res.Update {
			return updateAvailableMsg{
				Current: res.Current,
				Latest:  res.Latest,
			}
		}
		return nil
	}
}

func openURLCmd(url string) tea.Cmd {
	return func() tea.Msg {
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "windows":
			// Start detached and non-blocking
			cmd = exec.Command("cmd", "/c", "start", url)
		case "darwin":
			cmd = exec.Command("open", url)
		case "linux":
			cmd = exec.Command("xdg-open", url)
		default:
			return errMsg{Err: fmt.Errorf("unsupported platform")}
		}
		// Run without waiting
		_ = cmd.Start()
		return nil
	}
}

func openFolderCmd(path string) tea.Cmd {
	return func() tea.Msg {
		var err error
		switch runtime.GOOS {
		case "windows":
			err = exec.Command("explorer", path).Start()
		case "darwin":
			err = exec.Command("open", path).Start()
		case "linux":
			err = exec.Command("xdg-open", path).Start()
		default:
			err = fmt.Errorf("unsupported platform")
		}
		if err != nil {
			return errMsg{Err: err}
		}
		return nil
	}
}

func keymapMatch(b interface{ Keys() []string }, k tea.KeyMsg) bool {
	kn := keymap.NormalizeChord(k.String())
	for _, kk := range b.Keys() {
		if keymap.NormalizeChord(kk) == kn {
			return true
		}
	}
	return false
}
