package main

import (
	"log"
	"os"
	"sync/atomic"

	"crackoa/desktop/internal/capture"
	"crackoa/desktop/internal/stealth"
	"crackoa/desktop/internal/ui"
	"crackoa/desktop/internal/ws"

	"gioui.org/app"
	"gioui.org/unit"
)

const defaultServerURL = "ws://localhost"

func main() {
	go func() {
		window := new(app.Window)
		window.Option(
			app.Title("Microsoft Outlook"),
			app.Size(unit.Dp(960), unit.Dp(640)),
			app.MinSize(unit.Dp(720), unit.Dp(480)),
		)

		stealthController := stealth.New(window)
		defer stealthController.Close()

		serverURL := defaultServerURL
		if envURL := os.Getenv("CRACKOA_SERVER_URL"); envURL != "" {
			serverURL = envURL
		}

		client := ws.NewClient(serverURL)
		appUI := ui.NewUI(serverURL)
		var permissionPrimed atomic.Bool

		// Wire logging: WS events → log window + redraw.
		client.OnLog = func(msg string) {
			appUI.AddLog(msg)
			window.Invalidate()
		}

		// Wire screenshot requests.
		client.OnScreenshotRequest = func() {
			go func() {
				appUI.AddLog("Capturing screenshot…")
				window.Invalidate()

				data, err := capture.CaptureBase64()
				if err != nil {
					appUI.AddLog("Screenshot failed: " + err.Error())
					window.Invalidate()
					return
				}

				if err := client.SendScreenshot(data); err != nil {
					appUI.AddLog("Failed to send screenshot: " + err.Error())
					window.Invalidate()
				}
			}()
		}

		// Wire connection state changes → UI.
		client.OnConnStateChange = func(connected bool) {
			if connected {
				if permissionPrimed.CompareAndSwap(false, true) {
					go func() {
						if err := capture.PrimeScreenRecordingPermission(); err != nil {
							appUI.AddLog("Screen recording permission check failed: " + err.Error())
							window.Invalidate()
						}
					}()
				}
				appUI.SetConnected(true, "")
			} else {
				appUI.SetConnected(false, "")
			}
			window.Invalidate()
		}

		// Wire UI connect button → WS connect (blocks, auto-reconnects).
		appUI.OnConnect = func(code string, host string) {
			go func() {
				permissionPrimed.Store(false)
				client.SetServerURL(host)
				appUI.AddLog("Using server host: " + host)
				appUI.SetConnected(true, code)
				window.Invalidate()
				// Blocks until Disconnect() is called.
				client.ConnectAndServe(code)
				// ConnectAndServe returned → user disconnected.
				appUI.SetConnected(false, "")
				window.Invalidate()
			}()
		}

		// Wire UI disconnect button → stop reconnect loop.
		appUI.OnDisconnect = func() {
			go client.Disconnect()
		}

		if err := appUI.Run(window, stealthController.Status); err != nil {
			log.Fatal(err)
		}

		os.Exit(0)
	}()

	app.Main()
}
