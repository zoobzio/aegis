package capitan

import (
	"aegis/pipz"
)

// PipzEventSink implements the pipz.EventEmitter interface
// It broadcasts pipz events through capitan's event system
type PipzEventSink struct{}

// CreatePipzEventSink creates an event sink for pipz
func CreatePipzEventSink() pipz.EventEmitter {
	return &PipzEventSink{}
}

// EmitProcessorRegistered broadcasts when a processor is registered
func (p *PipzEventSink) EmitProcessorRegistered(event pipz.ProcessorRegisteredEvent) {
	// Use BroadcastSync to ensure handlers run before metadata is cached
	// This is critical for emergent behaviors like auto-registering tags
	BroadcastSync[pipz.ProcessorRegisteredEvent, pipz.ProcessorRegisteredEvent](event)
}


// EmitContractCreated broadcasts when a contract is created
func (p *PipzEventSink) EmitContractCreated(event pipz.ContractCreatedEvent) {
	// Use Broadcast for discovery - multiple systems may track contracts
	Broadcast[pipz.ContractCreatedEvent, pipz.ContractCreatedEvent](event)
}