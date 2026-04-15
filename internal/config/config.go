package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"

	"github.com/programmersd21/kairo/internal/util"
)

const appName = "kairo"

type Config struct {
	App     AppConfig     `toml:"app"`
	Theme   ThemeConfig   `toml:"theme"`
	Storage StorageConfig `toml:"storage"`
	Sync    SyncConfig    `toml:"sync"`
	Plugins PluginsConfig `toml:"plugins"`
	Keymap  KeymapConfig  `toml:"keymap"`
}

type AppConfig struct {
	Theme    string `toml:"theme"`
	VimMode  bool   `toml:"vim_mode"`
	ShowHelp bool   `toml:"show_help"`
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
	NewTask    string `toml:"new_task"`
	EditTask   string `toml:"edit_task"`
	DeleteTask string `toml:"delete_task"`
	OpenTask   string `toml:"open_task"`
	Back       string `toml:"back"`
	Quit       string `toml:"quit"`

	ViewInbox    string `toml:"view_inbox"`
	ViewToday    string `toml:"view_today"`
	ViewUpcoming string `toml:"view_upcoming"`
	ViewTag      string `toml:"view_tag"`
	ViewPriority string `toml:"view_priority"`
	CycleTheme   string `toml:"cycle_theme"`
	Help         string `toml:"help"`
}

func Default() Config {
	return Config{
		App: AppConfig{
			Theme:    "midnight",
			VimMode:  false,
			ShowHelp: true,
		},
		Theme:   ThemeConfig{},
		Storage: StorageConfig{Path: ""},
		Sync: SyncConfig{
			Enabled:  false,
			RepoPath: "",
			Remote:   "origin",
			Branch:   "main",
			Strategy: "ours",
			AutoPush: true,
		},
		Plugins: PluginsConfig{
			Enabled: true,
			Dir:     "",
		},
		Keymap: KeymapConfig{
			Palette:      "ctrl+p",
			NewTask:      "n",
			EditTask:     "e",
			DeleteTask:   "d",
			OpenTask:     "enter",
			Back:         "esc",
			Quit:         "q",
			ViewInbox:    "1",
			ViewToday:    "2",
			ViewUpcoming: "3",
			ViewTag:      "4",
			ViewPriority: "5",
			CycleTheme:   "t",
			Help:         "?",
		},
	}
}

func (c Config) Save() error {
	path, err := configPath()
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

	path, err := configPath()
	if err != nil {
		return cfg, err
	}

	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return cfg, err
	}
	if err := toml.Unmarshal(b, &cfg); err != nil {
		return cfg, err
	}

	cfg.App.Theme = strings.TrimSpace(cfg.App.Theme)
	cfg.Sync.Strategy = strings.ToLower(strings.TrimSpace(cfg.Sync.Strategy))
	if cfg.Sync.Strategy == "" {
		cfg.Sync.Strategy = "ours"
	}
	return cfg, nil
}

func configPath() (string, error) {
	d, err := util.AppDataDir(appName)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(d, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(d, "config.toml"), nil
}
