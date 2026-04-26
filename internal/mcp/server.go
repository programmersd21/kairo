package mcp

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/programmersd21/kairo/internal/buildinfo"
	"github.com/programmersd21/kairo/internal/service"
)

var globalSvc service.TaskService

func NewServer(svc service.TaskService) *server.MCPServer {
	globalSvc = svc
	s := server.NewMCPServer(
		"kairo-mcp",
		buildinfo.EffectiveVersion(),
		server.WithToolCapabilities(true),
	)

	s.AddTool(mcp.NewTool("kairo_create_task",
		mcp.WithDescription("Create a new task"),
		mcp.WithString("title", mcp.Required(), mcp.Description("Title of the task")),
		mcp.WithString("description", mcp.Description("Markdown description")),
		mcp.WithNumber("priority", mcp.Description("0 (highest) to 3 (lowest)")),
		mcp.WithString("status", mcp.Description("'todo', 'doing', or 'done'")),
		mcp.WithString("deadline", mcp.Description("RFC3339 formatted date-time string")),
		mcp.WithString("tags", mcp.Description("Comma-separated list of tags")),
	), CreateTaskHandler)

	s.AddTool(mcp.NewTool("kairo_update_task",
		mcp.WithDescription("Update an existing task"),
		mcp.WithString("id", mcp.Required(), mcp.Description("ID of the task to update")),
		mcp.WithString("title", mcp.Description("New title")),
		mcp.WithString("description", mcp.Description("New markdown description")),
		mcp.WithNumber("priority", mcp.Description("0-3")),
		mcp.WithString("status", mcp.Description("'todo', 'doing', or 'done'")),
		mcp.WithString("deadline", mcp.Description("RFC3339 formatted date-time string (or empty string to clear)")),
		mcp.WithString("tags", mcp.Description("Comma-separated list of tags")),
	), UpdateTaskHandler)

	s.AddTool(mcp.NewTool("kairo_list_tasks",
		mcp.WithDescription("List all tasks"),
	), ListTasksHandler)

	s.AddTool(mcp.NewTool("kairo_delete_task",
		mcp.WithDescription("Delete a task by ID"),
		mcp.WithString("id", mcp.Required()),
	), DeleteTaskHandler)

	s.AddTool(mcp.NewTool("kairo_set_theme",
		mcp.WithDescription("Change the UI theme of Kairo"),
		mcp.WithString("theme", mcp.Required(), mcp.Description("Name of the theme (e.g. catppuccin, dracula, nord, midnight, etc.)")),
	), SetThemeHandler)

	s.AddTool(mcp.NewTool("kairo_plugin_list",
		mcp.WithDescription("List all installed plugins"),
	), PluginListHandler)

	s.AddTool(mcp.NewTool("kairo_plugin_get",
		mcp.WithDescription("Get the source code of a plugin"),
		mcp.WithString("name", mcp.Required(), mcp.Description("Name of the plugin file (e.g. hello.lua)")),
	), PluginGetHandler)

	s.AddTool(mcp.NewTool("kairo_plugin_write",
		mcp.WithDescription("Create or update a plugin's source code"),
		mcp.WithString("name", mcp.Required(), mcp.Description("Name of the plugin file (e.g. myplugin.lua)")),
		mcp.WithString("content", mcp.Required(), mcp.Description("Lua source code for the plugin")),
	), PluginWriteHandler)

	s.AddTool(mcp.NewTool("kairo_plugin_delete",
		mcp.WithDescription("Delete a plugin"),
		mcp.WithString("name", mcp.Required(), mcp.Description("Name of the plugin file")),
	), PluginDeleteHandler)

	return s
}
