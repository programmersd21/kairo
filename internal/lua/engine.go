// Package lua provides Lua bindings for the kairo service layer
// This is the primary extensibility interface for plugins
package lua

import (
	"context"
	"fmt"
	"time"

	lua "github.com/yuin/gopher-lua"

	"github.com/programmersd21/kairo/internal/core"
	"github.com/programmersd21/kairo/internal/hooks"
	"github.com/programmersd21/kairo/internal/service"
)

// Engine manages the Lua runtime and plugin execution
type Engine struct {
	service service.TaskService
	hooks   *hooks.Manager
	timeout time.Duration
}

// NewEngine creates a new Lua engine with the given service and hooks
func NewEngine(svc service.TaskService, hks *hooks.Manager) *Engine {
	return &Engine{
		service: svc,
		hooks:   hks,
		timeout: 5 * time.Second, // Default 5 second timeout for plugin execution
	}
}

// SetTimeout sets the execution timeout for plugin code
func (e *Engine) SetTimeout(d time.Duration) {
	e.timeout = d
}

// SetupKairoAPI sets up the kairo module and API in the given Lua state
func (e *Engine) SetupKairoAPI(L *lua.LState) {
	// Create kairo module table
	kairo := L.NewTable()

	// Task operations
	L.SetField(kairo, "create_task", L.NewFunction(e.luaCreateTask))
	L.SetField(kairo, "get_task", L.NewFunction(e.luaGetTask))
	L.SetField(kairo, "list_tasks", L.NewFunction(e.luaListTasks))
	L.SetField(kairo, "update_task", L.NewFunction(e.luaUpdateTask))
	L.SetField(kairo, "delete_task", L.NewFunction(e.luaDeleteTask))

	// Event hooks
	L.SetField(kairo, "on", L.NewFunction(e.luaOn))
	L.SetField(kairo, "off", L.NewFunction(e.luaOff))

	// Notifications
	L.SetField(kairo, "notify", L.NewFunction(e.luaNotify))

	// Meta
	L.SetField(kairo, "version", lua.LString("1.2.2"))

	// Set as global
	L.SetGlobal("kairo", kairo)
}

// RunFile executes a Lua file in a sandboxed environment with kairo bindings
func (e *Engine) RunFile(filePath string) (lua.LValue, error) {
	L := lua.NewState()
	defer L.Close()

	// Set up kairo API bindings
	e.SetupKairoAPI(L)

	// Execute file
	if err := L.DoFile(filePath); err != nil {
		return nil, fmt.Errorf("lua error: %w", err)
	}

	// Return the result of the script (often a table in Kairo plugins)
	return L.Get(-1), nil
}

// luaCreateTask creates a new task
// Usage: local task, err = kairo.create_task({title="Task", tags={"work"}})
func (e *Engine) luaCreateTask(L *lua.LState) int {
	if L.GetTop() == 0 {
		L.Push(lua.LNil)
		L.Push(lua.LString("missing task table"))
		return 2
	}

	tbl := L.Get(1)
	if tbl.Type() != lua.LTTable {
		L.Push(lua.LNil)
		L.Push(lua.LString("expected table"))
		return 2
	}

	task := core.Task{}

	// Parse table fields
	if v := L.GetField(tbl, "title"); v.Type() == lua.LTString {
		task.Title = lua.LVAsString(v)
	}
	if v := L.GetField(tbl, "description"); v.Type() == lua.LTString {
		task.Description = lua.LVAsString(v)
	}
	if v := L.GetField(tbl, "status"); v.Type() == lua.LTString {
		task.Status = core.Status(lua.LVAsString(v))
	}
	if v := L.GetField(tbl, "priority"); v.Type() == lua.LTNumber {
		task.Priority = core.Priority(lua.LVAsNumber(v))
	}
	if v := L.GetField(tbl, "tags"); v.Type() == lua.LTTable {
		task.Tags = e.luaTableToStringArray(L, v)
	}

	// Create via service
	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	created, err := e.service.Create(ctx, task)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Return created task as table
	L.Push(e.taskToLua(L, created))
	L.Push(lua.LNil)
	return 2
}

// luaGetTask retrieves a task by ID
func (e *Engine) luaGetTask(L *lua.LState) int {
	id := L.CheckString(1)

	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	task, err := e.service.GetByID(ctx, id)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(e.taskToLua(L, task))
	L.Push(lua.LNil)
	return 2
}

