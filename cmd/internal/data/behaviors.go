package data

import (
	"aegis/catalog"
)

// RegisterDataBehaviors registers behaviors that the data domain cares about
// This demonstrates how users would register behaviors during app initialization
// These behaviors tell the framework which tags to extract
func RegisterDataBehaviors() {
	// Register scope extraction behavior - tells framework we care about "scope" tags
	scopePipeline := catalog.GetScopePipeline[MetadataTestUser]()
	scopePipeline.Register(catalog.FieldScope, func(input catalog.ScopeInput[MetadataTestUser]) catalog.ScopeOutput {
		// In a real app, this would extract scope from field tags
		return catalog.ScopeOutput{Scope: "field"}
	})
	
	// Register validation behavior - tells framework we care about "validate" tags  
	validationPipeline := catalog.GetValidationPipeline[MetadataTestUser]()
	validationPipeline.Register(catalog.FormatValidation, func(input catalog.ValidationInput[MetadataTestUser]) catalog.ValidationOutput[MetadataTestUser] {
		// In a real app, this would validate based on field tags
		return catalog.ValidationOutput[MetadataTestUser]{Error: nil}
	})
}