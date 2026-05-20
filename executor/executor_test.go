package executor_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yadukrishnan2004/antrelay-sdk/executor"
	"github.com/yadukrishnan2004/antrelay-sdk/registry"
	"github.com/yadukrishnan2004/antrelay-sdk/task"
)


func makeRegistry(handlers map[string]task.HandlerFunc) *registry.Registry {
	r := registry.New()
	for name, h := range handlers {
		_ = r.Register(name, h)
	}
	return r
}

// successHandler simulates a handler that works correctly.
func successHandler(_ context.Context, input []byte) ([]byte, error) {
	return []byte(`{"status":"ok"}`), nil
}

// failureHandler simulates a handler that returns a business error.
func failureHandler(_ context.Context, input []byte) ([]byte, error) {
	return nil, errors.New("payment declined")
}

// panicHandler simulates a badly written handler that panics.
func panicHandler(_ context.Context, input []byte) ([]byte, error) {
	panic("something went very wrong")
}

func TestExecute(t *testing.T) {
	t.Run("returns successful result when handler succeeds", func(t *testing.T) {
		r := makeRegistry(map[string]task.HandlerFunc{
			"processOrder": successHandler,
		})
		ex := executor.New(r)
		tk := task.New("processOrder", []byte(`{}`), "orders-queue", 3)

		result := ex.Executor(context.Background(), tk)

		assert.True(t, result.Success)
		assert.Empty(t, result.Error)
		assert.Equal(t, []byte(`{"status":"ok"}`), result.Output)
		assert.Equal(t, tk.ID, result.TaskID)
	})

	t.Run("returns failed result when handler returns error", func(t *testing.T) {
		r := makeRegistry(map[string]task.HandlerFunc{
			"processOrder": failureHandler,
		})
		ex := executor.New(r)
		tk := task.New("processOrder", []byte(`{}`), "orders-queue", 3)

		result := ex.Executor(context.Background(), tk)

		assert.False(t, result.Success)
		assert.Equal(t, "payment declined", result.Error)
		assert.Nil(t, result.Output)
	})

	t.Run("returns failed result when no handler is registered", func(t *testing.T) {
		r := registry.New()
		ex := executor.New(r)
		tk := task.New("unknownFunction", []byte(`{}`), "orders-queue", 3)

		result := ex.Executor(context.Background(), tk)

		assert.False(t, result.Success)
		assert.Contains(t, result.Error, "lookup failed")
	})

	t.Run("recovers from panic and returns failed result", func(t *testing.T) {
		r := makeRegistry(map[string]task.HandlerFunc{
			"processOrder": panicHandler,
		})
		ex := executor.New(r)
		tk := task.New("processOrder", []byte(`{}`), "orders-queue", 3)

		// this should NOT panic the test process
		result := ex.Executor(context.Background(), tk)

		assert.False(t, result.Success)
		assert.Contains(t, result.Error, "panicked")
	})

	t.Run("passes input correctly to handler", func(t *testing.T) {
		var receivedInput []byte

		captureHandler := func(_ context.Context, input []byte) ([]byte, error) {
			receivedInput = input
			return []byte(`{}`), nil
		}

		r := makeRegistry(map[string]task.HandlerFunc{
			"processOrder": captureHandler,
		})
		ex := executor.New(r)
		input := []byte(`{"orderId":"123"}`)
		tk := task.New("processOrder", input, "orders-queue", 3)

		ex.Executor(context.Background(), tk)

		assert.Equal(t, input, receivedInput)
	})

	t.Run("result taskID matches the task that was executed", func(t *testing.T) {
		r := makeRegistry(map[string]task.HandlerFunc{
			"processOrder": successHandler,
		})
		ex := executor.New(r)
		tk := task.New("processOrder", []byte(`{}`), "orders-queue", 3)

		result := ex.Executor(context.Background(), tk)

		assert.Equal(t, tk.ID, result.TaskID)
	})
}