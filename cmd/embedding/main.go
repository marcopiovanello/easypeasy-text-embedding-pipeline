package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/marcopiovanello/easypeasyocr/internal/activities"
	"github.com/marcopiovanello/easypeasyocr/pkg/utils"
	"go.temporal.io/sdk/worker"
)

var (
	postgresURL         = os.Getenv("POSTGRES_URL")
	embeddingServiceURL = os.Getenv("EMBEDDING_SERVICE_URL")
	temporalServerAddr  = os.Getenv("TEMPORAL_ADDR")
	temporalNamespace   = os.Getenv("TEMPORAL_NAMESPACE")
	taskQueueLen        = os.Getenv("TASK_QUEUE_LEN")
)

func main() {
	queueLen, err := strconv.Atoi(taskQueueLen)
	if err != nil {
		log.Fatalln(err)
	}
	if queueLen <= 0 {
		queueLen = 8
	}

	w, stop, err := utils.NewTemporalWorker(utils.TemporalWorkerOpts{
		Address:            temporalServerAddr,
		Namespace:          temporalNamespace,
		MaxConcurrentTasks: queueLen,
	})
	if err != nil {
		log.Fatalln(err)
	}
	defer stop()

	pgConfig, err := pgxpool.ParseConfig(postgresURL)
	if err != nil {
		log.Fatalln("failed parsing database datasource", postgresURL, err.Error())
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), pgConfig)
	if err != nil {
		log.Fatalln("failed creating pgpool", err.Error())
	}

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalln("failed connecting to postgresql", err.Error())
	}

	act := activities.NewEmbeddingActivity(pool, http.DefaultClient, embeddingServiceURL)

	w.RegisterActivity(act.EmbedText)
	w.RegisterActivity(act.PersistToPgvector)

	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatalln("unable to start worker", err)
	}
}
