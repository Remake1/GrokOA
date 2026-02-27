package room

import (
	"crypto/rand"
	"math/big"
	"sync"
	"time"

	"github.com/coder/websocket"
)

const (
	codeLength = 4
	codeChars  = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // no I, O, 0, 1 to avoid confusion
	maxRetries = 10

	DefaultGracePeriod = 30 * time.Second
	cleanupInterval    = 10 * time.Second
)

type Room struct {
	Code      string
	CreatedAt time.Time

	mu           sync.Mutex
	webConn      *websocket.Conn
	desktopConn  *websocket.Conn
	clientLeftAt *time.Time // set when web client disconnects, cleared on rejoin
	reconnected  chan struct{}
}

func (r *Room) SetWebConn(conn *websocket.Conn) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.webConn = conn
	r.clientLeftAt = nil
}

func (r *Room) WebConn() *websocket.Conn {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.webConn
}

func (r *Room) ClearWebConn() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.webConn = nil

	now := time.Now()
	r.clientLeftAt = &now
}

func (r *Room) ClientLeftAt() *time.Time {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.clientLeftAt
}

func (r *Room) SetDesktopConn(conn *websocket.Conn) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.desktopConn = conn
}

func (r *Room) DesktopConn() *websocket.Conn {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.desktopConn
}

// Reconnected returns a channel that is closed when the client reconnects.
// Used by the grace-period goroutine to abort cleanup.
func (r *Room) Reconnected() <-chan struct{} {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.reconnected
}

// NotifyReconnected signals that the client has reconnected.
func (r *Room) NotifyReconnected() {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Close the old channel to unblock any waiting goroutine, then create a new one.
	select {
	case <-r.reconnected:
		// already closed
	default:
		close(r.reconnected)
	}

	r.reconnected = make(chan struct{})
}

// IsAbandoned returns true if both connections are nil.
func (r *Room) IsAbandoned() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.webConn == nil && r.desktopConn == nil
}

type Manager struct {
	mu          sync.Mutex
	rooms       map[string]*Room
	gracePeriod time.Duration
	stopCleanup chan struct{}
}

func NewManager(gracePeriod time.Duration) *Manager {
	if gracePeriod <= 0 {
		gracePeriod = DefaultGracePeriod
	}

	m := &Manager{
		rooms:       make(map[string]*Room),
		gracePeriod: gracePeriod,
		stopCleanup: make(chan struct{}),
	}

	go m.cleanupLoop()

	return m
}

func (m *Manager) GracePeriod() time.Duration {
	return m.gracePeriod
}

func (m *Manager) CreateRoom() (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for range maxRetries {
		code, err := generateCode()
		if err != nil {
			return "", err
		}

		if _, exists := m.rooms[code]; exists {
			continue
		}

		m.rooms[code] = &Room{
			Code:        code,
			CreatedAt:   time.Now(),
			reconnected: make(chan struct{}),
		}

		return code, nil
	}

	return "", ErrCodeCollision
}

func (m *Manager) GetRoom(code string) (*Room, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	r, ok := m.rooms[code]

	return r, ok
}

func (m *Manager) DeleteRoom(code string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.rooms, code)
}

// Stop terminates the background cleanup goroutine.
func (m *Manager) Stop() {
	close(m.stopCleanup)
}

// cleanupLoop periodically removes abandoned rooms whose grace period has expired.
func (m *Manager) cleanupLoop() {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.evictExpired()
		case <-m.stopCleanup:
			return
		}
	}
}

func (m *Manager) evictExpired() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()

	for code, room := range m.rooms {
		leftAt := room.ClientLeftAt()
		if leftAt == nil {
			continue
		}

		if now.Sub(*leftAt) > m.gracePeriod {
			// Close desktop if still connected.
			if dc := room.DesktopConn(); dc != nil {
				dc.Close(websocket.StatusGoingAway, "room expired")
			}

			delete(m.rooms, code)
		}
	}
}

func generateCode() (string, error) {
	b := make([]byte, codeLength)

	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(codeChars))))
		if err != nil {
			return "", err
		}

		b[i] = codeChars[n.Int64()]
	}

	return string(b), nil
}
