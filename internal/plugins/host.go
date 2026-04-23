package plugins

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	lua "github.com/yuin/gopher-lua"

	"github.com/programmersd21/kairo/internal/core"
	klua "github.com/programmersd21/kairo/internal/lua"
	"github.com/programmersd21/kairo/internal/service"
)

type PluginInfo struct {
	ID          string
	Name        string
	Description string
	Author      string
	Version     string
	Path        string
}

type CommandInfo struct {
	PluginID string
	ID       string // full ID: plugin:<pluginID>:<cmdID>
	Title    string
	Hint     string
}

type ViewInfo struct {
	PluginID string
	ID       string // full ID: plugin:<pluginID>:<viewID>
	Title    string
	Filter   core.Filter
}

type handlerRef struct {
	Path  string
	CmdID string
}

type Host struct {
	service service.TaskService
	dir     string

	mu       sync.RWMutex
	plugins  []PluginInfo
	cmds     []CommandInfo
	views    []ViewInfo
	handlers map[string]handlerRef // fullID -> handler reference
	lastErr  error

	watcher *fsnotify.Watcher

	// For API feedback
	notifyFunc func(string, bool)
}

func New(svc service.TaskService, dir string) *Host {
	return &Host{
		service: svc,
		dir:     dir,
	}
}

func (h *Host) Enabled() bool { return strings.TrimSpace(h.dir) != "" }

func (h *Host) SetNotifyFunc(f func(string, bool)) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.notifyFunc = f
}

func (h *Host) LastError() error {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.lastErr
}

func (h *Host) Commands() []CommandInfo {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return append([]CommandInfo(nil), h.cmds...)
}

func (h *Host) Plugins() []PluginInfo {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return append([]PluginInfo(nil), h.plugins...)
}

func (h *Host) Views() []ViewInfo {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return append([]ViewInfo(nil), h.views...)
}

func (h *Host) LoadAll() error {
	if !h.Enabled() {
		return nil
	}
	if err := os.MkdirAll(h.dir, 0o755); err != nil {
		return err
	}
	ents, err := os.ReadDir(h.dir)
	if err != nil {
		return err
	}
	sort.Slice(ents, func(i, j int) bool { return ents[i].Name() < ents[j].Name() })

	var plugins []PluginInfo
	var cmds []CommandInfo
	var views []ViewInfo
	handlers := map[string]handlerRef{}

	for _, ent := range ents {
		if ent.IsDir() || !strings.HasSuffix(strings.ToLower(ent.Name()), ".lua") {
			continue
		}
		path := filepath.Join(h.dir, ent.Name())
		info, pc, pv, ph, err := h.loadOne(path)
		if err != nil {
			h.setErr(err)
			continue
		}
		plugins = append(plugins, info)
		cmds = append(cmds, pc...)
		views = append(views, pv...)
		for k, v := range ph {
			handlers[k] = v
		}
	}

	h.mu.Lock()
	h.plugins = plugins
	h.cmds = cmds
	h.views = views
	h.handlers = handlers
	h.lastErr = nil
	h.mu.Unlock()
	return nil
}

func (h *Host) DeletePlugin(id string) error {
	h.mu.RLock()
	var path string
	for _, p := range h.plugins {
		if p.ID == id {
			path = p.Path
			break
		}
	}
	h.mu.RUnlock()

	if path == "" {
		return errors.New("plugin not found")
	}

	if err := os.Remove(path); err != nil {
		return err
	}

	return h.LoadAll()
}

func (h *Host) setErr(err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.lastErr = err
}

func (h *Host) RunCommand(ctx context.Context, fullID string) error {
	h.mu.RLock()
	ref, ok := h.handlers[fullID]
	h.mu.RUnlock()
	if !ok {
		return errors.New("plugin command not found")
	}

	L := lua.NewState()
	defer L.Close()

	eng := klua.NewEngine(h.service, h.service.Hooks())
	eng.SetupKairoAPI(L)

	if err := L.DoFile(ref.Path); err != nil {
		return err
	}
	ret := L.Get(-1)
	tbl, ok := ret.(*lua.LTable)
	if !ok {
		return errors.New("plugin must return a table")
	}
	ctbl, _ := tbl.RawGetString("commands").(*lua.LTable)
	if ctbl == nil {
		return errors.New("plugin has no commands")
	}

	run := lua.LNil
	ctbl.ForEach(func(_ lua.LValue, v lua.LValue) {
		if run.Type() == lua.LTFunction {
			return
		}
		c, ok := v.(*lua.LTable)
		if !ok {
			return
		}
		if luaToString(c.RawGetString("id")) != ref.CmdID {
			return
		}
		run = c.RawGetString("run")
	})
	if run.Type() != lua.LTFunction {
		return errors.New("plugin command missing run()")
	}
	L.Push(run)
	return L.PCall(0, 0, nil)
}

