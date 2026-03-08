package geminirepository

import (
	"context"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	aiservice "api/internal/service/ai"

	"google.golang.org/genai"
)

var (
	ErrMissingAPIKey = errors.New("gemini api key is required")
	ErrInvalidModel  = errors.New("model is required")
	ErrInvalidPrompt = errors.New("prompt is required")
)

type Repository struct {
	client *genai.Client
}

func NewRepository(ctx context.Context, apiKey string) (*Repository, error) {
	trimmedAPIKey := strings.TrimSpace(apiKey)
	if trimmedAPIKey == "" {
		return nil, ErrMissingAPIKey
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  trimmedAPIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("create gemini client: %w", err)
	}

	return &Repository{client: client}, nil
}

func (r *Repository) StreamChat(
	ctx context.Context,
	request aiservice.ChatRequest,
	onChunk aiservice.StreamChunkHandler,
) error {
	if strings.TrimSpace(request.Model) == "" {
		return ErrInvalidModel
	}

	if strings.TrimSpace(request.Prompt) == "" {
		return ErrInvalidPrompt
	}

	if onChunk == nil {
		return errors.New("stream chunk handler is required")
	}

	parts := make([]*genai.Part, 0, len(request.ImagePaths)+1)
	parts = append(parts, genai.NewPartFromText(request.Prompt))

	for _, imagePath := range request.ImagePaths {
		data, mimeType, err := readImageFile(imagePath)
		if err != nil {
			return fmt.Errorf("prepare image %q: %w", imagePath, err)
		}

		parts = append(parts, genai.NewPartFromBytes(data, mimeType))
	}

	contents := []*genai.Content{
		genai.NewContentFromParts(parts, genai.RoleUser),
	}

	for result, err := range r.client.Models.GenerateContentStream(
		ctx,
		strings.TrimSpace(request.Model),
		contents,
		nil,
	) {
		if err != nil {
			return fmt.Errorf("stream gemini response: %w", err)
		}

		if result == nil {
			continue
		}

		delta := result.Text()
		if delta == "" {
			continue
		}

		if err := onChunk(delta); err != nil {
			return fmt.Errorf("handle streamed delta: %w", err)
		}
	}

	return nil
}

func readImageFile(path string) ([]byte, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", fmt.Errorf("read image file: %w", err)
	}

	contentType := http.DetectContentType(data)
	if !strings.HasPrefix(contentType, "image/") {
		if byExt := mime.TypeByExtension(strings.ToLower(filepath.Ext(path))); strings.HasPrefix(byExt, "image/") {
			contentType = byExt
		} else {
			contentType = "application/octet-stream"
		}
	}

	return data, contentType, nil
}
