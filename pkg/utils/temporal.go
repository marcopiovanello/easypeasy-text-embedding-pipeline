package utils

import (
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

type TemporalWorkerOpts struct {
	Address            string
	Namespace          string
	TaskQueue          string
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

	w := worker.New(c, opts.TaskQueue, worker.Options{
		MaxConcurrentActivityExecutionSize: opts.MaxConcurrentTasks,
	})

	return w, c.Close, nil
}
