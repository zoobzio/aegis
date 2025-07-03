package security

import (
	"fmt"
	"regexp"
	"strings"
	
	"aegis/catalog"
)

// Common validation patterns
var (
	emailPattern      = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	ssnPattern        = regexp.MustCompile(`^\d{3}-?\d{2}-?\d{4}$`)
	phonePattern      = regexp.MustCompile(`^[\d\s\-\.\(\)\+]+$`)
	creditCardPattern = regexp.MustCompile(`^\d{4}[\s\-]?\d{4}[\s\-]?\d{4}[\s\-]?\d{4}$`)
	alphaPattern      = regexp.MustCompile(`^[a-zA-Z]+$`)
	alphanumPattern   = regexp.MustCompile(`^[a-zA-Z0-9]+$`)
)

// RegisterStandardValidators registers validation behaviors for common types
// This uses the pipz-based validation pipeline
func RegisterStandardValidators() {
	// For now, register with the old system until we have type-specific registration
	// In the future, this would be:
	// pipeline := catalog.GetValidationPipeline[User]()
	// pipeline.Register(catalog.RequiredValidation, validateRequired)
	
	// Register field validators for common validation rules
	catalog.RegisterFieldValidator("required", validateRequired)
	catalog.RegisterFieldValidator("email", validateEmail) 
	catalog.RegisterFieldValidator("ssn", validateSSN)
	catalog.RegisterFieldValidator("phone", validatePhone)
	catalog.RegisterFieldValidator("creditcard", validateCreditCard)
	catalog.RegisterFieldValidator("alpha", validateAlpha)
	catalog.RegisterFieldValidator("alphanum", validateAlphanum)
	catalog.RegisterFieldValidator("min", validateMin)
	catalog.RegisterFieldValidator("max", validateMax)
}

// Validator functions

func validateRequired(value any) error {
	if value == nil {
		return fmt.Errorf("field is required")
	}
	
	// Check for zero values
	switch v := value.(type) {
	case string:
		if strings.TrimSpace(v) == "" {
			return fmt.Errorf("field is required")
		}
	case []any:
		if len(v) == 0 {
			return fmt.Errorf("field is required")
		}
	case map[string]any:
		if len(v) == 0 {
			return fmt.Errorf("field is required")
		}
	}
	
	return nil
}

func validateEmail(value any) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("email validation requires string type")
	}
	
	if !emailPattern.MatchString(str) {
		return fmt.Errorf("invalid email format")
	}
	
	return nil
}

func validateSSN(value any) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("SSN validation requires string type")
	}
	
	if !ssnPattern.MatchString(str) {
		return fmt.Errorf("invalid SSN format")
	}
	
	return nil
}

func validatePhone(value any) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("phone validation requires string type")
	}
	
	// Remove common formatting
	cleaned := strings.Map(func(r rune) rune {
		if (r >= '0' && r <= '9') || r == '+' {
			return r
		}
		return -1
	}, str)
	
	// Check length (10-15 digits allowing for country codes)
	if len(cleaned) < 10 || len(cleaned) > 15 {
		return fmt.Errorf("phone number must be 10-15 digits")
	}
	
	return nil
}

func validateCreditCard(value any) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("credit card validation requires string type")
	}
	
	// Basic format check
	if !creditCardPattern.MatchString(str) {
		return fmt.Errorf("invalid credit card format")
	}
	
	// Could add Luhn algorithm here for real validation
	
	return nil
}

func validateAlpha(value any) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("alpha validation requires string type")
	}
	
	if !alphaPattern.MatchString(str) {
		return fmt.Errorf("field must contain only letters")
	}
	
	return nil
}

func validateAlphanum(value any) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("alphanum validation requires string type")
	}
	
	if !alphanumPattern.MatchString(str) {
		return fmt.Errorf("field must contain only letters and numbers")
	}
	
	return nil
}

func validateMin(value any) error {
	// Min/max validators need parameters, which the current system doesn't support
	// When we move to pipz, we can pass parameters through the input type
	switch value.(type) {
	case int, int8, int16, int32, int64:
		return nil
	case uint, uint8, uint16, uint32, uint64:
		return nil
	case float32, float64:
		return nil
	case string:
		return nil
	default:
		return fmt.Errorf("min validation requires numeric or string type")
	}
}

func validateMax(value any) error {
	// Similar to validateMin
	switch value.(type) {
	case int, int8, int16, int32, int64:
		return nil
	case uint, uint8, uint16, uint32, uint64:
		return nil
	case float32, float64:
		return nil
	case string:
		return nil
	default:
		return fmt.Errorf("max validation requires numeric or string type")
	}
}