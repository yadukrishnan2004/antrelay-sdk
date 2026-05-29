package executor_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yadukrishnan2004/antrelay-sdk/executor"
	"github.com/yadukrishnan2004/antrelay-sdk/registry"
	"github.com/yadukrishnan2004/antrelay-sdk/task"
)

func makeRegistry(handlers map[string]interface{}) *registry.Registry {
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
		r := makeRegistry(map[string]interface{}{
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
		r := makeRegistry(map[string]interface{}{
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
		r := makeRegistry(map[string]interface{}{
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

		r := makeRegistry(map[string]interface{}{
			"processOrder": captureHandler,
		})
		ex := executor.New(r)
		input := []byte(`{"orderId":"123"}`)
		tk := task.New("processOrder", input, "orders-queue", 3)

		ex.Executor(context.Background(), tk)

		assert.Equal(t, input, receivedInput)
	})

	t.Run("result taskID matches the task that was executed", func(t *testing.T) {
		r := makeRegistry(map[string]interface{}{
			"processOrder": successHandler,
		})
		ex := executor.New(r)
		tk := task.New("processOrder", []byte(`{}`), "orders-queue", 3)

		result := ex.Executor(context.Background(), tk)

		assert.Equal(t, tk.ID, result.TaskID)
	})

	t.Run("returns error when handler exceeds timeout", func(t *testing.T) {
		slowHandler := func(ctx context.Context, input []byte) ([]byte, error) {
			select {
			case <-ctx.Done():
				return nil, ctx.Err() // respects cancellation
			case <-time.After(5 * time.Second): // simulates slow work
				return []byte(`{}`), nil
			}
		}

		r := makeRegistry(map[string]interface{}{
			"slowHandler": slowHandler,
		})

		// create executor with very short timeout
		ex := executor.New(r).WithTimeout(100 * time.Millisecond)
		tk := task.New("slowHandler", []byte(`{}`), "orders-queue", 3)

		result := ex.Executor(context.Background(), tk)

		assert.False(t, result.Success)
		assert.Contains(t, result.Error, "context deadline exceeded")
	})

	// --- NEW DYNAMIC REFLECTION TESTS ---

	t.Run("dynamic reflection: invokes a function with no context and custom types", func(t *testing.T) {
		type InputStruct struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}
		type OutputStruct struct {
			Greeting string `json:"greeting"`
		}

		customFn := func(in InputStruct) (OutputStruct, error) {
			return OutputStruct{Greeting: "Hello " + in.Name}, nil
		}

		r := makeRegistry(map[string]interface{}{
			"greet": customFn,
		})
		ex := executor.New(r)
		tk := task.New("greet", []byte(`{"name":"Yadhu","age":22}`), "orders-queue", 3)

		result := ex.Executor(context.Background(), tk)

		assert.True(t, result.Success)
		assert.Equal(t, []byte(`{"greeting":"Hello Yadhu"}`), result.Output)
	})

	t.Run("dynamic reflection: invokes a function with multiple parameters as JSON array", func(t *testing.T) {
		customFn := func(ctx context.Context, name string, amount float64) (string, error) {
			return name + " charged " + fmt.Sprintf("%.2f", amount), nil
		}

		r := makeRegistry(map[string]interface{}{
			"charge": customFn,
		})
		ex := executor.New(r)
		tk := task.New("charge", []byte(`["Alice",99.95]`), "orders-queue", 3)

		result := ex.Executor(context.Background(), tk)

		assert.True(t, result.Success)
		assert.Equal(t, []byte(`"Alice charged 99.95"`), result.Output)
	})

	t.Run("dynamic reflection: invokes a function with multiple return values", func(t *testing.T) {
		customFn := func(a int, b int) (int, int, error) {
			return a + b, a * b, nil
		}

		r := makeRegistry(map[string]interface{}{
			"math": customFn,
		})
		ex := executor.New(r)
		tk := task.New("math", []byte(`[3,4]`), "orders-queue", 3)

		result := ex.Executor(context.Background(), tk)

		assert.True(t, result.Success)
		assert.Equal(t, []byte(`[7,12]`), result.Output)
	})
}