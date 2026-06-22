package rag

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"

	"github.com/bugfan/empty015/internal/models"
	"github.com/bugfan/empty015/internal/store"
	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
)

const (
	chunkSize    = 800 // characters per chunk
	chunkOverlap = 100 // overlapping characters between chunks
)

// EmbedClient wraps the OpenAI-compatible embedding API
type EmbedClient struct {
	client    *openai.Client
	model     string
	isGemini  bool
	geminiKey string
}

// providerBaseURLs mirrors the same map as in chat package to avoid import cycle
var providerBaseURLs = map[string]string{
	"openai":     "https://api.openai.com/v1",
	"gemini":     "https://generativelanguage.googleapis.com/v1beta/openai/",
	"zhipu":      "https://open.bigmodel.cn/api/paas/v4/",
	"volcengine": "https://ark.cn-beijing.volces.com/api/v3/",
	"deepseek":   "https://api.deepseek.com",
	"qwen":       "https://dashscope.aliyuncs.com/compatible-mode/v1",
	"moonshot":   "https://api.moonshot.cn/v1",
}

var providerEmbedModels = map[string]string{
	"openai":     "text-embedding-3-small",
	"gemini":     "gemini-embedding-2",
	"zhipu":      "embedding-3",
	"volcengine": "doubao-embedding-large-text-240915",
	"qwen":       "text-embedding-v3",
}

// NewEmbedClient creates an embedding client based on configuration
func NewEmbedClient(cfg *models.Config) (*EmbedClient, error) {
	// Use embed-specific API key if provided, otherwise fall back to main key
	apiKey := cfg.EmbedAPIKey
	if apiKey == "" {
		apiKey = cfg.APIKey
	}
	if apiKey == "" {
		return nil, fmt.Errorf("API key is not configured")
	}

	// Resolve embedding model
	embedModel := cfg.EmbedModel
	if embedModel == "" {
		if m, ok := providerEmbedModels[cfg.Provider]; ok {
			embedModel = m
		}
	}
	if embedModel == "" {
		return nil, fmt.Errorf("embedding model is not configured for provider '%s'", cfg.Provider)
	}

	// Resolve base URL for embedding
	baseURL := cfg.BaseURL
	if baseURL == "" {
		// Try to infer the embed provider's URL from the model name (especially for hybrid setups)
		if strings.HasPrefix(embedModel, "text-embedding-3") || embedModel == "text-embedding-ada-002" {
			baseURL = "https://api.openai.com/v1"
		} else if strings.HasPrefix(embedModel, "embedding-") {
			baseURL = "https://open.bigmodel.cn/api/paas/v4/"
		} else if strings.HasPrefix(embedModel, "text-embedding-004") {
			baseURL = "https://generativelanguage.googleapis.com/v1beta/openai/"
		} else if strings.HasPrefix(embedModel, "doubao-embedding") {
			baseURL = "https://ark.cn-beijing.volces.com/api/v3/"
		}

		// Fallback to chat provider's base URL if we couldn't infer from the embedding model
		if baseURL == "" {
			if u, ok := providerBaseURLs[cfg.Provider]; ok {
				baseURL = u
			}
		}
	}

	config := openai.DefaultConfig(apiKey)
	if baseURL != "" {
		config.BaseURL = baseURL
	}

	if embedModel == "" {
		return nil, fmt.Errorf("embedding model is not configured for provider '%s'. Please set it in Settings", cfg.Provider)
	}

	return &EmbedClient{
		client:    openai.NewClientWithConfig(config),
		model:     embedModel,
		isGemini:  cfg.Provider == "gemini",
		geminiKey: apiKey,
	}, nil
}

func (c *EmbedClient) GetEmbedding(ctx context.Context, text string) ([]float32, error) {
	if c.isGemini {
		return c.getGeminiEmbedding(ctx, text)
	}

	req := openai.EmbeddingRequest{
		Input: []string{text},
		Model: openai.EmbeddingModel(c.model),
	}
	resp, err := c.client.CreateEmbeddings(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("embedding error: %v", err)
	}
	return resp.Data[0].Embedding, nil
}

