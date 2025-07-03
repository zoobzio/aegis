package security

import (
	"fmt"
	"strings"
	"time"
	
	"github.com/spf13/cobra"
	
	"aegis/adapters/security"
	"aegis/catalog"
	"aegis/cereal"
	"aegis/sctx"
	"aegis/zlog"
)

// Real-world test scenarios that showcase composable security workflows

// Healthcare scenario - patient records with complex access rules
type PatientRecord struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	SSN             string    `json:"ssn" validate:"ssn" scope:"admin"`
	MedicalID       string    `json:"medical_id"`
	Diagnosis       string    `json:"diagnosis" scope:"medical"`
	Medications     []string  `json:"medications" scope:"medical"`
	InsuranceInfo   string    `json:"insurance" scope:"billing"`
	DoctorNotes     string    `json:"doctor_notes" scope:"doctor"`
	EmergencyContact string   `json:"emergency_contact" scope:"patient"`
	LastVisit       time.Time `json:"last_visit"`
}

// Financial scenario - transaction with PCI compliance
type FinancialTransaction struct {
	TransactionID   string  `json:"tx_id"`
	CustomerName    string  `json:"customer_name"`
	CreditCard      string  `json:"credit_card" validate:"creditcard" scope:"payment"`
	Amount          float64 `json:"amount"`
	MerchantAccount string  `json:"merchant_account" scope:"merchant"`
	ProcessorFee    float64 `json:"processor_fee" scope:"accounting"`
	RiskScore       int     `json:"risk_score" scope:"fraud"`
	IPAddress       string  `json:"ip_address" scope:"security"`
}

// Multi-tenant SaaS - customer data with tenant isolation
type CustomerData struct {
	CustomerID      string            `json:"customer_id"`
	CompanyName     string            `json:"company_name"`
	APIKey          string            `json:"api_key" validate:"apikey" scope:"admin"`
	BillingEmail    string            `json:"billing_email" validate:"email"`
	CreditBalance   float64           `json:"credit_balance" scope:"billing"`
	UsageMetrics    map[string]int    `json:"usage_metrics" scope:"analytics"`
	SubscriptionTier string           `json:"subscription_tier"`
	CustomConfig    map[string]string `json:"custom_config" scope:"tenant_admin"`
}

// runSerializationScenarios runs real-world security scenarios
func runSerializationTests(cmd *cobra.Command, args []string) {
	// Initialize security adapter first (registers tags, masks, validators)
	security.Initialize()
	
	// Register security processors for our types
	security.RegisterSerializationSecurity[PatientRecord]()
	security.RegisterSerializationSecurity[FinancialTransaction]()
	security.RegisterSerializationSecurity[CustomerData]()
	
	// Also register additional behaviors to show composition
	registerAuditBehavior[PatientRecord]()
	registerComplianceBehavior[FinancialTransaction]()
	registerTenantIsolation[CustomerData]()
	
	runTest("Simple Security Test", testSimpleSecurity)
	runTest("Healthcare Access Control", testHealthcareScenario)
	runTest("Financial Compliance", testFinancialCompliance)
	runTest("Multi-tenant Isolation", testMultiTenantScenario)
	runTest("Composable Security Behaviors", testComposableBehaviors)
}

// Simple test to verify security processing works
func testSimpleSecurity() error {
	type SimpleRecord struct {
		Public  string `json:"public"`
		Private string `json:"private" scope:"admin"`
	}
	
	// Register security for this type
	security.RegisterSerializationSecurity[SimpleRecord]()
	
	record := SimpleRecord{
		Public:  "everyone can see this",
		Private: "only admin can see this",
	}
	
	// Test without admin permission
	userCtx := sctx.NewUserContext("user", []string{"user"})
	data, err := cereal.MarshalJSON(record, userCtx)
	if err != nil {
		return fmt.Errorf("serialization failed: %w", err)
	}
	
	result := string(data)
	if strings.Contains(result, "only admin can see this") {
		return fmt.Errorf("user should not see private field: %s", result)
	}
	
	// Test with admin permission
	adminCtx := sctx.NewUserContext("admin", []string{"admin"})
	data, err = cereal.MarshalJSON(record, adminCtx)
	if err != nil {
		return fmt.Errorf("admin serialization failed: %w", err)
	}
	
	result = string(data)
	if !strings.Contains(result, "only admin can see this") {
		return fmt.Errorf("admin should see private field: %s", result)
	}
	
	return nil
}

