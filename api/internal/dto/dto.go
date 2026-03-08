package dto

import "encoding/json"

// AuthRequest is the request body for POST /api/auth.
type AuthRequest struct {
	Key string `json:"key"`
}

// WSMessage is a generic envelope for all WebSocket messages.
type WSMessage struct {
	Type string          `json:"type"`
	Raw  json.RawMessage `json:"-"`
}

func (m *WSMessage) UnmarshalJSON(data []byte) error {
	type alias WSMessage

	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}

	m.Type = a.Type
	m.Raw = data

	return nil
}

// RoomCreatedMsg is sent to the web client when a room is created.
type RoomCreatedMsg struct {
	Type string `json:"type"`
	Code string `json:"code"`
}

// ScreenshotMsg is sent to the web client when a screenshot is received.
type ScreenshotMsg struct {
	Type string `json:"type"`
	ID   string `json:"id"`
	Data string `json:"data"`
}

// ScreenshotDataMsg is sent by the desktop with base64-encoded screenshot data.
type ScreenshotDataMsg struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

// SimpleMsg is a general-purpose message with a type and optional message.
type SimpleMsg struct {
	Type    string `json:"type"`
	Message string `json:"message,omitempty"`
}

// AIChatRequestMsg is sent by the web client to start an AI chat stream.
type AIChatRequestMsg struct {
	Type          string   `json:"type"`
	Model         string   `json:"model"`
	Prompt        string   `json:"prompt"`
	ScreenshotIDs []string `json:"screenshot_ids"`
}

// AIChatChunkMsg is sent to the web client for each streamed AI response chunk.
type AIChatChunkMsg struct {
	Type  string `json:"type"`
	Delta string `json:"delta"`
}

// AIChatDoneMsg is sent to the web client when the AI stream finishes.
type AIChatDoneMsg struct {
	Type string `json:"type"`
}
