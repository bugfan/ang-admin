package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/bugfan/empty015/internal/chat"
	"github.com/bugfan/empty015/internal/models"
	"github.com/bugfan/empty015/internal/rag"
	"github.com/bugfan/empty015/internal/store"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler holds API dependencies
type Handler struct{}

// New creates a new API handler
func New() *Handler {
	return &Handler{}
}

// ---- Book Management ----

// ListBooks returns all indexed books
func (h *Handler) ListBooks(c *gin.Context) {
	var books []models.Book
	store.DB.Order("created_at desc").Find(&books)
	c.JSON(http.StatusOK, gin.H{"books": books})
}

// UploadBook handles file upload and RAG indexing
func (h *Handler) UploadBook(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}

	title := c.PostForm("title")
	if title == "" {
		title = header.Filename
	}

	fileType := rag.DetectFileType(header.Filename, nil)

	book := &models.Book{
		ID:       uuid.New().String(),
		Title:    title,
		FileName: header.Filename,
		FileSize: header.Size,
		FileType: fileType,
		Status:   "processing",
	}

	if err := store.DB.Create(book).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create book record"})
		return
	}

	// setError is a helper to record the error message and set status=error
	setError := func(bookID, msg string) {
		fmt.Printf("[KnowledgeAI] indexing error for book %s: %s\n", bookID, msg)
		store.DB.Model(&models.Book{}).Where("id = ?", bookID).Updates(map[string]interface{}{
			"status":    "error",
			"error_msg": msg,
		})
	}

	// Index in background
	go func(bookID string, fileData []byte, ft string) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		cfg, err := store.GetConfig()
		if err != nil {
			setError(bookID, fmt.Sprintf("failed to load config: %v", err))
			return
		}
		if cfg.APIKey == "" {
			setError(bookID, "API Key 未配置，请先在「配置中心」填写您的 API Key")
			return
		}

		embedClient, err := rag.NewEmbedClient(cfg)
		if err != nil {
			setError(bookID, fmt.Sprintf("Embedding 客户端初始化失败: %v", err))
			return
		}

		text, err := rag.ExtractText(header.Filename, fileData, ft)
		if err != nil {
			setError(bookID, fmt.Sprintf("文本提取失败: %v", err))
			return
		}
		if len([]rune(text)) < 50 {
			setError(bookID, "文档内容太少或为纯图片PDF（无法提取文字），请检查文档内容")
			return
		}

		count, err := rag.IndexBook(ctx, bookID, text, embedClient)
		if err != nil {
			setError(bookID, fmt.Sprintf("向量索引失败（已处理 %d 块）: %v", count, err))
			return
		}

		store.DB.Model(&models.Book{}).Where("id = ?", bookID).Updates(map[string]interface{}{
			"status":      "ready",
			"chunk_count": count,
			"error_msg":   "",
		})
		fmt.Printf("[KnowledgeAI] book %s indexed: %d chunks\n", bookID, count)
	}(book.ID, data, fileType)

	c.JSON(http.StatusOK, gin.H{"book": book, "message": "Book is being indexed in the background"})
}

// DeleteBook removes a book and all its RAG chunks
func (h *Handler) DeleteBook(c *gin.Context) {
	id := c.Param("id")

	var book models.Book
	if err := store.DB.First(&book, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
		return
	}

	// Delete all associated chunks
	if err := rag.DeleteBookChunks(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete book chunks"})
		return
	}

	// Delete the book record
	if err := store.DB.Delete(&book).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete book"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Book deleted successfully"})
}

// GetBookStatus returns the processing status of a book
func (h *Handler) GetBookStatus(c *gin.Context) {
	id := c.Param("id")
	var book models.Book
	if err := store.DB.First(&book, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
		return
	}
	c.JSON(http.StatusOK, book)
}

// ---- Chat ----

// ChatRequest is the chat API request body
type ChatRequest struct {
	Question string `json:"question" binding:"required"`
}

// Chat handles a Q&A question using RAG
func (h *Handler) Chat(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Question is required"})
		return
	}

	cfg, err := store.GetConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config"})
		return
	}

	if cfg.APIKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "LLM API key is not configured. Please go to Settings."})
		return
	}

	// Check if any books are indexed
	var bookCount int64
	store.DB.Model(&models.Book{}).Where("status = ?", "ready").Count(&bookCount)
	if bookCount == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No knowledge base documents available. Please upload and index some books first."})
		return
	}

	ctx := c.Request.Context()

	topK := cfg.TopK
	if topK <= 0 {
		topK = 5
	}
	scoreThresh := cfg.ScoreThresh
	if scoreThresh <= 0 {
		scoreThresh = 0.3
	}

	// RAG retrieval
	embedClient, err := rag.NewEmbedClient(cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Embedding client error: %v", err)})
		return
	}

	results, err := rag.Search(ctx, req.Question, topK, scoreThresh, embedClient)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Search error: %v", err)})
		return
	}

	// Generate answer
	llmClient, err := chat.NewLLMClient(cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("LLM client error: %v", err)})
		return
	}

	answer, sources, err := llmClient.Answer(ctx, req.Question, results)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("LLM error: %v", err)})
		return
	}

	// Save messages
	userMsg := &models.ChatMessage{
		ID:      uuid.New().String(),
		Role:    "user",
		Content: req.Question,
	}
	store.DB.Create(userMsg)

	sourcesJSON, _ := json.Marshal(sources)
	assistantMsg := &models.ChatMessage{
		ID:      uuid.New().String(),
		Role:    "assistant",
		Content: answer,
		Sources: string(sourcesJSON),
	}
	store.DB.Create(assistantMsg)

	c.JSON(http.StatusOK, gin.H{
		"answer":  answer,
		"sources": sources,
		"found":   len(results) > 0,
	})
}

// GetHistory returns recent chat history
func (h *Handler) GetHistory(c *gin.Context) {
	var messages []models.ChatMessage
	store.DB.Order("created_at asc").Limit(200).Find(&messages)
	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

// ClearHistory deletes all chat messages
func (h *Handler) ClearHistory(c *gin.Context) {
	store.DB.Where("1 = 1").Delete(&models.ChatMessage{})
	c.JSON(http.StatusOK, gin.H{"message": "History cleared"})
}

// ---- Configuration ----

// GetConfig returns the current configuration (with API key masked)
func (h *Handler) GetConfig(c *gin.Context) {
	cfg, err := store.GetConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Mask API key for security
	masked := *cfg
	if len(masked.APIKey) > 8 {
		masked.APIKey = masked.APIKey[:4] + "****" + masked.APIKey[len(masked.APIKey)-4:]
	} else if len(masked.APIKey) > 0 {
		masked.APIKey = "****"
	}

	c.JSON(http.StatusOK, masked)
}

// SaveConfig saves the LLM configuration
func (h *Handler) SaveConfig(c *gin.Context) {
	var cfg models.Config
	if err := c.ShouldBindJSON(&cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// If API key is masked, keep the existing one
	if cfg.APIKey == "****" || (len(cfg.APIKey) > 4 && cfg.APIKey[4:8] == "****") {
		existing, err := store.GetConfig()
		if err == nil {
			cfg.APIKey = existing.APIKey
		}
	}

	if cfg.TopK <= 0 {
		cfg.TopK = 5
	}
	if cfg.ScoreThresh <= 0 {
		cfg.ScoreThresh = 0.3
	}

	if err := store.SaveConfig(&cfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Configuration saved"})
}
