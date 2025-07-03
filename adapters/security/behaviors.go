package security

import (
	"fmt"
	"strings"
	"time"

	"aegis/catalog"
	"aegis/capitan"
	"aegis/pipz"
	"aegis/sctx"
	"aegis/zlog"
)

// Event type identifiers (must be comparable)
type SecurityViolationEventType string
type SecurityAuditEventType string

const (
	SecurityViolationEvent SecurityViolationEventType = "security.violation"
	SecurityAuditEvent     SecurityAuditEventType     = "security.audit"
)

// RegisterSecurityBehaviors registers common security behavior patterns
// These are type-agnostic behaviors that work based on struct tags
func RegisterSecurityBehaviors() {
	// Register security field processors for logging
	RegisterSecurityLoggingProcessors()
	
	// Register automatic security behavior for new types
	RegisterAutomaticSecurityBehaviors()
}

// CreateAccessControlBehavior creates a behavior that enforces object-level access
func CreateAccessControlBehavior[T any](requiredPermission string) catalog.SecurityProcessor[T] {
	return func(input catalog.SecurityInput[T]) catalog.SecurityOutput[T] {
		// Check if user has the required permission
		if !input.Context.HasPermission(requiredPermission) {
			// Emit security violation
			capitan.Emit[SecurityViolationEventType, sctx.SecurityViolationEvent](sctx.SecurityViolationEvent{
				UserID:   input.Context.UserID,
				Resource: catalog.GetTypeName[T](),
				Action:   "access_denied",
				Severity: "critical",
			})
			
			return catalog.SecurityOutput[T]{
				Data: input.Data,
				Error: &sctx.SecurityError{
					Type:     "forbidden",
					Resource: catalog.GetTypeName[T](),
					Action:   "access",
					Message:  fmt.Sprintf("permission '%s' required", requiredPermission),
				},
			}
		}
		
		return catalog.SecurityOutput[T]{Data: input.Data, Error: nil}
	}
}

// CreateAuditBehavior creates a behavior that logs access for audit purposes
func CreateAuditBehavior[T any]() catalog.SecurityProcessor[T] {
	return func(input catalog.SecurityInput[T]) catalog.SecurityOutput[T] {
		// Emit audit event with proper type signature
		capitan.Emit[SecurityAuditEventType, AuditEventData](AuditEventData{
			UserID:    input.Context.UserID,
			Resource:  catalog.GetTypeName[T](),
			Action:    "data_access",
			Timestamp: input.Context.Extensions["timestamp"].(string),
		})
		
		// Always allow but log
		return catalog.SecurityOutput[T]{Data: input.Data, Error: nil}
	}
}

// AuditEventData is the data for audit events
type AuditEventData struct {
	UserID    string `json:"user_id"`
	Resource  string `json:"resource"`
	Action    string `json:"action"`
	Timestamp string `json:"timestamp"`
}

// Helper to check if a field should be encrypted
func shouldEncrypt(field catalog.FieldMetadata) bool {
	// Check for explicit encrypt tag
	if encryptTag := field.Tags["encrypt"]; encryptTag != "" {
		return true
	}
	
	// Check for PII indicators
	validateTag := field.Tags["validate"]
	return validateTag == "ssn" || validateTag == "creditcard" || strings.Contains(validateTag, "pii")
}

