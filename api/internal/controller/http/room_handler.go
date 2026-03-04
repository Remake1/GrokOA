package http

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"api/internal/dto"
	roomservice "api/internal/service/room"
	screenshotservice "api/internal/service/screenshot"

	"github.com/coder/websocket"
	"github.com/rs/zerolog"
)

type tokenValidator interface {
	ValidateToken(tokenStr string) error
}

type RoomHandler struct {
	rooms       *roomservice.Manager
	screenshots *screenshotservice.Service
	auth        tokenValidator
	logger      zerolog.Logger
}

func NewRoomHandler(
	rooms *roomservice.Manager,
	screenshots *screenshotservice.Service,
	auth tokenValidator,
	logger zerolog.Logger,
) *RoomHandler {
	return &RoomHandler{
		rooms:       rooms,
		screenshots: screenshots,
		auth:        auth,
		logger:      logger,
	}
}

// --- Handlers ---

func (h *RoomHandler) HandleWebClient(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "token query parameter required"})

		return
	}

	if err := h.auth.ValidateToken(token); err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid or expired token"})

		return
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		h.logger.Error().Err(err).Msg("ws accept failed for web client")

		return
	}
	defer conn.CloseNow()

	roomCode := r.URL.Query().Get("room")

	if roomCode != "" {
		h.rejoinRoom(r.Context(), conn, roomCode)
	} else {
		h.createAndJoinRoom(r.Context(), conn)
	}
}

func (h *RoomHandler) createAndJoinRoom(ctx context.Context, conn *websocket.Conn) {
	code, err := h.rooms.CreateRoom()
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to create room")
		conn.Close(websocket.StatusInternalError, "failed to create room")

		return
	}

	room, _ := h.rooms.GetRoom(code)
	room.SetWebConn(conn)

	h.logger.Info().Str("room", code).Msg("room created by web client")

	if err := writeWSJSON(ctx, conn, dto.RoomCreatedMsg{Type: "room_created", Code: code}); err != nil {
		return
	}

	h.readWebClient(ctx, conn, room)
	h.handleClientDisconnect(room)
}

func (h *RoomHandler) rejoinRoom(ctx context.Context, conn *websocket.Conn, code string) {
	room, ok := h.rooms.GetRoom(code)
	if !ok {
		h.logger.Warn().Str("room", code).Msg("rejoin failed: room not found or expired")

		writeWSJSON(ctx, conn, dto.SimpleMsg{Type: "error", Message: "room not found or expired"})
		conn.Close(websocket.StatusNormalClosure, "room not found")

		return
	}

	room.SetWebConn(conn)
	room.NotifyReconnected()

	h.logger.Info().Str("room", code).Msg("web client reconnected to room")

	if err := writeWSJSON(ctx, conn, dto.RoomCreatedMsg{Type: "room_rejoined", Code: code}); err != nil {
		return
	}

	// Notify client about current desktop state.
	if room.DesktopConn() != nil {
		_ = writeWSJSON(ctx, conn, dto.SimpleMsg{Type: "desktop_connected"})
	}

	h.readWebClient(ctx, conn, room)
	h.handleClientDisconnect(room)
}

func (h *RoomHandler) handleClientDisconnect(room *roomservice.Room) {
	room.ClearWebConn()

	h.logger.Info().
		Str("room", room.Code).
		Dur("grace_period", h.rooms.GracePeriod()).
		Msg("web client disconnected, grace period started")

	// Notify desktop that client left.
	if dc := room.DesktopConn(); dc != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_ = writeWSJSON(ctx, dc, dto.SimpleMsg{Type: "client_disconnected"})
		cancel()
	}
}

