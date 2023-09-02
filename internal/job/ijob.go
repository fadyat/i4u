package job

import (
	"context"
)

type Job interface {
	Run(context.Context)
}

// Producer is core interface for jobs workflow.
// It's an entry point.
type Producer interface {

	// Produce starts jobs workflow. By any stage of the workflow
	// can occur an error, so it returns a channel with errors.
	Produce(context.Context) <-chan error
}
