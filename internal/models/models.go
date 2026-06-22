package models

import (
	"time"

	"gorm.io/gorm"
)

// Book represents an imported knowledge document
type Book struct {
	ID          string         `gorm:"primaryKey" json:"id"`
	Title       string         `json:"title"`
	FileName    string         `json:"file_name"`
	FileSize    int64          `json:"file_size"`
	FileType    string         `json:"file_type"` // pdf, txt, md
	ChunkCount  int            `json:"chunk_count"`
	Status      string         `json:"status"` // processing, ready, error
	ErrorMsg    string         `json:"error_msg,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// Chunk represents a text chunk with its embedding vector stored as JSON
type Chunk struct {
	ID        string         `gorm:"primaryKey" json:"id"`
	BookID    string         `gorm:"index" json:"book_id"`
	Content   string         `json:"content"`
	ChunkIdx  int            `json:"chunk_idx"`
	Embedding string         `gorm:"type:text" json:"-"` // JSON float64 array
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// ChatMessage represents a Q&A conversation message
type ChatMessage struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	Role      string    `json:"role"` // user, assistant
	Content   string    `json:"content"`
	Sources   string    `json:"sources,omitempty"` // JSON list of source book titles
	CreatedAt time.Time `json:"created_at"`
}

// Config stores LLM and embedding configuration
type Config struct {
	ID           uint    `gorm:"primaryKey"`
	Provider     string  `json:"provider"`     // openai, gemini, zhipu, volcengine, deepseek, qwen, moonshot, custom
	APIKey       string  `json:"api_key"`
	EmbedAPIKey  string  `json:"embed_api_key"` // optional: separate key for embedding (e.g. when using DeepSeek chat + OpenAI embed)
	BaseURL      string  `json:"base_url"`
	ChatModel    string  `json:"chat_model"`
	EmbedModel   string  `json:"embed_model"`
	TopK         int     `json:"top_k"`
	ScoreThresh  float64 `json:"score_thresh"`
}
