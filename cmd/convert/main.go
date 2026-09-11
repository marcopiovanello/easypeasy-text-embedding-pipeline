package main

import (
	"log"
	"os"
	"strconv"

	"github.com/marcopiovanello/easypeasyocr/internal/activities"
	"github.com/marcopiovanello/easypeasyocr/internal/domain"
	"github.com/marcopiovanello/easypeasyocr/pkg/utils"
	"go.temporal.io/sdk/worker"
)

var (
	s3Region           = os.Getenv("S3_REGION")
	s3Bucket           = os.Getenv("S3_BUCKET")
	s3Endpoint         = os.Getenv("S3_HOST")
	s3AccessKeyId      = os.Getenv("S3_AKID")
	s3SecretKeyId      = os.Getenv("S3_SKID")
	temporalServerAddr = os.Getenv("TEMPORAL_ADDR")
	temporalNamespace  = os.Getenv("TEMPORAL_NAMESPACE")
	taskQueueLen       = os.Getenv("TASK_QUEUE_LEN")
)

func main() {
	queueLen, err := strconv.Atoi(taskQueueLen)
	if err != nil {
		log.Println("defaulting to task queue lenght: 4")
		queueLen = 0
	}
	if queueLen <= 0 {
		queueLen = 4
	}

	w, stop, err := utils.NewTemporalWorker(utils.TemporalWorkerOpts{
		Address:            temporalServerAddr,
		Namespace:          temporalNamespace,
		TaskQueue:          domain.TaskQueueConvert,
		MaxConcurrentTasks: queueLen,
	})
	defer stop()

	s3Client, err := utils.NewS3Client(utils.S3ClientOpts{
		Region:       s3Region,
		Endpoint:     s3Endpoint,
		AccessKeyId:  s3AccessKeyId,
		SecretKeyId:  s3SecretKeyId,
		UsePathStyle: true,
	})
	if err != nil {
		log.Fatalln(err)
	}

	act := activities.NewOCRExtractionActivity(s3Client, s3Bucket)

	w.RegisterActivity(act.ConvertAndUpload)
	w.RegisterActivity(act.RunOCR)

	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatalln("unable to start worker", err)
	}
}
