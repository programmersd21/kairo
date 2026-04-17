package keymap

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"

	"github.com/programmersd21/kairo/internal/config"
)

type Keymap struct {
	Palette    key.Binding
	TaskSearch key.Binding
	NewTask    key.Binding
	EditTask   key.Binding
	DeleteTask key.Binding
	OpenTask   key.Binding
	Back       key.Binding
	Quit       key.Binding

	ViewInbox     key.Binding
	ViewToday     key.Binding
	ViewUpcoming  key.Binding
	ViewCompleted key.Binding
	ViewTag       key.Binding
	ViewPriority  key.Binding

	CycleTheme    key.Binding
	OpenPluginDir key.Binding
	ManagePlugins key.Binding
	ToggleStrike  key.Binding
	Help          key.Binding
}

func FromConfig(c config.KeymapConfig) Keymap {
	return Keymap{
		Palette:    bind(c.Palette, "palette", "command palette"),
		TaskSearch: bind(c.TaskSearch, "search", "search tasks"),
		NewTask:    bind(c.NewTask, "new", "new task"),
		EditTask:   bind(c.EditTask, "edit", "edit task"),
		DeleteTask: bind(c.DeleteTask, "delete", "delete task"),
		OpenTask:   bind(c.OpenTask, "open", "open task"),
		Back:       bind(c.Back, "back", "back"),
		Quit:       bind(c.Quit, "quit", "quit"),

		ViewInbox:     bind(c.ViewInbox, "inbox", "inbox view"),
		ViewToday:     bind(c.ViewToday, "today", "today view"),
		ViewUpcoming:  bind(c.ViewUpcoming, "upcoming", "upcoming view"),
		ViewCompleted: bind(c.ViewCompleted, "completed", "completed view"),
		ViewTag:       bind(c.ViewTag, "tag", "tag view"),
		ViewPriority:  bind(c.ViewPriority, "priority", "priority view"),

		CycleTheme:    bind(c.CycleTheme, "theme", "theme menu"),
		OpenPluginDir: bind(c.OpenPluginDir, "plugins-dir", "open plugins folder"),
		ManagePlugins: bind(c.ManagePlugins, "manage-plugins", "manage plugins"),
		ToggleStrike:  bind(c.ToggleStrike, "strike", "toggle completion with animation"),
		Help:          bind(c.Help, "help", "show help"),
	}
}

func bind(keys, helpKey, helpDesc string) key.Binding {
	ks := parseKeys(keys)
	return key.NewBinding(key.WithKeys(ks...), key.WithHelp(helpKey, helpDesc))
}

func parseKeys(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(strings.ToLower(p))
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
