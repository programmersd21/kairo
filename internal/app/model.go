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
	"github.com/programmersd21/kairo/internal/history"
	"github.com/programmersd21/kairo/internal/plugins"
	"github.com/programmersd21/kairo/internal/search"
	"github.com/programmersd21/kairo/internal/service"
	istats "github.com/programmersd21/kairo/internal/stats"
	ksync "github.com/programmersd21/kairo/internal/sync"
	"github.com/programmersd21/kairo/internal/ui/ai_panel"
	"github.com/programmersd21/kairo/internal/ui/detail"
	"github.com/programmersd21/kairo/internal/ui/editor"
	"github.com/programmersd21/kairo/internal/ui/focus"
	"github.com/programmersd21/kairo/internal/ui/help"
	"github.com/programmersd21/kairo/internal/ui/import_export_menu"
	"github.com/programmersd21/kairo/internal/ui/keymap"
	"github.com/programmersd21/kairo/internal/ui/onboarding"
	"github.com/programmersd21/kairo/internal/ui/palette"
	"github.com/programmersd21/kairo/internal/ui/plugin_menu"
	"github.com/programmersd21/kairo/internal/ui/render"
	"github.com/programmersd21/kairo/internal/ui/settings"
	"github.com/programmersd21/kairo/internal/ui/sidebar"
	"github.com/programmersd21/kairo/internal/ui/stats"
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
	ModeOnboarding
	ModeStats
	ModeProjectSwitcher
	ModeProjectSidebar
	ModeFocus
	ModeResultEdit
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

	mode           Mode
	sidebarVisible bool

	activeProject string
	views         []core.View
	activeIdx     int
	prevActiveIdx int
	tagFilter     FilterState // Replaced plain tagParam with proper state management
	priParam      *core.Priority

	list       tasklist.Model
	side       sidebar.Model
	pal        palette.Model
	det        detail.Model
	edit       *editor.Model
	hlp        help.Model
	onb        onboarding.Model
	tm         theme_menu.Model
	pm         plugin_menu.Model
	set        settings.Model
	iem        import_export_menu.Model
	stats      stats.Model
	foc        focus.Model
	aiPanel    ai_panel.Model
	aiClient   *ai.Client
	aiKey      string
	aiChan     chan ai_panel.AIChunkMsg
	mcpCmd     *exec.Cmd
	mcpRunning bool

	tagFilterInput textinput.Model // Input field for direct tag filtering in Tag View
	resultInput    textinput.Model

	palFullIdx      *search.Index
	palTasksIdx     *search.Index
	palProjectsIdx  *search.Index
	palTasksOnly    bool
	selectingParent bool

	tasks    []core.Task
	all      []core.Task
	tags     []string
	projects []string

	statusText string
	isErr      bool
	statusID   int

	updateAvailable *updateAvailableMsg

	syncEngine *ksync.Engine
	hist       *history.History

	plugHost *plugins.Host
	plugCh   chan struct{}
	statusCh chan statusMsg
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

type statsLoadedMsg struct {
	Data istats.DashboardData
}