// RegisterSecurityLoggingProcessors registers field processors that add security context to logs
func RegisterSecurityLoggingProcessors() {
	// Get the zlog field processing contract
	fieldContract := pipz.GetContract[zlog.ZlogFieldType, zlog.ZlogField, []zlog.ZlogField]()
	
	// Define custom field types for security context
	const (
		SecurityContextType zlog.ZlogFieldType = "security_context"
		SensitiveDataType   zlog.ZlogFieldType = "sensitive"
		AuditFieldType      zlog.ZlogFieldType = "audit"
	)
	
	// Register processor for security context fields
	// This adds user permissions and request context to logs
	fieldContract.Register(SecurityContextType, func(field zlog.ZlogField) []zlog.ZlogField {
		ctx, ok := field.Value.(sctx.SecurityContext)
		if !ok {
			return []zlog.ZlogField{field}
		}
		
		// Expand security context into multiple fields
		fields := []zlog.ZlogField{
			zlog.String("user_id", ctx.UserID),
			zlog.Strings("permissions", ctx.Permissions),
		}
		
		// Add optional fields from extensions
		if userType, ok := ctx.Extensions["user_type"].(string); ok {
			fields = append(fields, zlog.String("user_type", userType))
		}
		if requestID, ok := ctx.Extensions["request_id"].(string); ok {
			fields = append(fields, zlog.String("request_id", requestID))
		}
		if serviceChain, ok := ctx.Extensions["service_chain"].([]string); ok {
			fields = append(fields, zlog.String("service_chain", strings.Join(serviceChain, " -> ")))
		}
		
		return fields
	})
	
	// Register processor for sensitive data fields
	// This automatically redacts or hashes sensitive values
	fieldContract.Register(SensitiveDataType, func(field zlog.ZlogField) []zlog.ZlogField {
		// Determine masking strategy based on field key
		var maskedValue string
		
		switch field.Key {
		case "ssn":
			maskedValue = MaskSSN(fmt.Sprint(field.Value))
		case "credit_card":
			maskedValue = MaskCreditCard(fmt.Sprint(field.Value))
		case "email":
			maskedValue = MaskEmail(fmt.Sprint(field.Value))
		case "api_key":
			maskedValue = MaskAPIKey(fmt.Sprint(field.Value))
		default:
			// Generic masking for unknown sensitive fields
			maskedValue = "[REDACTED]"
		}
		
		return []zlog.ZlogField{
			zlog.String(field.Key, maskedValue),
			zlog.Bool(field.Key+"_redacted", true),
		}
	})
	
	// Register processor for audit fields
	// This adds compliance metadata to logs
	fieldContract.Register(AuditFieldType, func(field zlog.ZlogField) []zlog.ZlogField {
		action, ok := field.Value.(string)
		if !ok {
			return []zlog.ZlogField{field}
		}
		
		// Add audit metadata
		return []zlog.ZlogField{
			zlog.String("audit_action", action),
			zlog.Time("audit_timestamp", time.Now()),
			zlog.String("audit_compliance", "SOC2,HIPAA"), // Could be dynamic based on context
			zlog.Bool("audit_required", true),
		}
	})
}

// RegisterAutomaticSecurityBehaviors sets up automatic security behavior registration
// When new types are registered, if they have security-related tags, behaviors are auto-registered
func RegisterAutomaticSecurityBehaviors() {
	// Listen for new contract creation events
	capitan.Hook[pipz.ContractCreatedEvent, pipz.ContractCreatedEvent](func(event pipz.ContractCreatedEvent) error {
		// Check if this is a serialization pipeline
		if strings.Contains(event.ContractSignature, "SerializationKey") {
			// Extract the type name from the signature
			// Format is like: "SerializationKey:SerializationInput[TypeName]:SerializationOutput[TypeName]"
			parts := strings.Split(event.ContractSignature, "[")
			if len(parts) >= 2 {
				typePart := strings.Split(parts[1], "]")[0]
				
				// For each new serialization pipeline, we should check if the type has security tags
				// and register appropriate behaviors
				// This is tricky because we need the actual type, not just the name
				
				// For now, log that we detected a serialization pipeline
				zlog.Debug("Serialization pipeline created for type",
					zlog.String("type", typePart),
					zlog.String("signature", event.ContractSignature),
				)
				
				// TODO: We need a way to register security behaviors for this type
				// The challenge is that we have the type name as a string, not the actual type
			}
		}
		return nil
	})
}