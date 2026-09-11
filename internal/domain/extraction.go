package domain

type DocumentInput struct {
	DocumentID string
	S3Key      string
}

type UploadResult struct {
	S3Key string
}

type ConvertResult struct {
	Checksum    string
	S3KeyBitmap string
}

type OCRResult struct {
	Text string
}

type EmbedResult struct {
	Vector []float32
}

const (
	TaskQueueDB      = "db"
	TaskQueueConvert = "cpu-convert"
	TaskQueueOCR     = "cpu-ocr"
	TaskQueueEmbed   = "gpu-embed"
	TaskQueueMain    = "extraction"
)
