package ai

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

var (
	ErrProviderAlreadyRegistered = errors.New("provider already registered")
	ErrModelAlreadyRegistered    = errors.New("model already registered")
	ErrUnknownModel              = errors.New("model is not registered")
	ErrInvalidProvider           = errors.New("invalid provider")
	ErrInvalidModel              = errors.New("invalid model")
)

var GlobalRegistry = NewRegistry()

type Registry struct {
	mu              sync.RWMutex
	providers       map[ProviderName]Provider
	modelToProvider map[string]ProviderName
}

func NewRegistry() *Registry {
	return &Registry{
		providers:       make(map[ProviderName]Provider),
		modelToProvider: make(map[string]ProviderName),
	}
}

func (r *Registry) Register(name ProviderName, provider Provider, models []string) error {
	if strings.TrimSpace(string(name)) == "" || provider == nil {
		return ErrInvalidProvider
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.providers[name]; exists {
		return fmt.Errorf("%w: %q", ErrProviderAlreadyRegistered, name)
	}

	seen := make(map[string]struct{}, len(models))
	normalizedModels := make([]string, 0, len(models))
	for _, model := range models {
		normalized := normalizeModel(model)
		if normalized == "" {
			return ErrInvalidModel
		}

		if _, duplicate := seen[normalized]; duplicate {
			continue
		}

		if mappedTo, exists := r.modelToProvider[normalized]; exists {
			return fmt.Errorf("%w: model=%q provider=%q", ErrModelAlreadyRegistered, normalized, mappedTo)
		}

		seen[normalized] = struct{}{}
		normalizedModels = append(normalizedModels, normalized)
	}

	r.providers[name] = provider
	for _, model := range normalizedModels {
		r.modelToProvider[model] = name
	}

	return nil
}

func (r *Registry) Resolve(model string) (Provider, ProviderName, error) {
	normalizedModel := normalizeModel(model)
	if normalizedModel == "" {
		return nil, "", ErrInvalidModel
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	name, exists := r.modelToProvider[normalizedModel]
	if !exists {
		return nil, "", fmt.Errorf("%w: %q", ErrUnknownModel, normalizedModel)
	}

	provider := r.providers[name]
	if provider == nil {
		return nil, "", fmt.Errorf("%w: %q", ErrInvalidProvider, name)
	}

	return provider, name, nil
}

func normalizeModel(model string) string {
	return strings.ToLower(strings.TrimSpace(model))
}
