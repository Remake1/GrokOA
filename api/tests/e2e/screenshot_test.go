package e2e_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
)

func TestScreenshotWorkflow(t *testing.T) {
	baseURL, stop := startApp(t, "x")
	defer stop()

	t.Run("desktop sends screenshot and client receives it with id", func(t *testing.T) {
		token := getToken(t, baseURL, "x")
		wsBase := "ws" + strings.TrimPrefix(baseURL, "http")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// 1. Client creates room.
		clientConn, _, err := websocket.Dial(ctx, wsBase+"/api/ws/client?token="+token, nil)
		if err != nil {
			t.Fatalf("client dial: %v", err)
		}
		defer clientConn.CloseNow()

		_, data, err := clientConn.Read(ctx)
		if err != nil {
			t.Fatalf("read room_created: %v", err)
		}

		var roomMsg struct {
			Type string `json:"type"`
			Code string `json:"code"`
		}
		if err := json.Unmarshal(data, &roomMsg); err != nil {
			t.Fatalf("unmarshal room_created: %v", err)
		}

		t.Logf("room created: %s", roomMsg.Code)

		// 2. Desktop joins.
		desktopConn, _, err := websocket.Dial(ctx, wsBase+"/api/ws/desktop?code="+roomMsg.Code, nil)
		if err != nil {
			t.Fatalf("desktop dial: %v", err)
		}
		defer desktopConn.CloseNow()

		// Client reads desktop_connected.
		_, data, err = clientConn.Read(ctx)
		if err != nil {
			t.Fatalf("read desktop_connected: %v", err)
		}

		var connMsg struct{ Type string }
		_ = json.Unmarshal(data, &connMsg)

		if connMsg.Type != "desktop_connected" {
			t.Fatalf("expected desktop_connected, got %s", connMsg.Type)
		}

		// 3. Client requests screenshot.
		requestMsg, _ := json.Marshal(map[string]string{"type": "request_screenshot"})
		if err := clientConn.Write(ctx, websocket.MessageText, requestMsg); err != nil {
			t.Fatalf("write request_screenshot: %v", err)
		}

		// 4. Desktop receives take_screenshot command.
		_, data, err = desktopConn.Read(ctx)
		if err != nil {
			t.Fatalf("desktop read take_screenshot: %v", err)
		}

		var takeMsg struct{ Type string }
		if err := json.Unmarshal(data, &takeMsg); err != nil {
			t.Fatalf("unmarshal take_screenshot: %v", err)
		}

		if takeMsg.Type != "take_screenshot" {
			t.Fatalf("expected take_screenshot, got %s", takeMsg.Type)
		}

		t.Log("desktop received take_screenshot command")

		// 5. Desktop sends screenshot data (a small 1x1 red PNG).
		//nolint:lll
		pngBytes := []byte{
			0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, // PNG signature
			0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52, // IHDR chunk
			0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
			0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53,
			0xde, 0x00, 0x00, 0x00, 0x0c, 0x49, 0x44, 0x41, // IDAT chunk
			0x54, 0x08, 0xd7, 0x63, 0xf8, 0xcf, 0xc0, 0x00,
			0x00, 0x00, 0x03, 0x00, 0x01, 0x36, 0x28, 0x19,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, // IEND chunk
			0x44, 0xae, 0x42, 0x60, 0x82,
		}
		b64Data := base64.StdEncoding.EncodeToString(pngBytes)

		screenshotDataMsg, _ := json.Marshal(map[string]string{
			"type": "screenshot_data",
			"data": b64Data,
		})

		if err := desktopConn.Write(ctx, websocket.MessageText, screenshotDataMsg); err != nil {
			t.Fatalf("write screenshot_data: %v", err)
		}

		// 6. Client receives screenshot with ID and data.
		_, data, err = clientConn.Read(ctx)
		if err != nil {
			t.Fatalf("client read screenshot: %v", err)
		}

		var ssMsg struct {
			Type string `json:"type"`
			ID   string `json:"id"`
			Data string `json:"data"`
		}
		if err := json.Unmarshal(data, &ssMsg); err != nil {
			t.Fatalf("unmarshal screenshot: %v", err)
		}

		if ssMsg.Type != "screenshot" {
			t.Fatalf("expected type screenshot, got %s", ssMsg.Type)
		}

		if ssMsg.ID == "" {
			t.Fatal("expected non-empty screenshot ID")
		}

		if len(ssMsg.ID) != 8 {
			t.Fatalf("expected 8-char hex ID, got %q (%d chars)", ssMsg.ID, len(ssMsg.ID))
		}

		if ssMsg.Data != b64Data {
			t.Fatal("screenshot data mismatch")
		}

		t.Logf("client received screenshot id=%s (%d bytes base64)", ssMsg.ID, len(ssMsg.Data))

		// 7. Verify the screenshot was actually saved to disk.
		decoded, err := base64.StdEncoding.DecodeString(ssMsg.Data)
		if err != nil {
			t.Fatalf("decode returned screenshot data: %v", err)
		}

		if len(decoded) != len(pngBytes) {
			t.Fatalf("decoded size mismatch: got %d, want %d", len(decoded), len(pngBytes))
		}

		clientConn.Close(websocket.StatusNormalClosure, "done")
		desktopConn.Close(websocket.StatusNormalClosure, "done")
	})

	t.Run("client can request multiple screenshots", func(t *testing.T) {
		token := getToken(t, baseURL, "x")
		wsBase := "ws" + strings.TrimPrefix(baseURL, "http")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		clientConn, _, err := websocket.Dial(ctx, wsBase+"/api/ws/client?token="+token, nil)
		if err != nil {
			t.Fatalf("client dial: %v", err)
		}
		defer clientConn.CloseNow()

		// Read room_created.
		_, data, err := clientConn.Read(ctx)
		if err != nil {
			t.Fatalf("read room_created: %v", err)
		}

		var roomMsg struct{ Code string }
		_ = json.Unmarshal(data, &roomMsg)

		desktopConn, _, err := websocket.Dial(ctx, wsBase+"/api/ws/desktop?code="+roomMsg.Code, nil)
		if err != nil {
			t.Fatalf("desktop dial: %v", err)
		}
		defer desktopConn.CloseNow()

		// Read desktop_connected.
		_, _, _ = clientConn.Read(ctx)

		// Fake PNG for testing.
		fakePNG := base64.StdEncoding.EncodeToString([]byte("fake-png-data"))

		screenshotIDs := make(map[string]bool)

		for i := range 3 {
			// Client requests screenshot.
			reqMsg, _ := json.Marshal(map[string]string{"type": "request_screenshot"})
			if err := clientConn.Write(ctx, websocket.MessageText, reqMsg); err != nil {
				t.Fatalf("request %d: write: %v", i+1, err)
			}

			// Desktop reads take_screenshot.
			_, _, err = desktopConn.Read(ctx)
			if err != nil {
				t.Fatalf("request %d: desktop read: %v", i+1, err)
			}

			// Desktop sends screenshot.
			ssData, _ := json.Marshal(map[string]string{
				"type": "screenshot_data",
				"data": fakePNG,
			})
			if err := desktopConn.Write(ctx, websocket.MessageText, ssData); err != nil {
				t.Fatalf("request %d: desktop write: %v", i+1, err)
			}

			// Client receives screenshot.
			_, data, err = clientConn.Read(ctx)
			if err != nil {
				t.Fatalf("request %d: client read: %v", i+1, err)
			}

			var ssMsg struct {
				Type string `json:"type"`
				ID   string `json:"id"`
			}
			_ = json.Unmarshal(data, &ssMsg)

			if ssMsg.Type != "screenshot" {
				t.Fatalf("request %d: expected screenshot, got %s", i+1, ssMsg.Type)
			}

			if screenshotIDs[ssMsg.ID] {
				t.Fatalf("request %d: duplicate screenshot ID %s", i+1, ssMsg.ID)
			}

			screenshotIDs[ssMsg.ID] = true
			t.Logf("screenshot %d: id=%s", i+1, ssMsg.ID)
		}

		if len(screenshotIDs) != 3 {
			t.Fatalf("expected 3 unique screenshot IDs, got %d", len(screenshotIDs))
		}

		t.Log("all 3 screenshots received with unique IDs")
	})
}
