// Package api provides the CLI/HTTP API interface for external automation
// This is how external systems interact with Kairo
package api

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/programmersd21/kairo/internal/core"
	"github.com/programmersd21/kairo/internal/service"
)

// TaskAPI defines the external API for task operations
// This is the contract for CLI/HTTP endpoints
type TaskAPI struct {
	service service.TaskService
}

// New creates a new task API
func New(svc service.TaskService) *TaskAPI {
	return &TaskAPI{service: svc}
}

// Request is the base request structure for all API calls
type Request struct {
	Action  string          `json:"action"`
	Payload json.RawMessage `json:"payload"`
}

// Response is the base response structure
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Execute processes an API request and returns a response
func (api *TaskAPI) Execute(ctx context.Context, req Request) Response {
	switch req.Action {
	case "create":
		return api.handleCreate(ctx, req.Payload)
	case "get":
		return api.handleGet(ctx, req.Payload)
	case "update":
		return api.handleUpdate(ctx, req.Payload)
	case "delete":
		return api.handleDelete(ctx, req.Payload)
	case "delete_all":
		return api.handleDeleteAll(ctx)
	case "list":
		return api.handleList(ctx, req.Payload)
	case "list_tags":
		return api.handleListTags(ctx)
	case "cleanup":
		return api.cleanup(ctx)
	default:
		return Response{
			Success: false,
			Error:   fmt.Sprintf("unknown action: %s", req.Action),
		}
	}
}

func (api *TaskAPI) cleanup(ctx context.Context) Response {
	if err := api.service.Prune(ctx); err != nil {
		return Response{
			Success: false,
			Error:   err.Error(),
		}
	}
	return Response{
		Success: true,
		Data:    "database cleaned successfully",
	}
}

