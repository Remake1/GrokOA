package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrWrongKey     = errors.New("wrong key")
	ErrEmptyKey     = errors.New("key is required")
	ErrIssueToken   = errors.New("issue token")
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

type Service struct {
	accessKey string
	jwtSecret string
	tokenTTL  time.Duration
	now       func() time.Time
}

func NewService(accessKey, jwtSecret string, tokenTTL time.Duration) *Service {
	if tokenTTL <= 0 {
		tokenTTL = 4 * time.Hour
	}

	return &Service{
		accessKey: accessKey,
		jwtSecret: jwtSecret,
		tokenTTL:  tokenTTL,
		now:       time.Now,
	}
}

func (s *Service) Authorize(ctx context.Context, key string) (string, error) {
	_ = ctx

	if strings.TrimSpace(key) == "" {
		return "", ErrEmptyKey
	}

	if !hmac.Equal([]byte(key), []byte(s.accessKey)) {
		return "", ErrWrongKey
	}

	token, err := s.issueToken()
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrIssueToken, err)
	}

	return token, nil
}

func (s *Service) issueToken() (string, error) {
	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}

	now := s.now().UTC()
	claims := map[string]any{
		"iat": now.Unix(),
		"exp": now.Add(s.tokenTTL).Unix(),
	}

	headerPart, err := encodeSegment(header)
	if err != nil {
		return "", err
	}

	claimsPart, err := encodeSegment(claims)
	if err != nil {
		return "", err
	}

	signingInput := headerPart + "." + claimsPart

	mac := hmac.New(sha256.New, []byte(s.jwtSecret))
	if _, err := mac.Write([]byte(signingInput)); err != nil {
		return "", err
	}

	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return signingInput + "." + signature, nil
}

func encodeSegment(v any) (string, error) {
	payload, err := json.Marshal(v)
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(payload), nil
}

func (s *Service) ValidateToken(tokenStr string) error {
	parts := strings.SplitN(tokenStr, ".", 3)
	if len(parts) != 3 {
		return ErrInvalidToken
	}

	signingInput := parts[0] + "." + parts[1]

	mac := hmac.New(sha256.New, []byte(s.jwtSecret))
	if _, err := mac.Write([]byte(signingInput)); err != nil {
		return ErrInvalidToken
	}

	expectedSig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(parts[2]), []byte(expectedSig)) {
		return ErrInvalidToken
	}

	claimsJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return ErrInvalidToken
	}

	var claims map[string]any
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return ErrInvalidToken
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return ErrInvalidToken
	}

	if s.now().UTC().Unix() > int64(exp) {
		return ErrExpiredToken
	}

	return nil
}
