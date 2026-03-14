package ui

import (
	"image"
	"image/color"
	"strings"
	"sync"
	"time"
	"unicode"

	"crackoa/desktop/internal/ws"

	"gioui.org/app"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// colors
var (
	colBg         = color.NRGBA{R: 0x12, G: 0x12, B: 0x12, A: 0xFF} // gray-10  #121212
	colCard       = color.NRGBA{R: 0x2D, G: 0x2D, B: 0x2D, A: 0xFF} // gray-9.5 #2d2d2d
	colAccent     = color.NRGBA{R: 0x87, G: 0x1F, B: 0x23, A: 0xFF} // red-6    #871f23  (primary)
	colAccentHov  = color.NRGBA{R: 0xC5, G: 0x21, B: 0x33, A: 0xFF} // red-4    #C52133  (primary-light)
	colDanger     = color.NRGBA{R: 0xB7, G: 0x2A, B: 0x2E, A: 0xFF} // red-5    #B72A2E
	colTextPri    = color.NRGBA{R: 0xF5, G: 0xF5, B: 0xF5, A: 0xFF} // gray-1   #f5f5f5
	colTextSec    = color.NRGBA{R: 0x72, G: 0x72, B: 0x72, A: 0xFF} // gray-7   #727272
	colLogBg      = color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xFF} // black    #000000
	colLogText    = color.NRGBA{R: 0xD0, G: 0xD0, B: 0xD0, A: 0xFF} // gray-3   #d0d0d0
	colInputBg    = color.NRGBA{R: 0x22, G: 0x22, B: 0x22, A: 0xFF} // gray-9.7 #222222
	colInputBrd   = color.NRGBA{R: 0x57, G: 0x57, B: 0x57, A: 0xFF} // gray-8   #575757
	colStatusConn = color.NRGBA{R: 0x4A, G: 0xC2, B: 0x6B, A: 0xFF} // green-4  #4ac26b
)

// UI holds all mutable state for the Gio interface.
type UI struct {
	codeEditor widget.Editor
	hostEditor widget.Editor
	connectBtn widget.Clickable
	disconnBtn widget.Clickable
	logList    widget.List

	mu        sync.Mutex
	logs      []string
	connected bool
	roomCode  string
	errMsg    string

	// Callbacks set by the caller.
	OnConnect    func(code string, host string)
	OnDisconnect func()
}

// NewUI creates and returns a new UI instance.
func NewUI(defaultHost string) *UI {
	if strings.TrimSpace(defaultHost) == "" {
		defaultHost = "ws://localhost"
	}

	u := &UI{}
	u.codeEditor.SingleLine = true
	u.codeEditor.MaxLen = 4
	u.codeEditor.Filter = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	u.hostEditor.SingleLine = true
	u.hostEditor.SetText(defaultHost)
	u.logList.List.Axis = layout.Vertical
	return u
}

// AddLog appends a timestamped message to the log window. Thread-safe.
func (u *UI) AddLog(msg string) {
	ts := time.Now().Format("15:04:05")
	u.mu.Lock()
	u.logs = append(u.logs, ts+" │ "+msg)
	u.mu.Unlock()
}

// SetConnected updates the connection state shown in the UI. Thread-safe.
func (u *UI) SetConnected(connected bool, roomCode string) {
	u.mu.Lock()
	u.connected = connected
	u.roomCode = roomCode
	u.mu.Unlock()
}

// SetError shows an error message in the form area. Thread-safe.
func (u *UI) SetError(msg string) {
	u.mu.Lock()
	u.errMsg = msg
	u.mu.Unlock()
}

