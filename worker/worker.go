package worker

import (
	"context"
	"log/slog"
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
	backoff *backoff
}

func New(
	p poller.Poller,
	e *executor.Executor,
	r reporter.Reporter,
	queue string,
	pollInterval time.Duration,
)*Worker{
	return &Worker{
		poller: p,
		executor: e,
		reporter: r,
		queue: queue,
		pollWait: pollInterval,
		backoff: NewBackoff(pollInterval,30 * time.Second),
	}
}

//starting the worker loop

func (w *Worker) Run(ctx context.Context) {
    slog.Info("worker starting", "queue", w.queue)

    for {
        select {
        case <-ctx.Done():
            slog.Info("worker shutting down", "queue", w.queue)
            return
        default:
        }

        t, err := w.poller.Poll(w.queue)
        if err != nil {
            slog.Error("poll failed",
                "queue", w.queue,
                "error", err,
                "retry_in", w.backoff.current,
            )
            w.Sleep(ctx, w.backoff.Next())
            continue
        }

        if t == nil {
            w.Sleep(ctx, w.backoff.Next())
            continue
        }

        w.backoff.Reset()

        slog.Info("executing task",
            "task_id", t.ID,
            "function", t.FunctionName,
            "retry_count", t.RetryCount,
        )

        result := w.executor.Executor(ctx, t)

        if !result.Success && t.CanRetry() {
            t.IncrementRetry()
            slog.Warn("task failed, requeueing",
                "task_id", t.ID,
                "attempt", t.RetryCount,
                "max_retries", t.MaxRetries,
                "error", result.Error,
            )
            w.poller.Requeue(t)
            continue
        }

        if err := w.reporter.Report(result); err != nil {
            slog.Error("failed to report result",
                "task_id", t.ID,
                "error", err,
            )
        }
    }
}

func (w *Worker) Sleep(ctx context.Context,d time.Duration){
	select{
		case <-ctx.Done():
		case <-time.After(d):
	}
}