package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"

	"github.com/programmersd21/kairo/internal/util"
)

const appName = "kairo"

type Config struct {
	App      AppConfig      `toml:"app"`
	Projects ProjectsConfig `toml:"projects"`
	Tags     TagsConfig     `toml:"tags"`
	Edit     EditConfig     `toml:"edit"`
	List     ListConfig     `toml:"list"`
	Theme    ThemeConfig    `toml:"theme"`
	Storage  StorageConfig  `toml:"storage"`
	Sync     SyncConfig     `toml:"sync"`
	Plugins  PluginsConfig  `toml:"plugins"`
	Keymap   KeymapConfig   `toml:"keymap"`
}

type EditConfig struct {
	Preview bool `toml:"preview"`
}

type TagsConfig struct {
	Highlight map[string]string `toml:"highlight"`
}

type ProjectsConfig struct {
	Default    string `toml:"default"`
	Shortcut   string `toml:"shortcut"`
	ShowColumn string `toml:"show_column"` // "auto" | "always" | "never"
	Order      string `toml:"order"`       // "alphabetical" | "recent"
}

type ListConfig struct {
	Order  ListOrderConfig  `toml:"order"`
	Fields ListFieldsConfig `toml:"fields"`
}

type ListFieldsConfig struct {
	Due DueFieldConfig `toml:"due"`
}

type DueFieldConfig struct {
	Minimal bool `toml:"minimal"`
}

type ListOrderConfig struct {
	Right []string `toml:"right"`
}

func normalizeRightOrder(in []string) []string {
	// Default: tags first (variable width), then due, then priority.
	def := []string{"tags", "due", "priority"}
	if len(in) == 0 {
		return def
	}
	allowed := map[string]struct{}{
		"tags":          {},
		"due":           {},
		"priority":      {},
		"project":       {},
		"open_issue_id": {},
	}
	seen := map[string]struct{}{}
	out := make([]string, 0, len(in))
	for _, raw := range in {
		k := strings.ToLower(strings.TrimSpace(raw))
		if _, ok := allowed[k]; !ok {
			continue
		}
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, k)
	}
	if len(out) == 0 {
		return def
	}
	return out
}

type AppConfig struct {
	Theme               string   `toml:"theme"`
	ActiveProject       string   `toml:"active_project"`
	VimMode             bool     `toml:"vim_mode"`
	ShowHelp            bool     `toml:"show_help"`
	ShowID              bool     `toml:"show_id"`
	Rainbow             bool     `toml:"rainbow"`
	GeminiAPIKey        string   `toml:"gemini_api_key"`
	AIModel             string   `toml:"ai_model"`
	MCPEnabled          bool     `toml:"mcp_enabled"`
	MCPPort             string   `toml:"mcp_port"`
	Animations          bool     `toml:"animations"`
	OnboardingCompleted bool     `toml:"onboarding_completed"`
	RecentProjects      []string `toml:"recent_projects"`
}

type StorageConfig struct {
	Path string `toml:"path"`
}

type ThemeConfig struct {
	Bg      string `toml:"bg"`
	Fg      string `toml:"fg"`
	Muted   string `toml:"muted"`
	Border  string `toml:"border"`
	Accent  string `toml:"accent"`
	Good    string `toml:"good"`
	Warn    string `toml:"warn"`
	Bad     string `toml:"bad"`
	Overlay string `toml:"overlay"`
}

type SyncConfig struct {
	Enabled  bool   `toml:"enabled"`
	RepoPath string `toml:"repo_path"`
	Remote   string `toml:"remote"`
	Branch   string `toml:"branch"`
	Strategy string `toml:"strategy"`
	AutoPush bool   `toml:"auto_push"`
}

type PluginsConfig struct {
	Enabled bool   `toml:"enabled"`
	Dir     string `toml:"dir"`
}

type KeymapConfig struct {
	Palette    string `toml:"palette"`
	TaskSearch string `toml:"task_search"`
	NewTask    string `toml:"new_task"`
	EditTask   string `toml:"edit_task"`
	DeleteTask string `toml:"delete_task"`
	OpenTask   string `toml:"open_task"`
	Back       string `toml:"back"`
	Quit       string `toml:"quit"`

	ViewInbox       string `toml:"view_inbox"`
	ViewToday       string `toml:"view_today"`
	ViewUpcoming    string `toml:"view_upcoming"`
	ViewCompleted   string `toml:"view_completed"`
	ViewTag         string `toml:"view_tag"`
	ViewPriority    string `toml:"view_priority"`
	CycleTheme      string `toml:"cycle_theme"`
	OpenPluginDir   string `toml:"open_plugin_dir"`
	ManagePlugins   string `toml:"manage_plugins"`
	ToggleStrike    string `toml:"toggle_strike"`
	ToggleCollapse  string `toml:"toggle_collapse"`
	DuplicateTask   string `toml:"duplicate_task"`
	Help            string `toml:"help"`
	WelcomeTour     string `toml:"welcome_tour"`
	Issues          string `toml:"issues"`
	Discussions     string `toml:"discussions"`
	Changelog       string `toml:"changelog"`
	Settings        string `toml:"settings"`
	ImportExport    string `toml:"import_export"`
	AIPanelToggle   string `toml:"ai_panel_toggle"`
	Stats           string `toml:"stats"`
	ProjectSwitcher string `toml:"project_switcher"`
	Undo            string `toml:"undo"`
	Redo            string `toml:"redo"`
	FocusSidebar    string `toml:"focus_sidebar"`
	FocusList       string `toml:"focus_list"`
}

