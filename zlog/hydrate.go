package zlog

// ZlogFieldProcessorContract provides field processing without zlog importing pipz
// This interface is designed to be satisfied by pipz.ServiceContract with a thin adapter.
// The methods match pipz's signature exactly, just without the pipz.Processor type alias.
type ZlogFieldProcessorContract interface {
	Register(fieldType ZlogFieldType, processor func(ZlogField) []ZlogField)
	Process(fieldType ZlogFieldType, field ZlogField) ([]ZlogField, bool)
}

var fieldProcessorContract ZlogFieldProcessorContract

// SetFieldProcessorContract allows pipz to hydrate zlog with field processing capability
func SetFieldProcessorContract(contract ZlogFieldProcessorContract) {
	fieldProcessorContract = contract
	Info("Field processing capabilities enabled")
}

// RegisterFieldProcessor registers a field processor through hydrated contract
func RegisterFieldProcessor(fieldType ZlogFieldType, processor ZlogFieldProcessor) {
	if fieldProcessorContract == nil {
		Fatal("zlog not hydrated with ZlogFieldProcessorContract",
			String("required", "pipz must be imported"),
			String("violation", "nuclear architecture broken"),
		)
	}
	
	// Convert ZlogFieldProcessor to contract signature
	contractProcessor := func(field ZlogField) []ZlogField {
		return processor(field)
	}
	
	fieldProcessorContract.Register(fieldType, contractProcessor)
}

// processFields processes fields through hydrated contract
func processFields(fields []ZlogField) []ZlogField {
	if fieldProcessorContract == nil {
		// No field processing available - return fields as-is
		return fields
	}
	
	var processed []ZlogField
	for _, field := range fields {
		if result, exists := fieldProcessorContract.Process(field.Type, field); exists {
			// Custom processor handles this field type
			processed = append(processed, result...)
		} else {
			// No processor, keep field as-is
			processed = append(processed, field)
		}
	}
	
	return processed
}