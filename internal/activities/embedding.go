package activities

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/bytedance/sonic"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/marcopiovanello/easypeasyocr/internal/domain"
	llamautils "github.com/marcopiovanello/easypeasyocr/pkg/llama_utils"
	"github.com/pgvector/pgvector-go"
	"go.temporal.io/sdk/temporal"
)

type EmbeddingServiceOpts struct {
	URL    string
	ApiKey string
	Model  string
}

type EmbeddingActiviy struct {
	db                     *pgxpool.Pool
	httplient              *http.Client
	embeddingModel         string
	embeddingServiceURL    string
	embeddingServiceApiKey string
}

func NewEmbeddingActivity(
	db *pgxpool.Pool,
	httpClient *http.Client,
	embeddingOpts *EmbeddingServiceOpts,
) *EmbeddingActiviy {
	return &EmbeddingActiviy{
		db:                     db,
		httplient:              httpClient,
		embeddingModel:         embeddingOpts.Model,
		embeddingServiceURL:    embeddingOpts.URL,
		embeddingServiceApiKey: embeddingOpts.ApiKey,
	}
}

func (a *EmbeddingActiviy) EmbedText(ctx context.Context, ocr domain.OCRResult) (domain.EmbedResult, error) {
	body, err := sonic.Marshal(llamautils.NewEmbeddingRequest(ocr.Text))
	if err != nil {
		return domain.EmbedResult{}, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", a.embeddingServiceURL, bytes.NewReader(body))
	if err != nil {
		return domain.EmbedResult{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", a.embeddingServiceApiKey))

	resp, err := a.httplient.Do(req)
	if err != nil {
		return domain.EmbedResult{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return domain.EmbedResult{}, errors.New("embedding service overloaded")
	}

	var results llamautils.OpenAIEmbedVector

	if err := sonic.ConfigStd.NewDecoder(resp.Body).Decode(&results); err != nil {
		return domain.EmbedResult{}, temporal.NewNonRetryableApplicationError(
			"invalid response from embedding service",
			"DecodeError",
			err,
		)
	}

	return domain.EmbedResult{Vector: results[0].Embedding[0]}, nil
}

func (a *EmbeddingActiviy) PersistToPgvector(ctx context.Context, documentId string, text string, vector []float32) error {
	conn, err := a.db.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	queryCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	embedding := pgvector.NewVector(vector)

	_, err = conn.Exec(queryCtx, `
		INSERT INTO documents (document_id, text, embedding)
		VALUES ($1, $2, $3)
		ON CONFLICT (document_id) DO UPDATE
		SET text = EXCLUDED.text, embedding = EXCLUDED.embedding
	`, documentId, text, embedding)

	return err
}
