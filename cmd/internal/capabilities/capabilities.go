package capabilities

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"

	"aegis/capitan"
	"aegis/catalog" 
	"aegis/cereal"
	"aegis/pipz"
	"aegis/sctx"
	"aegis/zlog"
)

// Capability abstractions - hide implementation details

// Type-Safe Event Emission
func EmitEvent[EventType comparable, EventData any](eventData EventData) error {
	return capitan.Emit[EventType, EventData](eventData)
}

// Type-Safe Event Listening
func ListenForEvent[EventType comparable, EventData any](handler func(EventData) error) {
	capitan.Listen[EventType, EventData](handler)
}

// Multi-Handler Event Hooks
func HookIntoEvent[EventType comparable, EventData any](handler func(EventData) error) {
	capitan.Hook[EventType, EventData](handler)
}

// Broadcast to Multiple Handlers
func BroadcastEvent[EventType comparable, EventData any](eventData EventData) error {
	return capitan.Broadcast[EventType, EventData](eventData)
}

// Type-Safe Processing Contracts
func GetProcessingContract[KeyType comparable, Input, Output any]() ProcessingContract[KeyType, Input, Output] {
	return &processingContractImpl[KeyType, Input, Output]{
		contract: pipz.GetContract[KeyType, Input, Output](),
	}
}

type ProcessingContract[KeyType comparable, Input, Output any] interface {
	Register(key KeyType, processor func(Input) Output)
	Process(key KeyType, input Input) (Output, bool)
}

type processingContractImpl[KeyType comparable, Input, Output any] struct {
	contract *pipz.ServiceContract[KeyType, Input, Output]
}

func (p *processingContractImpl[KeyType, Input, Output]) Register(key KeyType, processor func(Input) Output) {
	p.contract.Register(key, processor)
}

func (p *processingContractImpl[KeyType, Input, Output]) Process(key KeyType, input Input) (Output, bool) {
	return p.contract.Process(key, input)
}

// Security Pipeline Registration
func RegisterSecurityPipeline[T any](key catalog.SecurityBehaviorKey, processor pipz.Processor[catalog.SecurityInput[T], catalog.SecurityOutput[T]]) {
	contract := catalog.GetSecurityContract[T]()
	contract.Register(key, processor)
}

// Validation Pipeline Registration
func RegisterValidationPipeline[T any](key catalog.ValidationBehaviorKey, processor pipz.Processor[catalog.ValidationInput[T], catalog.ValidationOutput[T]]) {
	contract := catalog.GetValidationContract[T]()
	contract.Register(key, processor)
}

// Secure Serialization
func SecurelySerialize[T any](data T, ctx sctx.SecurityContext) ([]byte, error) {
	return cereal.MarshalJSON(data, ctx)
}

func SecurelyDeserialize[T any](data []byte, ctx sctx.SecurityContext) (T, error) {
	return cereal.UnmarshalJSON[T](data, ctx)
}

// Structured Logging
func LogStructured(msg string, fields ...zlog.Field) {
	zlog.Info(msg, fields...)
}

func LogSecurityEvent(msg string, fields ...zlog.Field) {
	zlog.Warn(msg, fields...)
}

// Security Context Creation
func CreateUserContext(userID string, permissions []string) sctx.SecurityContext {
	return sctx.NewUserContext(userID, permissions)
}

func CreateSystemContext(component string, permissions []string) sctx.SecurityContext {
	return sctx.NewSystemContext(component, permissions)
}

// Helper to get capability name without revealing package
func getCapabilityName(fn interface{}) string {
	name := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
	parts := strings.Split(name, ".")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return "unknown"
}

// Test result formatting
func TestPassed(capability string) string {
	return fmt.Sprintf("✅ %s capability verified", capability)
}

func TestFailed(capability string, err error) string {
	return fmt.Sprintf("❌ %s capability failed: %v", capability, err)
}