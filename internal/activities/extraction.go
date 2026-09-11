package activities

import (
	"bytes"
	"context"
	"fmt"
	"image/color"
	"path"
	"time"
	"uuid"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/disintegration/gift"
	"github.com/disintegration/imaging"
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

	src, err := imaging.Decode(obj.Body)
	if err != nil {
		return domain.ConvertResult{}, err
	}

	g := gift.New(
		gift.Resize(src.Bounds().Dx()*2, src.Bounds().Dy()*2, gift.CubicResampling),
		gift.Grayscale(),
		gift.Threshold(60.0),
	)

	dst := imaging.New(g.Bounds(src.Bounds()).Dx(), g.Bounds(src.Bounds()).Dy(), color.White)
	g.Draw(dst, src)

	var imgbuf bytes.Buffer

	err = imaging.Encode(&imgbuf, dst, imaging.PNG)
	if err != nil {
		return domain.ConvertResult{}, err
	}

	var (
		filename  = fmt.Sprintf("%s-%s.%s", doc.DocumentID, uuid.NewV4().String(), "png")
		uploadKey = path.Join(path.Dir(doc.S3Key), filename)
	)

	uploadedObj, err := a.S3Client.PutObject(ctx, &s3.PutObjectInput{
		Body:   &imgbuf,
		Bucket: aws.String(a.BucketName),
		Key:    &uploadKey,
	})
	if err != nil {
		return domain.ConvertResult{}, err
	}

	return domain.ConvertResult{
		Checksum:    *uploadedObj.ChecksumCRC32,
		S3KeyBitmap: uploadKey,
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
