package ai

import "context"

type ProviderName string

const (
	ProviderOpenAI ProviderName = "openai"
	ProviderGemini ProviderName = "gemini"
)

type ChatRequest struct {
	Model      string
	Prompt     string
	ImagePaths []string
}

type StreamChunkHandler func(delta string) error

type Provider interface {
	StreamChat(ctx context.Context, request ChatRequest, onChunk StreamChunkHandler) error
}
