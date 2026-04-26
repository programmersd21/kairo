package ai

import (
	"context"
	"fmt"

	"google.golang.org/genai"
)

type ToolUseEvent struct {
	ToolName string
	Args     map[string]interface{}
}

type StreamChunk struct {
	Text    string
	Done    bool
	Err     error
	ToolUse *ToolUseEvent
	Refresh bool // Signals the UI to reload data live
}

type AppContext struct {
	ViewName string
	Data     string
}

type Client struct {
	genaiClient *genai.Client
	model       string
}

func NewClient(ctx context.Context, apiKey string, modelName string) (*Client, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("ai: %w", err)
	}

	if modelName == "" {
		modelName = "gemini-3.1-flash-lite-preview"
	}

	return &Client{
		genaiClient: client,
		model:       modelName,
	}, nil
}

func (c *Client) ChatStream(ctx context.Context, history []*genai.Content, userMsg string, appCtx AppContext) (<-chan StreamChunk, error) {
	ch := make(chan StreamChunk, 100)

	go func() {
		defer close(ch)
		defer func() {
			if r := recover(); r != nil {
				_ = r // absorb panics from HTTP client during teardown
			}
		}()

		systemPrompt := fmt.Sprintf(`You are Kairo Assistant, an expert productivity AI embedded in Kairo, a terminal-based task manager.
Current View: %s
Context Data: %s

You have TOTAL control over the user's tasks, projects, UI themes, and Lua plugins through tool calls. 
You can:
- Manage tasks (create, update, delete, list, tags, priority, status, deadline).
- Change the UI theme (e.g. catppuccin, dracula, nord, midnight, etc.) using 'set_theme'.
- Manage Lua plugins (list, read, write, delete) using 'plugin_*' actions.
- Configure AI settings and export data.

Be concise, direct, and action-oriented. Format output for a terminal: use plain text, avoid markdown headers, use simple bullet points (- ) only.`, appCtx.ViewName, appCtx.Data)

		kairoTools := GetKairoTools()

		config := &genai.GenerateContentConfig{
			SystemInstruction: &genai.Content{
				Parts: []*genai.Part{{Text: systemPrompt}},
			},
			Tools: []*genai.Tool{kairoTools},
		}

		switch c.model {
		case "gemini-3.1-flash-lite-preview":
			config.ThinkingConfig = &genai.ThinkingConfig{
				ThinkingLevel: genai.ThinkingLevel("MINIMAL"),
			}
		case "gemini-2.5-flash-lite":
			config.ThinkingConfig = &genai.ThinkingConfig{
				ThinkingBudget: genai.Ptr[int32](0),
			}
			config.Tools = append(config.Tools, &genai.Tool{
				GoogleSearch: &genai.GoogleSearch{},
			})
		}

		history = append(history, &genai.Content{
			Role:  "user",
			Parts: []*genai.Part{{Text: userMsg}},
		})

		// send is a helper that respects context cancellation.
		send := func(chunk StreamChunk) bool {
			select {
			case <-ctx.Done():
				return false
			case ch <- chunk:
				return true
			}
		}

		for i := 0; i < 10; i++ {
			if ctx.Err() != nil {
				return
			}

			iter := c.genaiClient.Models.GenerateContentStream(ctx, c.model, history, config)
			var toolCalls []*genai.Part

			for resp, err := range iter {
				if err != nil {
					send(StreamChunk{Err: err})
					return
				}

				if len(resp.Candidates) == 0 {
					continue
				}

				for _, part := range resp.Candidates[0].Content.Parts {
					if part.Text != "" {
						if !send(StreamChunk{Text: part.Text}) {
							return
						}
					}
					if part.FunctionCall != nil {
						toolCalls = append(toolCalls, part)
						if !send(StreamChunk{ToolUse: &ToolUseEvent{
							ToolName: part.FunctionCall.Name,
							Args:     part.FunctionCall.Args,
						}}) {
							return
						}
					}
				}
			}

			if len(toolCalls) == 0 {
				send(StreamChunk{Done: true})
				return
			}

			history = append(history, &genai.Content{
				Role:  "model",
				Parts: toolCalls,
			})

			var results []*genai.Part
			for _, tc := range toolCalls {
				res, err := ExecuteTool(ctx, tc.FunctionCall.Name, tc.FunctionCall.Args)
				if err != nil {
					res = map[string]interface{}{"error": err.Error()}
				}

				resMap, ok := res.(map[string]interface{})
				if !ok {
					// Fallback: wrap non-map results
					resMap = map[string]interface{}{"result": res}
				}

				results = append(results, &genai.Part{
					FunctionResponse: &genai.FunctionResponse{
						Name:     tc.FunctionCall.Name,
						Response: resMap,
					},
				})
			}

			// Trigger a live UI refresh after tool execution
			send(StreamChunk{Refresh: true})

			history = append(history, &genai.Content{
				Role:  "tool",
				Parts: results,
			})
		}

		send(StreamChunk{Err: fmt.Errorf("max tool iterations reached")})
	}()

	return ch, nil
}
