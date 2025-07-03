package pipz

import (
	"fmt"
	"sync"
)

// Processor is the universal function signature for processing pipeline stages
type Processor[Input, Output any] func(Input) Output

// Event types for pipz operations - concrete, type-safe structs
type ContractCreatedEvent struct {
	ContractSignature string `json:"contract_signature"`
	KeyTypeName       string `json:"key_type_name"`
	InputTypeName     string `json:"input_type_name"`
	OutputTypeName    string `json:"output_type_name"`
}

type ProcessorRegisteredEvent struct {
	ContractSignature string `json:"contract_signature"`
	KeyTypeName       string `json:"key_type_name"`
	KeyValue          string `json:"key_value"`
}


// ServiceContract provides type-safe processing using typed keys (no magic strings)
type ServiceContract[KeyType comparable, Input, Output any] struct {
	processors map[KeyType]Processor[Input, Output]
	keys       []KeyType // Registration order
	mu         sync.RWMutex
}

// Global contract registry using type signatures as keys
var contractRegistry = struct {
	contracts map[string]any
	mu        sync.RWMutex
}{
	contracts: make(map[string]any),
}

// EventEmitter interface for hydration (implemented by capitan)
type EventEmitter interface {
	EmitProcessorRegistered(event ProcessorRegisteredEvent)
	EmitContractCreated(event ContractCreatedEvent)
}

var eventEmitter EventEmitter

func SetEventEmitter(emitter EventEmitter) {
	eventEmitter = emitter
}


// GetContract returns a type-safe contract for specific KeyType/Input/Output combination
// Type signature becomes the registry key - 100% type safe, zero magic strings
func GetContract[KeyType comparable, Input, Output any]() *ServiceContract[KeyType, Input, Output] {
	// Use type signature string as contract key (no reflection as map key)
	signature := getContractSignature[KeyType, Input, Output]()
	
	contractRegistry.mu.RLock()
	if existing, exists := contractRegistry.contracts[signature]; exists {
		contractRegistry.mu.RUnlock()
		return existing.(*ServiceContract[KeyType, Input, Output])
	}
	contractRegistry.mu.RUnlock()
	
	// Create new contract for this exact type combination
	contract := &ServiceContract[KeyType, Input, Output]{
		processors: make(map[KeyType]Processor[Input, Output]),
	}
	
	contractRegistry.mu.Lock()
	contractRegistry.contracts[signature] = contract
	contractRegistry.mu.Unlock()
	
	// Emit contract creation event if emitter is hydrated
	if eventEmitter != nil {
		eventEmitter.EmitContractCreated(ContractCreatedEvent{
			ContractSignature: signature,
			KeyTypeName:       getTypeName[KeyType](),
			InputTypeName:     getTypeName[Input](),
			OutputTypeName:    getTypeName[Output](),
		})
	}
	
	return contract
}

// Type-safe functions using typed keys (no magic strings)

// Register adds a processor with type-safe key
func (c *ServiceContract[KeyType, Input, Output]) Register(key KeyType, processor Processor[Input, Output]) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	// Track registration order
	if _, exists := c.processors[key]; !exists {
		c.keys = append(c.keys, key)
	}
	c.processors[key] = processor
	
	// Emit event if emitter is hydrated
	if eventEmitter != nil {
		eventEmitter.EmitProcessorRegistered(ProcessorRegisteredEvent{
			ContractSignature: getContractSignature[KeyType, Input, Output](),
			KeyTypeName:       getTypeName[KeyType](),
			KeyValue:          formatKey(key),
		})
	}
}

// Unregister removes a processor by typed key
func (c *ServiceContract[KeyType, Input, Output]) Unregister(key KeyType) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	// Remove from processors
	delete(c.processors, key)
	
	// Remove from keys to maintain order
	for i, k := range c.keys {
		if k == key {
			c.keys = append(c.keys[:i], c.keys[i+1:]...)
			break
		}
	}
}

// Process runs input through specific processor with 100% type safety
func (c *ServiceContract[KeyType, Input, Output]) Process(key KeyType, input Input) (Output, bool) {
	c.mu.RLock()
	processor, exists := c.processors[key]
	c.mu.RUnlock()
	
	if !exists {
		var zero Output
		return zero, false
	}
	
	result := processor(input)
	return result, true
}

// HasProcessor checks if processor exists for key
func (c *ServiceContract[KeyType, Input, Output]) HasProcessor(key KeyType) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	_, exists := c.processors[key]
	return exists
}

// ListKeys returns all registered processor keys in registration order
func (c *ServiceContract[KeyType, Input, Output]) ListKeys() []KeyType {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	// Return a copy to prevent external modification
	result := make([]KeyType, len(c.keys))
	copy(result, c.keys)
	return result
}

// Helper functions for event emission and logging
func getContractSignature[K comparable, I, O any]() string {
	return getTypeName[K]() + ":" + getTypeName[I]() + ":" + getTypeName[O]()
}


func formatKey(key any) string {
	return fmt.Sprintf("%v", key)
}

// That's it! Pure type-safe contract system with zero magic strings or non-generic interfaces

// SimpleContract provides single-processor pipelines without keys
type SimpleContract[Input, Output any] struct {
	processor Processor[Input, Output]
	mu        sync.RWMutex
}

// GetSimpleContract returns a single-processor contract for the given types
// This is for use cases where you only need one processor per type pair
func GetSimpleContract[Input, Output any]() *SimpleContract[Input, Output] {
	signature := getSimpleContractSignature[Input, Output]()
	
	contractRegistry.mu.RLock()
	if existing, exists := contractRegistry.contracts[signature]; exists {
		contractRegistry.mu.RUnlock()
		return existing.(*SimpleContract[Input, Output])
	}
	contractRegistry.mu.RUnlock()
	
	// Create new simple contract
	contract := &SimpleContract[Input, Output]{}
	
	contractRegistry.mu.Lock()
	contractRegistry.contracts[signature] = contract
	contractRegistry.mu.Unlock()
	
	return contract
}

// SetProcessor sets the single processor for this contract
func (c *SimpleContract[Input, Output]) SetProcessor(processor Processor[Input, Output]) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.processor = processor
}

// Process runs the input through the processor (no key needed)
func (c *SimpleContract[Input, Output]) Process(input Input) (Output, bool) {
	c.mu.RLock()
	processor := c.processor
	c.mu.RUnlock()
	
	if processor == nil {
		var zero Output
		return zero, false
	}
	
	result := processor(input)
	return result, true
}

// HasProcessor returns true if a processor is set
func (c *SimpleContract[Input, Output]) HasProcessor() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.processor != nil
}

// Helper function for simple contract signatures
func getSimpleContractSignature[I, O any]() string {
	return "simple:" + getTypeName[I]() + ":" + getTypeName[O]()
}