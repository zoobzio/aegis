package security

import (
	"strings"
	"aegis/catalog"
)

// RegisterMaskFunctions registers all standard masking functions
func RegisterMaskFunctions() {
	catalog.RegisterMaskFunction("ssn", MaskSSN)
	catalog.RegisterMaskFunction("creditcard", MaskCreditCard)
	catalog.RegisterMaskFunction("email", MaskEmail)
	catalog.RegisterMaskFunction("phone", MaskPhone)
	catalog.RegisterMaskFunction("password", MaskPassword)
	catalog.RegisterMaskFunction("apikey", MaskAPIKey)
	catalog.RegisterMaskFunction("secret", MaskSecret)
}

// MaskSSN masks a social security number
func MaskSSN(value string) string {
	if len(value) < 4 {
		return "***-**-****"
	}
	// Show last 4 digits
	return "***-**-" + value[len(value)-4:]
}

// MaskCreditCard masks a credit card number
func MaskCreditCard(value string) string {
	// Remove spaces and dashes
	cleaned := strings.ReplaceAll(value, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	
	if len(cleaned) < 4 {
		return "****-****-****-****"
	}
	
	// Show last 4 digits
	return "****-****-****-" + cleaned[len(cleaned)-4:]
}

// MaskEmail masks an email address
func MaskEmail(value string) string {
	parts := strings.Split(value, "@")
	if len(parts) != 2 {
		return "***@***.***"
	}
	
	// Mask local part but keep first and last char
	local := parts[0]
	if len(local) > 2 {
		local = string(local[0]) + strings.Repeat("*", len(local)-2) + string(local[len(local)-1])
	} else {
		local = strings.Repeat("*", len(local))
	}
	
	return local + "@" + parts[1]
}

// MaskPhone masks a phone number
func MaskPhone(value string) string {
	// Remove non-digits
	cleaned := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, value)
	
	if len(cleaned) < 4 {
		return "(***) ***-****"
	}
	
	// Show last 4 digits
	return "(***) ***-" + cleaned[len(cleaned)-4:]
}

// MaskPassword always returns a fixed mask
func MaskPassword(value string) string {
	return "********"
}

// MaskAPIKey masks an API key
func MaskAPIKey(value string) string {
	if len(value) < 8 {
		return "****"
	}
	// Show first 4 and last 4 characters
	return value[:4] + "..." + value[len(value)-4:]
}

// MaskSecret provides a generic secret mask
func MaskSecret(value string) string {
	if value == "" {
		return ""
	}
	return "[REDACTED]"
}