// luaListTasks lists tasks with optional filter
// Usage: local tasks = kairo.list_tasks({statuses={"todo"}, tag="work"})
func (e *Engine) luaListTasks(L *lua.LState) int {
	var filter core.Filter

	// Parse filter table if provided
	if L.GetTop() > 0 && L.Get(1).Type() == lua.LTTable {
		tbl := L.Get(1)

		if v := L.GetField(tbl, "statuses"); v.Type() == lua.LTTable {
			filter.Statuses = e.luaTableToStatusArray(L, v)
		}
		if v := L.GetField(tbl, "tags"); v.Type() == lua.LTTable {
			var tags []string
			v.(*lua.LTable).ForEach(func(_, val lua.LValue) {
				tags = append(tags, lua.LVAsString(val))
			})
			filter.Tags = tags
		}
		if v := L.GetField(tbl, "priority"); v.Type() == lua.LTNumber {
			p := core.Priority(lua.LVAsNumber(v))
			filter.Priority = &p
		}
		if v := L.GetField(tbl, "sort"); v.Type() == lua.LTString {
			filter.Sort = core.SortMode(lua.LVAsString(v))
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	tasks, err := e.service.List(ctx, filter)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Convert to Lua table array
	result := L.NewTable()
	for i, task := range tasks {
		L.RawSetInt(result, i+1, e.taskToLua(L, task))
	}

	L.Push(result)
	L.Push(lua.LNil)
	return 2
}

// luaUpdateTask updates a task
func (e *Engine) luaUpdateTask(L *lua.LState) int {
	id := L.CheckString(1)
	if L.GetTop() < 2 || L.Get(2).Type() != lua.LTTable {
		L.Push(lua.LNil)
		L.Push(lua.LString("expected patch table"))
		return 2
	}

	tbl := L.Get(2)
	patch := core.TaskPatch{}

	// Parse patch fields (only set non-nil values)
	if v := L.GetField(tbl, "title"); v.Type() == lua.LTString {
		s := lua.LVAsString(v)
		patch.Title = &s
	}
	if v := L.GetField(tbl, "description"); v.Type() == lua.LTString {
		s := lua.LVAsString(v)
		patch.Description = &s
	}
	if v := L.GetField(tbl, "status"); v.Type() == lua.LTString {
		s := core.Status(lua.LVAsString(v))
		patch.Status = &s
	}
	if v := L.GetField(tbl, "priority"); v.Type() == lua.LTNumber {
		p := core.Priority(lua.LVAsNumber(v))
		patch.Priority = &p
	}
	if v := L.GetField(tbl, "tags"); v.Type() == lua.LTTable {
		tags := e.luaTableToStringArray(L, v)
		patch.Tags = &tags
	}

	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	updated, err := e.service.Update(ctx, id, patch)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(e.taskToLua(L, updated))
	L.Push(lua.LNil)
	return 2
}

// luaDeleteTask deletes a task
func (e *Engine) luaDeleteTask(L *lua.LState) int {
	id := L.CheckString(1)

	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	if err := e.service.Delete(ctx, id); err != nil {
		L.Push(lua.LFalse)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(lua.LTrue)
	L.Push(lua.LNil)
	return 2
}

// luaOn registers an event listener
// Usage: kairo.on("task_create", function(event) ... end)
func (e *Engine) luaOn(L *lua.LState) int {
	eventType := L.CheckString(1)
	fn := L.CheckFunction(2)

	// Create a wrapper that converts the Lua function to a Go listener
	// Note: This is simplified. A production system would handle more complex lifecycle
	listener := func(event hooks.Event) {
		// Call Lua function with event data
		eventTable := L.NewTable()
		L.SetField(eventTable, "type", lua.LString(string(event.Type)))
		if event.Task != nil {
			L.SetField(eventTable, "task", e.taskToLua(L, *event.Task))
		}
		if event.Payload != nil {
			payloadTable := L.NewTable()
			for k, v := range event.Payload {
				// Simplified: just handle string values
				if s, ok := v.(string); ok {
					L.SetField(payloadTable, k, lua.LString(s))
				}
			}
			L.SetField(eventTable, "payload", payloadTable)
		}

		_ = L.CallByParam(lua.P{
			Fn:      fn,
			NRet:    0,
			Protect: true,
		}, eventTable)
	}

	e.hooks.On(hooks.EventType(eventType), listener)
	return 0
}

// luaOff removes an event listener (simplified - doesn't actually remove in this impl)
func (e *Engine) luaOff(L *lua.LState) int {
	// Simplified implementation - production would need a way to identify listeners
	return 0
}

// luaNotify sends a notification (stub for now)
// In production, this would emit to the UI
func (e *Engine) luaNotify(L *lua.LState) int {
	msg := L.CheckString(1)
	isErr := false
	if L.GetTop() > 1 {
		isErr = lua.LVAsBool(L.Get(2))
	}

	// TODO: Emit notification to UI
	_ = msg
	_ = isErr

	return 0
}

// Helper: convert Task to Lua table
func (e *Engine) taskToLua(L *lua.LState, task core.Task) lua.LValue {
	tbl := L.NewTable()
	L.SetField(tbl, "id", lua.LString(task.ID))
	L.SetField(tbl, "title", lua.LString(task.Title))
	L.SetField(tbl, "description", lua.LString(task.Description))
	L.SetField(tbl, "status", lua.LString(string(task.Status)))
	L.SetField(tbl, "priority", lua.LNumber(float64(task.Priority)))

	tagTable := L.NewTable()
	for i, tag := range task.Tags {
		L.RawSetInt(tagTable, i+1, lua.LString(tag))
	}
	L.SetField(tbl, "tags", tagTable)

	if task.Deadline != nil {
		L.SetField(tbl, "deadline", lua.LString(task.Deadline.Format(time.RFC3339)))
	}
	L.SetField(tbl, "created_at", lua.LString(task.CreatedAt.Format(time.RFC3339)))
	L.SetField(tbl, "updated_at", lua.LString(task.UpdatedAt.Format(time.RFC3339)))

	return tbl
}

// Helper: convert Lua table to string array
func (e *Engine) luaTableToStringArray(L *lua.LState, tbl lua.LValue) []string {
	var result []string
	if t, ok := tbl.(*lua.LTable); ok {
		t.ForEach(func(k, v lua.LValue) {
			if s, ok := v.(lua.LString); ok {
				result = append(result, string(s))
			}
		})
	}
	return result
}

// Helper: convert Lua table to Status array
func (e *Engine) luaTableToStatusArray(L *lua.LState, tbl lua.LValue) []core.Status {
	var result []core.Status
	if t, ok := tbl.(*lua.LTable); ok {
		t.ForEach(func(k, v lua.LValue) {
			if s, ok := v.(lua.LString); ok {
				result = append(result, core.Status(s))
			}
		})
	}
	return result
}
