package moisten

import (
	"sync"

	"aegis/capitan"
	"aegis/catalog"
	"aegis/pipz"
	"aegis/zlog"
)

var (
	initialized bool
	initMutex   sync.Mutex
)

// zlogFieldProcessor wraps pipz contract to implement zlog's interface
// This thin adapter is only needed because Go won't auto-convert between:
//   - func(ZlogField) []ZlogField (what zlog expects)
//   - pipz.Processor[ZlogField, []ZlogField] (what pipz uses)
//
// Even though they're the same underlying type, Go requires explicit conversion.
// This proves that zlog field processing IS a pipz pipeline!
type zlogFieldProcessor struct {
	contract *pipz.ServiceContract[zlog.ZlogFieldType, zlog.ZlogField, []zlog.ZlogField]
}

func (z *zlogFieldProcessor) Register(fieldType zlog.ZlogFieldType, processor func(zlog.ZlogField) []zlog.ZlogField) {
	z.contract.Register(fieldType, pipz.Processor[zlog.ZlogField, []zlog.ZlogField](processor))
}

func (z *zlogFieldProcessor) Process(fieldType zlog.ZlogFieldType, field zlog.ZlogField) ([]zlog.ZlogField, bool) {
	return z.contract.Process(fieldType, field)
}

// BehaviorInitializer is a function that registers behaviors during initialization
type BehaviorInitializer func()

// ForTesting initializes the framework for testing purposes
// This function is idempotent - multiple calls will only initialize once
func ForTesting(behaviorInitializers ...BehaviorInitializer) {
	initMutex.Lock()
	defer initMutex.Unlock()

	// Already initialized, nothing to do
	if initialized {
		return
	}

	// 1. Hydrate pipz with capitan event emission
	pipzEventSink := capitan.CreatePipzEventSink()
	pipz.SetEventEmitter(pipzEventSink)

	// 2. Create pipz contract for zlog field processing
	fieldContract := pipz.GetContract[zlog.ZlogFieldType, zlog.ZlogField, []zlog.ZlogField]()

	// 3. Create inline adapter - just wraps pipz contract to match zlog's interface
	// This is needed because Go won't implicitly convert between function type aliases
	zlog.SetFieldProcessorContract(&zlogFieldProcessor{fieldContract})

	// 4. Hydrate zlog with capitan event emission
	eventSink := capitan.CreateZlogEventSink()
	zlog.SetEventSink(eventSink)

	// 5. Set up catalog to auto-register tags from behavior registrations
	// This enables emergent behavior - when behaviors care about tags, those tags are auto-registered
	catalog.AutoRegisterTagsFromBehaviors()

	// Hook into pipz events to notify catalog when processors are registered
	capitan.Hook[pipz.ProcessorRegisteredEvent](func(event pipz.ProcessorRegisteredEvent) error {
		// Bridge the event to catalog's handler
		if handler := catalog.GetPipzEventHandler(); handler != nil {
			handler.OnProcessorRegistered(event.ContractSignature, event.KeyTypeName, event.KeyValue)
		}
		return nil
	})

	// 6. Run behavior initializers BEFORE any user code runs
	// This ensures tags are registered before metadata extraction
	for _, initializer := range behaviorInitializers {
		initializer()
	}

	// 7. Configure zlog for testing
	zlog.Configure(zlog.DEBUG)

	// Log successful initialization
	zlog.Info("Framework initialized for testing",
		zlog.String("mode", "testing"),
		zlog.String("log_level", "DEBUG"),
	)

	// Mark as initialized
	initialized = true
}