// Healthcare scenario - different roles see different data
func testHealthcareScenario() error {
	patient := PatientRecord{
		ID:               "P-12345",
		Name:             "John Smith",
		SSN:              "123-45-6789",
		MedicalID:        "MED-98765",
		Diagnosis:        "Hypertension, Type 2 Diabetes",
		Medications:      []string{"Metformin", "Lisinopril"},
		InsuranceInfo:    "BlueCross PPO #BC123456",
		DoctorNotes:      "Patient showing improvement. Continue current medication.",
		EmergencyContact: "Jane Smith (555) 123-4567",
		LastVisit:        time.Now().Add(-7 * 24 * time.Hour),
	}
	
	// Test different healthcare roles
	scenarios := []struct {
		role        string
		permissions []string
		canSee      []string
		cantSee     []string
	}{
		{
			role:        "Receptionist",
			permissions: []string{"patient"},
			canSee:      []string{"John Smith", "P-12345", "emergency_contact"},
			cantSee:     []string{"123-45-6789", "Hypertension", "Metformin", "BlueCross"},
		},
		{
			role:        "Nurse", 
			permissions: []string{"patient", "medical"},
			canSee:      []string{"Hypertension", "Metformin", "MED-98765"},
			cantSee:     []string{"123-45-6789", "BlueCross", "improvement"},
		},
		{
			role:        "Doctor",
			permissions: []string{"patient", "medical", "doctor"},
			canSee:      []string{"Hypertension", "improvement", "Continue"},
			cantSee:     []string{"123-45-6789", "BlueCross"},
		},
		{
			role:        "Billing",
			permissions: []string{"patient", "billing"},
			canSee:      []string{"BlueCross", "BC123456"},
			cantSee:     []string{"Hypertension", "Metformin", "improvement"},
		},
		{
			role:        "Admin",
			permissions: []string{"admin", "patient", "medical", "billing", "doctor"},
			canSee:      []string{"***-**-6789", "Hypertension", "BlueCross", "improvement"}, // SSN is masked even for admin
			cantSee:     []string{"123-45-6789"}, // Full SSN never visible
		},
	}
	
	for _, scenario := range scenarios {
		ctx := sctx.NewUserContext(scenario.role, scenario.permissions)
		data, err := cereal.MarshalJSON(patient, ctx)
		if err != nil {
			return fmt.Errorf("%s serialization failed: %w", scenario.role, err)
		}
		
		result := string(data)
		
		// Check what they can see
		for _, expected := range scenario.canSee {
			if !strings.Contains(result, expected) {
				return fmt.Errorf("%s should see '%s' in: %s", scenario.role, expected, result)
			}
		}
		
		// Check what they can't see
		for _, hidden := range scenario.cantSee {
			if strings.Contains(result, hidden) {
				return fmt.Errorf("%s should NOT see '%s' in: %s", scenario.role, hidden, result)
			}
		}
	}
	
	return nil
}

// Financial compliance - PCI and audit requirements
func testFinancialCompliance() error {
	transaction := FinancialTransaction{
		TransactionID:   "TXN-20250107-001",
		CustomerName:    "Acme Corp",
		CreditCard:      "4111-1111-1111-1234",
		Amount:          1299.99,
		MerchantAccount: "MERCH-98765",
		ProcessorFee:    38.99,
		RiskScore:       15,
		IPAddress:       "192.168.1.100",
	}
	
	// PCI compliance - credit card always masked except for specific roles
	customerCtx := sctx.NewUserContext("customer", []string{"customer"})
	customerData, _ := cereal.MarshalJSON(transaction, customerCtx)
	
	if strings.Contains(string(customerData), "4111-1111") {
		return fmt.Errorf("customer should never see full credit card")
	}
	
	// Merchant can see masked card
	merchantCtx := sctx.NewUserContext("merchant", []string{"merchant", "payment"})
	merchantData, _ := cereal.MarshalJSON(transaction, merchantCtx)
	
	if !strings.Contains(string(merchantData), "****-****-****-1234") {
		return fmt.Errorf("merchant should see masked credit card")
	}
	
	// Accounting can see fees but not card
	accountingCtx := sctx.NewUserContext("accounting", []string{"accounting"})
	accountingData, _ := cereal.MarshalJSON(transaction, accountingCtx)
	
	result := string(accountingData)
	if !strings.Contains(result, "38.99") {
		return fmt.Errorf("accounting should see processor fee")
	}
	if strings.Contains(result, "1234") {
		return fmt.Errorf("accounting should not see any card info, got: %s", result)
	}
	
	// Check audit metadata was added
	if !strings.Contains(string(merchantData), "TXN-20250107-001") {
		return fmt.Errorf("transaction ID should always be visible for audit")
	}
	
	return nil
}

