package ai

import (
	"context"
	"errors"
	"testing"
)

type stubProvider struct{}

func (stubProvider) StreamChat(ctx context.Context, request ChatRequest, onChunk StreamChunkHandler) error {
	_ = ctx
	_ = request
	_ = onChunk
	return nil
}

func TestRegistryResolve(t *testing.T) {
	t.Parallel()

	registry := NewRegistry()
	provider := stubProvider{}

	if err := registry.Register(ProviderOpenAI, provider, []string{"gpt-5.3-codex"}); err != nil {
		t.Fatalf("register provider: %v", err)
	}

	resolvedProvider, name, err := registry.Resolve(" GPT-5.3-CODEX ")
	if err != nil {
		t.Fatalf("resolve provider: %v", err)
	}

	if name != ProviderOpenAI {
		t.Fatalf("expected provider name %q, got %q", ProviderOpenAI, name)
	}

	if _, ok := resolvedProvider.(stubProvider); !ok {
		t.Fatalf("unexpected provider type: %T", resolvedProvider)
	}
}

func TestRegistryDuplicateModel(t *testing.T) {
	t.Parallel()

	registry := NewRegistry()
	if err := registry.Register(ProviderOpenAI, stubProvider{}, []string{"gpt-5.3-codex"}); err != nil {
		t.Fatalf("register first provider: %v", err)
	}

	err := registry.Register(ProviderGemini, stubProvider{}, []string{"gpt-5.3-codex"})
	if !errors.Is(err, ErrModelAlreadyRegistered) {
		t.Fatalf("expected ErrModelAlreadyRegistered, got: %v", err)
	}
}

func TestRegistryUnknownModel(t *testing.T) {
	t.Parallel()

	registry := NewRegistry()
	if err := registry.Register(ProviderOpenAI, stubProvider{}, []string{"gpt-5.3-codex"}); err != nil {
		t.Fatalf("register provider: %v", err)
	}

	_, _, err := registry.Resolve("unknown-model")
	if !errors.Is(err, ErrUnknownModel) {
		t.Fatalf("expected ErrUnknownModel, got: %v", err)
	}
}
