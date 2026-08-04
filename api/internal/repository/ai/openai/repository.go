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

	params := responses.ResponseNewParams{
		Model: openai.ResponsesModel(strings.TrimSpace(request.Model)),
		Input: responses.ResponseNewParamsInputUnion{
			OfInputItemList: responses.ResponseInputParam{
				responses.ResponseInputItemParamOfMessage(inputContent, responses.EasyInputMessageRoleUser),
			},
		},
	}

	if reasoningEffort := reasoningEffortForModel(request.Model); reasoningEffort != "" {
		params.Reasoning = openai.ReasoningParam{Effort: reasoningEffort}
	}

	stream := r.client.Responses.NewStreaming(ctx, params)
	defer stream.Close()

	receivedOutput := false
	for stream.Next() {
		event := stream.Current()
		completed, outputReceived, err := handleStreamEvent(event, receivedOutput, onChunk)
		if err != nil {
			return err
		}
		receivedOutput = receivedOutput || outputReceived

		// response.completed is the terminal lifecycle event. Do not wait for the
		// underlying SSE connection to close, since a proxy may keep it alive.
		if completed {
			return nil
		}
	}

	if err := stream.Err(); err != nil {
		return fmt.Errorf("stream openai response: %w", err)
	}

	return errors.New("openai response stream ended before a terminal event")
}

func handleStreamEvent(
	event responses.ResponseStreamEventUnion,
	receivedOutput bool,
	onChunk aiservice.StreamChunkHandler,
) (completed bool, outputReceived bool, err error) {
	switch event.Type {
	case "response.output_text.delta", "response.refusal.delta":
		if event.Delta == "" {
			return false, false, nil
		}

		if err := onChunk(event.Delta); err != nil {
			return false, false, fmt.Errorf("handle streamed delta: %w", err)
		}

		return false, true, nil

	case "response.completed":
		// The completed response is a safe fallback if the API or an intermediary
		// delivered the terminal event but no text delta events.
		if !receivedOutput {
			if output := event.Response.OutputText(); output != "" {
				if err := onChunk(output); err != nil {
					return false, false, fmt.Errorf("handle completed response: %w", err)
				}
			}
		}

		return true, false, nil

	case "error":
		message := strings.TrimSpace(event.Message)
		if message == "" {
			message = "unknown streaming error"
		}

		return false, false, fmt.Errorf("openai stream error: %s", message)

	case "response.failed":
		message := strings.TrimSpace(event.Response.Error.Message)
		if message == "" {
			message = "response generation failed"
		}

		return false, false, fmt.Errorf("openai response failed: %s", message)

	case "response.incomplete":
		reason := strings.TrimSpace(event.Response.IncompleteDetails.Reason)
		if reason == "" {
			reason = "unknown reason"
		}

		return false, false, fmt.Errorf("openai response incomplete: %s", reason)
	}

	return false, false, nil
}

func reasoningEffortForModel(model string) openai.ReasoningEffort {
	switch strings.ToLower(strings.TrimSpace(model)) {
	case "gpt-5.6-sol", "gpt-5.6-terra":
		return openai.ReasoningEffortMedium
	default:
		return ""
	}
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
