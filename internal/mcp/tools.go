package mcp

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/programmersd21/kairo/internal/api"
)

// runAPI is a helper that routes arguments to the unified api.TaskAPI
func runAPI(ctx context.Context, action string, args map[string]interface{}) (*mcp.CallToolResult, error) {
	if args == nil {
		args = make(map[string]interface{})
	}

	// Convert "tags" from comma-separated string to slice if needed
	if tagsStr, ok := args["tags"].(string); ok && tagsStr != "" {
		tags := []string{}
		for _, t := range strings.Split(tagsStr, ",") {
			tag := strings.TrimSpace(t)
			if tag != "" {
				tags = append(tags, tag)
			}
		}
		args["tags"] = tags
	} else if ok && tagsStr == "" {
		args["tags"] = []string{}
	}

	b, _ := json.Marshal(args)
	taskAPI := api.New(globalSvc)
	resp := taskAPI.Execute(ctx, api.Request{Action: action, Payload: b})

	if !resp.Success {
		return mcp.NewToolResultError(resp.Error), nil
	}

	out, _ := json.MarshalIndent(resp.Data, "", "  ")
	return mcp.NewToolResultText(string(out)), nil
}

func CreateTaskHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, _ := req.Params.Arguments.(map[string]interface{})
	return runAPI(ctx, "create", args)
}

func UpdateTaskHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, _ := req.Params.Arguments.(map[string]interface{})
	return runAPI(ctx, "update", args)
}

func DeleteTaskHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, _ := req.Params.Arguments.(map[string]interface{})
	return runAPI(ctx, "delete", args)
}

func ListTasksHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, _ := req.Params.Arguments.(map[string]interface{})
	return runAPI(ctx, "list", args)
}

func SetThemeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, _ := request.Params.Arguments.(map[string]interface{})
	theme, _ := args["theme"].(string)

	apiClient := api.New(globalSvc)
	payload, _ := json.Marshal(map[string]interface{}{"theme": theme})
	resp := apiClient.Execute(ctx, api.Request{Action: "set_theme", Payload: payload})

	if !resp.Success {
		return mcp.NewToolResultError("Failed to set theme: " + resp.Error), nil
	}

	return mcp.NewToolResultText("Theme updated successfully to " + theme), nil
}

func PluginListHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	apiClient := api.New(globalSvc)
	resp := apiClient.Execute(ctx, api.Request{Action: "plugin_list", Payload: []byte("{}")})
	if !resp.Success {
		return mcp.NewToolResultError("Failed to list plugins: " + resp.Error), nil
	}
	pluginsBytes, _ := json.MarshalIndent(resp.Data, "", "  ")
	return mcp.NewToolResultText(string(pluginsBytes)), nil
}

func PluginGetHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, _ := request.Params.Arguments.(map[string]interface{})
	name, _ := args["name"].(string)
	apiClient := api.New(globalSvc)
	payload, _ := json.Marshal(map[string]interface{}{"name": name})
	resp := apiClient.Execute(ctx, api.Request{Action: "plugin_get", Payload: payload})
	if !resp.Success {
		return mcp.NewToolResultError("Failed to get plugin: " + resp.Error), nil
	}
	return mcp.NewToolResultText(resp.Data.(string)), nil
}

func PluginWriteHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, _ := request.Params.Arguments.(map[string]interface{})
	name, _ := args["name"].(string)
	content, _ := args["content"].(string)
	apiClient := api.New(globalSvc)
	payload, _ := json.Marshal(map[string]interface{}{"name": name, "content": content})
	resp := apiClient.Execute(ctx, api.Request{Action: "plugin_write", Payload: payload})
	if !resp.Success {
		return mcp.NewToolResultError("Failed to write plugin: " + resp.Error), nil
	}
	return mcp.NewToolResultText(resp.Data.(string)), nil
}

func PluginDeleteHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, _ := request.Params.Arguments.(map[string]interface{})
	name, _ := args["name"].(string)
	apiClient := api.New(globalSvc)
	payload, _ := json.Marshal(map[string]interface{}{"name": name})
	resp := apiClient.Execute(ctx, api.Request{Action: "plugin_delete", Payload: payload})
	if !resp.Success {
		return mcp.NewToolResultError("Failed to delete plugin: " + resp.Error), nil
	}
	return mcp.NewToolResultText(resp.Data.(string)), nil
}
