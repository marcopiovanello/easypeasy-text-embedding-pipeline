package activities

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

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
	body, err := sonic.Marshal(llamautils.NewEmbeddingRequestPreInstructed(
		a.embeddingModel,
		ocr.Text,
	))
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

	var vectors [][]float32
	if err := json.NewDecoder(resp.Body).Decode(&vectors); err != nil {
		return domain.EmbedResult{}, temporal.NewNonRetryableApplicationError(
			"invalid response from embedding service", "DecodeError", err)
	}

	return domain.EmbedResult{Vector: vectors[0]}, nil
}

func (a *EmbeddingActiviy) PersistToPgvector(ctx context.Context, documentID string, text string, vector []float32) error {
	embedding := pgvector.NewVector(vector)

	_, err := a.db.Exec(ctx, `
		INSERT INTO document_extraction_embeddings (document_id, text, embedding)
		VALUES ($1, $2, $3)
		ON CONFLICT (document_id) DO UPDATE
		SET text = EXCLUDED.text, embedding = EXCLUDED.embedding
	`, documentID, text, embedding)

	return err
}
