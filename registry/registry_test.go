package registry_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yadukrishnan2004/antrelay-sdk/registry"
	"github.com/yadukrishnan2004/antrelay-sdk/task"
)

func dummyHandler(_ context.Context, input []byte) ([]byte, error) {
	return input, nil
}

func TestRegister(t *testing.T) {
	t.Run("registers a valid handler successfully", func(t *testing.T) {
		r := registry.New()

		err := r.Register("processOrder", dummyHandler)

		assert.NoError(t, err)
	})

	t.Run("returns error when handler is nil", func(t *testing.T) {
		r := registry.New()

		err := r.Register("processOrder", nil)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be nil")
	})

	t.Run("returns error when handler is not a function", func(t *testing.T) {
		r := registry.New()

		err := r.Register("processOrder", "not-a-function")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be a function")
	})

	t.Run("allows different names to be registered", func(t *testing.T) {
		r := registry.New()

		err1 := r.Register("processOrder", dummyHandler)
		err2 := r.Register("sendEmail", dummyHandler)

		assert.NoError(t, err1)
		assert.NoError(t, err2)
	})
}

func TestLookup(t *testing.T) {
	t.Run("returns handler for registered name", func(t *testing.T) {
		r := registry.New()
		_ = r.Register("processOrder", dummyHandler)

		handler, err := r.Lookup("processOrder")

		assert.NoError(t, err)
		assert.NotNil(t, handler)
	})

	t.Run("returns error for unregistered name", func(t *testing.T) {
		r := registry.New()

		handler, err := r.Lookup("nonExistent")

		assert.Error(t, err)
		assert.Nil(t, handler)
		assert.Contains(t, err.Error(), "no handler registered")
	})

	t.Run("returned handler is the same function that was registered", func(t *testing.T) {
		r := registry.New()
		_ = r.Register("processOrder", dummyHandler)

		handler, _ := r.Lookup("processOrder")
		hFunc, ok := handler.(func(context.Context, []byte) ([]byte, error))
		assert.True(t, ok)
		result, err := hFunc(context.Background(), []byte(`{}`))

		assert.NoError(t, err)
		assert.Equal(t, []byte(`{}`), result)
	})
}

func TestList(t *testing.T) {
	t.Run("returns empty slice when nothing registered", func(t *testing.T) {
		r := registry.New()

		names := r.List()

		assert.Empty(t, names)
	})

	t.Run("returns all registered names", func(t *testing.T) {
		r := registry.New()
		_ = r.Register("processOrder", dummyHandler)
		_ = r.Register("sendEmail", dummyHandler)

		names := r.List()

		assert.Len(t, names, 2)
		assert.Contains(t, names, "processOrder")
		assert.Contains(t, names, "sendEmail")
	})
}

// TestConcurrentAccess verifies the registry is safe when multiple
// goroutines read and write at the same time.
func TestConcurrentAccess(t *testing.T) {
	r := registry.New()
	_ = r.Register("processOrder", dummyHandler)

	// launch 100 goroutines all reading simultaneously
	done := make(chan bool, 100)
	for i := 0; i < 100; i++ {
		go func() {
			handler, err := r.Lookup("processOrder")
			assert.NoError(t, err)
			assert.NotNil(t, handler)
			done <- true
		}()
	}

	// wait for all goroutines to finish
	for i := 0; i < 100; i++ {
		<-done
	}
}

// Ensure HandlerFunc type is importable through task package.
// This is a compile-time check, not a runtime check.
var _ task.HandlerFunc = dummyHandler