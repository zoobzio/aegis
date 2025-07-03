package extensibility

import (
	"fmt"

	"github.com/spf13/cobra"
	
	"aegis/catalog"
)

// runConventionTests runs all convention discovery tests
func runConventionTests(cmd *cobra.Command, args []string) {
	runTest("Interface Detection", TestInterfaceDetection)
	runTest("Convention Registration", TestConventionRegistration)
	runTest("Automatic Discovery", TestAutomaticDiscovery)
	runTest("Convention Composition", TestConventionComposition)
	runTest("Convention Evolution", TestConventionEvolution)
	runTest("Convention Chaining", TestConventionChaining)
	runTest("Convention Inheritance", TestConventionInheritance)
}

// Test types that implement conventions
type ConventionUser struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// Implement HasDefaults convention
func (u ConventionUser) Defaults() ConventionUser {
	return ConventionUser{
		ID:    "default-user",
		Email: "user@example.com", 
		Role:  "user",
	}
}

// Implement HasScope convention
func (u ConventionUser) Scope() string {
	switch u.Role {
	case "admin":
		return "admin:all"
	case "hr":
		return "hr:read"
	default:
		return "user:read"
	}
}

type AdminUser struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	Permissions []string `json:"permissions"`
}

// Only implement HasScope convention
func (a AdminUser) Scope() string {
	return "admin:all"
}

// EvolvingUser type for testing convention evolution
type EvolvingUser struct {
	ID    string
	Name  string
	Level int
}

// Implement HasScope interface
func (u EvolvingUser) Scope() string {
	if u.Level > 10 {
		return "power:user"
	}
	return "basic:user"
}

// Implement HasDefaults interface
func (u EvolvingUser) Defaults() EvolvingUser {
	if u.Name == "" {
		u.Name = "Anonymous"
	}
	if u.Level == 0 {
		u.Level = 1
	}
	return u
}

// ChainableUser for testing chaining conventions
type ChainableUser struct {
	ID         string
	Processed  []string
	Transforms int
}

// Base type with conventions
type BaseEntity struct {
	ID      string
	Created string
}

// ExtendedUser composed type
type ExtendedUser struct {
	BaseEntity
	Email string
	Role  string
}

// Implement HasDefaults for ExtendedUser
func (u ExtendedUser) Defaults() ExtendedUser {
	if u.ID == "" {
		u.ID = "generated-id"
	}
	if u.Created == "" {
		u.Created = "now"
	}
	if u.Role == "" {
		u.Role = "user"
	}
	return u
}

// TestInterfaceDetection verifies types can implement convention interfaces
func TestInterfaceDetection() error {
	// Test that framework can detect when types implement conventions
	
	// ConventionUser implements both HasDefaults and HasScope
	var user ConventionUser
	
	// Test interface detection - this would typically be done by reflection
	// but we'll test the convention system directly
	
	// Check if type can provide defaults
	defaultHandler, hasDefaults := catalog.GetDefaults[ConventionUser]()
	
	// Check if type can provide scope  
	scopeHandler, hasScope := catalog.GetScope[ConventionUser]()
	
	// For now, if no convention is registered, these will return false
	// But the types themselves implement the interfaces
	_ = defaultHandler
	_ = scopeHandler
	_ = hasDefaults
	_ = hasScope
	
	// Verify the actual interface methods work
	defaults := user.Defaults()
	if defaults.Email != "user@example.com" {
		return fmt.Errorf("interface method failed: expected user@example.com, got %s", defaults.Email)
	}
	
	scope := user.Scope()
	if scope != "user:read" {
		return fmt.Errorf("interface method failed: expected user:read, got %s", scope)
	}
	
	return nil
}

// TestConventionRegistration verifies convention handler registration
func TestConventionRegistration() error {
	// Test registering convention handlers for types that implement interfaces
	
	// Register a defaults handler for ConventionUser
	catalog.RegisterDefaults[ConventionUser](func(u ConventionUser) ConventionUser {
		// Custom defaults logic using the type's convention method
		return u.Defaults()
	})
	
	// Register a scope handler for ConventionUser
	catalog.RegisterScope[ConventionUser](func(u ConventionUser) string {
		// Custom scope logic using the type's convention method
		return u.Scope()
	})
	
	// Verify registration worked
	defaultHandler, exists := catalog.GetDefaults[ConventionUser]()
	if !exists {
		return fmt.Errorf("defaults convention not registered")
	}
	
	scopeHandler, exists := catalog.GetScope[ConventionUser]()
	if !exists {
		return fmt.Errorf("scope convention not registered")
	}
	
	// Test the registered handlers
	user := ConventionUser{Role: "admin"}
	defaults := defaultHandler(user)
	if defaults.Email != "user@example.com" {
		return fmt.Errorf("registered handler failed")
	}
	
	scope := scopeHandler(user)
	if scope != "admin:all" {
		return fmt.Errorf("scope handler failed: expected admin:all, got %s", scope)
	}
	
	return nil
}

// TestAutomaticDiscovery verifies convention-based capability discovery
func TestAutomaticDiscovery() error {
	// Test that framework can automatically discover and use conventions
	
	// Create different users with different roles
	adminUser := ConventionUser{ID: "1", Role: "admin"}
	hrUser := ConventionUser{ID: "2", Role: "hr"}  
	regularUser := ConventionUser{ID: "3", Role: "user"}
	
	// Get scope convention handler
	scopeHandler, exists := catalog.GetScope[ConventionUser]()
	if !exists {
		return fmt.Errorf("scope convention not available")
	}
	
	// Test automatic scope detection based on user role
	adminScope := scopeHandler(adminUser)
	if adminScope != "admin:all" {
		return fmt.Errorf("admin scope detection failed: got %s", adminScope)
	}
	
	hrScope := scopeHandler(hrUser)
	if hrScope != "hr:read" {
		return fmt.Errorf("hr scope detection failed: got %s", hrScope)
	}
	
	userScope := scopeHandler(regularUser)
	if userScope != "user:read" {
		return fmt.Errorf("user scope detection failed: got %s", userScope)
	}
	
	return nil
}

