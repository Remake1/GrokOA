package openairepository

import (
	"testing"

	openai "github.com/openai/openai-go/v3"
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
