package capitan

import (
	"aegis/zlog"
)

// zlogEventSink implements zlog.EventSink to integrate with Capitan
type zlogEventSink struct{}

func (s *zlogEventSink) EmitLogEvent(event zlog.LogEvent) {
	// Emit the strongly typed event - type system IS the distribution architecture!
	Emit[zlog.LogEventType, zlog.LogEvent](event)
}

// CreateZlogEventSink creates an EventSink for zlog (used by core orchestration)
func CreateZlogEventSink() zlog.EventSink {
	return &zlogEventSink{}
}