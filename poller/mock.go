package poller

import (
	"sync"
	"github.com/yadukrishnan2004/antrelay-sdk/task"
)


type MockPoller struct {
	mu    sync.Mutex
	tasks []*task.Task
	err   error
}

func NewMock() *MockPoller {
	return &MockPoller{
		tasks: make([]*task.Task, 0),
	}
}

func (m *MockPoller) QueueTask(t *task.Task) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.tasks = append(m.tasks, t)
}

func (m *MockPoller) SetError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.err = err
}

func (m *MockPoller) Poll(queue string) (*task.Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.err != nil {
		return nil, m.err
	}

	if len(m.tasks) == 0 {
		return nil, nil
	}

	// take the first task from the front of the slice
	// this is FIFO 
	next := m.tasks[0]
	m.tasks = m.tasks[1:]

	return next, nil
}

func (m *MockPoller) Len() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	return len(m.tasks)
}


func (m *MockPoller) Requeue(t *task.Task) {
    m.mu.Lock()
    defer m.mu.Unlock()

    // put it at the front so it gets picked up next
    m.tasks = append([]*task.Task{t}, m.tasks...)
}


var _ Poller = (*MockPoller)(nil)
