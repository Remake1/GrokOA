//go:build darwin && cgo

package stealth

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework AppKit
#include <dispatch/dispatch.h>
#import <AppKit/AppKit.h>

static int crackoa_is_main_bundle_app(void) {
	@autoreleasepool {
		NSString *path = [[NSBundle mainBundle] bundlePath];
		return [path.pathExtension isEqualToString:@"app"] ? 1 : 0;
	}
}

static BOOL crackoa_set_policy(NSApplicationActivationPolicy policy) {
	__block BOOL ok = NO;
	dispatch_block_t apply = ^{
		NSApplication *application = [NSApplication sharedApplication];
		ok = [application setActivationPolicy:policy];
	};

	if ([NSThread isMainThread]) {
		apply();
	} else {
		dispatch_sync(dispatch_get_main_queue(), apply);
	}

	return ok;
}

static int crackoa_enable_stealth_mode(void) {
	@autoreleasepool {
		BOOL ok = crackoa_set_policy(NSApplicationActivationPolicyAccessory);
		dispatch_block_t hide = ^{
			[[NSApplication sharedApplication] hide:nil];
		};
		if ([NSThread isMainThread]) {
			hide();
		} else {
			dispatch_sync(dispatch_get_main_queue(), hide);
		}
		return ok ? 1 : 0;
	}
}

static int crackoa_disable_stealth_mode(void) {
	@autoreleasepool {
		BOOL ok = crackoa_set_policy(NSApplicationActivationPolicyRegular);
		dispatch_block_t show = ^{
			NSApplication *application = [NSApplication sharedApplication];
			[application unhide:nil];
			[application activateIgnoringOtherApps:YES];
		};
		if ([NSThread isMainThread]) {
			show();
		} else {
			dispatch_sync(dispatch_get_main_queue(), show);
		}
		return ok ? 1 : 0;
	}
}
*/
import "C"

import (
	"errors"
	"sync"

	"golang.design/x/hotkey"
)

func startStealthHotkey(onChange func(enabled bool, status string)) (func() error, string, error) {
	if C.crackoa_is_main_bundle_app() == 0 {
		return nil, "Stealth mode works only when launched from the .app bundle.", errors.New("not running from .app bundle")
	}

	hk := hotkey.New([]hotkey.Modifier{hotkey.ModCtrl, hotkey.ModShift}, hotkey.KeyH)
	if err := hk.Register(); err != nil {
		return nil, "", err
	}

	initialStatus := "Stealth mode OFF. Press Ctrl+Shift+H to hide from Dock and Command-Tab."
	done := make(chan struct{})
	var (
		mu      sync.Mutex
		enabled bool
	)

	go func() {
		for {
			select {
			case <-done:
				return
			case _, ok := <-hk.Keydown():
				if !ok {
					return
				}

				mu.Lock()
				next := !enabled
				var toggleErr error
				if next {
					if C.crackoa_enable_stealth_mode() == 0 {
						toggleErr = errors.New("failed to enable stealth mode")
					}
				} else {
					if C.crackoa_disable_stealth_mode() == 0 {
						toggleErr = errors.New("failed to disable stealth mode")
					}
				}

				current := enabled
				if toggleErr == nil {
					enabled = next
					current = enabled
				}
				mu.Unlock()

				if toggleErr != nil {
					onChange(current, "Stealth mode toggle failed.")
					continue
				}

				if current {
					onChange(true, "Stealth mode ON. App hidden from Dock and Command-Tab.")
				} else {
					onChange(false, "Stealth mode OFF. App visible in Dock and Command-Tab.")
				}
			}
		}
	}()

	stop := func() error {
		close(done)

		mu.Lock()
		wasEnabled := enabled
		enabled = false
		mu.Unlock()

		if wasEnabled {
			_ = C.crackoa_disable_stealth_mode()
		}

		return hk.Unregister()
	}

	return stop, initialStatus, nil
}
