package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

// incomingMessage represents a JSON message received from the server.
type incomingMessage struct {
	Type    string `json:"type"`
	Message string `json:"message,omitempty"`
}

// screenshotPayload is the outgoing screenshot_data message.
type screenshotPayload struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

const (
	pingInterval   = 30 * time.Second
	reconnectMin   = 1 * time.Second
	reconnectMax   = 30 * time.Second
	reconnectScale = 2
)

// Client manages a WebSocket connection to the CrackOA server with
// automatic reconnection and keepalive pings.
type Client struct {
	serverURL string

	mu   sync.Mutex
	conn *websocket.Conn

	// OnLog is called for every notable event.
	OnLog func(msg string)
	// OnScreenshotRequest is called when the server asks for a screenshot.
	OnScreenshotRequest func()
	// OnConnStateChange is called when the connection state changes (true=connected, false=disconnected).
	OnConnStateChange func(connected bool)

	// stopCtx controls the entire lifecycle; cancelled by Disconnect().
	stopCtx    context.Context
	stopCancel context.CancelFunc
}

// NewClient creates a new WebSocket client targeting the given server base URL
// (e.g. "ws://localhost:8080").
func NewClient(serverURL string) *Client {
	return &Client{serverURL: serverURL}
}

// ConnectAndServe connects to the room and keeps the connection alive,
// automatically reconnecting on disconnect with exponential backoff.
// It blocks until Disconnect() is called.
func (c *Client) ConnectAndServe(code string) {
	ctx, cancel := context.WithCancel(context.Background())
	c.mu.Lock()
	c.stopCtx = ctx
	c.stopCancel = cancel
	c.mu.Unlock()

	backoff := reconnectMin

	for {
		err := c.dial(ctx, code)
		if err != nil {
			// If we were told to stop, exit.
			if ctx.Err() != nil {
				return
			}
			c.log("Connection failed: %v", err)
		} else {
			// Connection established — run read loop.
			backoff = reconnectMin // reset on successful connect
			c.readLoop(ctx)

			// If we were told to stop, exit.
			if ctx.Err() != nil {
				return
			}
			c.log("Connection lost")
		}

		c.notifyState(false)

		c.log("Reconnecting in %s…", backoff)
		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
		}

		backoff = backoff * reconnectScale
		if backoff > reconnectMax {
			backoff = reconnectMax
		}
	}
}

// Disconnect closes the connection and stops reconnection attempts.
func (c *Client) Disconnect() {
	c.mu.Lock()
	cancel := c.stopCancel
	conn := c.conn
	c.conn = nil
	c.stopCancel = nil
	c.mu.Unlock()

	if conn != nil {
		c.log("Disconnecting…")
		_ = conn.Close(websocket.StatusNormalClosure, "")
	}
	if cancel != nil {
		cancel()
	}
}

// SendScreenshot sends a screenshot_data message with the given Base64 PNG data.
func (c *Client) SendScreenshot(base64PNG string) error {
	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()

	if conn == nil {
		return fmt.Errorf("not connected")
	}

	payload := screenshotPayload{
		Type: "screenshot_data",
		Data: base64PNG,
	}

	c.log("→ Sending screenshot_data (%d bytes encoded)", len(base64PNG))

	if err := wsjson.Write(context.Background(), conn, payload); err != nil {
		c.log("Send error: %v", err)
		return err
	}

	c.log("→ Screenshot sent successfully")
	return nil
}

// Connected reports whether the client currently has an active connection.
func (c *Client) Connected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn != nil
}

// ── internal ────────────────────────────────────────────────────────

func (c *Client) dial(ctx context.Context, code string) error {
	url := fmt.Sprintf("%s/api/ws/desktop?code=%s", c.serverURL, code)
	c.log("Connecting to %s …", url)

	dialCtx, dialCancel := context.WithTimeout(ctx, 10*time.Second)
	defer dialCancel()

	conn, _, err := websocket.Dial(dialCtx, url, nil)
	if err != nil {
		return err
	}

	// Remove default read limit (screenshots can be large).
	conn.SetReadLimit(-1)

	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()

	c.log("Connected to room %s", code)
	c.notifyState(true)
	return nil
}

func (c *Client) readLoop(ctx context.Context) {
	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()
	if conn == nil {
		return
	}

	// Keepalive pings in background.
	pingDone := make(chan struct{})
	go func() {
		defer close(pingDone)
		ticker := time.NewTicker(pingInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := conn.Ping(ctx); err != nil {
					c.log("Ping failed: %v", err)
					return
				}
			}
		}
	}()

	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			c.log("Read error: %v", err)
			c.mu.Lock()
			c.conn = nil
			c.mu.Unlock()
			// Wait for ping goroutine to exit.
			<-pingDone
			return
		}

		var msg incomingMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			c.log("Invalid JSON: %v", err)
			continue
		}

		switch msg.Type {
		case "take_screenshot":
			c.log("← Received: take_screenshot")
			if c.OnScreenshotRequest != nil {
				c.OnScreenshotRequest()
			}
		case "client_disconnected":
			c.log("← Received: client_disconnected (web client left, waiting for reconnection)")
		case "error":
			c.log("← Received error: %s", msg.Message)
		default:
			c.log("← Received unknown message type: %s", msg.Type)
		}
	}
}

func (c *Client) notifyState(connected bool) {
	if c.OnConnStateChange != nil {
		c.OnConnStateChange(connected)
	}
}

func (c *Client) log(format string, args ...any) {
	if c.OnLog != nil {
		c.OnLog(fmt.Sprintf(format, args...))
	}
}
