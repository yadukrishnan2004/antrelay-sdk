package task

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// When a developer's API server says "start workflow processOrder",
// AntRelay creates a Task and puts it in a queue.
// The worker picks it up, finds the registered function by FunctionName,
// and executes it with Input as the argument.
type Task struct {
	ID           string
	FunctionName string
	Input        []byte
	Queue        string
	CreatedAt    time.Time
	MaxRetries   int
	RetryCount   int
}


// The worker sends this back to the server after execution.
// Success tells the server whether to move the workflow forward
// or to retry / mark it as failed.
type Result struct {
	TaskID    string
	Output    []byte
	Error     string
	Success   bool
	CreatedAt time.Time
}


// Every function the developer writes must match this signature.
// Context carries cancellation and timeout signals.

type HandlerFunc func(ctx context.Context, input []byte) ([]byte, error)

//creating new task with generated id

func New(functionName string, input []byte, queue string, maxRetries int) *Task{
	return &Task{
		ID: uuid.NewString(),
		FunctionName: functionName,
		Input: input,
		Queue: queue,
		CreatedAt: time.Now().UTC(),
		MaxRetries: maxRetries,
		RetryCount: 0,
	}
}


// NewResult creates a Result from a completed or failed task execution.

func NewResult(taskId string, output []byte, err error) *Result{
	r:=&Result{
		TaskID: taskId,
		Output: output,
		Success: err==nil,
		CreatedAt: time.Now().UTC(),
	}
	if err != nil{
		r.Error=err.Error()
	}

	return r
}

func (t *Task) CanRetry() bool{
	return t.RetryCount<t.MaxRetries
}

func (t *Task) IncrementRetry() {
	t.RetryCount++
}