func (c *EmbedClient) getGeminiEmbedding(ctx context.Context, text string) ([]float32, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:embedContent?key=%s", c.model, c.geminiKey)

	payloadObj := map[string]interface{}{
		"model": "models/" + c.model,
		"content": map[string]interface{}{
			"parts": []map[string]interface{}{
				{"text": text},
			},
		},
	}
	payloadBytes, _ := json.Marshal(payloadObj)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gemini API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Embedding struct {
			Values []float32 `json:"values"`
		} `json:"embedding"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse gemini response: %v", err)
	}

	if len(result.Embedding.Values) == 0 {
		return nil, fmt.Errorf("no embedding returned from gemini")
	}
	return result.Embedding.Values, nil
}

// Embed generates an embedding vector for the given text
func (e *EmbedClient) Embed(ctx context.Context, text string) ([]float64, error) {
	embedding, err := e.GetEmbedding(ctx, text)
	if err != nil {
		return nil, err
	}

	vec := make([]float64, len(embedding))
	for i, v := range embedding {
		vec[i] = float64(v)
	}
	return vec, nil
}

// ChunkText splits text into overlapping chunks
func ChunkText(text string) []string {
	text = strings.TrimSpace(text)
	if len(text) == 0 {
		return nil
	}

	var chunks []string
	runes := []rune(text)
	total := len(runes)

	for start := 0; start < total; {
		end := start + chunkSize
		if end > total {
			end = total
		}
		chunk := strings.TrimSpace(string(runes[start:end]))
		if len(chunk) > 20 {
			chunks = append(chunks, chunk)
		}
		if end >= total {
			break
		}
		start = end - chunkOverlap
		if start < 0 {
			start = 0
		}
	}
	return chunks
}

// IndexBook chunks, embeds and stores all chunks for a book
func IndexBook(ctx context.Context, bookID string, text string, embedClient *EmbedClient) (int, error) {
	chunks := ChunkText(text)
	if len(chunks) == 0 {
		return 0, fmt.Errorf("no text content found")
	}

	for i, content := range chunks {
		vec, err := embedClient.Embed(ctx, content)
		if err != nil {
			return i, fmt.Errorf("failed to embed chunk %d: %w", i, err)
		}

		vecJSON, err := json.Marshal(vec)
		if err != nil {
			return i, err
		}

		chunk := &models.Chunk{
			ID:        uuid.New().String(),
			BookID:    bookID,
			Content:   content,
			ChunkIdx:  i,
			Embedding: string(vecJSON),
		}
		if err := store.DB.Create(chunk).Error; err != nil {
			return i, fmt.Errorf("failed to store chunk: %w", err)
		}
	}

	return len(chunks), nil
}

// DeleteBookChunks removes all chunks for a given book
func DeleteBookChunks(bookID string) error {
	return store.DB.Where("book_id = ?", bookID).Delete(&models.Chunk{}).Error
}

// SearchResult is a retrieved chunk with similarity score
type SearchResult struct {
	Chunk     *models.Chunk
	Score     float64
	BookTitle string
}

// Search retrieves the top-K most similar chunks for a query
func Search(ctx context.Context, query string, topK int, scoreThresh float64, embedClient *EmbedClient) ([]SearchResult, error) {
	queryVec, err := embedClient.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}

	var chunks []models.Chunk
	if err := store.DB.Find(&chunks).Error; err != nil {
		return nil, err
	}

	type scoredChunk struct {
		chunk models.Chunk
		score float64
	}

	var scored []scoredChunk
	for _, c := range chunks {
		var vec []float64
		if err := json.Unmarshal([]byte(c.Embedding), &vec); err != nil {
			continue
		}
		score := cosineSimilarity(queryVec, vec)
		if score >= scoreThresh {
			scored = append(scored, scoredChunk{chunk: c, score: score})
		}
	}

	// Sort by score descending
	for i := 0; i < len(scored)-1; i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[j].score > scored[i].score {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}

	if topK > len(scored) {
		topK = len(scored)
	}
	scored = scored[:topK]

	// Load book titles
	bookTitles := map[string]string{}
	var books []models.Book
	store.DB.Find(&books)
	for _, b := range books {
		bookTitles[b.ID] = b.Title
	}

	results := make([]SearchResult, 0, len(scored))
	for _, s := range scored {
		c := s.chunk
		results = append(results, SearchResult{
			Chunk:     &c,
			Score:     s.score,
			BookTitle: bookTitles[c.BookID],
		})
	}
	return results, nil
}

func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}
