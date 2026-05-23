package reporter_test

import (
		"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yadukrishnan2004/antrelay-sdk/reporter"
	"github.com/yadukrishnan2004/antrelay-sdk/task"
)

func makeResult(taskID string, success bool) *task.Result {
	if success {
		return task.NewResult(taskID, []byte(`{}`), nil)
	}
	return task.NewResult(taskID, nil, errors.New("task failed"))
}

func TestMockReporter(t *testing.T) {

	t.Run("returns no error when reporting succeeds", func(t *testing.T) {
		r := reporter.NewMock()
		result := makeResult("task-001", true)

		err := r.Report(result)

		assert.NoError(t, err)
	})

	t.Run("stores reported result", func(t *testing.T) {
		r := reporter.NewMock()
		result := makeResult("task-001", true)

		r.Report(result)

		assert.Equal(t, 1, r.Len())
		assert.Equal(t, "task-001", r.Results()[0].TaskID)
	})

	t.Run("stores multiple results in order", func(t *testing.T) {
		r := reporter.NewMock()

		r.Report(makeResult("task-001", true))
		r.Report(makeResult("task-002", true))
		r.Report(makeResult("task-003", true))

		results := r.Results()
		assert.Len(t, results, 3)
		assert.Equal(t, "task-001", results[0].TaskID)
		assert.Equal(t, "task-002", results[1].TaskID)
		assert.Equal(t, "task-003", results[2].TaskID)
	})

	t.Run("returns error when error is set", func(t *testing.T) {
		r := reporter.NewMock()
		r.SetError(errors.New("server unreachable"))

		err := r.Report(makeResult("task-001", true))

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "server unreachable")
	})

	t.Run("does not store result when error is set", func(t *testing.T) {
		r := reporter.NewMock()
		r.SetError(errors.New("server unreachable"))

		r.Report(makeResult("task-001", true))

		assert.Equal(t, 0, r.Len())
	})

	t.Run("Len reflects stored result count", func(t *testing.T) {
		r := reporter.NewMock()

		assert.Equal(t, 0, r.Len())

		assert.Equal(t, 0, r.Len())

		r.Report(makeResult("task-002", true))
		assert.Equal(t, 1, r.Len())
	})

	t.Run("Results returns a copy not the internal slice", func(t *testing.T) {
		r := reporter.NewMock()
		r.Report(makeResult("task-001", true))

		snapshot := r.Results()
		snapshot[0] = nil // mutate the returned slice

		// internal state must be unaffected
		assert.NotNil(t, r.Results()[0])
	})
}