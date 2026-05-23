package poller

import "github.com/yadukrishnan2004/antrelay-sdk/task"


type Poller interface {
	Poll(queue string) (*task.Task, error)
	Requeue(t *task.Task)
}