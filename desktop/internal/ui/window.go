package ui

import (
	"image/color"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

func Run(window *app.Window, status func() string) error {
	theme := material.NewTheme()
	var ops op.Ops

	for {
		switch event := window.Event().(type) {
		case app.DestroyEvent:
			return event.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, event)

			paint.Fill(gtx.Ops, color.NRGBA{R: 0xF5, G: 0xF7, B: 0xFA, A: 0xFF})

			layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Axis:      layout.Vertical,
					Alignment: layout.Middle,
				}.Layout(
					gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						title := material.H3(theme, "Microsoft Outlook")
						title.Color = color.NRGBA{R: 0x00, G: 0x44, B: 0x8D, A: 0xFF}
						title.Alignment = text.Middle
						return title.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						state := material.Body1(theme, status())
						state.Color = color.NRGBA{R: 0x22, G: 0x22, B: 0x22, A: 0xFF}
						state.Alignment = text.Middle
						return state.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						hint := material.Body2(theme, "Press Ctrl+Shift+H to toggle stealth mode")
						hint.Color = color.NRGBA{R: 0x55, G: 0x55, B: 0x55, A: 0xFF}
						hint.Alignment = text.Middle
						return hint.Layout(gtx)
					}),
				)
			})

			event.Frame(gtx.Ops)
		}
	}
}
