package task_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yadukrishnan2004/antrelay-sdk/task"
)

func TestNew(t *testing.T){
	input:=[]byte(`{"orderId": "123"}`)

	t.Run("create a task with correct field",func(t *testing.T){
		tk:=task.New("processOrder", input, "orders-queue",3)

		assert.NotEmpty(t,tk.ID)
		assert.Equal(t,"processOrder", tk.FunctionName)
		assert.Equal(t, input,tk.Input)
		assert.Equal(t, "orders-queue", tk.Queue)
		assert.Equal(t,3,tk.MaxRetries)
		assert.Equal(t,0,tk.RetryCount)
		assert.False(t, tk.CreatedAt.IsZero())
	})

	t.Run("each task get a uniqu id",func(t *testing.T){
		tk1 := task.New("processOrder", input, "orders-queue", 3)
		tk2 := task.New("processOrder", input, "orders-queue", 3)

		assert.NotEqual(t, tk1.ID, tk2.ID)
	})
}

// testing the can retry test 

func TestCanRetry(t *testing.T){
	t.Run("can retry when retry are left", func(t *testing.T){
		tk:=task.New("processOrder", nil, "orders-queue",3)
		tk.RetryCount=1

		assert.True(t,tk.CanRetry())
	})

	t.Run("cannot retry when max retries is zero", func(t *testing.T){
		tk := task.New("processOrder", nil, "orders-queue", 0)

		assert.False(t, tk.CanRetry())
	})
}

// TestIncrementRetry verifies the retry counter increases correctly.
func TestIncrementRetry(t *testing.T) {
	t.Run("increments retry count by one", func(t *testing.T) {
		tk := task.New("processOrder", nil, "orders-queue", 3)

		tk.IncrementRetry()
		assert.Equal(t, 1, tk.RetryCount)

		tk.IncrementRetry()
		assert.Equal(t, 2, tk.RetryCount)
	})
}


// TestNewResult verifies result creation for both success and failure cases.
func TestNewResult(t *testing.T) {
	t.Run("creates successful result", func(t *testing.T) {
		output := []byte(`{"status": "charged"}`)
		result := task.NewResult("task-001", output, nil)

		assert.Equal(t, "task-001", result.TaskID)
		assert.Equal(t, output, result.Output)
		assert.True(t, result.Success)
		assert.Empty(t, result.Error)
		assert.False(t, result.CreatedAt.IsZero())
	})

	t.Run("creates failed result with error message", func(t *testing.T) {
		result := task.NewResult("task-001", nil, errors.New("payment declined"))

		assert.False(t, result.Success)
		assert.Equal(t, "payment declined", result.Error)
		assert.Nil(t, result.Output)
	})
}