func (m *Model) loadStatsCmd() tea.Cmd {
	return func() tea.Msg {
		tasks, _ := m.svc.ListAll(m.ctx)
		sessions, _ := m.svc.ListSessions(m.ctx)
		events, _ := m.svc.ListEvents(m.ctx)
		data := istats.ComputeDashboard(tasks, sessions, events)
		return statsLoadedMsg{Data: data}
	}
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

	resultInput := textinput.New()
	resultInput.Prompt = "Result: "
	resultInput.Placeholder = "Describe task outcome..."
	resultInput.CharLimit = 500
	resultInput.Width = 60

	m := &Model{
		ctx:                    ctx,
		cfg:                    cfg,
		svc:                    svc,
		km:                     km,
		thBuiltin:              thBuiltin,
		theme:                  th,
		s:                      s,
		mode:                   ModeList,
		activeProject:          cfg.App.ActiveProject,
		tagFilterInput:         tagInput,
		resultInput:            resultInput,
		RainbowAnimationOffset: 0,
	}
	m.list = tasklist.New(m.s, cfg.App.VimMode, cfg.App.Animations, m.km, cfg.List.Fields.Due.Minimal)
	m.list.SetTagsConfig(cfg.Tags.Highlight)
	m.list.SetRightOrder(cfg.List.Order.Right)
	m.side = sidebar.New(m.s)
	m.side.SetProjects(m.projects, cfg.Projects.Order, cfg.App.RecentProjects)
	m.side.SetActive(m.activeProject)
	m.pal = palette.New(m.s)
	m.det = detail.New(m.s)
	m.hlp = help.New(m.s, m.km)
	m.onb = onboarding.New(m.s, m.km)
	m.tm = theme_menu.New(m.s, nil)
	m.pm = plugin_menu.New(m.s)
	m.set = settings.New(m.s, cfg)
	m.iem = import_export_menu.New(m.s)
	m.stats = stats.New(m.s)
	m.foc = focus.New(m.s)
	m.aiPanel = ai_panel.New(m.s)
	m.aiChan = make(chan ai_panel.AIChunkMsg, 100)
	m.aiKey = cfg.App.GeminiAPIKey
	if m.aiKey != "" {
		m.aiClient, _ = ai.NewClient(ctx, m.aiKey, cfg.App.GeminiAPIKey)
		ai.SetService(svc)
	}

	// Initialize undo/redo history.
	m.hist = history.New(100)

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
			m.statusCh = make(chan statusMsg, 8)
			m.plugHost.SetNotifyFunc(func(msg string, isErr bool) {
				select {
				case m.statusCh <- statusMsg{Message: msg, IsErr: isErr}:
				default:
				}
			})
			_ = m.plugHost.LoadAll()

			// Now that plugins are loaded (and have had a chance to register hooks),
			// emit app_start so startup automation (like cleanup plugins) runs
			// in the same session (no restart required).
			m.svc.Hooks().AppStarted()

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

	if !m.cfg.App.OnboardingCompleted {
		m.mode = ModeOnboarding
	}

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
	if m.cfg.App.ActiveProject != "" {
		m.activeProject = m.cfg.App.ActiveProject
	}
	cmds := []tea.Cmd{m.loadTagsCmd(), m.loadProjectsCmd(), m.loadTasksCmd(), m.loadAllTasksCmd(), m.checkUpdateCmd(), m.cleanupTickCmd()}
	if m.cfg.App.Rainbow {
		m.rainbowAnimating = true
		cmds = append(cmds, m.rainbowTickCmd())
	}
	if m.plugCh != nil {
		cmds = append(cmds, m.listenPluginsCmd())
	}
	if m.statusCh != nil {
		cmds = append(cmds, m.listenStatusCmd())
	}
	if m.configCh != nil {
		cmds = append(cmds, m.listenConfigCmd())
	}
	if m.aiChan != nil {
		cmds = append(cmds, m.listenAICmd())
	}
	if m.mode == ModeOnboarding {
		cmds = append(cmds, m.onb.Init())
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
		return true
	case ModePalette, ModeTagFilter, ModeProjectSwitcher:
		return true
	case ModeImportExport:
		return true
	case ModeResultEdit:
		return true
	default:
		return false
	}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch x := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = x.Width, x.Height
		m.list.SetSize(x.Width, x.Height)
		m.onb.SetSize(x.Width, x.Height)
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
		m.statusID++
		return m, m.clearStatusCmd(m.statusID)

	case statusMsg:
		m.statusText = x.Message
		m.isErr = x.IsErr
		m.statusID++
		var cmds []tea.Cmd
		cmds = append(cmds, m.listenStatusCmd(), m.clearStatusCmd(m.statusID))
		if x.Refresh {
			m.rebuildComponentSizes()
			cmds = append(cmds, m.refreshCmd())

			// If in detail view, also reload the current task to reflect changes (undo/redo)
			if m.mode == ModeDetail {
				cmds = append(cmds, m.fetchOpenTaskCmd(m.det.Task().ID))
			}
		}
		return m, tea.Batch(cmds...)

	case clearStatusMsg:
		if x.ID == m.statusID {
			m.statusText = ""
			m.isErr = false
		}
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

	case projectsLoadedMsg:
		m.projects = x.Projects
		m.side.SetProjects(m.projects, m.cfg.Projects.Order, m.cfg.App.RecentProjects)
		m.side.SetActive(m.activeProject)
		m.rebuildProjectsIndex()
		return m, nil

	case statsLoadedMsg:
		m.stats.SetData(x.Data)
		return m, nil

	case allTasksLoadedMsg:
		m.all = x.Tasks
		m.list.SetAllTasks(m.all)
		m.rebuildPaletteIndex()
		return m, nil

	case onboarding.CloseMsg:
		m.mode = ModeList
		if !x.Skipped {
			m.cfg.App.OnboardingCompleted = true
			_ = m.cfg.Save()
		}
		if m.cfg.App.Animations {
			m.transitioning = m.cfg.App.Animations
			m.transitionStarted = time.Now()
			m.animationGen++
			return m, m.viewTransitionTickCmd()
		}
		return m, nil

	case palette.CloseMsg:
		if m.mode == ModePalette || m.mode == ModeProjectSwitcher {
			m.mode = ModeList
			if m.cfg.App.Animations {
				m.transitioning = m.cfg.App.Animations
				m.transitionStarted = time.Now()
				return m, m.viewTransitionTickCmd()
			}
			return m, nil
		}
		return m, nil

	case help.CloseMsg:
		if m.mode == ModeHelp {
			m.mode = ModeList
			if m.cfg.App.Animations {
				m.transitioning = m.cfg.App.Animations
				m.transitionStarted = time.Now()
				return m, m.viewTransitionTickCmd()
			}
			return m, nil
		}
		return m, nil

	case theme_menu.CloseMsg:
		if m.mode == ModeThemeMenu {
			m.mode = ModeList
			if m.cfg.App.Animations {
				m.transitioning = m.cfg.App.Animations
				m.transitionStarted = time.Now()
				return m, m.viewTransitionTickCmd()
			}
			return m, nil
		}
		return m, nil

	case settings.CloseMsg:
		if m.mode == ModeSettings {
			m.mode = ModeList
			if m.cfg.App.Animations {
				m.transitioning = m.cfg.App.Animations
				m.transitionStarted = time.Now()
				return m, m.viewTransitionTickCmd()
			}
			return m, nil
		}
		return m, nil

	case settings.ConfigChangedMsg:
		oldMCP := m.cfg.App.MCPEnabled
		m.cfg = x.Config
		m.aiKey = m.cfg.App.GeminiAPIKey
		if m.aiKey != "" {
			m.aiClient, _ = ai.NewClient(m.ctx, m.aiKey, m.cfg.App.AIModel)
			ai.SetService(m.svc)
		} else {
			m.aiClient = nil
		}
		m.km = keymap.FromConfig(m.cfg.Keymap)
		m.thBuiltin = theme.FindBuiltin(m.cfg.App.Theme)
		m.theme = applyThemeOverride(m.thBuiltin, m.cfg.Theme)
		m.refreshStyles()
		m.set.SetConfig(m.cfg)
		m.set.SetStyles(m.s)
		m.list.Animations = m.cfg.App.Animations // Sync animations setting to tasklist

		if m.edit != nil {
			m.edit.SetPreview(m.cfg.Edit.Preview)
		}

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
		if m.cfg.App.Animations {
			m.transitioning = m.cfg.App.Animations
			m.transitionStarted = time.Now()
			return m, m.viewTransitionTickCmd()
		}
		return m, nil

	case plugin_menu.CloseMsg:
		if m.mode == ModePluginMenu {
			m.mode = ModeList
			if m.cfg.App.Animations {
				m.transitioning = m.cfg.App.Animations
				m.transitionStarted = time.Now()
				return m, m.viewTransitionTickCmd()
			}
			return m, nil
		}
		return m, nil

	case import_export_menu.CloseMsg:
		if m.mode == ModeImportExport {
			m.mode = ModeList
			if m.cfg.App.Animations {
				m.transitioning = m.cfg.App.Animations
				m.transitionStarted = time.Now()
				return m, m.viewTransitionTickCmd()
			}
			return m, nil
		}
		return m, nil

	case import_export_menu.SelectMsg:
		if m.mode == ModeImportExport {
			m.mode = ModeList
			m.iem = m.iem.SetScope(x.Scope)
			var animCmd tea.Cmd
			if m.cfg.App.Animations {
				m.transitioning = m.cfg.App.Animations
				m.transitionStarted = time.Now()
				animCmd = m.viewTransitionTickCmd()
			}
			return m, tea.Batch(
				m.handleImportExportAction(x.Action, x.Path),
				animCmd,
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
			if m.cfg.App.Animations {
				m.transitioning = m.cfg.App.Animations
				m.transitionStarted = time.Now()
				m.animationGen++
				return m, m.viewTransitionTickCmd()
			}
			return m, nil
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
		if m.mode != ModePalette && m.mode != ModeProjectSwitcher {
			return m, nil
		}

		if m.selectingParent {
			m.selectingParent = false
			if m.edit != nil && x.Item.Kind == search.KindTask {
				m.edit.SetParentID(x.Item.ID)
			}
			m.mode = ModeEditor
			return m, nil
		}

		m.mode = ModeList
		switch x.Item.Kind {
		case search.KindTask:
			return m, m.fetchOpenTaskCmd(x.Item.ID)
		case search.KindTag:
			m.tagFilter.Set(x.Item.ID)
			m.setActiveView(core.ViewTag)
			m.transitioning = m.cfg.App.Animations
			if m.transitioning {
				m.transitionStarted = time.Now()
			}
			return m, tea.Batch(m.loadTasksCmd(), m.viewTransitionTickCmd())
		case search.KindCommand:
			return m, m.runCommand(x.Item.ID)
		}
		return m, nil

	case sidebar.SelectMsg:
		m.activeProject = x.Project
		m.updateRecentProject(m.activeProject)
		m.cfg.App.ActiveProject = m.activeProject
		_ = m.cfg.Save()
		m.side.SetProjects(m.projects, m.cfg.Projects.Order, m.cfg.App.RecentProjects)
		m.side.SetActive(m.activeProject)
		var animCmd tea.Cmd
		if m.cfg.App.Animations {
			m.transitioning = m.cfg.App.Animations
			m.transitionStarted = time.Now()
			m.animationGen++
			animCmd = m.viewTransitionTickCmd()
		}
		return m, tea.Batch(m.loadTasksCmd(), animCmd)

	case editor.CloseMsg:
		if m.mode == ModeEditor {
			m.mode = ModeList
			if m.cfg.App.Animations {
				m.transitioning = m.cfg.App.Animations
				m.transitionStarted = time.Now()
				m.animationGen++
				return m, m.viewTransitionTickCmd()
			}
			return m, nil
		}
		return m, nil

	case editor.SaveNewMsg:
		return m, tea.Batch(m.createTaskCmd(x.Task), func() tea.Msg { return editor.CloseMsg{} })

	case editor.SavePatchMsg:
		return m, tea.Batch(m.updateTaskCmd(x.ID, x.Patch), func() tea.Msg { return editor.CloseMsg{} })

	case editor.SelectParentMsg:
		m.palTasksOnly = true
		m.selectingParent = true
		m.applyPaletteIndex()
		m.pal.SetPlaceholder("Select parent task…")
		m.mode = ModePalette
		var animCmd tea.Cmd
		if m.cfg.App.Animations {
			m.transitioning = m.cfg.App.Animations
			m.transitionStarted = time.Now()
			animCmd = m.viewTransitionTickCmd()
		}
		return m, tea.Batch(m.pal.Open(), animCmd)

	case taskCreatedMsg:
		m.creatingTaskID = x.Task.ID
		m.creationStarted = time.Now()
		m.creationDuration = 800 * time.Millisecond
		m.animationGen++
		m.rebuildComponentSizes()
		return m, tea.Batch(
			m.refreshCmd(),
			m.syncIfEnabledCmd(),
			func() tea.Cmd {
				if m.cfg.App.Animations {
					return m.bloomAnimationTickCmd(x.Task.ID)
				}
				return nil
			}(),
		)

	case focus.SessionDoneMsg:
		return m, m.createFocusSessionCmd(x.Session)

	case focus.TickMsg:
		var cmd tea.Cmd
		m.foc, cmd = m.foc.Update(msg)
		// If rainbow is off, we still want the pulse to animate,
		// so we can increment the offset here too.
		if !m.cfg.App.Rainbow {
			m.RainbowAnimationOffset = (m.RainbowAnimationOffset + 1) % 7
		}
		return m, cmd

	case taskUpdatedMsg:
		m.rebuildComponentSizes()
		return m, m.refreshCmd()

	case string: // AI prompt from panel
		return m, tea.Batch(m.startAIStreamCmd(x), m.listenAICmd())

	case ai_panel.AIChunkMsg:
		var cmds []tea.Cmd
		var cmd tea.Cmd
		m.aiPanel, cmd = m.aiPanel.Update(x)
		cmds = append(cmds, cmd, m.listenAICmd())

		if x.Chunk.History != nil {
			m.aiPanel.History = x.Chunk.History
		}

		if x.Chunk.Refresh {
			m.statusText = "AI updated database"
			m.isErr = false
			m.statusID++
			cmds = append(cmds, m.refreshCmd(), m.syncIfEnabledCmd(), m.clearStatusCmd(m.statusID))
		}
		return m, tea.Batch(cmds...)

	case mcpStatusMsg:
		m.mcpRunning = x.Running
		return m, nil

	case taskDeletedMsg:
		// Deleting a task can orphan tags/projects; prune & reload so
		// UI stays accurate without requiring a restart.
		m.rebuildComponentSizes()
		return m, tea.Batch(m.pruneAndLoadTagsCmd(), m.loadProjectsCmd(), m.refreshCmd(), m.syncIfEnabledCmd())

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
		// Prune can remove orphan tags; refresh tags live so the palette/tag views
		// don't require an app restart.
		return m, tea.Batch(m.pruneAndLoadTagsCmd(), m.cleanupTickCmd())

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
		e := editor.New(m.s, editor.ModeEdit, x.Task, m.cfg.Edit.Preview)
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
		if m.mode == ModeOnboarding {
			var cmd tea.Cmd
			m.onb, cmd = m.onb.Update(msg)
			// Bubble up commands from onboarding
			if cmd != nil {
				return m, cmd
			}
			return m, nil
		}

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
				selected := m.list.GetSelectedTasks()
				var ids []string
				for _, t := range selected {
					ids = append(ids, t.ID)
				}
				m.mode = ModeList
				return m, m.deleteTasksCmd(ids)
			case "t":
				visible := m.list.GetVisibleTasks()
				var ids []string
				for _, item := range visible {
					ids = append(ids, item.ID)
				}
				m.mode = ModeList
				var animCmd tea.Cmd
				if m.cfg.App.Animations {
					m.transitioning = m.cfg.App.Animations
					m.transitionStarted = time.Now()
					m.animationGen++
					animCmd = m.viewTransitionTickCmd()
				}
				return m, tea.Batch(m.deleteTasksCmd(ids), animCmd)
			case "a":
				m.mode = ModeList
				var animCmd tea.Cmd
				if m.cfg.App.Animations {
					m.transitioning = m.cfg.App.Animations
					m.transitionStarted = time.Now()
					m.animationGen++
					animCmd = m.viewTransitionTickCmd()
				}
				return m, tea.Batch(m.deleteAllTasksCmd(), animCmd)
			case "n", "esc":
				m.mode = ModeList
				m.transitioning = m.cfg.App.Animations
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
				m.transitioning = m.cfg.App.Animations
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
				var animCmd tea.Cmd
				if m.cfg.App.Animations {
					m.transitioning = m.cfg.App.Animations
					m.transitionStarted = time.Now()
					m.animationGen++
					animCmd = m.viewTransitionTickCmd()
				}
				return m, tea.Batch(m.loadTasksCmd(), animCmd)
			case "ctrl+z":
				// Clear the entire input
				m.tagFilterInput.SetValue("")
				return m, nil
			case "esc":
				m.tagFilterInput.Blur()
				m.mode = ModeList
				if m.cfg.App.Animations {
					m.transitioning = m.cfg.App.Animations
					m.transitionStarted = time.Now()
					m.animationGen++
					return m, m.viewTransitionTickCmd()
				}
				return m, nil
			}
		}

		// Sidebar toggle
		if keymapMatch(m.km.ProjectSwitcher, km) {
			m.sidebarVisible = !m.sidebarVisible
			if m.sidebarVisible {
				m.mode = ModeProjectSidebar
				m.side.Focus(true)
			} else {
				m.mode = ModeList
				m.side.Focus(false)
			}
			m.rebuildComponentSizes()
			return m, nil
		}

		// Focus switching
		if m.sidebarVisible {
			if keymapMatch(m.km.FocusSidebar, km) {
				m.mode = ModeProjectSidebar
				m.side.Focus(true)
				return m, nil
			}
			if keymapMatch(m.km.FocusList, km) {
				m.mode = ModeList
				m.side.Focus(false)
				return m, nil
			}

			if km.String() == "tab" {
				if m.mode == ModeProjectSidebar {
					m.mode = ModeList
					m.side.Focus(false)
				} else {
					m.mode = ModeProjectSidebar
					m.side.Focus(true)
				}
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

			if keymapMatch(m.km.Undo, km) {
				op := m.hist.Undo()
				if op == nil {
					u, r := m.hist.Len()
					m.statusText = fmt.Sprintf("Nothing to undo (stack: %d/%d)", u, r)
					m.isErr = false
					m.statusID++
					return m, tea.Batch(m.listenStatusCmd(), m.clearStatusCmd(m.statusID))
				}
				return m, m.undoCmd(op)
			}
			if keymapMatch(m.km.Redo, km) {
				op := m.hist.Redo()
				if op == nil {
					u, r := m.hist.Len()
					m.statusText = fmt.Sprintf("Nothing to redo (stack: %d/%d)", u, r)
					m.isErr = false
					m.statusID++
					return m, tea.Batch(m.listenStatusCmd(), m.clearStatusCmd(m.statusID))
				}
				return m, m.redoCmd(op)
			}

			if km.String() == "ctrl+w" {
				m.onb = onboarding.New(m.s, m.km)
				m.onb.SetSize(m.width, m.height)
				m.mode = ModeOnboarding
				var animCmd tea.Cmd
				if m.cfg.App.Animations {
					m.transitioning = m.cfg.App.Animations
					m.transitionStarted = time.Now()
					m.animationGen++
					animCmd = m.viewTransitionTickCmd()
				}
				return m, tea.Batch(m.onb.Init(), animCmd)
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
					if m.cfg.App.Animations {
						m.transitioning = m.cfg.App.Animations
						m.transitionStarted = time.Now()
						return m, m.viewTransitionTickCmd()
					}
					return m, nil
				}
				if m.plugHost != nil {
					m.pm.SetPlugins(m.plugHost.Plugins())
					m.mode = ModePluginMenu
					if m.cfg.App.Animations {
						m.transitioning = m.cfg.App.Animations
						m.transitionStarted = time.Now()
						return m, m.viewTransitionTickCmd()
					}
					return m, nil
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
					var animCmd tea.Cmd
					if m.cfg.App.Animations {
						m.transitioning = m.cfg.App.Animations
						m.transitionStarted = time.Now()
						animCmd = m.viewTransitionTickCmd()
					}
					return m, tea.Batch(m.pal.Open(), animCmd)
				}
				if keymapMatch(m.km.TaskSearch, km) {
					m.palTasksOnly = true
					m.applyPaletteIndex()
					m.pal.SetPlaceholder("Search tasks…")
					m.mode = ModePalette
					var animCmd tea.Cmd
					if m.cfg.App.Animations {
						m.transitioning = m.cfg.App.Animations
						m.transitionStarted = time.Now()
						animCmd = m.viewTransitionTickCmd()
					}
					return m, tea.Batch(m.pal.Open(), animCmd)
				}
				if keymapMatch(m.km.CycleTheme, km) {
					m.mode = ModeThemeMenu
					if m.cfg.App.Animations {
						m.transitioning = m.cfg.App.Animations
						m.transitionStarted = time.Now()
						return m, m.viewTransitionTickCmd()
					}
					return m, nil
				}
				if keymapMatch(m.km.ImportExport, km) {
					m.mode = ModeImportExport
					if m.cfg.App.Animations {
						m.transitioning = m.cfg.App.Animations
						m.transitionStarted = time.Now()
						return m, m.viewTransitionTickCmd()
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
					if m.cfg.App.Animations {
						m.transitioning = m.cfg.App.Animations
						m.transitionStarted = time.Now()
						return m, m.viewTransitionTickCmd()
					}
					return m, nil
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
					if m.cfg.App.Animations {
						m.transitioning = m.cfg.App.Animations
						m.transitionStarted = time.Now()
						return m, m.viewTransitionTickCmd()
					}
					return m, nil
				}
				if keymapMatch(m.km.Stats, km) {
					m.mode = ModeStats
					var animCmd tea.Cmd
					if m.cfg.App.Animations {
						m.transitioning = m.cfg.App.Animations
						m.transitionStarted = time.Now()
						animCmd = m.viewTransitionTickCmd()
					}
					return m, tea.Batch(m.loadStatsCmd(), m.stats.Init(), animCmd)
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
						var animCmd tea.Cmd
						if m.cfg.App.Animations {
							m.transitioning = m.cfg.App.Animations
							m.transitionStarted = time.Now()
							animCmd = m.viewTransitionTickCmd()
						}
						return m, tea.Batch(m.loadTasksCmd(), animCmd)
					}
				}

				switch {
				case keymapMatch(m.km.ViewTag, km):
					m.tagFilterInput.SetValue(m.tagFilter.Value())
					m.tagFilterInput.Focus()
					m.mode = ModeTagFilter
					return m, nil
				case keymapMatch(m.km.Focus, km):
					if item, ok := m.list.Selected(); ok {
						m.foc.Task = &item.Task
					} else {
						m.foc.Task = nil
					}
					m.mode = ModeFocus
					return m, nil
				case km.String() == " ":
					// Handle space key for toggle collapse
					if item, ok := m.list.Selected(); ok && item.HasChildren {
						newCollapsed := !item.Collapsed
						patch := core.TaskPatch{Collapsed: &newCollapsed}
						return m, m.updateTaskCmd(item.ID, patch)
					}
				case km.String() == "tab":
					m.prevActiveIdx = m.activeIdx
					m.activeIdx = (m.activeIdx + 1) % len(m.views)
					var animCmd tea.Cmd
					if m.cfg.App.Animations {
						m.transitioning = m.cfg.App.Animations
						m.transitionStarted = time.Now()
						animCmd = m.viewTransitionTickCmd()
					}
					return m, tea.Batch(m.loadTasksCmd(), animCmd)
				case km.String() == "shift+tab":
					m.prevActiveIdx = m.activeIdx
					m.activeIdx--
					if m.activeIdx < 0 {
						m.activeIdx = len(m.views) - 1
					}
					var animCmd tea.Cmd
					if m.cfg.App.Animations {
						m.transitioning = m.cfg.App.Animations
						m.transitionStarted = time.Now()
						animCmd = m.viewTransitionTickCmd()
					}
					return m, tea.Batch(m.loadTasksCmd(), animCmd)
				case keymapMatch(m.km.NewTask, km):
					task := core.Task{Status: core.StatusTodo, Priority: core.P1}
					m.activeFilter().ApplyToTask(&task)
					e := editor.New(m.s, editor.ModeNew, task, m.cfg.Edit.Preview)
					m.edit = &e
					m.rebuildComponentSizes()
					m.mode = ModeEditor
					return m, m.edit.Init()
				case keymapMatch(m.km.EditTask, km):
					if item, ok := m.list.Selected(); ok {
						t := item.Task
						return m, m.fetchOpenEditCmd(t.ID)
					}
				case keymapMatch(m.km.DeleteTask, km):
					if _, ok := m.list.Selected(); ok {
						m.mode = ModeConfirmDelete
						return m, nil
					}
				case keymapMatch(m.km.OpenTask, km):
					if item, ok := m.list.Selected(); ok {
						t := item.Task
						return m, m.fetchOpenTaskCmd(t.ID)
					}
				case keymapMatch(m.km.ToggleCollapse, km):
					if item, ok := m.list.Selected(); ok && item.HasChildren {
						newCollapsed := !item.Collapsed
						patch := core.TaskPatch{Collapsed: &newCollapsed}
						return m, m.updateTaskCmd(item.ID, patch)
					}
				case keymapMatch(m.km.ToggleStrike, km):
					selected := m.list.GetSelectedTasks()
					if len(selected) == 1 && selected[0].Status == core.StatusDone {
						m.resultInput.SetValue(selected[0].Result)
						m.mode = ModeResultEdit
						m.resultInput.Focus()
						return m, nil
					}
					var cmds []tea.Cmd
					for _, t := range selected {
						newStatus := core.StatusDone
						if t.Status == core.StatusDone {
							newStatus = core.StatusTodo
						}
						cmds = append(cmds, m.updateTaskCmd(t.ID, core.TaskPatch{Status: &newStatus}))
					}
					return m, tea.Batch(cmds...)
				case keymapMatch(m.km.DuplicateTask, km):
					if item, ok := m.list.Selected(); ok {
						t := item.Task
						dup := t
						dup.ID = ""
						dup.Title = t.Title + " (copy)"
						e := editor.New(m.s, editor.ModeNew, dup, m.cfg.Edit.Preview)
						m.edit = &e
						m.rebuildComponentSizes()
						m.mode = ModeEditor
						return m, m.edit.Init()
					}
				}
			}

			if m.mode == ModeDetail {
				if keymapMatch(m.km.Back, km) {
					m.mode = ModeList
					if m.cfg.App.Animations {
						m.transitioning = m.cfg.App.Animations
						m.transitionStarted = time.Now()
						m.animationGen++
						return m, m.viewTransitionTickCmd()
					}
					return m, nil
				}
				if keymapMatch(m.km.EditTask, km) {
					return m, m.fetchOpenEditCmd(m.det.Task().ID)
				}
				if km.String() == " " {
					// Handle space key for toggle collapse
					t := m.det.Task()
					newCollapsed := !t.Collapsed
					patch := core.TaskPatch{Collapsed: &newCollapsed}
					return m, m.updateTaskCmd(t.ID, patch)
				}
				if keymapMatch(m.km.ToggleStrike, km) {
					t := m.det.Task()
					if t.Status == core.StatusDone {
						m.resultInput.SetValue(t.Result)
						m.mode = ModeResultEdit
						m.resultInput.Focus()
						return m, nil
					}
					m.animationGen++
					m.animationReverse = (t.Status == core.StatusDone)
					m.animatingTaskID = t.ID
					m.animationStarted = time.Now()
					m.animationDuration = 600 * time.Millisecond
					if m.cfg.App.Animations {
						return m, m.strikeAnimationTickCmd(t.ID)
					}
					newStatus := core.StatusDone
					if t.Status == core.StatusDone {
						newStatus = core.StatusTodo
					}
					return m, m.updateTaskCmd(t.ID, core.TaskPatch{Status: &newStatus})
				}
				if keymapMatch(m.km.ToggleCollapse, km) {
					t := m.det.Task()
					newCollapsed := !t.Collapsed
					patch := core.TaskPatch{Collapsed: &newCollapsed}
					return m, m.updateTaskCmd(t.ID, patch)
				}
			}
		}
	}

	// Delegate to active component.
	switch m.mode {
	case ModeResultEdit:
		if km, ok := msg.(tea.KeyMsg); ok {
			switch km.String() {
			case "enter":
				result := m.resultInput.Value()
				taskID := ""
				if m.mode == ModeDetail {
					taskID = m.det.Task().ID
				} else if item, ok := m.list.Selected(); ok {
					taskID = item.ID
				}
				if taskID != "" {
					patch := core.TaskPatch{Result: &result}
					m.mode = ModeList
					return m, m.updateTaskCmd(taskID, patch)
				}
				m.mode = ModeList
				return m, nil
			case "esc":
				m.mode = ModeList
				return m, nil
			default:
				var cmd tea.Cmd
				m.resultInput, cmd = m.resultInput.Update(msg)
				return m, cmd
			}
		}
		var cmd tea.Cmd
		m.resultInput, cmd = m.resultInput.Update(msg)
		return m, cmd
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
	case ModePalette, ModeProjectSwitcher:
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
	case ModeProjectSidebar:
		var cmd tea.Cmd
		m.side, cmd = m.side.Update(msg)
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
	case ModeFocus:
		if km, ok := msg.(tea.KeyMsg); ok {
			if keymapMatch(m.km.Back, km) {
				m.mode = ModeList
				return m, nil
			}
		}
		var cmd tea.Cmd
		m.foc, cmd = m.foc.Update(msg)
		return m, cmd
	case ModeSettings:
		var cmd tea.Cmd
		m.set, cmd = m.set.Update(msg)
		return m, cmd
	case ModeImportExport:
		var cmd tea.Cmd
		m.iem, cmd = m.iem.Update(msg)
		return m, cmd
	case ModeOnboarding:
		var cmd tea.Cmd
		m.onb, cmd = m.onb.Update(msg)
		return m, cmd
	case ModeStats:
		if km, ok := msg.(tea.KeyMsg); ok {
			if keymapMatch(m.km.Back, km) || km.String() == "q" {
				m.mode = ModeList
				if m.cfg.App.Animations {
					m.transitioning = m.cfg.App.Animations
					m.transitionStarted = time.Now()
					return m, m.viewTransitionTickCmd()
				}
				return m, nil
			}
		}
		var cmd tea.Cmd
		m.stats, cmd = m.stats.Update(msg)
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

	// Fill the viewport completely using the background color.
	// This ensures no terminal default background bleeds through.
	filled := render.FillViewport(content, m.width, m.height, m.s.Theme.Bg)

	// Return the painted string directly. `render.FillViewport` already
	// paints background for every line, inserts per-line EL sequences,
	// and appends ED at the end. Returning this raw painted string avoids
	// any reflow or re-rendering that `lipgloss.Place` may perform during
	// resizes which can re-introduce slivers in some terminals.
	return filled
}

