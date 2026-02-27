package screenshot

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

const idLength = 8

type Service struct {
	storageDir string
}

func NewService(storageDir string) (*Service, error) {
	if err := os.MkdirAll(storageDir, 0o755); err != nil {
		return nil, fmt.Errorf("create screenshots dir: %w", err)
	}

	return &Service{storageDir: storageDir}, nil
}

func (s *Service) Save(data []byte) (string, error) {
	id, err := generateID()
	if err != nil {
		return "", fmt.Errorf("generate screenshot id: %w", err)
	}

	path := filepath.Join(s.storageDir, id+".png")

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", fmt.Errorf("write screenshot file: %w", err)
	}

	return id, nil
}

func (s *Service) Path(id string) string {
	return filepath.Join(s.storageDir, id+".png")
}

func generateID() (string, error) {
	b := make([]byte, idLength/2)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}
