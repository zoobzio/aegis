package aegis

import (
	"context"
	"fmt"
	"sync"
)

type NodeFunction func(ctx context.Context, parameters []string) (string, error)

type FunctionRegistry struct {
	functions map[string]NodeFunction
	mu        sync.RWMutex
}

func NewFunctionRegistry() *FunctionRegistry {
	return &FunctionRegistry{
		functions: make(map[string]NodeFunction),
	}
}

func (fr *FunctionRegistry) Register(name string, fn NodeFunction) error {
	fr.mu.Lock()
	defer fr.mu.Unlock()
	
	if name == "" {
		return fmt.Errorf("function name cannot be empty")
	}
	
	if fn == nil {
		return fmt.Errorf("function cannot be nil")
	}
	
	if _, exists := fr.functions[name]; exists {
		return fmt.Errorf("function %s already registered", name)
	}
	
	fr.functions[name] = fn
	return nil
}

func (fr *FunctionRegistry) Unregister(name string) error {
	fr.mu.Lock()
	defer fr.mu.Unlock()
	
	if _, exists := fr.functions[name]; !exists {
		return fmt.Errorf("function %s not found", name)
	}
	
	delete(fr.functions, name)
	return nil
}

func (fr *FunctionRegistry) Execute(ctx context.Context, name string, parameters []string) (string, error) {
	fr.mu.RLock()
	fn, exists := fr.functions[name]
	fr.mu.RUnlock()
	
	if !exists {
		return "", fmt.Errorf("function %s not found", name)
	}
	
	return fn(ctx, parameters)
}

func (fr *FunctionRegistry) ListFunctions() []string {
	fr.mu.RLock()
	defer fr.mu.RUnlock()
	
	names := make([]string, 0, len(fr.functions))
	for name := range fr.functions {
		names = append(names, name)
	}
	return names
}

func (fr *FunctionRegistry) HasFunction(name string) bool {
	fr.mu.RLock()
	defer fr.mu.RUnlock()
	
	_, exists := fr.functions[name]
	return exists
}

func (fr *FunctionRegistry) Count() int {
	fr.mu.RLock()
	defer fr.mu.RUnlock()
	return len(fr.functions)
}