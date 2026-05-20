package executor

import (
	"context"
	"fmt"

	"github.com/yadukrishnan2004/antrelay-sdk/registry"
	"github.com/yadukrishnan2004/antrelay-sdk/task"
)

// Executor is responsible for running a task by looking up
// its handler in the registry and executing it safely.

type Executor struct {
	registry *registry.Registry
}

func New(r *registry.Registry) *Executor {
	return &Executor{
		registry: r,
	}
}

func (e *Executor) Executor(ctx context.Context, t *task.Task) *task.Result{
	handler,err:=e.registry.Lookup(t.FunctionName)
	if err != nil {
		return task.NewResult(t.ID, nil, fmt.Errorf("executor: lookup failed: %w", err))
	}

	output,err:=e.safeRun(ctx,handler,t.Input)
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