package e2e_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
)

func TestRoomCreation(t *testing.T) {
	baseURL, stop := startApp(t, "x")
	defer stop()

	t.Run("web client creates room and receives room code", func(t *testing.T) {
		token := getToken(t, baseURL, "x")
		wsURL := "ws" + strings.TrimPrefix(baseURL, "http") + "/ws/client?token=" + token

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		conn, resp, err := websocket.Dial(ctx, wsURL, nil)
		if err != nil {
			t.Fatalf("ws dial: %v", err)
		}
		defer conn.CloseNow()

		if resp.StatusCode != http.StatusSwitchingProtocols {
			t.Fatalf("expected 101, got %d", resp.StatusCode)
		}

		_, data, err := conn.Read(ctx)
		if err != nil {
			t.Fatalf("read room_created message: %v", err)
		}

		var msg struct {
			Type string `json:"type"`
			Code string `json:"code"`
		}
		if err := json.Unmarshal(data, &msg); err != nil {
			t.Fatalf("unmarshal room_created: %v, raw: %s", err, string(data))
		}

		if msg.Type != "room_created" {
			t.Fatalf("expected type %q, got %q", "room_created", msg.Type)
		}

		if len(msg.Code) != 4 {
			t.Fatalf("expected 4-char room code, got %q", msg.Code)
		}

		t.Logf("room created with code: %s", msg.Code)

		conn.Close(websocket.StatusNormalClosure, "done")
	})

	t.Run("web client without token gets rejected", func(t *testing.T) {
		wsURL := "ws" + strings.TrimPrefix(baseURL, "http") + "/ws/client"

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		conn, _, err := websocket.Dial(ctx, wsURL, nil)
		if err == nil {
			conn.CloseNow()
			t.Fatal("expected dial to fail without token, but it succeeded")
		}
	})

	t.Run("web client with invalid token gets rejected", func(t *testing.T) {
		wsURL := "ws" + strings.TrimPrefix(baseURL, "http") + "/ws/client?token=invalid.token.here"

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		conn, _, err := websocket.Dial(ctx, wsURL, nil)
		if err == nil {
			conn.CloseNow()
			t.Fatal("expected dial to fail with bad token, but it succeeded")
		}
	})

	t.Run("desktop joins room and client receives notification", func(t *testing.T) {
		token := getToken(t, baseURL, "x")
		wsBase := "ws" + strings.TrimPrefix(baseURL, "http")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		clientConn, _, err := websocket.Dial(ctx, wsBase+"/ws/client?token="+token, nil)
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

		desktopConn, _, err := websocket.Dial(ctx, wsBase+"/ws/desktop?code="+roomMsg.Code, nil)
		if err != nil {
			t.Fatalf("desktop dial: %v", err)
		}
		defer desktopConn.CloseNow()

		_, data, err = clientConn.Read(ctx)
		if err != nil {
			t.Fatalf("read desktop_connected: %v", err)
		}

		var connMsg struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(data, &connMsg); err != nil {
			t.Fatalf("unmarshal desktop_connected: %v", err)
		}

		if connMsg.Type != "desktop_connected" {
			t.Fatalf("expected type %q, got %q", "desktop_connected", connMsg.Type)
		}

		t.Log("desktop joined and client was notified")

		desktopConn.Close(websocket.StatusNormalClosure, "done")
		clientConn.Close(websocket.StatusNormalClosure, "done")
	})

	t.Run("desktop with invalid code gets rejected", func(t *testing.T) {
		wsURL := "ws" + strings.TrimPrefix(baseURL, "http") + "/ws/desktop?code=ZZZZ"

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		conn, _, err := websocket.Dial(ctx, wsURL, nil)
		if err == nil {
			conn.CloseNow()
			t.Fatal("expected dial to fail with invalid room code, but it succeeded")
		}
	})

	t.Run("client reconnects to room after disconnect", func(t *testing.T) {
		token := getToken(t, baseURL, "x")
		wsBase := "ws" + strings.TrimPrefix(baseURL, "http")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// 1. Create room.
		clientConn, _, err := websocket.Dial(ctx, wsBase+"/ws/client?token="+token, nil)
		if err != nil {
			t.Fatalf("client dial: %v", err)
		}

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

		roomCode := roomMsg.Code
		t.Logf("room created: %s", roomCode)

		// 2. Desktop joins.
		desktopConn, _, err := websocket.Dial(ctx, wsBase+"/ws/desktop?code="+roomCode, nil)
		if err != nil {
			t.Fatalf("desktop dial: %v", err)
		}
		defer desktopConn.CloseNow()

		// Read desktop_connected on client.
		_, data, err = clientConn.Read(ctx)
		if err != nil {
			t.Fatalf("read desktop_connected: %v", err)
		}

		var connectedMsg struct{ Type string }
		_ = json.Unmarshal(data, &connectedMsg)

		if connectedMsg.Type != "desktop_connected" {
			t.Fatalf("expected desktop_connected, got %s", connectedMsg.Type)
		}

		// 3. Client disconnects (simulating accidental drop).
		clientConn.Close(websocket.StatusGoingAway, "accidental disconnect")

		// Small delay to let server process the disconnect.
		time.Sleep(200 * time.Millisecond)

		// 4. Client reconnects with room code.
		clientConn2, _, err := websocket.Dial(ctx, wsBase+"/ws/client?token="+token+"&room="+roomCode, nil)
		if err != nil {
			t.Fatalf("reconnect dial: %v", err)
		}
		defer clientConn2.CloseNow()

		// Should receive room_rejoined.
		_, data, err = clientConn2.Read(ctx)
		if err != nil {
			t.Fatalf("read room_rejoined: %v", err)
		}

		var rejoinMsg struct {
			Type string `json:"type"`
			Code string `json:"code"`
		}
		if err := json.Unmarshal(data, &rejoinMsg); err != nil {
			t.Fatalf("unmarshal room_rejoined: %v", err)
		}

		if rejoinMsg.Type != "room_rejoined" {
			t.Fatalf("expected room_rejoined, got %s", rejoinMsg.Type)
		}

		if rejoinMsg.Code != roomCode {
			t.Fatalf("expected code %s, got %s", roomCode, rejoinMsg.Code)
		}

		// Should also receive desktop_connected since desktop is still there.
		_, data, err = clientConn2.Read(ctx)
		if err != nil {
			t.Fatalf("read desktop_connected after rejoin: %v", err)
		}

		var stateMsg struct{ Type string }
		_ = json.Unmarshal(data, &stateMsg)

		if stateMsg.Type != "desktop_connected" {
			t.Fatalf("expected desktop_connected after rejoin, got %s", stateMsg.Type)
		}

		t.Logf("client successfully reconnected to room %s", roomCode)

		clientConn2.Close(websocket.StatusNormalClosure, "done")
	})
}

func getToken(t *testing.T, baseURL, key string) string {
	t.Helper()

	statusCode, body := postAuth(t, baseURL, key)
	if statusCode != http.StatusOK {
		t.Fatalf("auth failed: status %d, body: %s", statusCode, string(body))
	}

	var resp map[string]string
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("unmarshal auth response: %v", err)
	}

	token, ok := resp["token"]
	if !ok || token == "" {
		t.Fatal("no token in auth response")
	}

	return token
}
