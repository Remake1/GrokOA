//go:build darwin

package capture

import (
	"fmt"
	"image"

	"github.com/kbinani/screenshot"
)

// PrimeScreenRecordingPermission performs a minimal screen capture attempt to
// trigger the macOS Screen Recording permission prompt without sending data.
func PrimeScreenRecordingPermission() error {
	if screenshot.NumActiveDisplays() == 0 {
		return fmt.Errorf("no active displays found")
	}

	bounds := screenshot.GetDisplayBounds(0)
	rect := image.Rect(bounds.Min.X, bounds.Min.Y, bounds.Min.X+1, bounds.Min.Y+1)
	if _, err := screenshot.CaptureRect(rect); err != nil {
		return fmt.Errorf("permission probe failed: %w", err)
	}

	return nil
}
