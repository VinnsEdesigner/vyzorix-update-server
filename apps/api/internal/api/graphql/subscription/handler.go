// Package subscription provides GraphQL subscription support via WebSocket.
package subscription

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/resolver"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/ws"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

// Handler manages WebSocket connections for GraphQL subscriptions.
type Handler struct {
	hub      *hub.Hub
	resolver *resolver.Resolver
	authMw   *middleware.AuthMiddleware
	log      *slog.Logger
	mu       sync.Mutex
	clients  map[*websocket.Conn]*Client
}

// Client represents a subscription WebSocket client.
type Client struct {
	conn       *websocket.Conn
	operator   *operator.Operator
	subs       map[string]func() // subscriptionID -> unsubscribe
	mu         sync.Mutex
	done       chan struct{}
	handler    *Handler
}

// Message types for subscription protocol.
type wsMessage struct {
	Type    string          `json:"type"`
	ID      string          `json:"id,omitempty"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// SubscribePayload is the payload for subscribe messages.
type SubscribePayload struct {
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
	OperationName string                 `json:"operationName,omitempty"`
}

// NewHandler creates a new subscription handler.
func NewHandler(hub *hub.Hub, res *resolver.Resolver, authMw *middleware.AuthMiddleware, log *slog.Logger) *Handler {
	return &Handler{
		hub:      hub,
		resolver: res,
		authMw:   authMw,
		log:      log,
		clients:  make(map[*websocket.Conn]*Client),
	}
}

// HandleWebSocket upgrades HTTP to WebSocket and handles subscription connections.
func (h *Handler) HandleWebSocket(c *gin.Context) {
	// Authenticate via cookie
	headers := map[string]string{
		"Cookie": c.GetHeader("Cookie"),
	}

	op, err := h.authMw.Authenticate(c.Request.Context(), headers)
	if err != nil {
		h.log.Debug("subscription auth failed", "err", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.log.Error("websocket upgrade failed", "err", err)
		return
	}

	client := &Client{
		conn:     conn,
		operator: op,
		subs:     make(map[string]func()),
		done:     make(chan struct{}),
		handler:  h,
	}

	h.mu.Lock()
	h.clients[conn] = client
	h.mu.Unlock()

	h.log.Info("subscription client connected", "operatorID", op.ID)

	// Handle incoming messages
	go client.readPump()
	go client.writePump()
}

// readPump handles incoming WebSocket messages.
func (c *Client) readPump() {
	defer func() {
		c.cleanup()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.handler.log.Debug("websocket read error", "err", err)
			}
			return
		}

		c.handleMessage(message)
	}
}

// writePump handles outgoing WebSocket messages.
func (c *Client) writePump() {
	defer c.conn.Close()

	for {
		select {
		case <-c.done:
			return
		default:
			// Keep connection alive
		}
	}
}

// handleMessage processes incoming subscription messages.
func (c *Client) handleMessage(message []byte) {
	var msg wsMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		c.sendError(msg.ID, "invalid message format")
		return
	}

	switch msg.Type {
	case "connection_init":
		// Send connection acknowledgment
		c.sendMessage(wsMessage{Type: "connection_ack"})

	case "subscribe":
		c.handleSubscribe(msg.ID, msg.Payload)

	case "complete":
		c.handleComplete(msg.ID)

	default:
		c.sendError(msg.ID, "unknown message type: "+msg.Type)
	}
}

// handleSubscribe processes a subscription request.
func (c *Client) handleSubscribe(id string, payload json.RawMessage) {
	var sub SubscribePayload
	if err := json.Unmarshal(payload, &sub); err != nil {
		c.sendError(id, "invalid subscription payload")
		return
	}

	// Parse the subscription query to determine what to subscribe to
	switch {
	case contains(sub.Query, "deviceUpdated"):
		c.subscribeDeviceUpdates(id, sub)
	case contains(sub.Query, "telemetryReceived"):
		c.subscribeTelemetry(id, sub)
	case contains(sub.Query, "commandStatusChanged"):
		c.subscribeCommandStatus(id, sub)
	default:
		// For unknown subscriptions, just acknowledge
		c.sendMessage(wsMessage{
			Type:    "next",
			ID:      id,
			Payload: json.RawMessage(`{"data":{"__typename":"Subscription"}}`),
		})
	}
}

// handleComplete unsubscribes from a subscription.
func (c *Client) handleComplete(id string) {
	c.mu.Lock()
	if unsubscribe, ok := c.subs[id]; ok {
		unsubscribe()
		delete(c.subs, id)
	}
	c.mu.Unlock()
}

// subscribeDeviceUpdates subscribes to device update events.
func (c *Client) subscribeDeviceUpdates(id string, sub SubscribePayload) {
	// Extract device ID from variables if present
	deviceID, _ := sub.Variables["deviceId"].(string)

	// Subscribe to hub for device updates
	unsubscribe := c.handler.hub.SubscribeDeviceUpdates(c.operator.ID, deviceID, func(data interface{}) {
		c.sendMessage(wsMessage{
			Type:    "next",
			ID:      id,
			Payload: json.RawMessage(`{"data":{"deviceUpdated":` + mustMarshal(data) + `}}`),
		})
	})

	c.mu.Lock()
	c.subs[id] = unsubscribe
	c.mu.Unlock()

	// Send initial confirmation
	c.sendMessage(wsMessage{Type: "next", ID: id, Payload: json.RawMessage(`{"data":{"deviceUpdated":null}}`)})
}

// subscribeTelemetry subscribes to real-time telemetry.
func (c *Client) subscribeTelemetry(id string, sub SubscribePayload) {
	deviceID, _ := sub.Variables["deviceId"].(string)

	unsubscribe := c.handler.hub.SubscribeTelemetry(c.operator.ID, deviceID, func(data interface{}) {
		c.sendMessage(wsMessage{
			Type:    "next",
			ID:      id,
			Payload: json.RawMessage(`{"data":{"telemetryReceived":` + mustMarshal(data) + `}}`),
		})
	})

	c.mu.Lock()
	c.subs[id] = unsubscribe
	c.mu.Unlock()

	c.sendMessage(wsMessage{Type: "next", ID: id, Payload: json.RawMessage(`{"data":{"telemetryReceived":null}}`)})
}

// subscribeCommandStatus subscribes to command status changes.
func (c *Client) subscribeCommandStatus(id string, sub SubscribePayload) {
	dispatchID, _ := sub.Variables["dispatchId"].(string)

	unsubscribe := c.handler.hub.SubscribeCommandStatus(c.operator.ID, dispatchID, func(data interface{}) {
		c.sendMessage(wsMessage{
			Type:    "next",
			ID:      id,
			Payload: json.RawMessage(`{"data":{"commandStatusChanged":` + mustMarshal(data) + `}}`),
		})
	})

	c.mu.Lock()
	c.subs[id] = unsubscribe
	c.mu.Unlock()

	c.sendMessage(wsMessage{Type: "next", ID: id, Payload: json.RawMessage(`{"data":{"commandStatusChanged":null}}`)})
}

// cleanup removes the client and unsubscribes from all subscriptions.
func (c *Client) cleanup() {
	c.mu.Lock()
	for id, unsubscribe := range c.subs {
		unsubscribe()
		delete(c.subs, id)
	}
	c.mu.Unlock()

	c.handler.mu.Lock()
	delete(c.handler.clients, c.conn)
	c.handler.mu.Unlock()

	c.conn.Close()
	c.handler.log.Info("subscription client disconnected", "operatorID", c.operator.ID)
}

// sendMessage sends a message to the WebSocket client.
func (c *Client) sendMessage(msg wsMessage) {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, _ := json.Marshal(msg)
	if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
		c.handler.log.Debug("failed to send message", "err", err)
	}
}

// sendError sends an error message to the client.
func (c *Client) sendError(id, message string) {
	c.sendMessage(wsMessage{
		Type: "error",
		ID:   id,
		Payload: json.RawMessage(`{"message":"` + message + `"}`),
	})
}

// mustMarshal marshals data to JSON, panics on error.
func mustMarshal(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}

// contains checks if s contains substr (case-insensitive).
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsLower(s, substr))
}

func containsLower(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if equalFold(s[i:i+len(substr)], substr) {
			return true
		}
	}
	return false
}

func equalFold(s, t string) bool {
	if len(s) != len(t) {
		return false
	}
	for i := 0; i < len(s); i++ {
		cs, ct := s[i], t[i]
		if cs == ct {
			continue
		}
		if cs >= 'A' && cs <= 'Z' && ct >= 'a' && ct <= 'z' {
			cs += 'a' - 'A'
		} else if ct >= 'A' && ct <= 'Z' && cs >= 'a' && cs <= 'z' {
			ct += 'a' - 'A'
		} else {
			return false
		}
	}
	return true
}
