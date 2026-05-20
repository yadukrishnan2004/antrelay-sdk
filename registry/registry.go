package registry

import (
	"fmt"
	"sync"

	"github.com/yadukrishnan2004/antrelay-sdk/task"
)


type Registry struct{
	mu sync.RWMutex
	handlers map[string]task.HandlerFunc
}

func New() *Registry{
	return &Registry{
		handlers: make(map[string]task.HandlerFunc),
	}
}

// Register stores a handler function under the given name

func (r *Registry) Register(name string , handler task.HandlerFunc)error{
	if name == ""{
		return fmt.Errorf("antrelay: handler name cannot be empty")
	}

	if handler == nil{
		return fmt.Errorf("antrelay: handler for %q cannot be nil", name)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _,exist:=r.handlers[name];exist{
		return fmt.Errorf("antrelay: handler %q is already registered", name)
	}

	r.handlers[name] = handler
	return nil
}


// Lookup retrieves a handler by name.
//
// Returns an error if no handler exists with that name.

func (r *Registry) Lookup(name string)(task.HandlerFunc,error){
	r.mu.Lock()
	defer r.mu.Unlock()

	handler,exists:=r.handlers[name]

	if !exists{
		return nil, fmt.Errorf("antrelay: no handler registered for %q", name)
	}

	return handler,nil
}

// List returns all registered handler names.
func (r *Registry) List() []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	names:=make([]string,0,len(r.handlers))
	for name:=range r.handlers{
		names=append(names, name)
	}
	return names
}


