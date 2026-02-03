package aegis

import (
	"context"
	"testing"
)

func TestNewFunctionRegistry(t *testing.T) {
	fr := NewFunctionRegistry()
	
	if fr.Count() != 0 {
		t.Errorf("expected 0 functions, got %d", fr.Count())
	}
	
	if len(fr.ListFunctions()) != 0 {
		t.Errorf("expected empty function list, got %v", fr.ListFunctions())
	}
}

func TestFunctionRegistryRegister(t *testing.T) {
	fr := NewFunctionRegistry()
	
	testFunc := func(ctx context.Context, params []string) (string, error) {
		return "test result", nil
	}
	
	err := fr.Register("test-func", testFunc)
	if err != nil {
		t.Errorf("failed to register function: %v", err)
	}
	
	if fr.Count() != 1 {
		t.Errorf("expected 1 function, got %d", fr.Count())
	}
	
	if !fr.HasFunction("test-func") {
		t.Error("expected function to be registered")
	}
	
	functions := fr.ListFunctions()
	if len(functions) != 1 || functions[0] != "test-func" {
		t.Errorf("expected ['test-func'], got %v", functions)
	}
}

func TestFunctionRegistryRegisterErrors(t *testing.T) {
	fr := NewFunctionRegistry()
	
	testFunc := func(ctx context.Context, params []string) (string, error) {
		return "test", nil
	}
	
	err := fr.Register("", testFunc)
	if err == nil {
		t.Error("expected error for empty function name")
	}
	
	err = fr.Register("test-func", nil)
	if err == nil {
		t.Error("expected error for nil function")
	}
	
	err = fr.Register("valid-func", testFunc)
	if err != nil {
		t.Errorf("failed to register valid function: %v", err)
	}
	
	err = fr.Register("valid-func", testFunc)
	if err == nil {
		t.Error("expected error for duplicate function name")
	}
}

func TestFunctionRegistryExecute(t *testing.T) {
	fr := NewFunctionRegistry()
	
	testFunc := func(ctx context.Context, params []string) (string, error) {
		if len(params) == 0 {
			return "no params", nil
		}
		return "got: " + params[0], nil
	}
	
	fr.Register("test-func", testFunc)
	
	ctx := context.Background()
	
	result, err := fr.Execute(ctx, "test-func", []string{})
	if err != nil {
		t.Errorf("failed to execute function: %v", err)
	}
	if result != "no params" {
		t.Errorf("expected 'no params', got '%s'", result)
	}
	
	result, err = fr.Execute(ctx, "test-func", []string{"hello"})
	if err != nil {
		t.Errorf("failed to execute function with params: %v", err)
	}
	if result != "got: hello" {
		t.Errorf("expected 'got: hello', got '%s'", result)
	}
	
	_, err = fr.Execute(ctx, "nonexistent", []string{})
	if err == nil {
		t.Error("expected error for nonexistent function")
	}
}

func TestFunctionRegistryUnregister(t *testing.T) {
	fr := NewFunctionRegistry()
	
	testFunc := func(ctx context.Context, params []string) (string, error) {
		return "test", nil
	}
	
	fr.Register("test-func", testFunc)
	
	if fr.Count() != 1 {
		t.Errorf("expected 1 function, got %d", fr.Count())
	}
	
	err := fr.Unregister("test-func")
	if err != nil {
		t.Errorf("failed to unregister function: %v", err)
	}
	
	if fr.Count() != 0 {
		t.Errorf("expected 0 functions after unregister, got %d", fr.Count())
	}
	
	if fr.HasFunction("test-func") {
		t.Error("function should not exist after unregister")
	}
	
	err = fr.Unregister("nonexistent")
	if err == nil {
		t.Error("expected error unregistering nonexistent function")
	}
}

func TestNodeFunctionMethods(t *testing.T) {
	node := NewNode("test-node", "Test Node", NodeTypeGeneric, "localhost:8080")
	
	if node.Functions == nil {
		t.Error("expected function registry to be initialized")
	}
	
	testFunc := func(ctx context.Context, params []string) (string, error) {
		return "node function result", nil
	}
	
	err := node.RegisterFunction("node-func", testFunc)
	if err != nil {
		t.Errorf("failed to register function on node: %v", err)
	}
	
	functions := node.ListFunctions()
	if len(functions) != 1 || functions[0] != "node-func" {
		t.Errorf("expected ['node-func'], got %v", functions)
	}
	
	ctx := context.Background()
	result, err := node.ExecuteFunction(ctx, "node-func", []string{})
	if err != nil {
		t.Errorf("failed to execute function on node: %v", err)
	}
	if result != "node function result" {
		t.Errorf("expected 'node function result', got '%s'", result)
	}
	
	err = node.UnregisterFunction("node-func")
	if err != nil {
		t.Errorf("failed to unregister function on node: %v", err)
	}
	
	functions = node.ListFunctions()
	if len(functions) != 0 {
		t.Errorf("expected empty function list, got %v", functions)
	}
}

func TestNodeExecutePeerFunction(t *testing.T) {
	node := NewNode("test-node", "Test Node", NodeTypeGeneric, "localhost:8080")
	
	ctx := context.Background()
	
	_, err := node.ExecutePeerFunction(ctx, "nonexistent-peer", "some-func", []string{})
	if err == nil {
		t.Error("expected error calling function on nonexistent peer")
	}
}