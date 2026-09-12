package activities

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/marcopiovanello/easypeasyocr/internal/domain"
	"github.com/marcopiovanello/easypeasyocr/pkg/utils"
	"go.temporal.io/sdk/activity"
)

type OCRExtractionActivity struct {
	S3Client   *s3.Client
	BucketName string
}

func NewOCRExtractionActivity(S3Client *s3.Client, BucketName string) *OCRExtractionActivity {
	return &OCRExtractionActivity{
		S3Client:   S3Client,
		BucketName: BucketName,
	}
}

func (a *OCRExtractionActivity) RunOCR(ctx context.Context, upload domain.ConvertResult) (domain.OCRResult, error) {
	obj, err := a.S3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &a.BucketName,
		Key:    &upload.S3KeyBitmap,
	})
	if err != nil {
		return domain.OCRResult{}, err
	}
	defer obj.Body.Close()

	// OCR Heartbeat for as it's a VLRT
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				activity.RecordHeartbeat(ctx, "still processing")
			}
		}
	}()

	text, err := utils.RunTesseract(obj.Body)
	if err != nil {
		return domain.OCRResult{}, err
	}

	return domain.OCRResult{Text: utils.CleanTextForEmbedding(text)}, nil
}
