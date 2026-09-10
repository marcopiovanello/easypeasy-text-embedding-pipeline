package activities

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path"
	"time"
	"uuid"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/h2non/bimg"
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

func (a *OCRExtractionActivity) ConvertAndUpload(ctx context.Context, doc domain.DocumentInput) (domain.ConvertResult, error) {
	obj, err := a.S3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(a.BucketName),
		Key:    aws.String(doc.S3Key),
	})
	if err != nil {
		return domain.ConvertResult{}, err
	}
	defer obj.Body.Close()

	imgBytes, err := io.ReadAll(obj.Body)
	if err != nil {
		return domain.ConvertResult{}, err
	}

	processedImageBytes, err := bimg.NewImage(imgBytes).Process(bimg.Options{
		Width:          2480,
		Interpolator:   bimg.Bicubic,
		Interpretation: bimg.InterpretationBW,
	})
	if err != nil {
		return domain.ConvertResult{}, err
	}

	var (
		filename  = fmt.Sprintf("%s-%s.%s", doc.DocumentID, uuid.NewV4().String(), "jpg")
		uploadKey = path.Join(path.Dir(doc.S3Key), filename)
	)

	uploadedObj, err := a.S3Client.PutObject(ctx, &s3.PutObjectInput{
		Body:   bytes.NewReader(processedImageBytes),
		Bucket: aws.String(a.BucketName),
		Key:    &uploadKey,
	})
	if err != nil {
		return domain.ConvertResult{}, err
	}

	return domain.ConvertResult{
		Checksum:    *uploadedObj.ChecksumCRC32,
		S3KeyBitmap: filename,
	}, nil
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

	return domain.OCRResult{Text: text}, nil
}