func (h *RoomHandler) HandleDesktop(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "code query parameter required"})

		return
	}

	room, ok := h.rooms.GetRoom(code)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "room not found"})

		return
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		h.logger.Error().Err(err).Msg("ws accept failed for desktop")

		return
	}
	defer conn.CloseNow()

	conn.SetReadLimit(10 << 20) // 10 MB – desktop sends base64-encoded screenshots.

	room.SetDesktopConn(conn)

	h.logger.Info().Str("room", code).Msg("desktop connected to room")

	if webConn := room.WebConn(); webConn != nil {
		_ = writeWSJSON(r.Context(), webConn, dto.SimpleMsg{Type: "desktop_connected"})
	}

	h.readDesktop(r.Context(), conn, room)

	room.SetDesktopConn(nil)

	if webConn := room.WebConn(); webConn != nil {
		_ = writeWSJSON(r.Context(), webConn, dto.SimpleMsg{Type: "desktop_disconnected"})
	}
}

// --- Read loops ---

func (h *RoomHandler) readWebClient(ctx context.Context, conn *websocket.Conn, room *roomservice.Room) {
	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
				websocket.CloseStatus(err) == websocket.StatusGoingAway {
				h.logger.Info().Str("room", room.Code).Msg("web client disconnected normally")
			}

			return
		}

		var msg dto.WSMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			_ = writeWSJSON(ctx, conn, dto.SimpleMsg{Type: "error", Message: "invalid message format"})

			continue
		}

		switch msg.Type {
		case "request_screenshot":
			desktopConn := room.DesktopConn()
			if desktopConn == nil {
				_ = writeWSJSON(ctx, conn, dto.SimpleMsg{Type: "error", Message: "desktop not connected"})

				continue
			}

			if err := writeWSJSON(ctx, desktopConn, dto.SimpleMsg{Type: "take_screenshot"}); err != nil {
				_ = writeWSJSON(ctx, conn, dto.SimpleMsg{Type: "error", Message: "failed to reach desktop"})
			}
		case "close_room":
			h.rooms.DeleteRoom(room.Code)
			h.logger.Info().Str("room", room.Code).Msg("room closed by client")

			if dc := room.DesktopConn(); dc != nil {
				dc.Close(websocket.StatusNormalClosure, "room closed")
			}

			return
		default:
			_ = writeWSJSON(ctx, conn, dto.SimpleMsg{Type: "error", Message: "unknown message type"})
		}
	}
}

func (h *RoomHandler) readDesktop(ctx context.Context, conn *websocket.Conn, room *roomservice.Room) {
	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
				websocket.CloseStatus(err) == websocket.StatusGoingAway {
				h.logger.Info().Str("room", room.Code).Msg("desktop disconnected normally")
			}

			return
		}

		var msg dto.WSMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			_ = writeWSJSON(ctx, conn, dto.SimpleMsg{Type: "error", Message: "invalid message format"})

			continue
		}

		switch msg.Type {
		case "screenshot_data":
			h.handleScreenshotData(ctx, data, room)
		default:
			_ = writeWSJSON(ctx, conn, dto.SimpleMsg{Type: "error", Message: "unknown message type"})
		}
	}
}

func (h *RoomHandler) handleScreenshotData(ctx context.Context, data []byte, room *roomservice.Room) {
	var msg dto.ScreenshotDataMsg
	if err := json.Unmarshal(data, &msg); err != nil {
		h.logger.Error().Err(err).Str("room", room.Code).Msg("invalid screenshot_data message")

		return
	}

	imgBytes, err := base64.StdEncoding.DecodeString(msg.Data)
	if err != nil {
		h.logger.Error().Err(err).Str("room", room.Code).Msg("invalid base64 in screenshot_data")

		return
	}

	id, err := h.screenshots.Save(imgBytes)
	if err != nil {
		h.logger.Error().Err(err).Str("room", room.Code).Msg("failed to save screenshot")

		return
	}

	h.logger.Info().Str("room", room.Code).Str("screenshot_id", id).Msg("screenshot saved")

	webConn := room.WebConn()
	if webConn == nil {
		return
	}

	_ = writeWSJSON(ctx, webConn, dto.ScreenshotMsg{
		Type: "screenshot",
		ID:   id,
		Data: msg.Data,
	})
}

// --- Helpers ---

func writeWSJSON(ctx context.Context, conn *websocket.Conn, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return conn.Write(ctx, websocket.MessageText, data)
}
