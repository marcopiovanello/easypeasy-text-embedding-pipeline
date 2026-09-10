package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mime"
	"net/http"
	"os"
	"path"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/marcopiovanello/easypeasyocr/internal/domain"
	"github.com/marcopiovanello/easypeasyocr/internal/workflow"
	"github.com/marcopiovanello/easypeasyocr/pkg/utils"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
)

var (
	s3Region           = os.Getenv("S3_REGION")
	s3Bucket           = os.Getenv("S3_BUCKET")
	s3Endpoint         = os.Getenv("S3_HOST")
	s3AccessKeyId      = os.Getenv("S3_AKID")
	s3SecretKeyId      = os.Getenv("S3_SKID")
	temporalServerAddr = os.Getenv("TEMPORAL_ADDR")
	temporalNamespace  = os.Getenv("TEMPORAL_NAMESPACE")
)

func main() {
	temporalClient, err := client.Dial(client.Options{
		HostPort:  temporalServerAddr,
		Namespace: temporalNamespace,
	})
	if err != nil {
		log.Fatalln(err)
	}
	defer temporalClient.Close()

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

	mux := http.NewServeMux()

	mux.HandleFunc("PUT /upload", steamFileAndRunWorkflow(temporalClient, s3Client))

	http.ListenAndServe(":8080", mux)
}

type workflowRunResponse struct {
	WorkflowId    string `json:"workflow_id"`
	WorkflowRunId string `json:"workflow_run_id"`
}

func steamFileAndRunWorkflow(temporalClient client.Client, s3Client *s3.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		defer r.Body.Close()

		if err := r.ParseMultipartForm(2 << 20); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()

		contentType := header.Header.Get("Content-Type")
		if contentType == "" {
			contentType = mime.TypeByExtension(header.Filename)
		}

		var (
			tenantId   = r.FormValue("tenant_id")
			documentId = r.FormValue("document_id")
		)

		key := path.Join("extractions", tenantId, documentId)

		_, err = s3Client.PutObject(r.Context(), &s3.PutObjectInput{
			ContentType: aws.String(contentType),
			Bucket:      aws.String(s3Bucket),
			Key:         aws.String(key),
			Body:        file,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		workflowId := fmt.Sprintf("extraction-%s-%s", documentId, tenantId)

		wr, err := temporalClient.ExecuteWorkflow(
			context.Background(),
			client.StartWorkflowOptions{
				ID:                    workflowId,
				WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
			},
			workflow.ExtractionWorkflow,
			domain.DocumentInput{
				DocumentID: documentId,
				S3Key:      key,
			},
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(workflowRunResponse{
			WorkflowId:    wr.GetID(),
			WorkflowRunId: wr.GetRunID(),
		})
	}
}
