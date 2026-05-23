package worker

import (
	"context"
	"log"
	"time"

	"github.com/yadukrishnan2004/antrelay-sdk/executor"
	"github.com/yadukrishnan2004/antrelay-sdk/poller"
	"github.com/yadukrishnan2004/antrelay-sdk/reporter"
)

type Worker struct {
	poller poller.Poller
	executor *executor.Executor
	reporter reporter.Reporter
	queue string
	pollWait time.Duration
}

func New(
	p poller.Poller,
	e *executor.Executor,
	r reporter.Reporter,
	queue string,
)*Worker{
	return &Worker{
		poller: p,
		executor: e,
		reporter: r,
		queue: queue,
		pollWait: 2*time.Second,
	}
}

//starting the worker loop

func (w *Worker) Run(ctx context.Context){
	log.Printf("worker:starting on queue %q",w.queue)

	for{

		select{
		case <-ctx.Done():
			log.Printf("Worker shutting down")
			return
		default:
		}

		//asking the worker poll to next work

		t,err:=w.poller.Poll(w.queue)
		if err!=nil{
			log.Printf("worker: poll error:%v",err)
			w.Sleep(ctx)
			continue
		}

		if t==nil{
			w.Sleep(ctx)
			continue
		}


		log.Printf("worker: executing task %s (%s)", t.ID, t.FunctionName)

		result := w.executor.Executor(ctx, t)

		if !result.Success && t.CanRetry(){
			t.IncrementRetry()
			log.Printf("worker: retrying task %s (attempt %d/%d)",
				t.ID, t.RetryCount, t.MaxRetries)
			w.poller.Requeue(t)
			continue
		}

		if err := w.reporter.Report(result); err != nil {
			log.Printf("worker: failed to report result for task %s: %v", t.ID, err)
		}
	}
}

func (w *Worker) Sleep(ctx context.Context){
	select{
		case <-ctx.Done():
		case <-time.After(w.pollWait):
	}
}