func (h *Host) Watch(ctx context.Context, onChange func()) error {
	if !h.Enabled() {
		return nil
	}
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	h.watcher = w
	if err := w.Add(h.dir); err != nil {
		_ = w.Close()
		return err
	}
	go func() {
		defer func() { _ = w.Close() }()
		debounce := time.NewTimer(0)
		if !debounce.Stop() {
			<-debounce.C
		}
		pending := false
		for {
			select {
			case <-ctx.Done():
				return
			case ev := <-w.Events:
				if ev.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename|fsnotify.Remove) == 0 {
					continue
				}
				if !strings.HasSuffix(strings.ToLower(ev.Name), ".lua") {
					continue
				}
				if !pending {
					pending = true
					debounce.Reset(250 * time.Millisecond)
				}
			case <-debounce.C:
				pending = false
				_ = h.LoadAll()
				if onChange != nil {
					onChange()
				}
			case err := <-w.Errors:
				if err != nil {
					h.setErr(err)
				}
			}
		}
	}()
	return nil
}

func (h *Host) loadOne(path string) (PluginInfo, []CommandInfo, []ViewInfo, map[string]handlerRef, error) {
	L := lua.NewState()
	defer L.Close()

	eng := klua.NewEngine(h.service, h.service.Hooks())
	eng.SetupKairoAPI(L)

	if err := L.DoFile(path); err != nil {
		return PluginInfo{}, nil, nil, nil, fmt.Errorf("plugin %s: %w", filepath.Base(path), err)
	}
	ret := L.Get(-1)
	tbl, ok := ret.(*lua.LTable)
	if !ok {
		return PluginInfo{}, nil, nil, nil, fmt.Errorf("plugin %s: must return a table", filepath.Base(path))
	}

	pluginID := luaToString(tbl.RawGetString("id"))
	if pluginID == "" {
		pluginID = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	}

	info := PluginInfo{
		ID:          pluginID,
		Name:        luaToString(tbl.RawGetString("name")),
		Description: luaToString(tbl.RawGetString("description")),
		Author:      luaToString(tbl.RawGetString("author")),
		Version:     luaToString(tbl.RawGetString("version")),
		Path:        path,
	}
	if info.Name == "" {
		info.Name = info.ID
	}

	var cmds []CommandInfo
	var views []ViewInfo
	handlers := map[string]handlerRef{}

	if ctbl, ok := tbl.RawGetString("commands").(*lua.LTable); ok {
		ctbl.ForEach(func(_ lua.LValue, v lua.LValue) {
			c, ok := v.(*lua.LTable)
			if !ok {
				return
			}
			id := luaToString(c.RawGetString("id"))
			title := luaToString(c.RawGetString("title"))
			hint := luaToString(c.RawGetString("hint"))
			run := c.RawGetString("run")
			if id == "" || title == "" || run.Type() != lua.LTFunction {
				return
			}
			full := "plugin:" + pluginID + ":" + id
			cmds = append(cmds, CommandInfo{PluginID: pluginID, ID: full, Title: title, Hint: hint})
			handlers[full] = handlerRef{Path: path, CmdID: id}
		})
	}

	if vtbl, ok := tbl.RawGetString("views").(*lua.LTable); ok {
		vtbl.ForEach(func(_ lua.LValue, v lua.LValue) {
			vt, ok := v.(*lua.LTable)
			if !ok {
				return
			}
			id := luaToString(vt.RawGetString("id"))
			title := luaToString(vt.RawGetString("title"))
			filterTbl, _ := vt.RawGetString("filter").(*lua.LTable)
			if id == "" || title == "" || filterTbl == nil {
				return
			}
			f := core.Filter{IncludeNilDeadline: true}
			if sts, ok := filterTbl.RawGetString("statuses").(*lua.LTable); ok {
				var ss []core.Status
				sts.ForEach(func(_ lua.LValue, vv lua.LValue) {
					s := strings.ToLower(luaToString(vv))
					if s != "" {
						ss = append(ss, core.Status(s))
					}
				})
				f.Statuses = ss
			}
			if tagsTbl, ok := filterTbl.RawGetString("tags").(*lua.LTable); ok {
				var tags []string
				tagsTbl.ForEach(func(_ lua.LValue, vv lua.LValue) {
					t := core.NormalizeTag(luaToString(vv))
					if t != "" {
						tags = append(tags, t)
					}
				})
				f.Tags = tags
			}
			if minP := luaToInt(filterTbl.RawGetString("min_priority")); minP >= 0 {
				p := core.Priority(minP).Clamp()
				f.Priority = &p
			}
			if s := luaToString(filterTbl.RawGetString("sort")); s != "" {
				f.Sort = core.SortMode(s)
			}
			full := "plugin:" + pluginID + ":" + id
			views = append(views, ViewInfo{PluginID: pluginID, ID: full, Title: title, Filter: f})
		})
	}

	return info, cmds, views, handlers, nil
}

func luaToString(v lua.LValue) string {
	if v == nil || v.Type() == lua.LTNil {
		return ""
	}
	return v.String()
}

func luaToInt(v lua.LValue) int {
	switch x := v.(type) {
	case lua.LNumber:
		return int(x)
	default:
		return -1
	}
}
