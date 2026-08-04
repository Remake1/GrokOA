package openairepository

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
)

func TestReasoningEffortForModel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		model string
		want  openai.ReasoningEffort
	}{
		{name: "sol", model: "gpt-5.6-sol", want: openai.ReasoningEffortMedium},
		{name: "terra", model: " GPT-5.6-TERRA ", want: openai.ReasoningEffortMedium},
		{name: "gpt 5.4 uses API default", model: "gpt-5.4", want: ""},
		{name: "gpt 5.4 mini uses API default", model: "gpt-5.4-mini", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := reasoningEffortForModel(tt.model); got != tt.want {
				t.Fatalf("reasoningEffortForModel(%q) = %q, want %q", tt.model, got, tt.want)
			}
		})
	}
}

func TestHandleStreamEvent(t *testing.T) {
	t.Parallel()

	t.Run("streams output text delta", func(t *testing.T) {
		event := streamEvent(t, `{"type":"response.output_text.delta","delta":"hello"}`)
		var output strings.Builder

		completed, received, err := handleStreamEvent(event, false, func(delta string) error {
			output.WriteString(delta)
			return nil
		})

		if err != nil {
			t.Fatalf("handleStreamEvent returned error: %v", err)
		}
		if completed {
			t.Fatal("delta event must not complete the stream")
		}
		if !received {
			t.Fatal("delta event must report received output")
		}
		if got := output.String(); got != "hello" {
			t.Fatalf("output = %q, want %q", got, "hello")
		}
	})

	t.Run("completes immediately and falls back to final output", func(t *testing.T) {
		event := streamEvent(t, `{
			"type":"response.completed",
			"response":{
				"output":[{
					"type":"message",
					"content":[{"type":"output_text","text":"final answer"}]
				}]
			}
		}`)
		var output strings.Builder

		completed, received, err := handleStreamEvent(event, false, func(delta string) error {
			output.WriteString(delta)
			return nil
		})

		if err != nil {
			t.Fatalf("handleStreamEvent returned error: %v", err)
		}
		if !completed {
			t.Fatal("response.completed must complete the stream")
		}
		if received {
			t.Fatal("terminal fallback must not be counted as a delta")
		}
		if got := output.String(); got != "final answer" {
			t.Fatalf("output = %q, want %q", got, "final answer")
		}
	})

	t.Run("does not duplicate final output after deltas", func(t *testing.T) {
		event := streamEvent(t, `{
			"type":"response.completed",
			"response":{
				"output":[{
					"type":"message",
					"content":[{"type":"output_text","text":"already streamed"}]
				}]
			}
		}`)
		called := false

		completed, _, err := handleStreamEvent(event, true, func(string) error {
			called = true
			return nil
		})

		if err != nil {
			t.Fatalf("handleStreamEvent returned error: %v", err)
		}
		if !completed {
			t.Fatal("response.completed must complete the stream")
		}
		if called {
			t.Fatal("completed response duplicated previously streamed output")
		}
	})

	t.Run("surfaces API error event", func(t *testing.T) {
		event := streamEvent(t, `{"type":"error","message":"rate limited"}`)

		_, _, err := handleStreamEvent(event, false, func(string) error { return nil })
		if err == nil || !strings.Contains(err.Error(), "rate limited") {
			t.Fatalf("error = %v, want rate limit message", err)
		}
	})

	t.Run("surfaces failed response", func(t *testing.T) {
		event := streamEvent(t, `{
			"type":"response.failed",
			"response":{"error":{"message":"model failed"}}
		}`)

		_, _, err := handleStreamEvent(event, false, func(string) error { return nil })
		if err == nil || !strings.Contains(err.Error(), "model failed") {
			t.Fatalf("error = %v, want failure message", err)
		}
	})

	t.Run("surfaces incomplete response", func(t *testing.T) {
		event := streamEvent(t, `{
			"type":"response.incomplete",
			"response":{"incomplete_details":{"reason":"max_output_tokens"}}
		}`)

		_, _, err := handleStreamEvent(event, false, func(string) error { return nil })
		if err == nil || !strings.Contains(err.Error(), "max_output_tokens") {
			t.Fatalf("error = %v, want incomplete reason", err)
		}
	})

	t.Run("propagates chunk handler error", func(t *testing.T) {
		event := streamEvent(t, `{"type":"response.output_text.delta","delta":"hello"}`)
		wantErr := errors.New("socket closed")

		_, _, err := handleStreamEvent(event, false, func(string) error { return wantErr })
		if !errors.Is(err, wantErr) {
			t.Fatalf("error = %v, want wrapped %v", err, wantErr)
		}
	})
}

func streamEvent(t *testing.T, raw string) responses.ResponseStreamEventUnion {
	t.Helper()

	var event responses.ResponseStreamEventUnion
	if err := json.Unmarshal([]byte(raw), &event); err != nil {
		t.Fatalf("unmarshal stream event: %v", err)
	}

	return event
}
