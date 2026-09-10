package llamautils

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type EmbeddingRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature int       `json:"temperature"`
}

func NewEmbeddingRequest(model, systemMessage, userMessage string) *EmbeddingRequest {
	return &EmbeddingRequest{
		Model: model,
		Messages: []Message{
			{
				Role:    "system",
				Content: systemMessage,
			},
			{
				Role:    "user",
				Content: userMessage,
			},
		},
		Temperature: 0,
	}
}

func NewEmbeddingRequestPreInstructed(model, userMessage string) *EmbeddingRequest {
	return NewEmbeddingRequest(
		model,
		"Estrai i dati dalla fattura e rispondi SOLO in JSON valido con questi campi: numero_fattura, data, fornitore, partita_iva, imponibile, iva, totale.",
		userMessage,
	)
}
