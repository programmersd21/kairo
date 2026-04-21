// Package hooks provides the event system for Kairo - allowing plugins to react to app events
package hooks

import (
	"sync"

	"github.com/programmersd21/kairo/internal/core"
)

// EventType defines the type of application event
type EventType string

const (
	EventTaskCreate EventType = "task_create"
	EventTaskUpdate EventType = "task_update"
	EventTaskDelete EventType = "task_delete"
	EventAppStart   EventType = "app_start"
	EventAppStop    EventType = "app_stop"
)

// Event represents an application event that plugins can subscribe to
type Event struct {
	Type    EventType
	Task    *core.Task             // Nil for app lifecycle events
	Patch   *core.TaskPatch        // Nil except for task_update
	Error   error                  // Nil if success
	Payload map[string]interface{} // Additional context
}

// Listener is a callback that handles an event
type Listener func(Event)

// Manager manages event hooks for the application
type Manager struct {
	mu        sync.RWMutex
	listeners map[EventType][]Listener
}

// New creates a new hook manager
func New() *Manager {
	return &Manager{
		listeners: make(map[EventType][]Listener),
	}
}

// On registers a listener for an event type
func (m *Manager) On(eventType EventType, listener Listener) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.listeners[eventType] = append(m.listeners[eventType], listener)
}

// Off removes all listeners for an event type
// Note: In this implementation, we only provide removal by event type
func (m *Manager) Off(eventType EventType) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.listeners[eventType] = nil
}

// Emit triggers all listeners for an event type
// Listeners are called synchronously in registration order
func (m *Manager) Emit(event Event) {
	m.mu.RLock()
	listeners := m.listeners[event.Type]
	m.mu.RUnlock()

	// Call listeners outside the lock to prevent deadlocks
	for _, listener := range listeners {
		// Recover from panics in listener to prevent one bad listener from crashing everything
		func() {
			defer func() {
				if r := recover(); r != nil {
					_ = r
				}
			}()
			listener(event)
		}()
	}
}

// TaskCreated emits a task creation event
func (m *Manager) TaskCreated(task core.Task) {
	m.Emit(Event{
		Type: EventTaskCreate,
		Task: &task,
	})
}

// TaskUpdated emits a task update event
func (m *Manager) TaskUpdated(task core.Task, patch core.TaskPatch) {
	m.Emit(Event{
		Type:  EventTaskUpdate,
		Task:  &task,
		Patch: &patch,
	})
}

// TaskDeleted emits a task deletion event
func (m *Manager) TaskDeleted(taskID string) {
	m.Emit(Event{
		Type: EventTaskDelete,
		Payload: map[string]interface{}{
			"task_id": taskID,
		},
	})
}

// AppStarted emits an app startup event
func (m *Manager) AppStarted() {
	m.Emit(Event{
		Type: EventAppStart,
	})
}

// AppStopped emits an app stop event
func (m *Manager) AppStopped() {
	m.Emit(Event{
		Type: EventAppStop,
	})
}
