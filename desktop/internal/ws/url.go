package ws

import (
	"fmt"
	"net/url"
	"strings"
)

// NormalizeServerURL accepts websocket URLs, HTTP URLs, and bare hosts,
// returning a websocket base URL suitable for dialing.
func NormalizeServerURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("server URL is required")
	}

	if !strings.Contains(raw, "://") {
		raw = "ws://" + raw
	}

	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("invalid server URL: %w", err)
	}

	switch u.Scheme {
	case "http":
		u.Scheme = "ws"
	case "https":
		u.Scheme = "wss"
	case "ws", "wss":
	default:
		return "", fmt.Errorf("unsupported scheme %q", u.Scheme)
	}

	if u.Host == "" {
		return "", fmt.Errorf("server host is required")
	}

	u.Path = strings.TrimRight(u.Path, "/")
	u.RawQuery = ""
	u.Fragment = ""

	return u.String(), nil
}
