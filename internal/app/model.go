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

	"github.com/fsnotify/fsnotify"
	"github.com/programmersd21/kairo/internal/ai"
	"github.com/programmersd21/kairo/internal/api"
	"github.com/programmersd21/kairo/internal/buildinfo"
	"github.com/programmersd21/kairo/internal/config"
	"github.com/programmersd21/kairo/internal/core"
	"github.com/programmersd21/kairo/internal/plugins"
	"github.com/programmersd21/kairo/internal/search"
	"github.com/programmersd21/kairo/internal/service"
	ksync "github.com/programmersd21/kairo/internal/sync"
	"github.com/programmersd21/kairo/internal/ui/ai_panel"
	"github.com/programmersd21/kairo/internal/ui/detail"
	"github.com/programmersd21/kairo/internal/ui/editor"
	"github.com/programmersd21/kairo/internal/ui/help"
	"github.com/programmersd21/kairo/internal/ui/import_export_menu"
	"github.com/programmersd21/kairo/internal/ui/keymap"
	"github.com/programmersd21/kairo/internal/ui/palette"
	"github.com/programmersd21/kairo/internal/ui/plugin_menu"
	"github.com/programmersd21/kairo/internal/ui/render"
	"github.com/programmersd21/kairo/internal/ui/settings"
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
	ModeSettings
	ModeImportExport
)

type Model struct {
	ctx context.Context

	cfg       config.Config
	svc       service.TaskService
	km        keymap.Keymap
	thBuiltin theme.Theme
	theme     theme.Theme
	s         styles.Styles

	width  int
	height int

	mode Mode

	views         []core.View
	activeIdx     int
	prevActiveIdx int
	tagFilter     FilterState // Replaced plain tagParam with proper state management
	priParam      *core.Priority

	list       tasklist.Model
	pal        palette.Model
	det        detail.Model
	edit       *editor.Model
	hlp        help.Model
	tm         theme_menu.Model
	pm         plugin_menu.Model
	set        settings.Model
	iem        import_export_menu.Model
	aiPanel    ai_panel.Model
	aiClient   *ai.Client
	aiKey      string
	aiChan     chan ai_panel.AIChunkMsg
	mcpCmd     *exec.Cmd
	mcpRunning bool

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
	configCh chan config.Config

	RainbowAnimationOffset int
	rainbowAnimating       bool
	animatingTaskID        string
	animationStarted       time.Time
	animationDuration      time.Duration
	animationReverse       bool

	creatingTaskID   string
	creationStarted  time.Time
	creationDuration time.Duration

	deletingTaskID string
	deleteStarted  time.Time
	deleteDuration time.Duration

	transitioning      bool
	transitionStarted  time.Time
	transitionProgress float64 // eased [0, 1] progress for view transitions

	// animationGen is incremented each time a new animation starts.
	// Tick messages carry the generation they were spawned under;
	// stale ticks (gen mismatch) are silently dropped in Update().
	animationGen int
}

func (m *Model) rainbowTickCmd() tea.Cmd {
	return tea.Tick(150*time.Millisecond, func(time.Time) tea.Msg {
		return rainbowTickMsg{}
	})
}

func (m *Model) cleanupTickCmd() tea.Cmd {
	return tea.Tick(1*time.Hour, func(time.Time) tea.Msg {
		return cleanupTickMsg{}
	})
}

