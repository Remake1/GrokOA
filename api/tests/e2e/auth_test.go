package e2e_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestAuthorization(t *testing.T) {
	baseURL, stop := startApp(t, "x")
	defer stop()

	t.Run("correct key returns jwt token", func(t *testing.T) {
		statusCode, body := postAuth(t, baseURL, "x")
		if statusCode != http.StatusOK {
			t.Fatalf("expected status %d, got %d, body: %s", http.StatusOK, statusCode, string(body))
		}

		var response map[string]any
		if err := json.Unmarshal(body, &response); err != nil {
			t.Fatalf("expected JSON response with token, unmarshal error: %v, body: %s", err, string(body))
		}

		token, ok := response["token"].(string)
		if !ok || token == "" {
			t.Fatalf("expected non-empty jwt token in response body, got: %s", string(body))
		}
	})

	t.Run("wrong key returns wrong key response", func(t *testing.T) {
		_, body := postAuth(t, baseURL, "wrong")
		if !strings.Contains(string(body), "wrong key") {
			t.Fatalf("expected response to contain %q, got: %s", "wrong key", string(body))
		}
	})
}

func TestAuthorizationRateLimit(t *testing.T) {
	baseURL, stop := startApp(t, "x")
	defer stop()

	t.Run("sixth failed attempt is rate limited", func(t *testing.T) {
		for range 5 {
			statusCode, body := postAuth(t, baseURL, "wrong")
			if statusCode != http.StatusUnauthorized {
				t.Fatalf("expected first five failures to return %d, got %d, body: %s", http.StatusUnauthorized, statusCode, string(body))
			}
		}

		statusCode, body := postAuth(t, baseURL, "wrong")
		if statusCode != http.StatusTooManyRequests {
			t.Fatalf("expected sixth failure to return %d, got %d, body: %s", http.StatusTooManyRequests, statusCode, string(body))
		}

		if !strings.Contains(string(body), "too many failed login attempts") {
			t.Fatalf("expected rate limit response body, got: %s", string(body))
		}
	})
}

func TestAuthorizationSuccessNotCountedAgainstFailedLoginLimit(t *testing.T) {
	baseURL, stop := startApp(t, "x")
	defer stop()

	for range 5 {
		statusCode, body := postAuth(t, baseURL, "wrong")
		if statusCode != http.StatusUnauthorized {
			t.Fatalf("expected failed attempt to return %d, got %d, body: %s", http.StatusUnauthorized, statusCode, string(body))
		}
	}

	statusCode, body := postAuth(t, baseURL, "x")
	if statusCode != http.StatusOK {
		t.Fatalf("expected correct key after five failures to return %d, got %d, body: %s", http.StatusOK, statusCode, string(body))
	}
}

func startApp(t *testing.T, accessKey string) (string, func()) {
	t.Helper()

	rootDir := projectRoot(t)
	port := freePort(t)
	baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)

	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, "go", "run", "./cmd/app/main.go")
	cmd.Dir = rootDir

	logFile, err := os.CreateTemp(t.TempDir(), "app-log-*.txt")
	if err != nil {
		t.Fatalf("create temp log file: %v", err)
	}

	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.Env = append(
		os.Environ(),
		"ACCESS_KEY="+accessKey,
		"JWT_SECRET=test-jwt-secret",
		"CONFIG_PATH=config.yaml",
		"HTTP_HOST=127.0.0.1",
		fmt.Sprintf("HTTP_PORT=%d", port),
		"SCREENSHOT_DIR="+t.TempDir(),
	)

	if err := cmd.Start(); err != nil {
		_ = logFile.Close()
		t.Fatalf("start app: %v", err)
	}

	waitForLiveEndpoint(t, baseURL, cmd, logFile)

	stop := func() {
		cancel()
		_ = cmd.Wait()
		_ = logFile.Close()
	}

	return baseURL, stop
}

func waitForLiveEndpoint(t *testing.T, baseURL string, cmd *exec.Cmd, logFile *os.File) {
	t.Helper()

	client := &http.Client{Timeout: 300 * time.Millisecond}
	deadline := time.Now().Add(15 * time.Second)

	for time.Now().Before(deadline) {
		resp, err := client.Get(baseURL + "/api/live")
		if err == nil {
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}

		if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
			logs := readLogs(t, logFile.Name())
			t.Fatalf("app exited before becoming ready; logs:\n%s", logs)
		}

		time.Sleep(150 * time.Millisecond)
	}

	logs := readLogs(t, logFile.Name())
	t.Fatalf("app did not become ready within timeout; logs:\n%s", logs)
}

func postAuth(t *testing.T, baseURL, key string) (int, []byte) {
	t.Helper()

	payload, err := json.Marshal(map[string]string{"key": key})
	if err != nil {
		t.Fatalf("marshal auth payload: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, baseURL+"/api/auth", bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("create auth request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("perform auth request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read auth response body: %v", err)
	}

	return resp.StatusCode, body
}

func freePort(t *testing.T) int {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("allocate free port: %v", err)
	}
	defer listener.Close()

	return listener.Addr().(*net.TCPAddr).Port
}

func projectRoot(t *testing.T) string {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}

	return filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", ".."))
}

func readLogs(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Sprintf("failed to read logs from %s: %v", path, err)
	}

	return string(content)
}