// Run drives the Gio event loop. Blocks until the window is closed.
func (u *UI) Run(window *app.Window, stealthStatus func() string) error {
	theme := material.NewTheme()
	var ops op.Ops

	for {
		switch event := window.Event().(type) {
		case app.DestroyEvent:
			return event.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, event)

			// Force uppercase on the editor content, preserving caret position.
			if txt := u.codeEditor.Text(); txt != strings.ToUpper(txt) {
				selStart, selEnd := u.codeEditor.Selection()
				u.codeEditor.SetText(strings.ToUpper(txt))
				u.codeEditor.SetCaret(selStart, selEnd)
			}

			// Handle button clicks.
			if u.connectBtn.Clicked(gtx) {
				code := strings.ToUpper(strings.TrimSpace(u.codeEditor.Text()))
				host, err := ws.NormalizeServerURL(u.hostEditor.Text())
				if len(code) == 4 && isAlphaNum(code) {
					if err != nil {
						u.mu.Lock()
						u.errMsg = "Enter a valid host, e.g. 172.16.3.88 or ws://localhost"
						u.mu.Unlock()
						continue
					}

					u.hostEditor.SetText(host)

					u.mu.Lock()
					u.errMsg = ""
					u.mu.Unlock()
					if u.OnConnect != nil {
						u.OnConnect(code, host)
					}
				} else {
					u.mu.Lock()
					u.errMsg = "Enter a valid 4-character room code"
					u.mu.Unlock()
				}
			}

			if u.disconnBtn.Clicked(gtx) {
				if u.OnDisconnect != nil {
					u.OnDisconnect()
				}
			}

			// Fill background.
			paint.Fill(gtx.Ops, colBg)

			// Main layout: vertical flex with padding.
			layout.UniformInset(unit.Dp(20)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					// Top card: form / status.
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return u.layoutCard(gtx, theme, stealthStatus)
					}),
					// Spacer.
					layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),
					// Log panel: fills remaining space.
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						return u.layoutLogPanel(gtx, theme)
					}),
				)
			})

			event.Frame(gtx.Ops)
		}
	}
}

// layoutCard draws the top card (form or connected status).
func (u *UI) layoutCard(gtx layout.Context, theme *material.Theme, stealthStatus func() string) layout.Dimensions {
	return drawRoundedRect(gtx, colCard, 12, func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(20)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
				// Title.
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					title := material.H5(theme, "CrackOA Desktop")
					title.Color = colTextPri
					title.Alignment = text.Middle
					return title.Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),
				// Stealth status.
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					s := material.Caption(theme, stealthStatus())
					s.Color = colTextSec
					s.Alignment = text.Middle
					return s.Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),
				// Form or connected state.
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					u.mu.Lock()
					connected := u.connected
					roomCode := u.roomCode
					errMsg := u.errMsg
					u.mu.Unlock()

					if connected {
						return u.layoutConnected(gtx, theme, roomCode)
					}
					return u.layoutForm(gtx, theme, errMsg)
				}),
			)
		})
	})
}

// layoutForm draws the room-code entry form.
func (u *UI) layoutForm(gtx layout.Context, theme *material.Theme, errMsg string) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.Body2(theme, "WebSocket Host")
			lbl.Color = colTextSec
			return lbl.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Max.X = gtx.Dp(unit.Dp(320))
			gtx.Constraints.Min.X = gtx.Dp(unit.Dp(320))
			return drawRoundedRect(gtx, colInputBg, 8, func(gtx layout.Context) layout.Dimensions {
				return drawBorder(gtx, colInputBrd, 8, 1, func(gtx layout.Context) layout.Dimensions {
					return layout.UniformInset(unit.Dp(10)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						ed := material.Editor(theme, &u.hostEditor, "ws://localhost")
						ed.Color = colTextPri
						ed.HintColor = colTextSec
						ed.TextSize = unit.Sp(16)
						return ed.Layout(gtx)
					})
				})
			})
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.Body2(theme, "Room Code")
			lbl.Color = colTextSec
			return lbl.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		// Input field.
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			// Constrain width.
			gtx.Constraints.Max.X = gtx.Dp(unit.Dp(200))
			gtx.Constraints.Min.X = gtx.Dp(unit.Dp(200))
			return drawRoundedRect(gtx, colInputBg, 8, func(gtx layout.Context) layout.Dimensions {
				// Border.
				return drawBorder(gtx, colInputBrd, 8, 1, func(gtx layout.Context) layout.Dimensions {
					return layout.UniformInset(unit.Dp(10)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						ed := material.Editor(theme, &u.codeEditor, "AB12")
						ed.Color = colTextPri
						ed.HintColor = colTextSec
						ed.TextSize = unit.Sp(20)
						return ed.Layout(gtx)
					})
				})
			})
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
		// Connect button.
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Max.X = gtx.Dp(unit.Dp(200))
			gtx.Constraints.Min.X = gtx.Dp(unit.Dp(200))
			return layoutButton(gtx, theme, &u.connectBtn, "Connect", colAccent, colAccentHov)
		}),
		// Error message.
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if errMsg == "" {
				return layout.Dimensions{}
			}
			return layout.Inset{Top: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				e := material.Caption(theme, errMsg)
				e.Color = colDanger
				e.Alignment = text.Middle
				return e.Layout(gtx)
			})
		}),
	)
}