func (m *Model) renderMainUI() string {
	// Calculate the width budget: when AI panel or sidebar is visible, the main UI
	// shrinks to make room. Both halves must fit within m.width.
	mainW := m.width
	aiPanelW := 0
	if m.aiPanel.Visible {
		aiPanelW = int(float64(m.width) * 0.35)
		if aiPanelW < 30 {
			aiPanelW = 30
		}
		mainW -= aiPanelW
	}

	sidebarW := 0
	if m.sidebarVisible {
		sidebarW = 25
		mainW -= sidebarW
	}

	head := m.renderHeaderWithWidth(mainW + sidebarW)
	foot := m.renderFooterWithWidth(mainW + sidebarW)

	hHeight := lipgloss.Height(head)
	fHeight := lipgloss.Height(foot)
	availableHeight := m.height - hHeight - fHeight
	if availableHeight < 0 {
		availableHeight = 0
	}

	// Update sizes dynamically
	if m.sidebarVisible {
		m.side.SetSize(sidebarW, availableHeight)
	}
	m.list.SetSize(mainW, availableHeight)
	m.det.ShowID = m.cfg.App.ShowID
	m.det.SetSize(mainW, availableHeight)
	m.pal.SetSize(mainW, availableHeight)
	m.pm.SetSize(mainW, availableHeight)
	m.set.SetSize(mainW, availableHeight)
	m.hlp.SetSize(mainW, availableHeight)
	m.hlp.AIEnabled = m.aiKey != ""
	m.tm.SetSize(mainW, availableHeight)
	m.iem.SetSize(mainW, availableHeight)
	m.stats.SetSize(mainW, availableHeight)
	m.foc.SetSize(mainW, availableHeight)
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
	case ModePalette, ModeProjectSwitcher:
		body = m.pal.View()
	case ModeHelp:
		body = m.hlp.View()
	case ModeThemeMenu:
		body = m.tm.View()
	case ModePluginMenu:
		body = m.pm.View()
	case ModeSettings:
		body = m.set.View()
	case ModeResultEdit:
		body = m.renderResultOverlay(availableHeight)
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
	case ModeFocus:
		body = m.foc.View()
	case ModeStats:
		body = m.stats.View()
	case ModeOnboarding:
		body = m.list.View()
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

	if m.sidebarVisible {
		sidebarContent := m.side.View()
		sidebarContent = lipgloss.NewStyle().
			Width(sidebarW).
			Height(availableHeight).
			Background(m.s.Theme.Bg).
			Border(lipgloss.NormalBorder(), false, true, false, false).
			BorderForeground(m.s.Theme.Border).
			Render(sidebarContent)
		// We join body instead of content because head/foot should span full width?
		// Actually, the user's screenshot shows head/foot spanning only the main area.
		// Let's re-join with head/foot separately if needed.
		content = lipgloss.JoinVertical(lipgloss.Left, head, lipgloss.JoinHorizontal(lipgloss.Top, sidebarContent, body), foot)
	}

	if m.mode == ModeOnboarding {
		content = lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			m.onb.View(),
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceBackground(m.s.Theme.Bg),
		)
	}

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
		args := []string{"mcp"}
		if m.cfg.App.MCPPort != "" {
			args = append(args, m.cfg.App.MCPPort)
		}
		cmd := exec.Command(exe, args...)
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
}

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
		isActive := false
		// During transition, we treat them all as inactive to show them "behind" the bubble
		if i == m.activeIdx && (!m.transitioning || m.prevActiveIdx == m.activeIdx) {
			isActive = true
		}

		title := utilTruncate(v.Title, maxTitleW)

		var rendered string
		if isActive {
			l := m.s.TagLeft.Foreground(m.s.Theme.Accent).Render()
			r := m.s.TagRight.Foreground(m.s.Theme.Accent).Render()

			activePill := lipgloss.NewStyle().
				Foreground(m.s.Theme.Bg).
				Background(m.s.Theme.Accent).
				Render(title)

			rendered = l + activePill + r
		} else {
			// Inactive tabs should be clean text, not pills.
			rendered = m.s.TabInactive.Render(title)
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

			l := m.s.TagLeft.Foreground(m.s.Theme.Accent).Render()
			r := m.s.TagRight.Foreground(m.s.Theme.Accent).Render()

			indicatorCenter := lipgloss.NewStyle().
				Foreground(m.s.Theme.Bg).
				Background(m.s.Theme.Accent).
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
	leftCap := m.s.TagLeft.Foreground(m.s.Theme.Accent).Render()
	rightCap := m.s.TagRight.Foreground(m.s.Theme.Accent).Render()

	count := fmt.Sprintf("%d tasks", len(m.tasks))
	taskCountPill := leftCap + pillStyle.Render(count) + rightCap
	countRow := lipgloss.PlaceHorizontal(innerW, lipgloss.Center, taskCountPill)

	logoRow := lipgloss.PlaceHorizontal(innerW, lipgloss.Center, logo)
	if m.activeProject != "" {
		projectPill := lipgloss.NewStyle().
			Foreground(m.s.Theme.Accent).
			Bold(true).
			Padding(0, 1).
			Render("PROJECT: " + m.activeProject)
		logoRow = lipgloss.JoinHorizontal(lipgloss.Center, logo, "  ", projectPill)
		logoRow = lipgloss.PlaceHorizontal(innerW, lipgloss.Center, logoRow)
	}

	headerContent := lipgloss.JoinVertical(lipgloss.Center, "", logoRow, "", tabRow, "", countRow)
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
	leftCap := m.s.TagLeft.Foreground(m.s.Theme.Accent).Render()
	rightCap := m.s.TagRight.Foreground(m.s.Theme.Accent).Render()

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
		delLeft := m.s.TagLeft.Foreground(m.s.Theme.Bad).Render()
		delRight := m.s.TagRight.Foreground(m.s.Theme.Bad).Render()
		delPill := delLeft + m.s.BadgeDelete.Render("DELETE?") + delRight
		left = " " + delPill + " " + makePill("y/enter confirm") + sep + makePill("t delete tab") + sep + makePill("a delete all") + sep + makePill("n/esc cancel")
	case ModeConfirmQuit:
		quitLeft := m.s.TagLeft.Foreground(m.s.Theme.Warn).Render()
		quitRight := m.s.TagRight.Foreground(m.s.Theme.Warn).Render()
		quitPill := quitLeft + m.s.BadgeQuit.Render("QUIT?") + quitRight
		left = " " + quitPill + " " + makePill("y/enter confirm") + sep + makePill("n/esc cancel")
	case ModeTagFilter:
		left = " " + makePill("enter apply") + sep + makePill("esc cancel") + sep + makePill("ctrl+z clear")
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
				left = " " + makePill("ctrl+s save") + sep + makePill("ctrl+p preview") + sep + makePill("esc cancel") + sep + makePill("tab nav")
			case ModePalette:
				left = " " + makePill("enter select") + sep + makePill("esc/p cancel") + sep + makePill(styles.IconUp+styles.IconDown+" nav")
			case ModeProjectSwitcher:
				left = " " + makePill("enter select") + sep + makePill("esc cancel") + sep + makePill(styles.IconUp+styles.IconDown+" nav")
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
					makePill(fk(m.km.ProjectSwitcher) + " project"),
					makePill(fk(m.km.NewTask) + " " + styles.IconNew + "new"),
					makePill("f " + styles.IconTag + "tag"),
					makePill(fk(m.km.ToggleStrike) + " " + styles.IconStrike + "done"),
					makePill(fk(m.km.Stats) + " stats"),
					makePill(fk(m.km.DeleteTask) + " " + styles.IconDelete + "delete"),
					makePill(fk(m.km.Settings) + " settings"),
				}

				// Contextual toggle for collapse/expand
				if item, ok := m.list.Selected(); ok && item.HasChildren {
					label := "collapse"
					if item.Collapsed {
						label = "expand"
					}
					items = append(items, makePill(fk(m.km.ToggleCollapse)+" "+label))
				}

				if m.aiKey != "" {
					items = append(items, makePill(fk(m.km.AIPanelToggle)+" assistant"))
				}
				items = append(items, makePill(fk(m.km.Help)+" "+styles.IconHelp+"help"))
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

		focusPill := ""
		if m.foc.Active && (m.mode == ModeList || m.mode == ModeFocus) {
			pulseStyle := m.s.BadgeDoing
			if m.foc.State == focus.StateShortBreak || m.foc.State == focus.StateLongBreak {
				pulseStyle = m.s.BadgeGood
			}

			text := "DEEP WORK"
			if m.foc.State != focus.StateFocus {
				text = "BREAK"
			}

			// Pulse effect using RainbowAnimationOffset
			if m.RainbowAnimationOffset%2 == 0 {
				text = "• " + text + " •"
			} else {
				text = "  " + text + "  "
			}

			focusPill = m.s.TagLeft.Foreground(pulseStyle.GetBackground()).Render() +
				pulseStyle.Render(text) +
				m.s.TagRight.Foreground(pulseStyle.GetBackground()).Render() + " "
		}

		mcpStatus := ""
		if m.mcpRunning {
			mcpStatus = makePill("MCP "+styles.IconSuccess) + " "
		}
		right = focusPill + mcpStatus + makePill(syncLogo+versionText) + " "
	}

	return render.BarLine(left, right, m.width, m.s.Theme.Bg)
}