func Default() Config {
	return Config{
		App: AppConfig{
			Theme:      "catppuccin",
			VimMode:    false,
			ShowHelp:   true,
			ShowID:     true,
			Rainbow:    false,
			AIModel:    "gemini-3.1-flash-lite-preview",
			MCPEnabled: false,
			MCPPort:    "8080",
			Animations: true,
		},
		Projects: ProjectsConfig{
			Default:    "default",
			Shortcut:   "p",
			ShowColumn: "auto",
			Order:      "alphabetical",
		},
		Edit: EditConfig{
			Preview: true,
		},
		List: ListConfig{
			Order: ListOrderConfig{
				Right: []string{"tags", "due", "priority", "open_issue_id"},
			},
			Fields: ListFieldsConfig{
				Due: DueFieldConfig{
					Minimal: true,
				},
			},
		},
		Theme: ThemeConfig{
			Bg:      "", // Use theme default
			Fg:      "",
			Muted:   "",
			Border:  "",
			Accent:  "",
			Good:    "",
			Warn:    "",
			Bad:     "",
			Overlay: "",
		},
		Storage: StorageConfig{
			Path: "kairo.db",
		},
		Sync: SyncConfig{
			Enabled:  false,
			RepoPath: "tasks-sync",
			Remote:   "origin",
			Branch:   "main",
			Strategy: "ours",
			AutoPush: true,
		},
		Plugins: PluginsConfig{
			Enabled: true,
			Dir:     "plugins",
		},
		Tags: TagsConfig{
			Highlight: make(map[string]string),
		},
		Keymap: KeymapConfig{
			Palette:         "ctrl+p",
			TaskSearch:      "/",
			NewTask:         "n",
			EditTask:        "e",
			DeleteTask:      "d",
			OpenTask:        "enter",
			Back:            "esc",
			Quit:            "q",
			ViewInbox:       "1",
			ViewToday:       "2",
			ViewUpcoming:    "3",
			ViewCompleted:   "4",
			ViewTag:         "f",
			ViewPriority:    "5",
			CycleTheme:      "t",
			OpenPluginDir:   "ctrl+g",
			ManagePlugins:   "p",
			ToggleStrike:    "z",
			ToggleCollapse:  "space",
			DuplicateTask:   "ctrl+d",
			Help:            "?",
			WelcomeTour:     "ctrl+w",
			Issues:          "i",
			Discussions:     "u",
			Changelog:       "c",
			Settings:        "ctrl+s",
			ImportExport:    "x",
			AIPanelToggle:   "ctrl+a",
			Stats:           "s",
			ProjectSwitcher: "ctrl+e",
			Undo:            "ctrl+z",
			Redo:            "ctrl+y",
			FocusSidebar:    "[",
			FocusList:       "]",
		},
	}
}

