package poller_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yadukrishnan2004/antrelay-sdk/poller"
	"github.com/yadukrishnan2004/antrelay-sdk/task"
)

func TestMockPoller(t *testing.T) {
	t.Run("returns nil when no tasks are queued", func(t *testing.T) {
		p := poller.NewMock()

		result, err := p.Poll("orders-queue")

		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("returns queued task when available", func(t *testing.T) {
		p := poller.NewMock()
		tk := task.New("processOrder", []byte(`{}`), "orders-queue", 3)
		p.QueueTask(tk)

		result, err := p.Poll("orders-queue")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, tk.ID, result.ID)
		assert.Equal(t, "processOrder", result.FunctionName)
	})

	t.Run("returns tasks in FIFO order", func(t *testing.T) {
		p := poller.NewMock()
		first := task.New("processOrder", []byte(`{}`), "orders-queue", 3)
		second := task.New("sendEmail", []byte(`{}`), "orders-queue", 3)

		p.QueueTask(first)
		p.QueueTask(second)

		result1, _ := p.Poll("orders-queue")
		result2, _ := p.Poll("orders-queue")

		assert.Equal(t, first.ID, result1.ID)
		assert.Equal(t, second.ID, result2.ID)
	})

	t.Run("returns nil after all tasks consumed", func(t *testing.T) {
		p := poller.NewMock()
		tk := task.New("processOrder", []byte(`{}`), "orders-queue", 3)
		p.QueueTask(tk)

		p.Poll("orders-queue") // consume the task

		result, err := p.Poll("orders-queue") // nothing left
		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("returns error when error is set", func(t *testing.T) {
		p := poller.NewMock()
		p.SetError(errors.New("connection refused"))

		result, err := p.Poll("orders-queue")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "connection refused")
	})

	t.Run("Len reflects remaining tasks", func(t *testing.T) {
		p := poller.NewMock()
		p.QueueTask(task.New("processOrder", []byte(`{}`), "orders-queue", 3))
		p.QueueTask(task.New("sendEmail", []byte(`{}`), "orders-queue", 3))

		assert.Equal(t, 2, p.Len())

		p.Poll("orders-queue")
		assert.Equal(t, 1, p.Len())

		p.Poll("orders-queue")
		assert.Equal(t, 0, p.Len())
	})
}