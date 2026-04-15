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
	"github.com/programmersd21/kairo/internal/storage"
)

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
	repo *storage.Repository
	dir  string

	mu       sync.RWMutex
	cmds     []CommandInfo
	views    []ViewInfo
	handlers map[string]handlerRef // fullID -> handler reference
	lastErr  error

	watcher *fsnotify.Watcher
}

func New(repo *storage.Repository, dir string) *Host { return &Host{repo: repo, dir: dir} }

func (h *Host) Enabled() bool { return strings.TrimSpace(h.dir) != "" }

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

	var cmds []CommandInfo
	var views []ViewInfo
	handlers := map[string]handlerRef{}

	for _, ent := range ents {
		if ent.IsDir() || !strings.HasSuffix(strings.ToLower(ent.Name()), ".lua") {
			continue
		}
		path := filepath.Join(h.dir, ent.Name())
		pc, pv, ph, err := h.loadOne(path)
		if err != nil {
			h.setErr(err)
			continue
		}
		cmds = append(cmds, pc...)
		views = append(views, pv...)
		for k, v := range ph {
			handlers[k] = v
		}
	}

	h.mu.Lock()
	h.cmds = cmds
	h.views = views
	h.handlers = handlers
	h.lastErr = nil
	h.mu.Unlock()
	return nil
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

	L := lua.NewState(lua.Options{SkipOpenLibs: true})
	defer L.Close()
	lua.OpenBase(L)
	lua.OpenTable(L)
	lua.OpenString(L)
	lua.OpenMath(L)

	h.registerAPI(L, ctx)

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

	var run lua.LValue = lua.LNil
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

func (h *Host) loadOne(path string) ([]CommandInfo, []ViewInfo, map[string]handlerRef, error) {
	L := lua.NewState(lua.Options{SkipOpenLibs: true})
	defer L.Close()
	lua.OpenBase(L)
	lua.OpenTable(L)
	lua.OpenString(L)
	lua.OpenMath(L)

	h.registerAPI(L, context.Background())

	if err := L.DoFile(path); err != nil {
		return nil, nil, nil, fmt.Errorf("plugin %s: %w", filepath.Base(path), err)
	}
	ret := L.Get(-1)
	tbl, ok := ret.(*lua.LTable)
	if !ok {
		return nil, nil, nil, fmt.Errorf("plugin %s: must return a table", filepath.Base(path))
	}

	pluginID := luaToString(tbl.RawGetString("id"))
	if pluginID == "" {
		pluginID = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
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
			if tag := luaToString(filterTbl.RawGetString("tag")); tag != "" {
				f.Tag = core.NormalizeTag(tag)
			}
			if minP := luaToInt(filterTbl.RawGetString("min_priority")); minP >= 0 {
				p := core.Priority(minP).Clamp()
				f.Priority = &p
			}
			full := "plugin:" + pluginID + ":" + id
			views = append(views, ViewInfo{PluginID: pluginID, ID: full, Title: title, Filter: f})
		})
	}

	return cmds, views, handlers, nil
}

func (h *Host) registerAPI(L *lua.LState, ctx context.Context) {
	k := L.NewTable()
	L.SetGlobal("kairo", k)

	L.SetField(k, "create_task", L.NewFunction(func(L *lua.LState) int {
		title := L.CheckString(1)
		desc := L.OptString(2, "")
		var tags []string
		if tbl, ok := L.Get(3).(*lua.LTable); ok {
			tbl.ForEach(func(_ lua.LValue, v lua.LValue) {
				tags = append(tags, core.NormalizeTag(luaToString(v)))
			})
		}
		priority := core.Priority(L.OptInt(4, int(core.P1))).Clamp()
		status := core.Status(strings.ToLower(L.OptString(5, string(core.StatusTodo))))
		task := core.Task{Title: title, Description: desc, Tags: tags, Priority: priority, Status: status}
		_, _ = h.repo.CreateTask(ctx, task)
		return 0
	}))

	L.SetField(k, "delete_task", L.NewFunction(func(L *lua.LState) int {
		id := L.CheckString(1)
		_ = h.repo.DeleteTask(ctx, id)
		return 0
	}))
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
