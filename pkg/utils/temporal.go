package utils

import (
	"github.com/marcopiovanello/easypeasyocr/internal/domain"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

type TemporalWorkerOpts struct {
	Address            string
	Namespace          string
	MaxConcurrentTasks int
}

func NewTemporalWorker(opts TemporalWorkerOpts) (worker.Worker, func(), error) {
	c, err := client.Dial(client.Options{
		HostPort:  opts.Address,
		Namespace: opts.Namespace,
	})
	if err != nil {
		return nil, nil, err
	}

	w := worker.New(c, domain.TaskQueueConvert, worker.Options{
		MaxConcurrentActivityExecutionSize: opts.MaxConcurrentTasks,
	})

	return w, c.Close, nil
}
