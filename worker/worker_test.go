package worker_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yadukrishnan2004/antrelay-sdk/executor"
	"github.com/yadukrishnan2004/antrelay-sdk/poller"
	"github.com/yadukrishnan2004/antrelay-sdk/registry"
	"github.com/yadukrishnan2004/antrelay-sdk/reporter"
	"github.com/yadukrishnan2004/antrelay-sdk/task"
	"github.com/yadukrishnan2004/antrelay-sdk/worker"
)

func setup(handlers map[string]task.HandlerFunc)(*worker.Worker,*poller.MockPoller,*reporter.MockReporter){
	r:=registry.New()
	for name,h:=range handlers{
		r.Register(name,h)
	}

	p := poller.NewMock()
	rep := reporter.NewMock()
	ex := executor.New(r)
	w := worker.New(p, ex, rep, "test-queue")

	return w, p, rep
}

func TestWorker(t *testing.T) {
	t.Run("executes and reports a task successfully", func(t *testing.T) {
		// arrange
		w, p, rep := setup(map[string]task.HandlerFunc{
			"processOrder": func(_ context.Context, input []byte) ([]byte, error) {
				return []byte(`{"status":"ok"}`), nil
			},
		})

		p.QueueTask(task.New("processOrder", []byte(`{}`), "test-queue", 3))

		ctx, cancel := context.WithCancel(context.Background())

				go func() {
			for rep.Len() == 0 {
			}
			cancel()
		}()

		w.Run(ctx)

		assert.Equal(t, 1, rep.Len())
		assert.True(t, rep.Results()[0].Success)
	})
}