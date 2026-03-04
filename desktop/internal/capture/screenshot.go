package capture

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/png"

	"github.com/kbinani/screenshot"
)

// CaptureBase64 takes a screenshot of the primary display and returns
// the image as a standard Base64-encoded PNG string.
func CaptureBase64() (string, error) {
	n := screenshot.NumActiveDisplays()
	if n == 0 {
		return "", fmt.Errorf("no active displays found")
	}

	bounds := screenshot.GetDisplayBounds(0)
	img, err := screenshot.CaptureRect(bounds)
	if err != nil {
		return "", fmt.Errorf("capture failed: %w", err)
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", fmt.Errorf("PNG encode failed: %w", err)
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}
