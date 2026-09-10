package workflow

import (
	"time"

	"github.com/marcopiovanello/easypeasyocr/internal/domain"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func ExtractionWorkflow(ctx workflow.Context, input domain.DocumentInput) error {
	convertCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		TaskQueue:           domain.TaskQueueConvert,
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval: 2 * time.Second,
			MaximumAttempts: 3,
		},
	})

	var convertRes domain.ConvertResult
	if err := workflow.ExecuteActivity(convertCtx, "ConvertAndUpload", input).Get(ctx, &convertRes); err != nil {
		return err
	}

	ocrCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		TaskQueue:           domain.TaskQueueOCR,
		StartToCloseTimeout: 15 * time.Minute,
		HeartbeatTimeout:    30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval: 5 * time.Second,
			MaximumAttempts: 3,
		},
	})

	var ocrRes domain.OCRResult
	if err := workflow.ExecuteActivity(ocrCtx, "RunOCR", convertRes).Get(ctx, &ocrRes); err != nil {
		return err
	}

	embedCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		TaskQueue:           domain.TaskQueueEmbed,
		StartToCloseTimeout: 3 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumAttempts:    5,
		},
	})

	var embedRes domain.EmbedResult
	if err := workflow.ExecuteActivity(embedCtx, "EmbedText", ocrRes).Get(ctx, &embedRes); err != nil {
		return err
	}

	dbCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		TaskQueue:           domain.TaskQueueDB,
		StartToCloseTimeout: 1 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval: time.Second,
			MaximumAttempts: 5,
		},
	})

	return workflow.ExecuteActivity(dbCtx, "PersistToPgvector", input.DocumentID, ocrRes.Text, embedRes.Vector).Get(ctx, nil)
}
