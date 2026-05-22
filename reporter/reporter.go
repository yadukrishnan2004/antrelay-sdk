package reporter

import "github.com/yadukrishnan2004/antrelay-sdk/task"

type Reporter interface {
	Report(result *task.Result) error
}