func New(ctx context.Context, cfg config.Config, svc service.TaskService) (tea.Model, error) {
	thBuiltin := theme.FindBuiltin(cfg.App.Theme)
	th := applyThemeOverride(thBuiltin, cfg.Theme)
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
		thBuiltin:              thBuiltin,
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
	m.tm = theme_menu.New(m.s, nil)
	m.pm = plugin_menu.New(m.s)
	m.set = settings.New(m.s, cfg)
	m.iem = import_export_menu.New(m.s)
	m.aiPanel = ai_panel.New(m.s)
	m.aiChan = make(chan ai_panel.AIChunkMsg, 100)
	m.aiKey = cfg.App.GeminiAPIKey
	if m.aiKey != "" {
		m.aiClient, _ = ai.NewClient(ctx, m.aiKey, cfg.App.AIModel)
		ai.SetService(svc)
	}

	// Config watcher.
	m.configCh = make(chan config.Config, 8)
	cPath, err := config.ConfigPath()
	if err == nil {
		watcher, err := fsnotify.NewWatcher()
		if err == nil {
			_ = watcher.Add(filepath.Dir(cPath))
			go func() {
				for {
					select {
					case event, ok := <-watcher.Events:
						if !ok {
							return
						}
						// Watch for writes or renames (some editors save via rename) to the config file
						if (event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Rename)) && filepath.Base(event.Name) == "config.toml" {
							time.Sleep(100 * time.Millisecond) // Wait for write to stabilize
							newCfg, err := config.Load()
							if err == nil {
								select {
								case m.configCh <- newCfg:
								default:
								}
							}
						}
					case <-ctx.Done():
						_ = watcher.Close()
						return
					}
				}
			}()
		}
	}

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

			// If the configured theme is a plugin theme, apply it now that plugins are loaded
			for _, pt := range m.plugHost.Themes() {
				if pt.Name == m.cfg.App.Theme {
					m.thBuiltin = pt
					m.theme = applyThemeOverride(pt, m.cfg.Theme)
					m.refreshStyles()
					break
				}
			}

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
	cmds := []tea.Cmd{m.loadTagsCmd(), m.loadTasksCmd(), m.loadAllTasksCmd(), m.checkUpdateCmd(), m.cleanupTickCmd()}
	if m.cfg.App.Rainbow {
		m.rainbowAnimating = true
		cmds = append(cmds, m.rainbowTickCmd())
	}
	if m.plugCh != nil {
		cmds = append(cmds, m.listenPluginsCmd())
	}
	if m.configCh != nil {
		cmds = append(cmds, m.listenConfigCmd())
	}
	if m.aiChan != nil {
		cmds = append(cmds, m.listenAICmd())
	}
	if m.cfg.App.MCPEnabled {
		cmds = append(cmds, m.startMCPCmd())
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
	case ModeImportExport:
		// Import/Export menu has a file path input field
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
		m.list.SetSize(x.Width, x.Height)
		m.pal.SetSize(x.Width, x.Height)
		m.det.SetSize(x.Width, x.Height)
		m.tm.SetSize(x.Width, x.Height)
		m.pm.SetSize(x.Width, x.Height)
		m.set.SetSize(x.Width, x.Height)
		m.iem.SetSize(x.Width, x.Height)
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
		m.list.SetAllTasks(m.all)
		m.rebuildPaletteIndex()
		return m, nil

	case palette.CloseMsg:
		if m.mode == ModePalette {
			m.mode = ModeList
			m.transitioning = true
			m.transitionStarted = time.Now()
			return m, m.viewTransitionTickCmd()
		}
		return m, nil

	case help.CloseMsg:
		if m.mode == ModeHelp {
			m.mode = ModeList
			m.transitioning = true
			m.transitionStarted = time.Now()
			return m, m.viewTransitionTickCmd()
		}
		return m, nil

	case theme_menu.CloseMsg:
		if m.mode == ModeThemeMenu {
			m.mode = ModeList
			m.transitioning = true
			m.transitionStarted = time.Now()
			return m, m.viewTransitionTickCmd()
		}
		return m, nil

	case settings.CloseMsg:
		if m.mode == ModeSettings {
			m.mode = ModeList
			m.transitioning = true
			m.transitionStarted = time.Now()
			return m, m.viewTransitionTickCmd()
		}
		return m, nil

	case settings.ConfigChangedMsg:
		oldMCP := m.cfg.App.MCPEnabled
		m.cfg = x.Config
		m.aiKey = m.cfg.App.GeminiAPIKey
		if m.aiKey != "" {
			m.aiClient, _ = ai.NewClient(m.ctx, m.aiKey, m.cfg.App.AIModel)
			ai.SetService(m.svc)
		}
		m.km = keymap.FromConfig(m.cfg.Keymap)
		m.thBuiltin = theme.FindBuiltin(m.cfg.App.Theme)
		m.theme = applyThemeOverride(m.thBuiltin, m.cfg.Theme)
		m.refreshStyles()
		m.set.SetConfig(m.cfg)
		m.set.SetStyles(m.s)

		m.rebuildViews()
		m.rebuildPaletteIndex()

		var mcpCmd tea.Cmd
		if m.cfg.App.MCPEnabled && !oldMCP {
			mcpCmd = m.startMCPCmd()
		} else if !m.cfg.App.MCPEnabled && oldMCP {
			mcpCmd = m.stopMCPCmd()
		}

		// If config changed externally or internally, we might need to restart/update components
		if m.cfg.Sync.Enabled && m.syncEngine == nil {
			m.syncEngine = ksync.New(m.svc.Repo(), m.cfg.Sync.RepoPath, m.cfg.Sync.Remote, m.cfg.Sync.Branch, ksync.Strategy(m.cfg.Sync.Strategy), m.cfg.Sync.AutoPush)
		} else if !m.cfg.Sync.Enabled {
			m.syncEngine = nil
		}

		cmds := []tea.Cmd{m.listenConfigCmd()}
		if mcpCmd != nil {
			cmds = append(cmds, mcpCmd)
		}
		// Restart rainbow ticker if it was just enabled and isn't already running
		if m.cfg.App.Rainbow && !m.rainbowAnimating {
			m.rainbowAnimating = true
			cmds = append(cmds, m.rainbowTickCmd())
		}

		// Continue listening for more config changes
		return m, tea.Batch(cmds...)

	case theme_menu.SelectMsg:
		m.theme = x.Theme
		m.cfg.App.Theme = x.Theme.Name
		_ = m.cfg.Save()
		m.refreshStyles()
		m.mode = ModeList
		m.transitioning = true
		m.transitionStarted = time.Now()
		return m, m.viewTransitionTickCmd()

	case plugin_menu.CloseMsg:
		if m.mode == ModePluginMenu {
			m.mode = ModeList
			m.transitioning = true
			m.transitionStarted = time.Now()
			return m, m.viewTransitionTickCmd()
		}
		return m, nil

	case import_export_menu.CloseMsg:
		if m.mode == ModeImportExport {
			m.mode = ModeList
			m.transitioning = true
			m.transitionStarted = time.Now()
			return m, m.viewTransitionTickCmd()
		}
		return m, nil

	case import_export_menu.SelectMsg:
		if m.mode == ModeImportExport {
			m.mode = ModeList
			m.transitioning = true
			m.transitionStarted = time.Now()
			return m, tea.Batch(
				m.handleImportExportAction(x.Action, x.Path),
				m.viewTransitionTickCmd(),
			)
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

	case plugin_menu.TransitionMsg:
		if m.mode == ModePluginMenu {
			m.transitioning = true
			m.transitionStarted = time.Now()
			m.animationGen++
			return m, m.viewTransitionTickCmd()
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
			m.prevActiveIdx = m.activeIdx
			m.setActiveView(core.ViewTag)
			m.rebuildComponentSizes() // Recalculate layout when filter changes
			m.transitioning = true
			m.transitionStarted = time.Now()
			return m, tea.Batch(m.loadTasksCmd(), m.viewTransitionTickCmd())
		case search.KindCommand:
			return m, m.runCommand(x.Item.ID)
		}
		return m, nil

	case editor.CloseMsg:
		if m.mode == ModeEditor {
			m.edit = nil
			m.mode = ModeList
			m.transitioning = true
			m.transitionStarted = time.Now()
			m.animationGen++
			return m, m.viewTransitionTickCmd()
		}
		return m, nil

	case editor.SaveNewMsg:
		return m, tea.Batch(m.createTaskCmd(x.Task), func() tea.Msg { return editor.CloseMsg{} })

	case editor.SavePatchMsg:
		return m, tea.Batch(m.updateTaskCmd(x.ID, x.Patch), func() tea.Msg { return editor.CloseMsg{} })

	case taskCreatedMsg:
		m.creatingTaskID = x.Task.ID
		m.creationStarted = time.Now()
		m.creationDuration = 800 * time.Millisecond
		m.animationGen++
		return m, tea.Batch(
			m.loadTagsCmd(),
			m.loadTasksCmd(),
			m.loadAllTasksCmd(),
			m.syncIfEnabledCmd(),
			m.bloomAnimationTickCmd(x.Task.ID),
		)

	case taskUpdatedMsg:
		return m, tea.Batch(m.loadTagsCmd(), m.loadTasksCmd(), m.loadAllTasksCmd(), m.syncIfEnabledCmd())

	case string: // AI prompt from panel
		return m, m.startAIStreamCmd(x)

	case ai_panel.AIChunkMsg:
		var cmds []tea.Cmd
		var cmd tea.Cmd
		m.aiPanel, cmd = m.aiPanel.Update(x)
		cmds = append(cmds, cmd, m.listenAICmd())

		if x.Chunk.Refresh {
			cmds = append(cmds, m.loadTagsCmd(), m.loadTasksCmd(), m.loadAllTasksCmd(), m.syncIfEnabledCmd())
		}
		return m, tea.Batch(cmds...)

	case mcpStatusMsg:
		m.mcpRunning = x.Running
		return m, nil

	case taskDeletedMsg:
		return m, tea.Batch(m.loadTagsCmd(), m.loadTasksCmd(), m.loadAllTasksCmd(), m.syncIfEnabledCmd())

	case rainbowTickMsg:
		if !m.cfg.App.Rainbow {
			m.rainbowAnimating = false
			return m, nil
		}
		m.rainbowAnimating = true
		// Linear rainbow animation: increment offset each tick
		m.RainbowAnimationOffset = (m.RainbowAnimationOffset + 1) % 7 // 7 colors in rainbow
		return m, m.rainbowTickCmd()

	case cleanupTickMsg:
		_ = m.svc.Prune(m.ctx)
		return m, m.cleanupTickCmd()

	case deleteAnimationTickMsg:
		if m.deletingTaskID != x.TaskID || x.Gen != m.animationGen {
			return m, nil
		}
		elapsed := time.Since(m.deleteStarted)
		if elapsed >= m.deleteDuration {
			taskID := m.deletingTaskID
			m.deletingTaskID = ""
			return m, m.deleteTaskCmd(taskID)
		}
		return m, m.deleteAnimationTickCmd(x.TaskID)

	case strikeAnimationTickMsg:
		// Drop stale ticks from a previous animation cycle
		if m.animatingTaskID != x.TaskID || x.Gen != m.animationGen {
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

	case bloomAnimationTickMsg:
		// Drop stale ticks from a previous animation cycle
		if m.creatingTaskID != x.TaskID || x.Gen != m.animationGen {
			return m, nil
		}
		elapsed := time.Since(m.creationStarted)
		if elapsed >= m.creationDuration {
			m.creatingTaskID = ""
			return m, nil
		}
		return m, m.bloomAnimationTickCmd(x.TaskID)

	case viewTransitionTickMsg:
		if !m.transitioning || x.Gen != m.animationGen {
			return m, nil
		}
		elapsed := time.Since(m.transitionStarted)
		duration := 600 * time.Millisecond
		if elapsed >= duration {
			m.transitioning = false
			m.transitionProgress = 1.0
			m.prevActiveIdx = m.activeIdx
			return m, nil
		}
		raw := float64(elapsed) / float64(duration)
		m.transitionProgress = render.Linear(raw)
		return m, m.viewTransitionTickCmd()

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
		// AI Panel priority
		if m.aiPanel.Visible {
			// AIPanelToggle still toggles
			if keymapMatch(m.km.AIPanelToggle, km) {
				m.aiPanel.Visible = false
				return m, nil
			}

			var cmd tea.Cmd
			m.aiPanel, cmd = m.aiPanel.Update(msg)
			if cmd != nil {
				return m, cmd
			}
			// AI panel intercepts all keys except ctrl+c
			if km.String() != "ctrl+c" {
				return m, nil
			}
		}

		if m.mode == ModeConfirmDelete {
			switch km.String() {
			case "y", "enter":
				if t, ok := m.list.Selected(); ok {
					m.mode = ModeList
					m.deletingTaskID = t.ID
					m.deleteStarted = time.Now()
					m.deleteDuration = 600 * time.Millisecond
					m.animationGen++
					return m, m.deleteAnimationTickCmd(t.ID)
				}
			case "a":
				m.mode = ModeList
				m.transitioning = true
				m.transitionStarted = time.Now()
				m.animationGen++
				return m, tea.Batch(m.deleteAllTasksCmd(), m.viewTransitionTickCmd())
			case "n", "esc":
				m.mode = ModeList
				m.transitioning = true
				m.transitionStarted = time.Now()
				m.animationGen++
				return m, m.viewTransitionTickCmd()
			}
		}

		if m.mode == ModeConfirmQuit {
			switch km.String() {
			case "y", "enter":
				return m, tea.Quit
			case "n", "esc":
				m.mode = ModeList
				m.transitioning = true
				m.transitionStarted = time.Now()
				m.animationGen++
				return m, m.viewTransitionTickCmd()
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

				// Validate
				parts := core.ParseTags(tagValue)
				allValid := true
				for _, p := range parts {
					found := false
					for _, t := range m.tags {
						if t == p {
							found = true
							break
						}
					}
					if !found {
						allValid = false
						break
					}
				}

				if !allValid && tagValue != "" {
					return m, nil // Don't submit
				}

				m.tagFilterInput.Blur()
				// Handle clear
				if tagValue == "" {
					m.tagFilter.Clear()
				} else {
					m.tagFilter.Set(tagValue)
				}

				m.setActiveView(core.ViewTag)
				m.rebuildComponentSizes()
				m.mode = ModeList
				m.transitioning = true
				m.transitionStarted = time.Now()
				m.animationGen++
				return m, tea.Batch(m.loadTasksCmd(), m.viewTransitionTickCmd())
			case "ctrl+u":
				// Clear the entire input
				m.tagFilterInput.SetValue("")
				return m, nil
			case "esc":
				m.tagFilterInput.Blur()
				m.mode = ModeList
				m.transitioning = true
				m.transitionStarted = time.Now()
				m.animationGen++
				return m, m.viewTransitionTickCmd()
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

			if keymapMatch(m.km.AIPanelToggle, km) {
				m.aiPanel.Toggle()
				m.aiPanel.SetSize(m.width, m.height)
				if m.aiPanel.Visible {
					return m, m.aiPanel.Init()
				}
				return m, nil
			}

			// Sub-menu specific toggles/actions that should work even in the menu themselves (to close them)
			if keymapMatch(m.km.ManagePlugins, km) {
				if m.mode == ModePluginMenu {
					m.mode = ModeList
					m.transitioning = true
					m.transitionStarted = time.Now()
					return m, m.viewTransitionTickCmd()
				}
				if m.plugHost != nil {
					m.pm.SetPlugins(m.plugHost.Plugins())
					m.mode = ModePluginMenu
					m.transitioning = true
					m.transitionStarted = time.Now()
					return m, m.viewTransitionTickCmd()
				}
				return m, nil
			}

			// Primary mode utility keys
			if m.mode == ModeList || m.mode == ModeDetail {
				if keymapMatch(m.km.Palette, km) {
					m.palTasksOnly = false
					m.applyPaletteIndex()
					m.pal.SetPlaceholder("Search tasks, commands, tags…")
					m.mode = ModePalette
					m.transitioning = true
					m.transitionStarted = time.Now()
					return m, tea.Batch(m.pal.Open(), m.viewTransitionTickCmd())
				}
				if keymapMatch(m.km.TaskSearch, km) {
					m.palTasksOnly = true
					m.applyPaletteIndex()
					m.pal.SetPlaceholder("Search tasks…")
					m.mode = ModePalette
					m.transitioning = true
					m.transitionStarted = time.Now()
					return m, tea.Batch(m.pal.Open(), m.viewTransitionTickCmd())
				}
				if keymapMatch(m.km.CycleTheme, km) {
					m.mode = ModeThemeMenu
					m.transitioning = true
					m.transitionStarted = time.Now()
					return m, m.viewTransitionTickCmd()
				}
				if keymapMatch(m.km.ImportExport, km) {
					m.mode = ModeImportExport
					m.transitioning = true
					m.transitionStarted = time.Now()
					return m, m.viewTransitionTickCmd()
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
					m.transitioning = true
					m.transitionStarted = time.Now()
					return m, m.viewTransitionTickCmd()
				}
				if keymapMatch(m.km.Issues, km) {
					return m, openURLCmd("https://github.com/programmersd21/kairo/issues")
				}
				if keymapMatch(m.km.Discussions, km) {
					return m, openURLCmd("https://github.com/programmersd21/kairo/discussions")
				}
				if keymapMatch(m.km.Changelog, km) {
					return m, openURLCmd("https://github.com/programmersd21/kairo/blob/main/CHANGELOG.md")
				}
				if keymapMatch(m.km.Settings, km) {
					m.mode = ModeSettings
					m.transitioning = true
					m.transitionStarted = time.Now()
					return m, m.viewTransitionTickCmd()
				}
			}

			if m.mode == ModeList {
				// Dynamic view switching (1-9)
				if len(km.String()) == 1 && km.String() >= "1" && km.String() <= "9" {
					digit := int(km.String()[0] - '0')
					idx := digit - 1
					if idx >= 0 && idx < len(m.views) {
						m.prevActiveIdx = m.activeIdx
						m.activeIdx = idx
						m.tagFilter.Clear()
						m.rebuildComponentSizes()
						m.transitioning = true
						m.transitionStarted = time.Now()
						return m, tea.Batch(m.loadTasksCmd(), m.viewTransitionTickCmd())
					}
				}

				switch {
				case km.String() == "f":
					m.tagFilterInput.SetValue(m.tagFilter.Value())
					m.tagFilterInput.Focus()
					m.mode = ModeTagFilter
					return m, nil
				case km.String() == "tab":
					m.prevActiveIdx = m.activeIdx
					m.activeIdx = (m.activeIdx + 1) % len(m.views)
					m.transitioning = true
					m.transitionStarted = time.Now()
					return m, tea.Batch(m.loadTasksCmd(), m.viewTransitionTickCmd())
				case km.String() == "shift+tab":
					m.prevActiveIdx = m.activeIdx
					m.activeIdx--
					if m.activeIdx < 0 {
						m.activeIdx = len(m.views) - 1
					}
					m.transitioning = true
					m.transitionStarted = time.Now()
					return m, tea.Batch(m.loadTasksCmd(), m.viewTransitionTickCmd())
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
						m.animationGen++
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
					m.transitioning = true
					m.transitionStarted = time.Now()
					m.animationGen++
					return m, m.viewTransitionTickCmd()
				}
				if keymapMatch(m.km.EditTask, km) {
					return m, m.fetchOpenEditCmd(m.det.Task().ID)
				}
				if keymapMatch(m.km.ToggleStrike, km) {
					t := m.det.Task()
					m.animationGen++
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

		// Validate tags in real-time
		input := m.tagFilterInput.Value()
		if input != "" {
			parts := core.ParseTags(input)
			valid := true
			for _, p := range parts {
				found := false
				for _, t := range m.tags {
					if t == p {
						found = true
						break
					}
				}
				if !found {
					valid = false
					break
				}
			}
			if !valid {
				m.tagFilterInput.TextStyle = m.s.BadgeDelete
			} else {
				m.tagFilterInput.TextStyle = m.s.Text
			}
		} else {
			m.tagFilterInput.TextStyle = m.s.Text
		}

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
	case ModeSettings:
		var cmd tea.Cmd
		m.set, cmd = m.set.Update(msg)
		return m, cmd
	case ModeImportExport:
		var cmd tea.Cmd
		m.iem, cmd = m.iem.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *Model) View() string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}

	// Pass down view transition progress for list cascading animation
	m.list.ViewTransitioning = m.transitioning
	m.list.ViewTransitionProgress = m.transitionProgress

	m.list.DeletingTaskID = m.deletingTaskID
	if m.deletingTaskID != "" {
		elapsed := time.Since(m.deleteStarted)
		raw := float64(elapsed) / float64(m.deleteDuration)
		m.list.DeleteProgress = render.EaseOutQuad(raw)
	}

	content := m.renderMainUI()

	// Final rendering pipeline: FillViewport guarantees that every cell in
	// the width×height viewport has the background color applied.
	// It pads lines, fills missing rows, and—critically—re-applies the
	// background ANSI sequence after every SGR reset (\x1b[0m), which is
	// the root cause of terminal default background bleeding through.
	return render.FillViewport(content, m.width, m.height, m.s.Theme.Bg)
}

func (m *Model) renderMainUI() string {
	// Calculate the width budget: when AI panel is visible, the main UI
	// shrinks to make room. Both halves must fit within m.width.
	mainW := m.width
	aiPanelW := 0
	if m.aiPanel.Visible {
		aiPanelW = int(float64(m.width) * 0.35)
		if aiPanelW < 30 {
			aiPanelW = 30
		}
		mainW = m.width - aiPanelW
		if mainW < 40 {
			mainW = 40
			aiPanelW = m.width - mainW
		}
	}

	head := m.renderHeaderWithWidth(mainW)
	foot := m.renderFooterWithWidth(mainW)

	hHeight := lipgloss.Height(head)
	fHeight := lipgloss.Height(foot)
	availableHeight := m.height - hHeight - fHeight
	if availableHeight < 0 {
		availableHeight = 0
	}

	// Update sizes dynamically — use mainW so components don't overflow
	m.list.SetSize(mainW, availableHeight)
	m.det.SetSize(mainW, availableHeight)
	m.pal.SetSize(mainW, availableHeight)
	m.pm.SetSize(mainW, availableHeight)
	m.set.SetSize(mainW, availableHeight)
	m.hlp.SetSize(mainW, availableHeight)
	m.tm.SetSize(mainW, availableHeight)
	m.iem.SetSize(mainW, availableHeight)
	if m.edit != nil {
		m.edit.SetSize(mainW, availableHeight)
	}
	if m.aiPanel.Visible {
		m.aiPanel.SetSizeExact(aiPanelW, availableHeight)
	}

	// Sync animation state to tasklist
	if m.animatingTaskID != "" {
		m.list.SetAnimation(m.animatingTaskID, m.animationStarted, m.animationDuration, m.animationReverse)
	}
	if m.creatingTaskID != "" {
		m.list.SetCreationAnimation(m.creatingTaskID, m.creationStarted, m.creationDuration)
	}

	var body string
	switch m.mode {
	case ModeList, ModeConfirmDelete:
		body = m.list.View()
	case ModeDetail:
		body = m.det.View()
	case ModePalette:
		body = m.pal.View()
	case ModeHelp:
		body = m.hlp.View()
	case ModeThemeMenu:
		body = m.tm.View()
	case ModePluginMenu:
		body = m.pm.View()
	case ModeSettings:
		body = m.set.View()
	case ModeTagFilter:
		body = m.renderTagFilterOverlay(availableHeight)
	case ModeEditor:
		if m.edit != nil {
			body = m.edit.View()
		} else {
			body = m.list.View()
		}
	case ModeImportExport:
		body = m.iem.View()
	default:
		body = m.list.View()
	}

	// Ensure body fills its allocated height.
	body = lipgloss.NewStyle().
		Height(availableHeight).
		Width(mainW).
		Background(m.s.Theme.Bg).
		Render(body)

	// Cinematic "vertical split" reveal (masking effect)
	if m.transitioning && m.transitionProgress < 1.0 {
		lines := strings.Split(body, "\n")
		mid := availableHeight / 2
		revealHalf := int(float64(mid) * m.transitionProgress)

		emptyLine := lipgloss.NewStyle().
			Width(mainW).
			Background(m.s.Theme.Bg).
			Render(strings.Repeat(" ", mainW))

		for i := 0; i < len(lines); i++ {
			dist := i - mid
			if dist < 0 {
				dist = -dist
			}
			if dist > revealHalf {
				lines[i] = emptyLine
			}
		}
		body = strings.Join(lines, "\n")
	}

	content := lipgloss.JoinVertical(lipgloss.Left, head, body, foot)
	if m.aiPanel.Visible {
		return lipgloss.JoinHorizontal(lipgloss.Top, content, m.aiPanel.View())
	}
	return content
}

func (m *Model) startAIStreamCmd(prompt string) tea.Cmd {
	return func() tea.Msg {
		defer func() {
			if r := recover(); r != nil {
				_ = r // absorb panics from network/channel teardown
			}
		}()

		if m.aiClient == nil {
			if m.aiKey == "" {
				return ai_panel.AIChunkMsg{Chunk: ai.StreamChunk{Err: fmt.Errorf("API key not set. Go to settings (ctrl+s) to add it")}}
			}
			var err error
			m.aiClient, err = ai.NewClient(m.ctx, m.aiKey, m.cfg.App.AIModel)
			if err != nil {
				return ai_panel.AIChunkMsg{Chunk: ai.StreamChunk{Err: err}}
			}
		}

		appCtx := ai.AppContext{
			ViewName: m.views[m.activeIdx].Title,
			Data:     fmt.Sprintf("Tasks: %d, Tags: %v", len(m.all), m.tags),
		}

		ch, err := m.aiClient.ChatStream(m.ctx, m.aiPanel.History, prompt, appCtx)
		if err != nil {
			return ai_panel.AIChunkMsg{Chunk: ai.StreamChunk{Err: err}}
		}

		for chunk := range ch {
			select {
			case <-m.ctx.Done():
				return nil
			case m.aiChan <- ai_panel.AIChunkMsg{Chunk: chunk}:
			}
		}
		return nil
	}
}

func (m *Model) listenAICmd() tea.Cmd {
	return func() tea.Msg {
		return <-m.aiChan
	}
}

type mcpStatusMsg struct {
	Running bool
}

func (m *Model) startMCPCmd() tea.Cmd {
	return func() tea.Msg {
		if m.mcpRunning {
			return nil
		}
		exe, _ := os.Executable()
		cmd := exec.Command(exe, "mcp")
		if err := cmd.Start(); err != nil {
			return mcpStatusMsg{Running: false}
		}
		m.mcpCmd = cmd
		m.mcpRunning = true
		return mcpStatusMsg{Running: true}
	}
}

func (m *Model) stopMCPCmd() tea.Cmd {
	return func() tea.Msg {
		if !m.mcpRunning || m.mcpCmd == nil {
			return nil
		}
		_ = m.mcpCmd.Process.Kill()
		_ = m.mcpCmd.Wait()
		m.mcpCmd = nil
		m.mcpRunning = false
		return mcpStatusMsg{Running: false}
	}
}

func (m *Model) rebuildComponentSizes() {
	// Component sizing is now handled dynamically in renderMainUI
}

// Add to Model struct:
// RainbowAnimationOffset int
// And inside New():
// m.RainbowAnimationOffset = 0

// renderHeaderWithWidth renders the header at a specific width.
func (m *Model) renderHeaderWithWidth(w int) string {
	saved := m.width
	m.width = w
	result := m.renderHeader()
	m.width = saved
	return result
}

// renderFooterWithWidth renders the footer at a specific width.
func (m *Model) renderFooterWithWidth(w int) string {
	saved := m.width
	m.width = w
	result := m.renderFooter()
	m.width = saved
	return result
}

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

	// Header has Padding(0, 1), so inner width is m.width - 2
	innerW := m.width - 2
	if innerW < 0 {
		innerW = 0
	}

	// Tabs
	tabs := []string{}
	tabWidths := make([]int, len(m.views))
	tabOffsets := make([]int, len(m.views))
	currentOffset := 0

	// Dynamic truncation to ensure tabs fit in the terminal width
	fixedTabW := 4 // 2 caps + 2 padding
	totalNaturalWidth := 0
	for _, v := range m.views {
		totalNaturalWidth += lipgloss.Width(v.Title) + fixedTabW
	}

	maxTitleW := 999
	if totalNaturalWidth > innerW && len(m.views) > 0 {
		availTextW := innerW - (len(m.views) * fixedTabW)
		if availTextW < 0 {
			availTextW = 0
		}
		maxTitleW = availTextW / len(m.views)
	}

	for i, v := range m.views {
		style := m.s.TabInactive
		isActive := false
		// During transition, we treat them all as inactive to show them "behind" the bubble
		if i == m.activeIdx && (!m.transitioning || m.prevActiveIdx == m.activeIdx) {
			style = m.s.TabActive
			isActive = true
		}

		title := utilTruncate(v.Title, maxTitleW)

		var rendered string
		if isActive {
			l := lipgloss.NewStyle().Foreground(m.s.Theme.Accent).Background(m.s.Theme.Bg).Render("")
			r := lipgloss.NewStyle().Foreground(m.s.Theme.Accent).Background(m.s.Theme.Bg).Render("")
			rendered = l + style.Render(title) + r
		} else {
			// Use background colored pill ends so the spacing matches the active tab
			l := lipgloss.NewStyle().Foreground(m.s.Theme.Bg).Background(m.s.Theme.Bg).Render("")
			r := lipgloss.NewStyle().Foreground(m.s.Theme.Bg).Background(m.s.Theme.Bg).Render("")
			rendered = l + style.Render(title) + r
		}

		tabs = append(tabs, rendered)
		tabWidths[i] = lipgloss.Width(rendered)
		tabOffsets[i] = currentOffset
		currentOffset += tabWidths[i]
	}
	inactiveTabRow := lipgloss.JoinHorizontal(lipgloss.Left, tabs...)
	tabRow := lipgloss.PlaceHorizontal(innerW, lipgloss.Center, inactiveTabRow)

	// Animated Indicator (Bubble effect)
	if m.transitioning && m.prevActiveIdx != m.activeIdx {
		// ... (keep the existing logic, just ensure tabRow is constructed correctly)
		pIdx := m.prevActiveIdx
		aIdx := m.activeIdx
		if pIdx >= 0 && pIdx < len(tabOffsets) && aIdx >= 0 && aIdx < len(tabOffsets) {
			currentPos := int(float64(tabOffsets[pIdx]) + float64(tabOffsets[aIdx]-tabOffsets[pIdx])*m.transitionProgress)
			currentWidth := int(float64(tabWidths[pIdx]) + float64(tabWidths[aIdx]-tabWidths[pIdx])*m.transitionProgress)

			// Calculate center offset for the bubble
			totalTabWidth := currentOffset
			startOffset := (innerW - totalTabWidth) / 2
			if startOffset < 0 {
				startOffset = 0
			}

			indicatorTextW := currentWidth - 2
			if indicatorTextW < 0 {
				indicatorTextW = 0
			}

			l := lipgloss.NewStyle().Foreground(m.s.Theme.Accent).Background(m.s.Theme.Bg).Render("")
			r := lipgloss.NewStyle().Foreground(m.s.Theme.Accent).Background(m.s.Theme.Bg).Render("")

			indicatorCenter := m.s.TabActive.
				Width(indicatorTextW).
				MaxHeight(1).
				Align(lipgloss.Center).
				Render(utilTruncate(m.views[aIdx].Title, indicatorTextW))

			indicator := l + indicatorCenter + r

			repeatCount := startOffset + currentPos
			if repeatCount < 0 {
				repeatCount = 0
			}
			spacer := strings.Repeat(" ", repeatCount)
			tabRow = spacer + indicator

			// To prevent flickering, we ensure the tabRow always has the same total width
			actualWidth := lipgloss.Width(indicator)
			remaining := currentOffset - currentPos - actualWidth
			if remaining > 0 {
				tabRow += strings.Repeat(" ", remaining)
			}
		}
	} else {
		// If not transitioning, ensure tabRow is centered
		tabRow = lipgloss.PlaceHorizontal(innerW, lipgloss.Center, tabRow)
	}

	// 3. Task Count Pill (Bottom Row)
	pillStyle := lipgloss.NewStyle().
		Foreground(m.s.Theme.Bg).
		Background(m.s.Theme.Accent).
		Bold(true).
		Padding(0, 1)
	leftCap := lipgloss.NewStyle().Foreground(m.s.Theme.Accent).Background(m.s.Theme.Bg).Render("")
	rightCap := lipgloss.NewStyle().Foreground(m.s.Theme.Accent).Background(m.s.Theme.Bg).Render("")

	count := fmt.Sprintf("%d tasks", len(m.tasks))
	taskCountPill := leftCap + pillStyle.Render(count) + rightCap
	countRow := lipgloss.PlaceHorizontal(innerW, lipgloss.Center, taskCountPill)

	headerContent := lipgloss.JoinVertical(lipgloss.Center, "", logo, "", tabRow, "", countRow)
	return m.s.Header.Width(m.width).Render(headerContent)
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

	// Style for the pill container (accent background, theme background text)
	pillStyle := lipgloss.NewStyle().
		Foreground(m.s.Theme.Bg).
		Background(m.s.Theme.Accent).
		Bold(true).
		Padding(0, 1)

	// Unicode pill ends for circular appearance
	leftCap := lipgloss.NewStyle().Foreground(m.s.Theme.Accent).Background(m.s.Theme.Bg).Render("")
	rightCap := lipgloss.NewStyle().Foreground(m.s.Theme.Accent).Background(m.s.Theme.Bg).Render("")

	// Separator between pills
	sep := lipgloss.NewStyle().
		Background(m.s.Theme.Bg).
		Render(" ")

	makePill := func(text string) string {
		return leftCap + pillStyle.Render(text) + rightCap
	}

	left := ""
	// Critical prompts are always shown regardless of ShowHelp setting
	switch m.mode {
	case ModeConfirmDelete:
		delLeft := lipgloss.NewStyle().Foreground(m.s.Theme.Bad).Background(m.s.Theme.Bg).Render("")
		delRight := lipgloss.NewStyle().Foreground(m.s.Theme.Bad).Background(m.s.Theme.Bg).Render("")
		delPill := delLeft + m.s.BadgeDelete.Render("DELETE?") + delRight
		left = " " + delPill + " " + makePill("y/enter confirm") + sep + makePill("a delete all") + sep + makePill("n/esc cancel")
	case ModeConfirmQuit:
		quitLeft := lipgloss.NewStyle().Foreground(m.s.Theme.Warn).Background(m.s.Theme.Bg).Render("")
		quitRight := lipgloss.NewStyle().Foreground(m.s.Theme.Warn).Background(m.s.Theme.Bg).Render("")
		quitPill := quitLeft + m.s.BadgeQuit.Render("QUIT?") + quitRight
		left = " " + quitPill + " " + makePill("y/enter confirm") + sep + makePill("n/esc cancel")
	case ModeTagFilter:
		left = " " + makePill("enter apply") + sep + makePill("esc cancel") + sep + makePill("ctrl+u clear")
	default:
		// Only show help pills if ShowHelp is enabled in config
		if m.cfg.App.ShowHelp {
			switch m.mode {
			case ModeDetail:
				items := []string{
					makePill(fk(m.km.Back) + " " + styles.IconBack + "back"),
					makePill(fk(m.km.EditTask) + " " + styles.IconEdit + "edit"),
					makePill(fk(m.km.Palette) + " " + styles.IconPalette + "palette"),
					makePill(fk(m.km.Help) + " " + styles.IconHelp + "help"),
					makePill(fk(m.km.Issues) + " " + styles.IconIssues + "issues"),
					makePill(fk(m.km.Discussions) + " " + styles.IconDiscuss + "discussions"),
					makePill(fk(m.km.Changelog) + " " + styles.IconChangelog + "changelog"),
				}
				left = " " + strings.Join(items, sep)
			case ModeEditor:
				left = " " + makePill("ctrl+s save") + sep + makePill("esc cancel") + sep + makePill("tab nav")
			case ModePalette:
				left = " " + makePill("enter select") + sep + makePill("esc/p cancel") + sep + makePill(styles.IconUp+styles.IconDown+" nav")
			case ModeHelp:
				left = " " + makePill("esc/q/"+fk(m.km.Help)+" cancel")
			case ModeThemeMenu:
				left = " " + makePill("enter select") + sep + makePill("esc/q/"+fk(m.km.CycleTheme)+" cancel") + sep + makePill(styles.IconUp+styles.IconDown+" nav")
			case ModeSettings:
				left = " " + makePill("esc/ctrl+s close") + sep + makePill("enter toggle") + sep + makePill(styles.IconUp+styles.IconDown+" nav")
			case ModePluginMenu:
				left = " " + makePill("enter detail") + sep + makePill("u uninstall") + sep + makePill("o open") + sep + makePill("r reload") + sep + makePill("p/"+fk(m.km.ManagePlugins)+" cancel")
			default:
				items := []string{
					makePill(fk(m.km.Palette) + " " + styles.IconPalette + "palette"),
					makePill(fk(m.km.NewTask) + " " + styles.IconNew + "new"),
					makePill("f " + styles.IconTag + "tag"),
					makePill(fk(m.km.ToggleStrike) + " " + styles.IconStrike + "done"),
					makePill(fk(m.km.DeleteTask) + " " + styles.IconDelete + "delete"),
					makePill(fk(m.km.Settings) + " settings"),
					makePill(fk(m.km.AIPanelToggle) + " assistant"),
					makePill(fk(m.km.Help) + " " + styles.IconHelp + "help"),
				}
				left = " " + strings.Join(items, sep)
			}
		}
	}

	right := ""
	if m.statusText != "" {
		icon := styles.IconInfo
		if m.isErr {
			icon = styles.IconError
		}
		right = makePill(icon+" "+m.statusText) + " "
	} else {
		syncLogo := ""
		if m.syncEngine != nil && m.syncEngine.Enabled() {
			syncLogo = styles.IconSync + " "
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
			versionText = fmt.Sprintf("Update: %s → %s", cur, lat)
		}
		mcpStatus := ""
		if m.mcpRunning {
			mcpStatus = makePill("MCP "+styles.IconSuccess) + " "
		}
		right = mcpStatus + makePill(syncLogo+versionText) + " "
	}

	return render.BarLine(left, right, m.width, m.s.Theme.Bg)
}

// renderTagFilterOverlay renders the tag filter input modal
func (m *Model) renderTagFilterOverlay(h int) string {
	// Create filter input modal
	inputLabel := m.s.Title.Render("Filter by Tag (space/comma separated)")
	input := lipgloss.NewStyle().Padding(0, 1).Render(m.tagFilterInput.View())

	// Show hints for invalid tags
	hintText := "Available: " + strings.Join(m.tags, ", ")
	inputVal := m.tagFilterInput.Value()
	if inputVal != "" {
		parts := core.ParseTags(inputVal)
		var invalid []string
		for _, p := range parts {
			found := false
			for _, t := range m.tags {
				if t == p {
					found = true
					break
				}
			}
			if !found {
				invalid = append(invalid, p)
			}
		}
		if len(invalid) > 0 {
			hintText = m.s.BadgeDelete.Render("Invalid tags: " + strings.Join(invalid, ", "))
		}
	}
	hint := m.s.Muted.Render(hintText)

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
			m.prevActiveIdx = m.activeIdx
			m.activeIdx = i
			return
		}
	}
}

func (m *Model) rebuildViews() {
	base := core.DefaultViews(time.Now())
	var pluginThemes []theme.Theme
	if m.plugHost != nil {
		for _, v := range m.plugHost.Views() {
			base = append(base, core.View{ID: core.ViewID(v.ID), Title: v.Title, Filter: v.Filter})
		}
		pluginThemes = m.plugHost.Themes()
	}
	m.views = base
	if m.activeIdx >= len(m.views) {
		m.activeIdx = 0
	}

	// Re-initialize components that depend on plugin data
	m.tm = theme_menu.New(m.s, pluginThemes)

	// If current theme is a plugin theme, refresh it (in case it changed)
	for _, pt := range pluginThemes {
		if pt.Name == m.cfg.App.Theme {
			m.thBuiltin = pt
			m.theme = applyThemeOverride(pt, m.cfg.Theme)
			m.refreshStyles()
			break
		}
	}
}

func (m *Model) activeFilter() core.Filter {
	v := m.views[m.activeIdx]
	f := v.Filter

	// Apply dynamic parameters if it's a built-in view that supports them
	if v.ID == core.ViewTag {
		tags := core.ParseTags(m.tagFilter.Value())
		if len(tags) == 0 {
			m.tagFilter.Clear()
			m.mode = ModeList
		} else {
			m.tagFilter.Set(strings.Join(tags, " "))
		}
		f.Tags = tags
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

func (m *Model) deleteAllTasksCmd() tea.Cmd {
	return func() tea.Msg {
		if err := m.svc.DeleteAll(m.ctx); err != nil {
			return errMsg{Err: err}
		}
		return taskUpdatedMsg{} // Trigger reload
	}
}

func (m *Model) strikeAnimationTickCmd(taskID string) tea.Cmd {
	gen := m.animationGen
	return tea.Tick(16*time.Millisecond, func(time.Time) tea.Msg {
		return strikeAnimationTickMsg{TaskID: taskID, Gen: gen}
	})
}

func (m *Model) bloomAnimationTickCmd(taskID string) tea.Cmd {
	gen := m.animationGen
	return tea.Tick(16*time.Millisecond, func(time.Time) tea.Msg {
		return bloomAnimationTickMsg{TaskID: taskID, Gen: gen}
	})
}

func (m *Model) deleteAnimationTickCmd(taskID string) tea.Cmd {
	gen := m.animationGen
	return tea.Tick(16*time.Millisecond, func(time.Time) tea.Msg {
		return deleteAnimationTickMsg{TaskID: taskID, Gen: gen}
	})
}

func (m *Model) viewTransitionTickCmd() tea.Cmd {
	gen := m.animationGen
	return tea.Tick(16*time.Millisecond, func(time.Time) tea.Msg {
		return viewTransitionTickMsg{Gen: gen}
	})
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
	m.list.SetAllTasks(m.all)
	m.pal = palette.New(m.s)
	m.det = detail.New(m.s)
	m.hlp = help.New(m.s, m.km)
	var pluginThemes []theme.Theme
	if m.plugHost != nil {
		pluginThemes = m.plugHost.Themes()
	}
	m.tm = theme_menu.New(m.s, pluginThemes)
	m.pm = plugin_menu.New(m.s)
	m.iem = import_export_menu.New(m.s)
	m.aiPanel.SetStyles(m.s)

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
		search.Item{ID: "cmd:import-export", Kind: search.KindCommand, Title: "Import/Export", Hint: "x"},
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

func utilTruncate(s string, w int) string {
	if w <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= w {
		return s
	}
	if w <= 1 {
		return "…"
	}
	r := []rune(s)
	if len(r) <= w-1 {
		return string(r)
	}
	return string(r[:w-1]) + "…"
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
	case "cmd:import-export":
		m.mode = ModeImportExport
		return nil
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

func (m *Model) listenConfigCmd() tea.Cmd {
	return func() tea.Msg {
		cfg := <-m.configCh
		return settings.ConfigChangedMsg{Config: cfg}
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
func (m *Model) handleImportExportAction(action import_export_menu.Action, path string) tea.Cmd {
	return func() tea.Msg {
		taskAPI := api.New(m.svc)
		var resp api.Response

		if action.IsExport() {
			req := api.Request{
				Action:  "export",
				Payload: []byte(fmt.Sprintf(`{"format":"%s"}`, action.Format())),
			}
			resp = taskAPI.Execute(m.ctx, req)
			if resp.Success {
				data, ok := resp.Data.(string)
				if !ok {
					return errMsg{Err: errors.New("invalid response from API")}
				}
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil && filepath.Dir(path) != "." {
					return errMsg{Err: err}
				}
				if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
					return errMsg{Err: err}
				}
				m.statusText = fmt.Sprintf("Exported to %s", path)
				m.isErr = false
				return nil
			}
		} else {
			data, err := os.ReadFile(path)
			if err != nil {
				return errMsg{Err: err}
			}
			req := api.Request{
				Action:  "import",
				Payload: []byte(fmt.Sprintf(`{"format":"%s", "data":%q}`, action.Format(), string(data))),
			}
			resp = taskAPI.Execute(m.ctx, req)
			if resp.Success {
				msg, _ := resp.Data.(string)
				m.statusText = msg
				m.isErr = false
				return taskUpdatedMsg{} // Trigger reload
			}
		}

		if resp.Error != "" {
			return errMsg{Err: errors.New(resp.Error)}
		}
		return nil
	}
}
