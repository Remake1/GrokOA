package openairepository

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	aiservice "api/internal/service/ai"

	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"
)

var (
	ErrMissingAPIKey = errors.New("openai api key is required")
	ErrInvalidModel  = errors.New("model is required")
	ErrInvalidPrompt = errors.New("prompt is required")
)

type Repository struct {
	client openai.Client
}

func NewRepository(apiKey string) (*Repository, error) {
	trimmedAPIKey := strings.TrimSpace(apiKey)
	if trimmedAPIKey == "" {
		return nil, ErrMissingAPIKey
	}

	return &Repository{
		client: openai.NewClient(option.WithAPIKey(trimmedAPIKey)),
	}, nil
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

	inputContent := make(responses.ResponseInputMessageContentListParam, 0, len(request.ImagePaths)+1)
	inputContent = append(inputContent, responses.ResponseInputContentUnionParam{
		OfInputText: &responses.ResponseInputTextParam{Text: request.Prompt},
	})

	for _, imagePath := range request.ImagePaths {
		dataURL, err := imageFileToDataURL(imagePath)
		if err != nil {
			return fmt.Errorf("prepare image %q: %w", imagePath, err)
		}

		inputContent = append(inputContent, responses.ResponseInputContentUnionParam{
			OfInputImage: &responses.ResponseInputImageParam{
				Detail:   responses.ResponseInputImageDetailAuto,
				ImageURL: openai.String(dataURL),
			},
		})
	}

	stream := r.client.Responses.NewStreaming(ctx, responses.ResponseNewParams{
		Model: openai.ResponsesModel(strings.TrimSpace(request.Model)),
		Input: responses.ResponseNewParamsInputUnion{
			OfInputItemList: responses.ResponseInputParam{
				responses.ResponseInputItemParamOfMessage(inputContent, responses.EasyInputMessageRoleUser),
			},
		},
	})
	defer stream.Close()

	for stream.Next() {
		event := stream.Current()
		if event.Type != "response.output_text.delta" {
			continue
		}

		delta := event.AsResponseOutputTextDelta().Delta
		if delta == "" {
			continue
		}

		if err := onChunk(delta); err != nil {
			return fmt.Errorf("handle streamed delta: %w", err)
		}
	}

	if err := stream.Err(); err != nil {
		return fmt.Errorf("stream openai response: %w", err)
	}

	return nil
}

func imageFileToDataURL(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read image file: %w", err)
	}

	contentType := http.DetectContentType(data)
	if !strings.HasPrefix(contentType, "image/") {
		if byExt := mime.TypeByExtension(strings.ToLower(filepath.Ext(path))); strings.HasPrefix(byExt, "image/") {
			contentType = byExt
		} else {
			contentType = "application/octet-stream"
		}
	}

	encoded := base64.StdEncoding.EncodeToString(data)
	return fmt.Sprintf("data:%s;base64,%s", contentType, encoded), nil
}
