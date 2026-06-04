package hub

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/VinnsEdesigner/vyzorix-update-server/models"
	"github.com/VinnsEdesigner/vyzorix-update-server/storage"
	"github.com/gorilla/websocket"
)

const writeTimeout = 10 * time.Second
const pongWait = 70 * time.Second
const pingPeriod = 30 * time.Second

type Client struct {
	DeviceID string
	Conn     *websocket.Conn
	Send     chan models.CommandFrame
	Hub      *Hub
}
type Hub struct {
	log       *slog.Logger
	store     *storage.Store
	mu        sync.RWMutex
	clients   map[string]*Client
	dashboard chan []byte
}

func New(log *slog.Logger, st *storage.Store) *Hub {
	return &Hub{log: log, store: st, clients: map[string]*Client{}, dashboard: make(chan []byte, 256)}
}
func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	if old := h.clients[c.DeviceID]; old != nil {
		close(old.Send)
		_ = old.Conn.Close()
	}
	h.clients[c.DeviceID] = c
	h.mu.Unlock()
	_ = h.store.SetOnline(context.Background(), c.DeviceID, true)
	h.log.Info("device websocket online", "deviceId", c.DeviceID)
}
func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	if h.clients[c.DeviceID] == c {
		delete(h.clients, c.DeviceID)
		close(c.Send)
		_ = h.store.SetOnline(context.Background(), c.DeviceID, false)
	}
	h.mu.Unlock()
	h.log.Info("device websocket offline", "deviceId", c.DeviceID)
}
func (h *Hub) Send(deviceID string, frame models.CommandFrame) bool {
	h.mu.RLock()
	c := h.clients[deviceID]
	h.mu.RUnlock()
	if c == nil {
		return false
	}
	select {
	case c.Send <- frame:
		return true
	default:
		return false
	}
}
func (h *Hub) Online(deviceID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.clients[deviceID] != nil
}
func (h *Hub) BroadcastTelemetry(raw []byte) {
	select {
	case h.dashboard <- raw:
	default:
	}
}
