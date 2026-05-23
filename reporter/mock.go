package reporter

import (
	"sync"

	"github.com/yadukrishnan2004/antrelay-sdk/task"
)

type MockReporter struct {
	mu      sync.Mutex
	results []*task.Result
	err     error
}

func NewMock() *MockReporter {
	return &MockReporter{
		results: make([]*task.Result, 0),
	}
}

func (m *MockReporter) Report(result *task.Result) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.err != nil {
		return m.err
	}

	m.results = append(m.results, result)
	return nil
}

func (m *MockReporter) SetError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.err = err
}

func (m *MockReporter) Results() []*task.Result {
	m.mu.Lock()
	defer m.mu.Unlock()

	out := make([]*task.Result, len(m.results))
	copy(out, m.results)
	return out
}

func (m *MockReporter) Len() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	return len(m.results)
}



var _ Reporter = (*MockReporter)(nil)
