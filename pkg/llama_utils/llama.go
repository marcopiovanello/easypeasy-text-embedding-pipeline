package llamautils

type NativeEmbeddingRequest struct {
	Content string `json:"content"`
}

func NewEmbeddingRequest(content string) *NativeEmbeddingRequest {
	return &NativeEmbeddingRequest{
		Content: content,
	}
}

type OpenAIEmbed struct {
	Index     int         `json:"index"`
	Embedding [][]float32 `json:"embedding"`
}

type OpenAIEmbedVector = []OpenAIEmbed
