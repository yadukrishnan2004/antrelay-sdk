package executor

import (
	"context"
	"fmt"
	"time"

	"github.com/yadukrishnan2004/antrelay-sdk/registry"
	"github.com/yadukrishnan2004/antrelay-sdk/task"
)

// Executor is responsible for running a task by looking up
// its handler in the registry and executing it safely.

type Executor struct {
	registry *registry.Registry
	handlerTimeout time.Duration
}

func New(r *registry.Registry) *Executor {
	return &Executor{
		registry: r,
		handlerTimeout: 30 * time.Second,
	}
}

// WithTimeout returns a new Executor with a custom handler timeout.
func (e *Executor) WithTimeout(d time.Duration) *Executor{
	return &Executor{
		registry: e.registry,
		handlerTimeout: d,
	}
}

func (e *Executor) Executor(ctx context.Context, t *task.Task) *task.Result{
	handler,err:=e.registry.Lookup(t.FunctionName)
	if err != nil {
		return task.NewResult(t.ID, nil, fmt.Errorf("executor: lookup failed: %w", err))
	}

    // create a child context that cancels after handlerTimeout
    timeoutCtx, cancel := context.WithTimeout(ctx, e.handlerTimeout)
    defer cancel()

    output, err := e.safeRun(timeoutCtx, handler, t.Input)
	return task.NewResult(t.ID, output, err)
}


func (e *Executor) safeRun(
	ctx context.Context,
	handler task.HandlerFunc,
	input []byte,
)(output []byte,err error){
	defer func(){
		if r:=recover();r != nil{
			err = fmt.Errorf("executor: handler panicked: %v", r)
			output = nil
		}
	}()

	return handler(ctx, input)
}