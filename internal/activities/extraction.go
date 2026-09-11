package activities

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"io"
	"path"
	"time"
	"uuid"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/marcopiovanello/easypeasyocr/internal/domain"
	"github.com/marcopiovanello/easypeasyocr/pkg/utils"
	"go.temporal.io/sdk/activity"
	"gocv.io/x/gocv"
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

	imgMat := gocv.NewMat()
	defer imgMat.Close()

	if err := gocv.IMDecodeIntoMat(imgBytes, gocv.IMReadGrayScale, &imgMat); err != nil {
		return domain.ConvertResult{}, err
	}

	// da capire se serve perché è utile con doc a bassa risoluzione ma fa casino con quelli scansionati ad alta risoluzione
	// imgResizedMat := gocv.NewMat()
	// defer imgResizedMat.Close()
	// if err := gocv.Resize(imgMat, &imgResizedMat, image.Point{X: 0, Y: 0}, 2.0, 2.0, gocv.InterpolationCubic); err != nil {
	// 	return domain.ConvertResult{}, err
	// }

	blurred := gocv.NewMat()
	defer blurred.Close()

	gocv.GaussianBlur(imgMat, &blurred, image.Point{X: 0, Y: 0}, 3, 3, gocv.BorderDefault)

	sharpened := gocv.NewMat()
	defer sharpened.Close()

	gocv.AddWeighted(imgMat, 1.5, blurred, -0.5, 0, &sharpened)

	imgBin := gocv.NewMat()
	defer imgBin.Close()

	gocv.Threshold(sharpened, &imgBin, 0, 255, gocv.ThresholdBinary|gocv.ThresholdOtsu)

	/*
		// In realtà non servirebbe nemmeno perché tesseract (LSTM) o un vision model sono sufficientemente intelligenti
		// per distinguere i caratteri delle tabelle.

		// Inverto i bit dell'immagine (testo bianco su sfondo nero) perché le trasformazioni morfologiche vanno solo
		// sui pixel bianchi
			imgInv := gocv.NewMat()
			defer imgInv.Close()
			gocv.BitwiseNot(imgBin, &imgInv)

			horizontalStructure := gocv.GetStructuringElement(gocv.MorphRect, image.Point{X: 40, Y: 1})
			horizontalLines := gocv.NewMat()
			defer horizontalLines.Close()

			gocv.MorphologyEx(imgInv, &horizontalLines, gocv.MorphOpen, horizontalStructure)

			verticalStructure := gocv.GetStructuringElement(gocv.MorphRect, image.Point{X: 1, Y: 40})
			verticalLines := gocv.NewMat()
			defer verticalLines.Close()
			gocv.MorphologyEx(imgInv, &verticalLines, gocv.MorphOpen, verticalStructure)

			tableMask := gocv.NewMat()
			defer tableMask.Close()

			// sbianco
			gocv.BitwiseOr(horizontalLines, verticalLines, &tableMask)

			tableMaskInv := gocv.NewMat()
			defer tableMaskInv.Close()
			gocv.BitwiseNot(tableMask, &tableMaskInv)

			imgClean := gocv.NewMat()
			defer imgClean.Close()
			gocv.BitwiseOr(imgBin, tableMask, &imgClean)
	*/

	buf, err := gocv.IMEncode(".png", imgBin)
	if err != nil {
		return domain.ConvertResult{}, err
	}
	defer buf.Close()

	//gc
	optimizedBytes := make([]byte, len(buf.GetBytes()))
	copy(optimizedBytes, buf.GetBytes())

	var (
		filename  = fmt.Sprintf("%s-%s.%s", doc.DocumentID, uuid.NewV4().String(), "png")
		uploadKey = path.Join(path.Dir(doc.S3Key), filename)
	)

	uploadedObj, err := a.S3Client.PutObject(ctx, &s3.PutObjectInput{
		ContentType: aws.String("image/png"),
		Body:        bytes.NewReader(optimizedBytes),
		Bucket:      aws.String(a.BucketName),
		Key:         &uploadKey,
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

	return domain.OCRResult{Text: utils.CleanTextForEmbedding(text)}, nil
}
