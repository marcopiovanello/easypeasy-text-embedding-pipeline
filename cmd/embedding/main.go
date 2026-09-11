package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/marcopiovanello/easypeasyocr/internal/activities"
	"github.com/marcopiovanello/easypeasyocr/internal/domain"
	"github.com/marcopiovanello/easypeasyocr/pkg/utils"
	pgxvec "github.com/pgvector/pgvector-go/pgx"
	"go.temporal.io/sdk/worker"
)

var (
	postgresURL         = os.Getenv("POSTGRES_URL")
	embeddingServiceURL = os.Getenv("EMBEDDING_SERVICE_URL")
	temporalServerAddr  = os.Getenv("TEMPORAL_ADDR")
	temporalNamespace   = os.Getenv("TEMPORAL_NAMESPACE")
	taskQueueLen        = os.Getenv("TASK_QUEUE_LEN")
	llamaCppApiKey      = os.Getenv("LLAMA_CPP_API_KEY")
	embeddingModel      = os.Getenv("EMBEDDING_MODEL")
)

func main() {
	queueLen, err := strconv.Atoi(taskQueueLen)
	if err != nil {
		log.Println("defaulting to task queue lenght: 8")
	}
	if queueLen <= 0 {
		queueLen = 8
	}

	wEmbed, stop, err := utils.NewTemporalWorker(utils.TemporalWorkerOpts{
		Address:            temporalServerAddr,
		Namespace:          temporalNamespace,
		TaskQueue:          domain.TaskQueueEmbed,
		MaxConcurrentTasks: queueLen,
	})
	if err != nil {
		log.Fatalln(err)
	}
	defer stop()

	wDB, stop, err := utils.NewTemporalWorker(utils.TemporalWorkerOpts{
		Address:            temporalServerAddr,
		Namespace:          temporalNamespace,
		TaskQueue:          domain.TaskQueueDB,
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

	pgConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		return pgxvec.RegisterTypes(ctx, conn)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), pgConfig)
	if err != nil {
		log.Fatalln("failed creating pgpool", err.Error())
	}

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalln("failed connecting to postgresql", err.Error())
	}

	act := activities.NewEmbeddingActivity(pool, http.DefaultClient, &activities.EmbeddingServiceOpts{
		URL:    embeddingServiceURL,
		ApiKey: llamaCppApiKey,
		Model:  embeddingModel,
	})

	log.Println(llamaCppApiKey, embeddingServiceURL)

	wEmbed.RegisterActivity(act.EmbedText)
	wDB.RegisterActivity(act.PersistToPgvector)

	go wEmbed.Run(worker.InterruptCh())
	go wDB.Run(worker.InterruptCh())

	cmdCtx, _ := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	<-cmdCtx.Done()
}
