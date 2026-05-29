package registry

import (
	"fmt"
	"reflect"
	"sync"
)

type Registry struct {
	mu       sync.RWMutex
	handlers map[string]interface{}
}

func New() *Registry {
	return &Registry{
		handlers: make(map[string]interface{}),
	}
}

// Register stores a handler function under the given name
func (r *Registry) Register(name string, handler interface{}) error {
	if name == "" {
		return fmt.Errorf("antrelay: handler name cannot be empty")
	}

	if handler == nil {
		return fmt.Errorf("antrelay: handler for %q cannot be nil", name)
	}

	// Verify that the handler is a function
	val := reflect.ValueOf(handler)
	if val.Kind() != reflect.Func {
		return fmt.Errorf("antrelay: handler for %q must be a function, got %s", name, val.Kind())
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exist := r.handlers[name]; exist {
		return fmt.Errorf("antrelay: handler %q is already registered", name)
	}

	r.handlers[name] = handler
	return nil
}

// Lookup retrieves a handler by name.
//
// Returns an error if no handler exists with that name.
func (r *Registry) Lookup(name string) (interface{}, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	handler, exists := r.handlers[name]

	if !exists {
		return nil, fmt.Errorf("antrelay: no handler registered for %q", name)
	}

	return handler, nil
}

// List returns all registered handler names.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.handlers))
	for name := range r.handlers {
		names = append(names, name)
	}
	return names
}


