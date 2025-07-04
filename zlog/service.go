package zlog

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// Private concrete logger service layer
var zlog *zZlog

// zZlog is the minimal service layer - only field processing + optional events
type zZlog struct {
	level LogLevel // Current log level
	// ZlogField processing handled through hydration
	eventSink EventSink    // Optional event emission
	mu        sync.RWMutex // Protect concurrent access
}

// EventSink interface for optional event emission (avoids circular deps)
type EventSink interface {
	EmitLogEvent(event LogEvent)
}

// LogEvent structure for event emission
type LogEvent struct {
	Level     string      `json:"level"`
	Message   string      `json:"message"`
	Fields    []ZlogField `json:"fields"`
	Timestamp time.Time   `json:"timestamp"`
}

// LogEventType is a comparable type for capitan event emission
type LogEventType struct{}

// ZlogFieldProcessor processes a field and returns transformed fields (matched to pipz)
type ZlogFieldProcessor func(ZlogField) []ZlogField

// Configure sets up zlog with config (replaces Register)
func Configure(level LogLevel) {
	zlog.mu.Lock()
	defer zlog.mu.Unlock()
	zlog.level = level
}

// SetEventSink enables optional event emission
func SetEventSink(sink EventSink) {
	zlog.mu.Lock()
	defer zlog.mu.Unlock()
	zlog.eventSink = sink
}

// RegisterZlogFieldProcessor moved to hydration.go

// processFields processes fields through hydrated contract (moved to hydration.go)
func (z *zZlog) processFields(fields []ZlogField) []ZlogField {
	return processFields(fields)
}

// emitEvent emits optional event if sink is available
func (z *zZlog) emitEvent(level, msg string, fields []ZlogField) {
	z.mu.RLock()
	sink := z.eventSink
	z.mu.RUnlock()

	if sink != nil {
		sink.EmitLogEvent(LogEvent{
			Level:     level,
			Message:   msg,
			Fields:    fields,
			Timestamp: time.Now(),
		})
	}
}

// shouldLog checks if message should be logged at given level
func (z *zZlog) shouldLog(level LogLevel) bool {
	z.mu.RLock()
	current := z.level
	z.mu.RUnlock()
	return level >= current
}

// Buffer pool for zero-allocation console output
var bufferPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

func getBuffer() *bytes.Buffer {
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	return buf
}

func putBuffer(buf *bytes.Buffer) {
	if buf.Cap() > 64*1024 { // Don't pool huge buffers
		return
	}
	bufferPool.Put(buf)
}

// writeConsole writes directly to stdout
func writeConsole(level, msg string, fields []ZlogField) {
	buf := getBuffer()
	defer putBuffer(buf)

	// Format: 2024-01-01T10:00:00Z LEVEL message key=value
	buf.WriteString(time.Now().Format(time.RFC3339))
	buf.WriteByte(' ')
	buf.WriteString(level)
	buf.WriteByte(' ')
	buf.WriteString(msg)

	// Add fields
	for _, field := range fields {
		buf.WriteByte(' ')
		buf.WriteString(field.Key)
		buf.WriteByte('=')
		formatZlogFieldValue(buf, field)
	}
	buf.WriteByte('\n')

	os.Stdout.Write(buf.Bytes())
}

// formatZlogFieldValue formats field value into buffer
func formatZlogFieldValue(buf *bytes.Buffer, field ZlogField) {
	switch field.Type {
	case StringType:
		value := field.Value.(string)
		if strings.Contains(value, " ") {
			buf.WriteByte('"')
			buf.WriteString(value)
			buf.WriteByte('"')
		} else {
			buf.WriteString(value)
		}
	case IntType:
		buf.WriteString(fmt.Sprintf("%d", field.Value.(int)))
	case Int64Type:
		buf.WriteString(fmt.Sprintf("%d", field.Value.(int64)))
	case Float64Type:
		buf.WriteString(fmt.Sprintf("%.2f", field.Value.(float64)))
	case BoolType:
		buf.WriteString(fmt.Sprintf("%t", field.Value.(bool)))
	case ErrorType:
		if field.Value != nil {
			err := field.Value.(error)
			buf.WriteByte('"')
			buf.WriteString(err.Error())
			buf.WriteByte('"')
		} else {
			buf.WriteString(`"<nil>"`)
		}
	case DurationType:
		duration := field.Value.(time.Duration)
		buf.WriteString(duration.String())
	case TimeType:
		t := field.Value.(time.Time)
		buf.WriteString(t.Format(time.RFC3339))
	default:
		buf.WriteString(fmt.Sprintf("%v", field.Value))
	}
}

// init sets up default zlog (field processing provided through hydration)
func init() {
	zlog = &zZlog{
		level: INFO,
	}
}
