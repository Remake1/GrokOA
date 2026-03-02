//go:build !darwin || !cgo

package stealth

import "errors"

func startStealthHotkey(onChange func(enabled bool, status string)) (func() error, string, error) {
	return nil, "Stealth mode is only available on macOS app bundles.", errors.New("stealth mode unsupported on this platform")
}