func (c Config) Save() error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}
	b, err := toml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func Load() (Config, error) {
	cfg := Default()

	path, err := ConfigPath()
	if err != nil {
		return cfg, err
	}

	// Try multiple locations
	var b []byte
	found := false

	// 1. Primary path (AppData/Roaming/kairo/config.toml or ~/.config/kairo/config.toml)
	if data, err := os.ReadFile(path); err == nil {
		b = data
		found = true
	}

	// 2. Fallback: ~/.kairo/config.toml (traditional CLI location)
	if !found {
		if home, err := os.UserHomeDir(); err == nil {
			fallback := filepath.Join(home, ".kairo", "config.toml")
			if data, err := os.ReadFile(fallback); err == nil {
				b = data
				found = true
			}
		}
	}

	// 3. Fallback: ~/.config/kairo/config.toml (explicit if AppDataDir failed to find it there)
	if !found {
		if home, err := os.UserHomeDir(); err == nil {
			fallback := filepath.Join(home, ".config", "kairo", "config.toml")
			if data, err := os.ReadFile(fallback); err == nil {
				b = data
				found = true
			}
		}
	}

	if !found {
		return cfg, nil
	}

	if err := toml.Unmarshal(b, &cfg); err != nil {
		return cfg, err
	}

	// Merge defaults for empty keybindings
	defaults := Default()
	if cfg.Keymap.Palette == "" {
		cfg.Keymap.Palette = defaults.Keymap.Palette
	}
	if cfg.Keymap.TaskSearch == "" {
		cfg.Keymap.TaskSearch = defaults.Keymap.TaskSearch
	}
	if cfg.Keymap.NewTask == "" {
		cfg.Keymap.NewTask = defaults.Keymap.NewTask
	}
	if cfg.Keymap.EditTask == "" {
		cfg.Keymap.EditTask = defaults.Keymap.EditTask
	}
	if cfg.Keymap.DeleteTask == "" {
		cfg.Keymap.DeleteTask = defaults.Keymap.DeleteTask
	}
	if cfg.Keymap.OpenTask == "" {
		cfg.Keymap.OpenTask = defaults.Keymap.OpenTask
	}
	if cfg.Keymap.Back == "" {
		cfg.Keymap.Back = defaults.Keymap.Back
	}
	if cfg.Keymap.Quit == "" {
		cfg.Keymap.Quit = defaults.Keymap.Quit
	}
	if cfg.Keymap.ViewInbox == "" {
		cfg.Keymap.ViewInbox = defaults.Keymap.ViewInbox
	}
	if cfg.Keymap.ViewToday == "" {
		cfg.Keymap.ViewToday = defaults.Keymap.ViewToday
	}
	if cfg.Keymap.ViewUpcoming == "" {
		cfg.Keymap.ViewUpcoming = defaults.Keymap.ViewUpcoming
	}
	if cfg.Keymap.ViewCompleted == "" {
		cfg.Keymap.ViewCompleted = defaults.Keymap.ViewCompleted
	}
	if cfg.Keymap.ViewTag == "" {
		cfg.Keymap.ViewTag = defaults.Keymap.ViewTag
	}
	if cfg.Keymap.ViewPriority == "" {
		cfg.Keymap.ViewPriority = defaults.Keymap.ViewPriority
	}
	if cfg.Keymap.CycleTheme == "" {
		cfg.Keymap.CycleTheme = defaults.Keymap.CycleTheme
	}
	if cfg.Keymap.OpenPluginDir == "" {
		cfg.Keymap.OpenPluginDir = defaults.Keymap.OpenPluginDir
	}
	if cfg.Keymap.ManagePlugins == "" {
		cfg.Keymap.ManagePlugins = defaults.Keymap.ManagePlugins
	}
	if cfg.Keymap.ProjectSwitcher == "" {
		cfg.Keymap.ProjectSwitcher = defaults.Keymap.ProjectSwitcher
	}
	if cfg.Keymap.ToggleStrike == "" {
		cfg.Keymap.ToggleStrike = defaults.Keymap.ToggleStrike
	}
	if cfg.Keymap.ToggleCollapse == "" {
		cfg.Keymap.ToggleCollapse = defaults.Keymap.ToggleCollapse
	}
	if cfg.Keymap.DuplicateTask == "" {
		cfg.Keymap.DuplicateTask = defaults.Keymap.DuplicateTask
	}
	if cfg.Keymap.Help == "" {
		cfg.Keymap.Help = defaults.Keymap.Help
	}
	if cfg.Keymap.Issues == "" {
		cfg.Keymap.Issues = defaults.Keymap.Issues
	}
	if cfg.Keymap.Discussions == "" {
		cfg.Keymap.Discussions = defaults.Keymap.Discussions
	}
	if cfg.Keymap.Changelog == "" {
		cfg.Keymap.Changelog = defaults.Keymap.Changelog
	}
	if cfg.Keymap.Settings == "" {
		cfg.Keymap.Settings = defaults.Keymap.Settings
	}
	if cfg.Keymap.ImportExport == "" {
		cfg.Keymap.ImportExport = defaults.Keymap.ImportExport
	}
	if cfg.Keymap.Stats == "" {
		cfg.Keymap.Stats = defaults.Keymap.Stats
	}

	appDir, _ := util.AppDataDir(appName)

	// Helper to resolve relative to app data dir
	resolve := func(p *string) {
		if *p != "" && !filepath.IsAbs(*p) {
			*p = filepath.Join(appDir, *p)
		}
	}

	resolve(&cfg.Storage.Path)
	resolve(&cfg.Sync.RepoPath)
	resolve(&cfg.Plugins.Dir)

	if cfg.Storage.Path == "" {
		cfg.Storage.Path = filepath.Join(appDir, "kairo.db")
	}
	if cfg.Plugins.Dir == "" {
		cfg.Plugins.Dir = filepath.Join(appDir, "plugins")
	}

	cfg.App.Theme = strings.TrimSpace(cfg.App.Theme)
	cfg.Sync.Strategy = strings.ToLower(strings.TrimSpace(cfg.Sync.Strategy))
	if cfg.Sync.Strategy == "" {
		cfg.Sync.Strategy = "ours"
	}

	// Validate/sanitize list ordering.
	cfg.List.Order.Right = normalizeRightOrder(cfg.List.Order.Right)

	// Keymap migrations.
	migrated := false
	switch strings.ToLower(strings.TrimSpace(cfg.Keymap.ManagePlugins)) {
	case "ctrl+alt+g", "alt+ctrl+g":
		// Legacy default; plugin manager is now bound to "p".
		cfg.Keymap.ManagePlugins = "p"
		migrated = true
	}
	if migrated {
		// Best-effort: keep the on-disk config in sync with new defaults.
		_ = cfg.Save()
	}
	return cfg, nil
}

func ConfigPath() (string, error) {
	d, err := util.AppDataDir(appName)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(d, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(d, "config.toml"), nil
}