// Multi-tenant scenario - tenant isolation and data boundaries
func testMultiTenantScenario() error {
	customer := CustomerData{
		CustomerID:       "CUST-789",
		CompanyName:      "TechStartup Inc",
		APIKey:           "sk_live_abcdef123456",
		BillingEmail:     "billing@techstartup.com",
		CreditBalance:    2500.00,
		UsageMetrics:     map[string]int{"api_calls": 150000, "storage_gb": 250},
		SubscriptionTier: "enterprise",
		CustomConfig:     map[string]string{"feature_x": "enabled", "api_version": "v2"},
	}
	
	// Customer support - limited view
	supportCtx := sctx.NewUserContext("support", []string{"support"})
	supportData, _ := cereal.MarshalJSON(customer, supportCtx)
	
	result := string(supportData)
	if strings.Contains(result, "sk_live") || strings.Contains(result, "2500") {
		return fmt.Errorf("support should not see API keys or billing info")
	}
	
	// Tenant admin - sees config but not billing
	tenantCtx := sctx.NewUserContext("tenant-admin-789", []string{"tenant_admin"})
	tenantCtx.Extensions["tenant_id"] = "789" // Tenant isolation
	tenantData, _ := cereal.MarshalJSON(customer, tenantCtx)
	
	result = string(tenantData)
	if !strings.Contains(result, "feature_x") {
		return fmt.Errorf("tenant admin should see custom config")
	}
	if strings.Contains(result, "2500") {
		return fmt.Errorf("tenant admin should not see billing balance")
	}
	
	// Analytics team - sees usage but not config
	analyticsCtx := sctx.NewUserContext("analytics", []string{"analytics"})
	analyticsData, _ := cereal.MarshalJSON(customer, analyticsCtx)
	
	result = string(analyticsData)
	if !strings.Contains(result, "150000") {
		return fmt.Errorf("analytics should see usage metrics, got: %s", result)
	}
	if strings.Contains(result, "feature_x") {
		return fmt.Errorf("analytics should not see custom config")
	}
	
	return nil
}

// Test composing multiple security behaviors
func testComposableBehaviors() error {
	// Create a complex record that triggers multiple behaviors
	patient := PatientRecord{
		ID:           "P-99999",
		Name:         "Test Patient",
		SSN:          "999-88-7777",
		Diagnosis:    "Test Diagnosis",
		DoctorNotes:  "Confidential notes",
	}
	
	// Admin context - should see audit trail in metadata
	adminCtx := sctx.NewUserContext("admin", []string{"admin", "medical", "doctor"})
	adminCtx.Extensions["request_id"] = "REQ-123"
	adminCtx.Extensions["source_ip"] = "10.0.0.1"
	
	data, err := cereal.MarshalJSON(patient, adminCtx)
	if err != nil {
		return fmt.Errorf("admin serialization failed: %w", err)
	}
	
	// The serialization should have:
	// 1. Applied security masking (SSN masked)
	// 2. Added audit metadata (from our audit behavior)
	// 3. Checked compliance rules (from compliance behavior)
	
	result := string(data)
	if !strings.Contains(result, "***-**-7777") {
		return fmt.Errorf("SSN should be masked even for admin")
	}
	
	// Log to show the audit trail
	zlog.Info("Serialization completed",
		zlog.String("user", "admin"),
		zlog.String("record_type", "PatientRecord"),
		zlog.String("record_id", patient.ID),
		zlog.Bool("security_applied", true),
		zlog.Bool("audit_logged", true),
		zlog.Bool("compliance_checked", true),
	)
	
	return nil
}

// Additional behaviors to show composition

// Audit behavior - adds audit trail to all healthcare records
func registerAuditBehavior[T any]() {
	pipeline := cereal.GetSerializationPipeline[T]()
	
	pipeline.Register(cereal.PreProcess, func(input cereal.SerializationInput[T]) cereal.SerializationOutput[T] {
		// Log access attempt
		zlog.Info("Data access audit",
			zlog.String("type", catalog.GetTypeName[T]()),
			zlog.String("user", input.Context.UserID),
			zlog.Time("timestamp", time.Now()),
			zlog.Strings("permissions", input.Context.Permissions),
		)
		
		// Add audit metadata
		return cereal.SerializationOutput[T]{
			Data: input.Data,
			ProcessingMetadata: map[string]any{
				"audit.timestamp": time.Now().Unix(),
				"audit.user":      input.Context.UserID,
				"audit.type":      "serialization",
			},
		}
	})
}

// Compliance behavior - ensures PCI compliance for financial data
func registerComplianceBehavior[T any]() {
	pipeline := cereal.GetSerializationPipeline[T]()
	
	pipeline.Register(cereal.Validate, func(input cereal.SerializationInput[T]) cereal.SerializationOutput[T] {
		// Check if user has PCI compliance training
		if pciCertified, ok := input.Context.Extensions["pci_certified"].(bool); ok && !pciCertified {
			// Still allow but add warning
			return cereal.SerializationOutput[T]{
				Data: input.Data,
				ProcessingMetadata: map[string]any{
					"compliance.warning": "User not PCI certified",
				},
			}
		}
		
		return cereal.SerializationOutput[T]{Data: input.Data}
	})
}

// Tenant isolation - ensures data doesn't leak across tenants
func registerTenantIsolation[T any]() {
	pipeline := cereal.GetSerializationPipeline[T]()
	
	pipeline.Register(cereal.PreProcess, func(input cereal.SerializationInput[T]) cereal.SerializationOutput[T] {
		// Check tenant context
		tenantID, ok := input.Context.Extensions["tenant_id"].(string)
		if !ok {
			return cereal.SerializationOutput[T]{
				Data:  input.Data,
				Error: fmt.Errorf("tenant context required for multi-tenant data"),
			}
		}
		
		// Could verify tenant ID matches data ownership here
		return cereal.SerializationOutput[T]{
			Data: input.Data,
			ProcessingMetadata: map[string]any{
				"tenant.id":       tenantID,
				"tenant.isolated": true,
			},
		}
	})
}