// TestConventionComposition verifies multiple conventions work together
func TestConventionComposition() error {
	// Test using multiple conventions together
	
	// Register conventions for AdminUser (only has scope)
	catalog.RegisterScope[AdminUser](func(a AdminUser) string {
		return a.Scope()
	})
	
	// Test composition - AdminUser has scope but not defaults
	_, hasDefaults := catalog.GetDefaults[AdminUser]()
	if hasDefaults {
		return fmt.Errorf("AdminUser should not have defaults convention")
	}
	
	scopeHandler, hasScope := catalog.GetScope[AdminUser]()
	if !hasScope {
		return fmt.Errorf("AdminUser should have scope convention")
	}
	
	// Test that scope works
	admin := AdminUser{ID: "admin-1"}
	scope := scopeHandler(admin)
	if scope != "admin:all" {
		return fmt.Errorf("admin scope failed: got %s", scope)
	}
	
	// Test that different types have different conventions
	userScopeHandler, _ := catalog.GetScope[ConventionUser]()
	userDefaultsHandler, _ := catalog.GetDefaults[ConventionUser]()
	
	if userScopeHandler == nil || userDefaultsHandler == nil {
		return fmt.Errorf("ConventionUser should have both conventions")
	}
	
	// Verify type isolation - each type gets its own convention space
	user := ConventionUser{Role: "user"}
	userScope := userScopeHandler(user)
	if userScope == scope {
		return fmt.Errorf("convention types should be isolated")
	}
	
	return nil
}

// TestConventionEvolution verifies conventions can evolve over time
func TestConventionEvolution() error {
	// Initially, register just one convention
	catalog.RegisterScope[EvolvingUser](func(u EvolvingUser) string {
		return u.Scope()
	})
	
	// Verify initial state
	_, hasDefaults := catalog.GetDefaults[EvolvingUser]()
	scopeHandler, hasScope := catalog.GetScope[EvolvingUser]()
	
	if hasDefaults {
		return fmt.Errorf("should not have defaults initially")
	}
	
	if !hasScope {
		return fmt.Errorf("should have scope convention")
	}
	
	// Test initial scope
	user := EvolvingUser{Level: 15}
	scope := scopeHandler(user)
	if scope != "power:user" {
		return fmt.Errorf("initial scope failed: got %s", scope)
	}
	
	// Later, add defaults convention (simulating evolution)
	catalog.RegisterDefaults[EvolvingUser](func(u EvolvingUser) EvolvingUser {
		return u.Defaults()
	})
	
	// Verify evolved state
	defaultsHandler, hasDefaults := catalog.GetDefaults[EvolvingUser]()
	if !hasDefaults {
		return fmt.Errorf("should have defaults after evolution")
	}
	
	// Test defaults
	emptyUser := EvolvingUser{}
	withDefaults := defaultsHandler(emptyUser)
	if withDefaults.Name != "Anonymous" || withDefaults.Level != 1 {
		return fmt.Errorf("defaults not applied correctly")
	}
	
	return nil
}

// TestConventionChaining verifies conventions can chain/compose
func TestConventionChaining() error {
	
	// Register multiple transformation conventions
	catalog.RegisterTransformer[ChainableUser](func(u ChainableUser) ChainableUser {
		u.Processed = append(u.Processed, "sanitize")
		u.Transforms++
		return u
	})
	
	// Get transformer and test
	transformer, exists := catalog.GetTransformer[ChainableUser]()
	if !exists {
		return fmt.Errorf("transformer not registered")
	}
	
	// Apply transformation - need to type assert since GetTransformer returns any
	user := ChainableUser{ID: "123"}
	
	// Type assert the transformer
	tf, ok := transformer.(func(ChainableUser) ChainableUser)
	if !ok {
		return fmt.Errorf("transformer has wrong type")
	}
	
	transformed := tf(user)
	
	if len(transformed.Processed) != 1 || transformed.Processed[0] != "sanitize" {
		return fmt.Errorf("transformation not applied")
	}
	
	if transformed.Transforms != 1 {
		return fmt.Errorf("transform count incorrect")
	}
	
	// Chain another transformation
	transformed2 := tf(transformed)
	if transformed2.Transforms != 2 {
		return fmt.Errorf("chaining failed")
	}
	
	return nil
}

// TestConventionInheritance verifies convention patterns with embedded types
func TestConventionInheritance() error {
	// Register convention for extended type
	catalog.RegisterDefaults[ExtendedUser](func(u ExtendedUser) ExtendedUser {
		return u.Defaults()
	})
	
	// Test defaults on composed type
	defaultsHandler, exists := catalog.GetDefaults[ExtendedUser]()
	if !exists {
		return fmt.Errorf("defaults not registered for ExtendedUser")
	}
	
	empty := ExtendedUser{}
	withDefaults := defaultsHandler(empty)
	
	// Verify both embedded and direct fields get defaults
	if withDefaults.ID != "generated-id" {
		return fmt.Errorf("embedded field default failed")
	}
	
	if withDefaults.Role != "user" {
		return fmt.Errorf("direct field default failed")
	}
	
	return nil
}