// layoutConnected draws the connected status with a disconnect button.
func (u *UI) layoutConnected(gtx layout.Context, theme *material.Theme, code string) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			dot := material.Body1(theme, "● Connected to room "+code)
			dot.Color = colStatusConn
			dot.Alignment = text.Middle
			return dot.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Max.X = gtx.Dp(unit.Dp(200))
			gtx.Constraints.Min.X = gtx.Dp(unit.Dp(200))
			return layoutButton(gtx, theme, &u.disconnBtn, "Disconnect", colDanger, colDanger)
		}),
	)
}

// layoutLogPanel draws the log window.
func (u *UI) layoutLogPanel(gtx layout.Context, theme *material.Theme) layout.Dimensions {
	return drawRoundedRect(gtx, colLogBg, 10, func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(12)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				// Header.
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					h := material.Caption(theme, "LOG")
					h.Color = colTextSec
					h.Font.Weight = font.Bold
					return h.Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
				// Log entries.
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					u.mu.Lock()
					logs := make([]string, len(u.logs))
					copy(logs, u.logs)
					u.mu.Unlock()

					// Auto-scroll to bottom.
					if len(logs) > 0 {
						u.logList.List.ScrollTo(len(logs) - 1)
					}

					return material.List(theme, &u.logList).Layout(gtx, len(logs), func(gtx layout.Context, i int) layout.Dimensions {
						line := material.Caption(theme, logs[i])
						line.Color = colLogText
						line.Font.Typeface = "monospace"
						return line.Layout(gtx)
					})
				}),
			)
		})
	})
}

// ── Helpers ──────────────────────────────────────────────────────────

func isAlphaNum(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func drawRoundedRect(gtx layout.Context, bg color.NRGBA, radius int, w layout.Widget) layout.Dimensions {
	macro := op.Record(gtx.Ops)
	dims := w(gtx)
	call := macro.Stop()

	rrect := clip.RRect{
		Rect: image.Rectangle{Max: dims.Size},
		NW:   radius, NE: radius, SE: radius, SW: radius,
	}
	stack := rrect.Push(gtx.Ops)
	paint.Fill(gtx.Ops, bg)
	call.Add(gtx.Ops)
	stack.Pop()
	return dims
}

func drawBorder(gtx layout.Context, col color.NRGBA, radius, width int, w layout.Widget) layout.Dimensions {
	macro := op.Record(gtx.Ops)
	dims := w(gtx)
	call := macro.Stop()

	rrect := clip.RRect{
		Rect: image.Rectangle{Max: dims.Size},
		NW:   radius, NE: radius, SE: radius, SW: radius,
	}
	stack := clip.Stroke{Path: rrect.Path(gtx.Ops), Width: float32(width)}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	stack.Pop()

	call.Add(gtx.Ops)
	return dims
}

func layoutButton(gtx layout.Context, theme *material.Theme, click *widget.Clickable, label string, bg, bgHov color.NRGBA) layout.Dimensions {
	btnColor := bg
	if click.Hovered() {
		btnColor = bgHov
	}
	return drawRoundedRect(gtx, btnColor, 8, func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(12)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return click.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				lbl := material.Body1(theme, label)
				lbl.Color = colTextPri
				lbl.Alignment = text.Middle
				return lbl.Layout(gtx)
			})
		})
	})
}
