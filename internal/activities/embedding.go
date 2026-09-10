package activities

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/bytedance/sonic"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/marcopiovanello/easypeasyocr/internal/domain"
	"github.com/pgvector/pgvector-go"
	"go.temporal.io/sdk/temporal"
)

type EmbeddingActiviy struct {
	db                  *pgxpool.Pool
	httplient           *http.Client
	embeddingServiceURL string
}

func NewEmbeddingActivity(db *pgxpool.Pool, httpClient *http.Client, embeddingServiceURL string) *EmbeddingActiviy {
	return &EmbeddingActiviy{
		db:                  db,
		httplient:           httpClient,
		embeddingServiceURL: embeddingServiceURL,
	}
}

func (a *EmbeddingActiviy) EmbedText(ctx context.Context, ocr domain.OCRResult) (domain.EmbedResult, error) {
	body, _ := sonic.Marshal(map[string]any{
		"inputs": ocr.Text,
	})

	req, err := http.NewRequestWithContext(ctx, "POST", a.embeddingServiceURL+"/embed", bytes.NewReader(body))
	if err != nil {
		return domain.EmbedResult{}, err
	}
	req.Header.Set("Content-Type", "application/json")

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
