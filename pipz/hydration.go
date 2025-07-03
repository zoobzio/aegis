package pipz

import (
	"reflect"
)

// TypeNameProvider provides type names without pipz importing catalog (non-generic service layer)
type TypeNameProvider interface {
	GetTypeNameFromType(typ reflect.Type) string
}

// EventProvider provides event emission capability without pipz importing capitan (non-generic service layer)
type EventProvider interface {
	EmitTypedEvent(eventType reflect.Type, eventData []byte)
}

var (
	typeNameProvider TypeNameProvider
	eventProvider    EventProvider
)


// SetTypeNameProvider allows catalog to hydrate pipz with type name capability
func SetTypeNameProvider(provider TypeNameProvider) {
	typeNameProvider = provider
}

// SetEventProvider allows capitan to hydrate pipz with event capability
func SetEventProvider(provider EventProvider) {
	eventProvider = provider
}

// getTypeName is now defined in types.go to avoid circular dependencies

// emitEvent emits event through hydrated provider (package-level generic API)  
func emitEvent[T any](event T) {
	if eventProvider != nil {
		typ := reflect.TypeOf(event)
		// For now, use simple string representation (could be enhanced with cereal)
		eventData := []byte(typ.String())
		eventProvider.EmitTypedEvent(typ, eventData)
	}
}