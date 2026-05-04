package ai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/programmersd21/kairo/internal/api"
	"github.com/programmersd21/kairo/internal/service"
	"google.golang.org/genai"
)

var globalSvc service.TaskService

func SetService(svc service.TaskService) {
	globalSvc = svc
}

func GetKairoTools() *genai.Tool {
	return &genai.Tool{
		FunctionDeclarations: []*genai.FunctionDeclaration{
			{
				Name:        "kairo_api",
				Description: "Execute a command against the Kairo API. This gives you TOTAL control over the app, tasks, settings, themes, plugins, and database. You can manage tasks fully, change the UI theme, manage Lua plugins, configure AI settings, and export/import data.",
				Parameters: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"action": {
							Type:        genai.TypeString,
							Description: "Action to perform: create, get, update, delete, delete_all, list, list_tags, export, import, cleanup, configure-ai, set_theme, plugin_list, plugin_get, plugin_write, plugin_delete",
						},
						"payload": {
							Type:        genai.TypeObject,
							Description: "JSON payload for the action. For 'set_theme', use 'theme' (string). For 'plugin_list', use {}. For 'plugin_get/delete', use 'name' (string). For 'plugin_write', use 'name' and 'content' (string). Tasks use title, description, tags, priority, status, deadline, recurrence (none|weekly|monthly), recurrence_weekly (array of strings), recurrence_monthly (number).",
						},
					},
					Required: []string{"action", "payload"},
				},
			},
		},
	}
}

func ExecuteTool(ctx context.Context, name string, args map[string]interface{}) (interface{}, error) {
	if globalSvc == nil {
		return nil, fmt.Errorf("service not initialized")
	}

	if name == "kairo_api" {
		action, ok := args["action"].(string)
		if !ok {
			return nil, fmt.Errorf("action is required and must be a string")
		}

		payloadObj, _ := args["payload"].(map[string]interface{})
		if payloadObj == nil {
			payloadObj = make(map[string]interface{})
		}

		b, err := json.Marshal(payloadObj)
		if err != nil {
			return nil, fmt.Errorf("invalid payload: %v", err)
		}

		taskAPI := api.New(globalSvc)
		resp := taskAPI.Execute(ctx, api.Request{Action: action, Payload: b})

		if !resp.Success {
			return nil, fmt.Errorf("api error: %s", resp.Error)
		}

		return resp.Data, nil
	}

	return nil, fmt.Errorf("unknown tool: %s", name)
}
