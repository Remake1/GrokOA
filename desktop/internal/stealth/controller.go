package stealth

import (
	"fmt"
	"sync"

	"gioui.org/app"
)

type Controller struct {
	mu     sync.RWMutex
	status string
	stop   func() error
}

func New(window *app.Window) *Controller {
	controller := &Controller{
		status: "Stealth mode is unavailable.",
	}

	stop, initialStatus, err := startStealthHotkey(func(enabled bool, status string) {
		controller.mu.Lock()
		controller.status = status
		controller.mu.Unlock()
		window.Invalidate()
	})

	if err != nil {
		controller.mu.Lock()
		controller.status = fmt.Sprintf("Stealth mode unavailable: %v", err)
		controller.mu.Unlock()
		return controller
	}

	controller.mu.Lock()
	controller.stop = stop
	controller.status = initialStatus
	controller.mu.Unlock()
	return controller
}

func (c *Controller) Close() {
	c.mu.RLock()
	stop := c.stop
	c.mu.RUnlock()
	if stop != nil {
		_ = stop()
	}
}

func (c *Controller) Status() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.status
}
