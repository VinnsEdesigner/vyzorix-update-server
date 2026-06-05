package hub

import (
	"context"
	"log/slog"
	"sync"

	"github.com/VinnsEdesigner/vyzorix-update-server/models"
	"github.com/VinnsEdesigner/vyzorix-update-server/storage"
)

// Hub manages WebSocket connections and routes messages between devices and dashboard.
type Hub struct {
	log       *slog.Logger
	store     *storage.Store
	mu        sync.RWMutex
	clients   map[string]*Client
	register  chan *Client
	unreg     chan *Client
	broadcast chan []byte
}

// New creates a new Hub instance.
func New(log *slog.Logger, st *storage.Store) *Hub {
	return &Hub{
		log:       log,
		store:     st,
		clients:   make(map[string]*Client),
		register:  make(chan *Client),
		unreg:     make(chan *Client),
		broadcast: make(chan []byte, 256),
	}
}

// Run starts the hub's event loop in a background goroutine.
// It handles client registration, unregistration, and telemetry broadcasting.
func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case c := <-h.register:
			h.mu.Lock()
			if old := h.clients[c.DeviceID]; old != nil {
				close(old.Send)
				_ = old.Conn.Close()
			}
			c.log = h.log
			h.clients[c.DeviceID] = c
			h.mu.Unlock()
			_ = h.store.SetOnline(context.Background(), c.DeviceID, true)
			h.log.Info("device websocket online", "deviceId", c.DeviceID)
		case c := <-h.unreg:
			h.mu.Lock()
			if h.clients[c.DeviceID] == c {
				delete(h.clients, c.DeviceID)
				close(c.Send)
				_ = h.store.SetOnline(context.Background(), c.DeviceID, false)
			}
			h.mu.Unlock()
			h.log.Info("device websocket offline", "deviceId", c.DeviceID)
		case raw := <-h.broadcast:
			h.mu.RLock()
			for _, c := range h.clients {
				select {
				case c.Send <- models.CommandFrame{Type: "broadcast", Args: raw}:
				default:
				}
			}
			h.mu.RUnlock()
			_ = raw
		}
	}
}

// Register adds a client to the hub.
func (h *Hub) Register(c *Client) { h.register <- c }

// Unregister removes a client from the hub.
func (h *Hub) Unregister(c *Client) { h.unreg <- c }

// Send delivers a command frame to a specific device.
// Returns true if the device is connected and frame was queued.
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

// Online checks if a device is currently connected.
func (h *Hub) Online(deviceID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.clients[deviceID] != nil
}

// Clients returns a copy of the current clients map.
func (h *Hub) Clients() map[string]*Client {
	h.mu.RLock()
	defer h.mu.RUnlock()
	out := make(map[string]*Client, len(h.clients))
	for k, v := range h.clients {
		out[k] = v
	}
	return out
}

// GetClient retrieves a specific client by device ID.
func (h *Hub) GetClient(deviceID string) *Client {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.clients[deviceID]
}

// ClientCount returns the number of currently connected device clients.
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// Store returns the data store.
func (h *Hub) Store() *storage.Store {
	return h.store
}

// BroadcastTelemetry sends telemetry data to connected clients.
func (h *Hub) BroadcastTelemetry(raw []byte) {
	select {
	case h.broadcast <- raw:
	default:
	}
}
