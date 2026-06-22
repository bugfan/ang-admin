package chat

import (
	"context"
	"fmt"
	"strings"

	"github.com/bugfan/empty015/internal/models"
	"github.com/bugfan/empty015/internal/rag"
	"github.com/sashabaranov/go-openai"
)

const systemPrompt = `You are a precise knowledge assistant. Your ONLY job is to answer questions based on the provided context excerpts from the user's knowledge base.

STRICT RULES:
1. ONLY use information from the provided [CONTEXT] sections below. 
2. If the question cannot be answered from the provided context, respond with exactly: "I'm sorry, I cannot find relevant information about this topic in the provided knowledge base."
3. Do NOT use any external knowledge, assumptions, or hallucinations.
4. Always cite which source (book/document title) the information comes from.
5. Be concise and accurate. Quote the source material when appropriate.`

// ProviderDefaults holds provider-specific default values
type ProviderDefaults struct {
	BaseURL    string
	ChatModel  string
	EmbedModel string
}

// KnownProviders maps provider name to their defaults
var KnownProviders = map[string]ProviderDefaults{
	"openai": {
		BaseURL:    "https://api.openai.com/v1",
		ChatModel:  "gpt-4o-mini",
		EmbedModel: "text-embedding-3-small",
	},
	"gemini": {
		BaseURL:    "https://generativelanguage.googleapis.com/v1beta/openai/",
		ChatModel:  "gemini-2.0-flash",
		EmbedModel: "text-embedding-004",
	},
	"zhipu": {
		BaseURL:    "https://open.bigmodel.cn/api/paas/v4/",
		ChatModel:  "glm-4-flash",
		EmbedModel: "embedding-3",
	},
	"volcengine": {
		BaseURL:    "https://ark.cn-beijing.volces.com/api/v3/",
		ChatModel:  "doubao-1-5-pro-32k-250115",
		EmbedModel: "doubao-embedding-large-text-240915",
	},
	"deepseek": {
		BaseURL:    "https://api.deepseek.com",
		ChatModel:  "deepseek-chat",
		EmbedModel: "", // DeepSeek has no embedding; user must use another provider
	},
	"qwen": {
		BaseURL:    "https://dashscope.aliyuncs.com/compatible-mode/v1",
		ChatModel:  "qwen-plus",
		EmbedModel: "text-embedding-v3",
	},
	"moonshot": {
		BaseURL:    "https://api.moonshot.cn/v1",
		ChatModel:  "moonshot-v1-8k",
		EmbedModel: "", // Moonshot has no embedding
	},
	"custom": {
		BaseURL:    "",
		ChatModel:  "",
		EmbedModel: "",
	},
}

// ResolveBaseURL returns the effective base URL for a provider config
func ResolveBaseURL(cfg *models.Config) string {
	if cfg.BaseURL != "" {
		return cfg.BaseURL
	}
	if defaults, ok := KnownProviders[cfg.Provider]; ok {
		return defaults.BaseURL
	}
	return ""
}

// LLMClient wraps the chat completion API
type LLMClient struct {
	client *openai.Client
	model  string
}

// NewLLMClient creates an LLM chat client based on configuration
func NewLLMClient(cfg *models.Config) (*LLMClient, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("API key is not configured. Please configure it in Settings.")
	}

	baseURL := ResolveBaseURL(cfg)
	config := openai.DefaultConfig(cfg.APIKey)
	if baseURL != "" {
		config.BaseURL = baseURL
	}

	chatModel := cfg.ChatModel
	if chatModel == "" {
		if defaults, ok := KnownProviders[cfg.Provider]; ok {
			chatModel = defaults.ChatModel
		}
		if chatModel == "" {
			chatModel = "gpt-4o-mini"
		}
	}

	return &LLMClient{
		client: openai.NewClientWithConfig(config),
		model:  chatModel,
	}, nil
}

// Answer generates a RAG-grounded answer for the user question
func (l *LLMClient) Answer(ctx context.Context, question string, results []rag.SearchResult) (string, []string, error) {
	if len(results) == 0 {
		return "I'm sorry, I cannot find relevant information about this topic in the provided knowledge base.", nil, nil
	}

	// Build context from retrieved chunks
	var contextParts []string
	sourceSet := map[string]bool{}
	for i, r := range results {
		contextParts = append(contextParts, fmt.Sprintf("[CONTEXT %d - Source: %s]\n%s", i+1, r.BookTitle, r.Chunk.Content))
		sourceSet[r.BookTitle] = true
	}

	var sources []string
	for s := range sourceSet {
		if s != "" {
			sources = append(sources, s)
		}
	}

	contextBlock := strings.Join(contextParts, "\n\n---\n\n")

	userMessage := fmt.Sprintf(`Please answer the following question using ONLY the context provided below.

%s

---

Question: %s`, contextBlock, question)

	resp, err := l.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: l.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: userMessage,
			},
		},
		Temperature: 0.1,
		MaxTokens:   1500,
	})
	if err != nil {
		return "", nil, fmt.Errorf("LLM error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", nil, fmt.Errorf("no response from LLM")
	}

	answer := resp.Choices[0].Message.Content
	return answer, sources, nil
}
