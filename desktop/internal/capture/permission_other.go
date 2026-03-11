//go:build !darwin

package capture

// PrimeScreenRecordingPermission is a no-op on non-macOS platforms.
func PrimeScreenRecordingPermission() error {
	return nil
}