// renderResultOverlay renders the task result input modal
func (m *Model) renderResultOverlay(h int) string {
	inputLabel := m.s.Title.Render("Task Result")
	input := lipgloss.NewStyle().Padding(0, 1).Render(m.resultInput.View())

	modal := lipgloss.JoinVertical(lipgloss.Left,
		inputLabel,
		input,
		m.s.Muted.Render("Describe the outcome of this task"),
	)

	overlayW := min(60, m.width-4)
	card := m.s.Overlay.Width(overlayW).Render(modal)

	return lipgloss.Place(m.width, h, lipgloss.Center, lipgloss.Center, card,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceBackground(m.s.Theme.Bg),
	)
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

	overlayW := min(60, m.width-4)
	card := m.s.Overlay.Width(overlayW).Render(modal)

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

	// Apply project filter
	if m.activeProject != "" {
		f.Project = m.activeProject
	}

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

func (m *Model) loadProjectsCmd() tea.Cmd {
	return func() tea.Msg {
		projects, err := m.svc.ListProjects(m.ctx)
		if err != nil {
			return errMsg{Err: err}
		}
		return projectsLoadedMsg{Projects: projects}
	}
}

func (m *Model) refreshCmd() tea.Cmd {
	return tea.Batch(
		m.loadTasksCmd(),
		m.loadAllTasksCmd(),
		m.loadTagsCmd(),
		m.loadProjectsCmd(),
	)
}

func (m *Model) updateRecentProject(project string) {
	if project == "" {
		return
	}
	recent := m.cfg.App.RecentProjects
	// Remove if already exists
	for i, p := range recent {
		if p == project {
			recent = append(recent[:i], recent[i+1:]...)
			break
		}
	}
	// Prepend
	recent = append([]string{project}, recent...)
	// Limit to 50
	if len(recent) > 50 {
		recent = recent[:50]
	}
	m.cfg.App.RecentProjects = recent
}

func (m *Model) pruneAndLoadTagsCmd() tea.Cmd {
	return func() tea.Msg {
		_ = m.svc.Prune(m.ctx)
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
		m.hist.Record(history.CreateOperation(history.OpCreate, "", []string{created.ID}, nil, []core.Task{created.DeepCopy()}))
		return taskCreatedMsg{Task: created}
	}
}

func (m *Model) updateTaskCmd(id string, p core.TaskPatch) tea.Cmd {
	return func() tea.Msg {
		before, err := m.svc.GetByID(m.ctx, id)
		if err != nil {
			return errMsg{Err: err}
		}
		updated, err := m.svc.Update(m.ctx, id, p)
		if err != nil {
			return errMsg{Err: err}
		}

		opType := history.OpUpdate
		if p.Status != nil {
			opType = history.OpToggleStatus
		}
		m.hist.Record(history.CreateOperation(opType, "", []string{id}, []core.Task{before.DeepCopy()}, []core.Task{updated.DeepCopy()}))
		return taskUpdatedMsg{Task: updated}
	}
}

func (m *Model) createFocusSessionCmd(s core.FocusSession) tea.Cmd {
	return func() tea.Msg {
		if err := m.svc.Repo().CreateFocusSession(m.ctx, s); err != nil {
			return errMsg{Err: err}
		}
		return nil
	}
}

func (m *Model) deleteTaskCmd(id string) tea.Cmd {
	return func() tea.Msg {
		before, err := m.svc.GetByID(m.ctx, id)
		if err != nil {
			return errMsg{Err: err}
		}
		if err := m.svc.Delete(m.ctx, id); err != nil {
			return errMsg{Err: err}
		}
		m.hist.Record(history.CreateOperation(history.OpDelete, "", []string{id}, []core.Task{before.DeepCopy()}, nil))
		return taskDeletedMsg{ID: id}
	}
}

func (m *Model) deleteAllTasksCmd() tea.Cmd {
	return func() tea.Msg {
		before, _ := m.svc.ListAll(m.ctx)
		var ids []string
		var beforeCopy []core.Task
		for _, t := range before {
			ids = append(ids, t.ID)
			beforeCopy = append(beforeCopy, t.DeepCopy())
		}
		if err := m.svc.DeleteAll(m.ctx); err != nil {
			return errMsg{Err: err}
		}
		m.hist.Record(history.CreateOperation(history.OpBulkDelete, "Delete All", ids, beforeCopy, nil))
		return taskUpdatedMsg{} // Trigger reload
	}
}

func (m *Model) undoCmd(op *history.Operation) tea.Cmd {
	return func() tea.Msg {
		if err := m.applyOperation(op, true); err != nil {
			return errMsg{Err: err}
		}

		return statusMsg{
			Message: fmt.Sprintf("Undone: %s", history.GetOperationDescription(op)),
			IsErr:   false,
			Refresh: true,
		}
	}
}

func (m *Model) redoCmd(op *history.Operation) tea.Cmd {
	return func() tea.Msg {
		if err := m.applyOperation(op, false); err != nil {
			return errMsg{Err: err}
		}

		return statusMsg{
			Message: fmt.Sprintf("Redone: %s", history.GetOperationDescription(op)),
			IsErr:   false,
			Refresh: true,
		}
	}
}

func (m *Model) applyOperation(op *history.Operation, undo bool) error {
	switch op.Type {
	case history.OpCreate:
		if undo {
			return m.svc.DeleteTasks(m.ctx, op.TaskIDs)
		} else {
			for _, t := range op.After {
				if err := m.svc.UpsertTask(m.ctx, t); err != nil {
					return err
				}
			}
		}
	case history.OpDelete, history.OpBulkDelete:
		if undo {
			for _, t := range op.Before {
				if err := m.svc.UpsertTask(m.ctx, t); err != nil {
					return err
				}
			}
		} else {
			for _, id := range op.TaskIDs {
				if err := m.svc.Delete(m.ctx, id); err != nil {
					return err
				}
			}
		}
	default:
		tasks := op.After
		if undo {
			tasks = op.Before
		}
		for _, t := range tasks {
			t.UpdatedAt = time.Now()
			if err := m.svc.UpsertTask(m.ctx, t); err != nil {
				return err
			}
		}
	}
	return nil
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
	m.list = tasklist.New(m.s, m.cfg.App.VimMode, m.cfg.App.Animations, m.km, m.cfg.List.Fields.Due.Minimal)
	m.list.SetTagsConfig(m.cfg.Tags.Highlight)
	m.list.SetRightOrder(m.cfg.List.Order.Right)
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
		search.Item{ID: "cmd:onboarding", Kind: search.KindCommand, Title: "Welcome Tour", Hint: "ctrl+w"},
	)

	if m.aiKey != "" {
		items = append(items, search.Item{ID: "cmd:ai", Kind: search.KindCommand, Title: "AI Assistant", Hint: "ctrl+a"})
	}

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
		if t.OpenIssueID != "" {
			hint += " • " + t.OpenIssueID
		}
		if t.Responsible != "" {
			hint += " • @" + t.Responsible
		}
		title := t.Title
		if t.Result != "" {
			title = t.Title + " [✓ " + t.Result + "]"
		}
		items = append(items, search.Item{ID: t.ID, Kind: search.KindTask, Title: title, Desc: t.Description, Hint: hint})
	}

	m.palFullIdx = search.NewIndex(items)

	taskItems := make([]search.Item, 0, len(m.all))
	for _, t := range m.all {
		hint := string(t.Status)
		if t.Deadline != nil {
			hint += " • due " + t.Deadline.Local().Format("Jan 2")
		}
		if t.OpenIssueID != "" {
			hint += " • " + t.OpenIssueID
		}
		if t.Responsible != "" {
			hint += " • @" + t.Responsible
		}
		title := t.Title
		if t.Result != "" {
			title = t.Title + " [✓ " + t.Result + "]"
		}
		taskItems = append(taskItems, search.Item{ID: t.ID, Kind: search.KindTask, Title: title, Desc: t.Description, Hint: hint})
	}
	m.palTasksIdx = search.NewIndex(taskItems)
	m.rebuildProjectsIndex()
	m.applyPaletteIndex()
}

