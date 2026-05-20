package poller

import "github.com/yadukrishnan2004/antrelay-sdk/task"


type Poller interface {
	// Poll asks the task source for the next available task
	// on the given queue.
	Poll(queue string) (*task.Task, error)
}