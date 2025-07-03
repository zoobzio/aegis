package capitan

import (
	"aegis/cereal"
	"aegis/sctx"
)

// COMMAND PATTERN - Single handler per command type

// Emit sends a strongly typed command to its single handler
// Use this for commands where exactly one service should respond
func Emit[EventType comparable, EventData any](eventData EventData) error {
	// Serialize event data to bytes (using package context)
	ctx := sctx.NewPackageContext("capitan", []string{"system:all"})
	eventBytes, err := cereal.MarshalJSON(eventData, ctx)
	if err != nil {
		return err
	}

	// Get the appropriate event contract based on type signature
	eventContract := GetEventContract[EventType, func([]byte) error]()
	
	// Use zero value of EventType as the key (like pipz pattern)
	var eventType EventType
	return eventContract.EmitEvent(eventType, eventBytes)
}

// Listen registers THE SINGLE handler for a specific command type
// Registering a new handler REPLACES any existing handler
// Use this for commands, state machines, and single-owner patterns
func Listen[EventType comparable, EventData any](handler func(EventData) error) {
	byteHandler := func(eventBytes []byte) error {
		ctx := sctx.NewPackageContext("capitan", []string{"system:all"})
		eventData, err := cereal.UnmarshalJSON[EventData](eventBytes, ctx)
		if err != nil {
			return err
		}
		return handler(eventData)
	}
	
	eventContract := GetEventContract[EventType, func([]byte) error]()
	// Use zero value of EventType as the key
	var eventType EventType
	eventContract.RegisterHandler(eventType, byteHandler)
}

// EVENT PATTERN - Multiple handlers per event type

// NOTE: Hook and Broadcast are defined in hooks.go
// Hook - registers a handler that will be called along with all other hooks
// Broadcast - emits an event to all registered hooks asynchronously
