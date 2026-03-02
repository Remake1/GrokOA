package main

import (
	"log"
	"os"

	"crackoa/desktop/internal/stealth"
	"crackoa/desktop/internal/ui"
	"gioui.org/app"
	"gioui.org/unit"
)

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

		if err := ui.Run(window, stealthController.Status); err != nil {
			log.Fatal(err)
		}

		os.Exit(0)
	}()

	app.Main()
}
