package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadUsesSecretFiles(t *testing.T) {
	t.Setenv("DOTENV_PATH", filepath.Join(t.TempDir(), "missing.env"))
	t.Setenv("CONFIG_PATH", filepath.Join("..", "config.yaml"))
	t.Setenv("ACCESS_KEY", "")
	t.Setenv("JWT_SECRET", "")
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("GEMINI_API_KEY", "")

	secretDir := t.TempDir()
	accessKeyPath := filepath.Join(secretDir, "access_key")
	jwtSecretPath := filepath.Join(secretDir, "jwt_secret")
	openAIKeyPath := filepath.Join(secretDir, "openai_api_key")

	if err := os.WriteFile(accessKeyPath, []byte(" swarm-access-key \n"), 0o600); err != nil {
		t.Fatalf("write access key secret: %v", err)
	}

	if err := os.WriteFile(jwtSecretPath, []byte("swarm-jwt-secret\n"), 0o600); err != nil {
		t.Fatalf("write jwt secret: %v", err)
	}

	if err := os.WriteFile(openAIKeyPath, []byte("openai-secret"), 0o600); err != nil {
		t.Fatalf("write openai key: %v", err)
	}

	t.Setenv("ACCESS_KEY_FILE", accessKeyPath)
	t.Setenv("JWT_SECRET_FILE", jwtSecretPath)
	t.Setenv("OPENAI_API_KEY_FILE", openAIKeyPath)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.Auth.AccessKey != "swarm-access-key" {
		t.Fatalf("unexpected access key: %q", cfg.Auth.AccessKey)
	}

	if cfg.Auth.JWTSecret != "swarm-jwt-secret" {
		t.Fatalf("unexpected jwt secret: %q", cfg.Auth.JWTSecret)
	}

	if cfg.AI.OpenAI.APIKey != "openai-secret" {
		t.Fatalf("unexpected openai key: %q", cfg.AI.OpenAI.APIKey)
	}

	if cfg.AI.Gemini.APIKey != "" {
		t.Fatalf("expected empty gemini key, got %q", cfg.AI.Gemini.APIKey)
	}
}
