package ai

import (
	openai "github.com/sashabaranov/go-openai"
)

// NewGeminiClient returns a Client that calls Gemini via its OpenAI-compatible endpoint.
// Drop-in replacement for NewOpenAIClient — no new dependencies needed.
func NewGeminiClient(apiKey, model string) Client {
	if model == "" {
		model = "gemini-2.0-flash"
	}

	cfg := openai.DefaultConfig(apiKey)
	cfg.BaseURL = "https://generativelanguage.googleapis.com/v1beta/openai/"

	return &OpenAIClient{
		client: openai.NewClientWithConfig(cfg),
		model:  model,
	}
}
