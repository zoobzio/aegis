package capitan

import (
	"reflect"
	"sync"
	
	"aegis/pipz"
)

// EventContract provides type-safe event handling using pipz contracts  
type EventContract[EventType comparable, HandlerType any] struct {
	contract *pipz.ServiceContract[EventType, []byte, error]
}

// Global event registry using type signatures as keys (like pipz)
var eventRegistry = struct {
	contracts map[string]any
	mu        sync.RWMutex
}{
	contracts: make(map[string]any),
}

// GetEventContract returns a type-safe event contract for specific Event/Handler combination
func GetEventContract[EventType comparable, HandlerType any]() *EventContract[EventType, HandlerType] {
	// Use type signature string as contract key (same pattern as pipz)
	signature := getEventSignature[EventType, HandlerType]()
	
	eventRegistry.mu.RLock()
	if existing, exists := eventRegistry.contracts[signature]; exists {
		eventRegistry.mu.RUnlock()
		return existing.(*EventContract[EventType, HandlerType])
	}
	eventRegistry.mu.RUnlock()
	
	// Create new event contract using pipz
	contract := &EventContract[EventType, HandlerType]{
		contract: pipz.GetContract[EventType, []byte, error](),
	}
	
	eventRegistry.mu.Lock()
	eventRegistry.contracts[signature] = contract
	eventRegistry.mu.Unlock()
	
	return contract
}

// RegisterHandler adds a typed event handler to the contract
func (e *EventContract[EventType, HandlerType]) RegisterHandler(eventType EventType, handler func([]byte) error) {
	processor := pipz.Processor[[]byte, error](handler)
	e.contract.Register(eventType, processor)
}

// EmitEvent emits typed event through the contract
func (e *EventContract[EventType, HandlerType]) EmitEvent(eventType EventType, eventBytes []byte) error {
	result, exists := e.contract.Process(eventType, eventBytes)
	if !exists {
		return nil // No handlers registered, not an error
	}
	return result
}

// HasHandler checks if handler exists for event type
func (e *EventContract[EventType, HandlerType]) HasHandler(eventType EventType) bool {
	return e.contract.HasProcessor(eventType)
}

// ListEvents returns all registered event types  
func (e *EventContract[EventType, HandlerType]) ListEvents() []EventType {
	return e.contract.ListKeys()
}

// Helper function to generate type signatures
func getEventSignature[EventType, HandlerType any]() string {
	return getTypeName[EventType]() + ":" + getTypeName[HandlerType]()
}

// TypeNameProvider for hydration (like pipz pattern)
type TypeNameProvider interface {
	GetTypeNameFromType(typ reflect.Type) string
}

var typeNameProvider TypeNameProvider

func SetTypeNameProvider(provider TypeNameProvider) {
	typeNameProvider = provider
}

func getTypeName[T any]() string {
	if typeNameProvider == nil {
		// Fallback to simple type name for now
		return reflect.TypeOf((*T)(nil)).Elem().String()
	}
	typ := reflect.TypeOf((*T)(nil)).Elem()
	return typeNameProvider.GetTypeNameFromType(typ)
}