package main

import (
	"log"
	"os"

	"github.com/marcopiovanello/easypeasyocr/internal/domain"
	"github.com/marcopiovanello/easypeasyocr/internal/workflow"
	"github.com/marcopiovanello/easypeasyocr/pkg/utils"
	"go.temporal.io/sdk/worker"
)

var (
	temporalServerAddr = os.Getenv("TEMPORAL_ADDR")
	temporalNamespace  = os.Getenv("TEMPORAL_NAMESPACE")
)

func main() {
	w, stop, err := utils.NewTemporalWorker(utils.TemporalWorkerOpts{
		Address:   temporalServerAddr,
		Namespace: temporalNamespace,
		TaskQueue: domain.TaskQueueMain,
	})
	if err != nil {
		log.Fatalln(err)
	}
	defer stop()

	w.RegisterWorkflow(workflow.ExtractionWorkflow)

	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatalln("unable to start worker", err)
	}
}