func (m *Model) rebuildProjectsIndex() {
	// Filter projects to only include those with tasks
	projectHasTasks := make(map[string]bool)
	for _, task := range m.all {
		if task.Project != "" {
			projectHasTasks[task.Project] = true
		}
	}

	items := make([]search.Item, 0, len(m.projects)+1)
	items = append(items, search.Item{ID: "", Kind: search.KindProject, Title: "<All Projects>", Hint: "show all tasks"})
	for _, p := range m.projects {
		// Only add project to switcher if it has tasks
		if projectHasTasks[p] {
			items = append(items, search.Item{ID: p, Kind: search.KindProject, Title: p, Hint: "project"})
		}
	}
	m.palProjectsIdx = search.NewIndex(items)
}

func (m *Model) applyPaletteIndex() {
	if m.mode == ModeProjectSwitcher {
		if m.palProjectsIdx != nil {
			m.pal.SetIndex(m.palProjectsIdx)
		}
		return
	}
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
		e := editor.New(m.s, editor.ModeNew, core.Task{Status: core.StatusTodo, Priority: core.P1}, m.cfg.Edit.Preview)
		m.edit = &e
		m.rebuildComponentSizes()
		m.mode = ModeEditor
		return m.edit.Init()
	case "cmd:theme":
		m.mode = ModeThemeMenu
		return nil
	case "cmd:ai":
		m.aiPanel.Toggle()
		m.aiPanel.SetSize(m.width, m.height)
		if m.aiPanel.Visible {
			return m.aiPanel.Init()
		}
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
	case "cmd:onboarding":
		m.onb = onboarding.New(m.s, m.km)
		m.onb.SetSize(m.width, m.height)
		m.mode = ModeOnboarding
		return m.onb.Init()
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

func (m *Model) listenStatusCmd() tea.Cmd {
	return func() tea.Msg {
		return <-m.statusCh
	}
}

func (m *Model) clearStatusCmd(id int) tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return clearStatusMsg{ID: id}
	})
}
func (m *Model) handleImportExportAction(action import_export_menu.Action, path string) tea.Cmd {
	return func() tea.Msg {
		taskAPI := api.New(m.svc)
		var resp api.Response

		if action.IsExport() {
			scope := ""
			if m.activeProject != "" && m.iem.ScopeValue() == import_export_menu.ScopeCurrentProject {
				scope = m.activeProject
			}
			req := api.Request{
				Action:  "export",
				Payload: []byte(fmt.Sprintf(`{"format":"%s","project":"%s"}`, action.Format(), scope)),
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