// TaskDTO is the data transfer object for tasks (matches Lua and JSON schema)
type TaskDTO struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Priority    int      `json:"priority"`
	Status      string   `json:"status"`
	Deadline    *string  `json:"deadline,omitempty"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

// toDTO converts a core.Task to a DTO for serialization
func toDTO(t core.Task) TaskDTO {
	dto := TaskDTO{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		Tags:        t.Tags,
		Priority:    int(t.Priority),
		Status:      string(t.Status),
		CreatedAt:   t.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   t.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if t.Deadline != nil {
		s := t.Deadline.Format("2006-01-02T15:04:05Z")
		dto.Deadline = &s
	}
	return dto
}

// handleCreate processes a create request
func (api *TaskAPI) handleCreate(ctx context.Context, payload json.RawMessage) Response {
	type CreatePayload struct {
		Title       string   `json:"title"`
		Description string   `json:"description,omitempty"`
		Tags        []string `json:"tags,omitempty"`
		Priority    *int     `json:"priority,omitempty"`
		Status      string   `json:"status,omitempty"`
	}

	var p CreatePayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return Response{
			Success: false,
			Error:   fmt.Sprintf("invalid payload: %v", err),
		}
	}

	if p.Title == "" {
		return Response{
			Success: false,
			Error:   "title is required",
		}
	}

	task := core.Task{
		Title:       p.Title,
		Description: p.Description,
		Tags:        p.Tags,
		Status:      core.StatusTodo,
	}

	if p.Status != "" {
		task.Status = core.Status(p.Status)
	}
	if p.Priority != nil {
		task.Priority = core.Priority(*p.Priority)
	}

	created, err := api.service.Create(ctx, task)
	if err != nil {
		return Response{
			Success: false,
			Error:   err.Error(),
		}
	}

	return Response{
		Success: true,
		Data:    toDTO(created),
	}
}

// handleGet processes a get request
func (api *TaskAPI) handleGet(ctx context.Context, payload json.RawMessage) Response {
	type GetPayload struct {
		ID string `json:"id"`
	}

	var p GetPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return Response{
			Success: false,
			Error:   "invalid payload",
		}
	}

	if p.ID == "" {
		return Response{
			Success: false,
			Error:   "id is required",
		}
	}

	task, err := api.service.GetByID(ctx, p.ID)
	if err != nil {
		return Response{
			Success: false,
			Error:   err.Error(),
		}
	}

	return Response{
		Success: true,
		Data:    toDTO(task),
	}
}

// handleUpdate processes an update request
func (api *TaskAPI) handleUpdate(ctx context.Context, payload json.RawMessage) Response {
	type UpdatePayload struct {
		ID          string   `json:"id"`
		Title       *string  `json:"title,omitempty"`
		Description *string  `json:"description,omitempty"`
		Tags        []string `json:"tags,omitempty"`
		Priority    *int     `json:"priority,omitempty"`
		Status      *string  `json:"status,omitempty"`
	}

	var p UpdatePayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return Response{
			Success: false,
			Error:   "invalid payload",
		}
	}

	if p.ID == "" {
		return Response{
			Success: false,
			Error:   "id is required",
		}
	}

	patch := core.TaskPatch{
		Title:       p.Title,
		Description: p.Description,
	}

	if len(p.Tags) > 0 {
		patch.Tags = &p.Tags
	}

	if p.Priority != nil {
		pr := core.Priority(*p.Priority)
		patch.Priority = &pr
	}

	if p.Status != nil {
		s := core.Status(*p.Status)
		patch.Status = &s
	}

	updated, err := api.service.Update(ctx, p.ID, patch)
	if err != nil {
		return Response{
			Success: false,
			Error:   err.Error(),
		}
	}

	return Response{
		Success: true,
		Data:    toDTO(updated),
	}
}

// handleDelete processes a delete request
func (api *TaskAPI) handleDelete(ctx context.Context, payload json.RawMessage) Response {
	type DeletePayload struct {
		ID string `json:"id"`
	}

	var p DeletePayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return Response{
			Success: false,
			Error:   "invalid payload",
		}
	}

	if p.ID == "" {
		return Response{
			Success: false,
			Error:   "id is required",
		}
	}

	if err := api.service.Delete(ctx, p.ID); err != nil {
		return Response{
			Success: false,
			Error:   err.Error(),
		}
	}

	return Response{
		Success: true,
		Data: map[string]string{
			"id": p.ID,
		},
	}
}

// handleDeleteAll processes a delete_all request
func (api *TaskAPI) handleDeleteAll(ctx context.Context) Response {
	if err := api.service.DeleteAll(ctx); err != nil {
		return Response{
			Success: false,
			Error:   err.Error(),
		}
	}

	return Response{
		Success: true,
		Data:    "all tasks deleted successfully",
	}
}

// handleList processes a list request
func (api *TaskAPI) handleList(ctx context.Context, payload json.RawMessage) Response {
	type ListPayload struct {
		Statuses []string `json:"statuses,omitempty"`
		Tags     []string `json:"tags,omitempty"`
		Priority *int     `json:"priority,omitempty"`
		Sort     string   `json:"sort,omitempty"`
	}

	var p ListPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		p = ListPayload{} // Use defaults if unmarshal fails
	}

	filter := core.Filter{
		Tags: p.Tags,
		Sort: core.SortMode(p.Sort),
	}

	// Convert status strings
	for _, s := range p.Statuses {
		filter.Statuses = append(filter.Statuses, core.Status(s))
	}

	// Convert priority
	if p.Priority != nil {
		pr := core.Priority(*p.Priority)
		filter.Priority = &pr
	}

	tasks, err := api.service.List(ctx, filter)
	if err != nil {
		return Response{
			Success: false,
			Error:   err.Error(),
		}
	}

	dtos := make([]TaskDTO, len(tasks))
	for i, t := range tasks {
		dtos[i] = toDTO(t)
	}

	return Response{
		Success: true,
		Data:    dtos,
	}
}

// handleListTags lists all tags
func (api *TaskAPI) handleListTags(ctx context.Context) Response {
	tags, err := api.service.ListTags(ctx)
	if err != nil {
		return Response{
			Success: false,
			Error:   err.Error(),
		}
	}

	return Response{
		Success: true,
		Data:    tags,
	}
}
