package capitan

import (
	"sync"
	
	"aegis/cereal"
	"aegis/sctx"
)

// HookContract manages multiple handlers for broadcast events
type HookContract[EventType comparable, HandlerType any] struct {
	// Each event type can have multiple handlers
	handlers map[EventType][]func([]byte) error
	mu       sync.RWMutex
}

// Global hook registry - separate from single-handler events
var hookRegistry = struct {
	contracts map[string]any
	mu        sync.RWMutex
}{
	contracts: make(map[string]any),
}

// GetHookContract returns a multi-handler contract for broadcasting
func GetHookContract[EventType comparable, HandlerType any]() *HookContract[EventType, HandlerType] {
	signature := getEventSignature[EventType, HandlerType]()
	
	hookRegistry.mu.RLock()
	if existing, exists := hookRegistry.contracts[signature]; exists {
		hookRegistry.mu.RUnlock()
		return existing.(*HookContract[EventType, HandlerType])
	}
	hookRegistry.mu.RUnlock()
	
	// Create new hook contract
	contract := &HookContract[EventType, HandlerType]{
		handlers: make(map[EventType][]func([]byte) error),
	}
	
	hookRegistry.mu.Lock()
	hookRegistry.contracts[signature] = contract
	hookRegistry.mu.Unlock()
	
	return contract
}

// AddHook adds a handler to the broadcast list
func (h *HookContract[EventType, HandlerType]) AddHook(eventType EventType, handler func([]byte) error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	h.handlers[eventType] = append(h.handlers[eventType], handler)
}

// BroadcastEvent sends event to all registered hooks
func (h *HookContract[EventType, HandlerType]) BroadcastEvent(eventType EventType, eventBytes []byte) error {
	h.mu.RLock()
	handlers := h.handlers[eventType]
	h.mu.RUnlock()
	
	// Broadcast to all handlers asynchronously
	for _, handler := range handlers {
		go func(h func([]byte) error) {
			// Ignore errors in broadcast mode
			_ = h(eventBytes)
		}(handler)
	}
	
	return nil
}

// Hook registers a handler that will be called along with all other hooks
func Hook[EventType comparable, EventData any](handler func(EventData) error) {
	byteHandler := func(eventBytes []byte) error {
		ctx := sctx.NewPackageContext("capitan", []string{"system:all"})
		eventData, err := cereal.UnmarshalJSON[EventData](eventBytes, ctx)
		if err != nil {
			return err
		}
		return handler(eventData)
	}
	
	hookContract := GetHookContract[EventType, func([]byte) error]()
	var eventType EventType
	hookContract.AddHook(eventType, byteHandler)
}

// Broadcast emits an event to all registered hooks (async)
func Broadcast[EventType comparable, EventData any](eventData EventData) error {
	// Serialize event data
	ctx := sctx.NewPackageContext("capitan", []string{"system:all"})
	eventBytes, err := cereal.MarshalJSON(eventData, ctx)
	if err != nil {
		return err
	}
	
	// Get hook contract and broadcast
	hookContract := GetHookContract[EventType, func([]byte) error]()
	var eventType EventType
	return hookContract.BroadcastEvent(eventType, eventBytes)
}

// BroadcastSync emits an event to all registered hooks synchronously
func BroadcastSync[EventType comparable, EventData any](eventData EventData) error {
	// Serialize event data
	ctx := sctx.NewPackageContext("capitan", []string{"system:all"})
	eventBytes, err := cereal.MarshalJSON(eventData, ctx)
	if err != nil {
		return err
	}
	
	// Get hook contract
	hookContract := GetHookContract[EventType, func([]byte) error]()
	var eventType EventType
	
	// Get handlers
	hookContract.mu.RLock()
	handlers := hookContract.handlers[eventType]
	hookContract.mu.RUnlock()
	
	// Execute all handlers synchronously
	for _, handler := range handlers {
		if err := handler(eventBytes); err != nil {
			// In sync mode, we return the first error
			return err
		}
	}
	
	return nil
}

// CountHooks returns number of hooks registered for an event type
func CountHooks[EventType comparable, EventData any]() int {
	hookContract := GetHookContract[EventType, func([]byte) error]()
	var eventType EventType
	
	hookContract.mu.RLock()
	defer hookContract.mu.RUnlock()
	
	return len(hookContract.handlers[eventType])
}