package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/yadukrishnan2004/antrelay-sdk/registry"
	"github.com/yadukrishnan2004/antrelay-sdk/task"
)

// Executor is responsible for running a task by looking up
// its handler in the registry and executing it safely.
type Executor struct {
	registry       *registry.Registry
	handlerTimeout time.Duration
}

func New(r *registry.Registry) *Executor {
	return &Executor{
		registry:       r,
		handlerTimeout: 30 * time.Second,
	}
}

// WithTimeout returns a new Executor with a custom handler timeout.
func (e *Executor) WithTimeout(d time.Duration) *Executor {
	return &Executor{
		registry:       e.registry,
		handlerTimeout: d,
	}
}

func (e *Executor) Executor(ctx context.Context, t *task.Task) *task.Result {
	handler, err := e.registry.Lookup(t.FunctionName)
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
	handler interface{},
	input []byte,
)(output []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("executor: handler panicked: %v", r)
			output = nil
		}
	}()

	// If handler matches task.HandlerFunc exactly, run it directly for efficiency
	if h, ok := handler.(task.HandlerFunc); ok {
		return h(ctx, input)
	}
	if h, ok := handler.(func(context.Context, []byte) ([]byte, error)); ok {
		return h(ctx, input)
	}

	// Otherwise, call it dynamically using reflection (like Temporal)
	handlerVal := reflect.ValueOf(handler)
	handlerType := handlerVal.Type()
	numIn := handlerType.NumIn()

	inArgs := make([]reflect.Value, numIn)
	startIdx := 0

	// Bind context.Context if it is the first argument
	if numIn > 0 && handlerType.In(0).Implements(reflect.TypeOf((*context.Context)(nil)).Elem()) {
		inArgs[0] = reflect.ValueOf(ctx)
		startIdx = 1
	}

	remainingParams := numIn - startIdx
	if remainingParams > 0 {
		var rawArgs []json.RawMessage
		isArray := false
		if len(input) > 0 {
			if json.Unmarshal(input, &rawArgs) == nil {
				isArray = true
			}
		}

		if isArray && len(rawArgs) == remainingParams {
			for i := 0; i < remainingParams; i++ {
				paramType := handlerType.In(startIdx + i)
				if paramType == reflect.TypeOf([]byte(nil)) {
					inArgs[startIdx+i] = reflect.ValueOf(rawArgs[i])
				} else {
					paramVal := reflect.New(paramType)
					if err := json.Unmarshal(rawArgs[i], paramVal.Interface()); err != nil {
						return nil, fmt.Errorf("executor: failed to unmarshal argument %d: %w", i, err)
					}
					inArgs[startIdx+i] = paramVal.Elem()
				}
			}
		} else if remainingParams == 1 {
			paramType := handlerType.In(startIdx)
			if paramType == reflect.TypeOf([]byte(nil)) {
				inArgs[startIdx] = reflect.ValueOf(input)
			} else {
				paramVal := reflect.New(paramType)
				inputData := input
				if len(inputData) == 0 {
					inputData = []byte("null")
				}
				if err := json.Unmarshal(inputData, paramVal.Interface()); err != nil {
					return nil, fmt.Errorf("executor: failed to unmarshal argument: %w", err)
				}
				inArgs[startIdx] = paramVal.Elem()
			}
		} else {
			return nil, fmt.Errorf("executor: expected %d arguments, but input is not a JSON array of that size", remainingParams)
		}
	}

	// Dynamically invoke the function
	outVals := handlerVal.Call(inArgs)

	numOut := handlerType.NumOut()
	var errVal error

	// If the last return value is an error type, extract it
	if numOut > 0 {
		lastType := handlerType.Out(numOut - 1)
		if lastType.Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			if !outVals[numOut-1].IsNil() {
				errVal = outVals[numOut-1].Interface().(error)
			}
			numOut-- // Exclude error from processed outputs
		}
	}

	if errVal != nil {
		return nil, errVal
	}

	// Serialize result outputs
	if numOut == 1 {
		retVal := outVals[0].Interface()
		if byteSlice, ok := retVal.([]byte); ok {
			return byteSlice, nil
		}
		outBytes, err := json.Marshal(retVal)
		if err != nil {
			return nil, fmt.Errorf("executor: failed to marshal return value: %w", err)
		}
		return outBytes, nil
	} else if numOut > 1 {
		results := make([]interface{}, numOut)
		for i := 0; i < numOut; i++ {
			results[i] = outVals[i].Interface()
		}
		outBytes, err := json.Marshal(results)
		if err != nil {
			return nil, fmt.Errorf("executor: failed to marshal return values: %w", err)
		}
		return outBytes, nil
	}

	return nil, nil
}