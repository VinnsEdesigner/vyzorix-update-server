// Package subscription provides GraphQL subscription support via WebSocket.
package subscription

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
	"github.com/gorilla/websocket"
)

// Client represents a subscription WebSocket client.
type Client struct {
	ctx       context.Context
	conn      *websocket.Conn
	operator  *operator.Operator
	orgID     string
	subs      map[string]func()
	done      chan struct{}
	handler   *Handler
	presenter *Presenter
	mu        sync.Mutex
}

// NewClient creates a new subscription client.
func NewClient(conn *websocket.Conn, op *operator.Operator, orgID string, handler *Handler, presenter *Presenter) *Client {
	return &Client{
		conn:      conn,
		operator:  op,
		orgID:     orgID,
		subs:      make(map[string]func()),
		done:      make(chan struct{}),
		ctx:       context.Background(),
		handler:   handler,
		presenter: presenter,
	}
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

// readPump handles incoming WebSocket messages.
func (c *Client) readPump() {
	defer func() {
		c.cleanup()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.presenter.LogMessageError(c.ctx, err)
			}

			return
		}

		c.handleMessage(message)
	}
}

// writePump handles outgoing WebSocket messages.
func (c *Client) writePump() {
	defer func() { _ = c.conn.Close() }()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.done:
			return
		case <-ticker.C:
			// Send periodic ping to keep connection alive.
			if err := c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
				return
			}

			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
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
		// Send connection acknowledgment.
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

	// Parse the subscription query to determine what to subscribe to.
	switch {
	case contains(sub.Query, "deviceUpdated"):
		c.subscribeDeviceUpdates(id, sub)
	case contains(sub.Query, "telemetryReceived"):
		c.subscribeTelemetry(id, sub)
	case contains(sub.Query, "commandStatusChanged"):
		c.subscribeCommandStatus(id, sub)
	case contains(sub.Query, "organizationEvent"):
		c.subscribeOrganizationEvent(id, sub)
	case contains(sub.Query, "memberEvent"):
		c.subscribeMemberEvent(id, sub)
	default:
		// For unknown subscriptions, just acknowledge.
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
	// Extract device ID from variables if present.
	deviceID, _ := sub.Variables["deviceId"].(string)

	// Check hub availability.
	if c.handler.hub == nil {
		c.sendError(id, "subscription service unavailable")
		return
	}

	// Subscribe to hub for device updates.
	unsubscribe := c.handler.hub.SubscribeDeviceUpdates(c.operator.ID, deviceID, func(data interface{}) error {
		c.sendMessage(wsMessage{
			Type:    "next",
			ID:      id,
			Payload: json.RawMessage(`{"data":{"deviceUpdated":` + mustMarshal(data) + `}}`),
		})
		return nil
	})

	c.mu.Lock()
	c.subs[id] = unsubscribe
	c.mu.Unlock()

	c.presenter.AuditSubscribe(c.ctx, c.operator, "device_updates", deviceID)

	// Send initial confirmation.
	c.sendMessage(wsMessage{Type: "next", ID: id, Payload: json.RawMessage(`{"data":{"deviceUpdated":null}}`)})
}

// subscribeTelemetry subscribes to real-time telemetry.
func (c *Client) subscribeTelemetry(id string, sub SubscribePayload) {
	deviceID, _ := sub.Variables["deviceId"].(string)

	// Check hub availability.
	if c.handler.hub == nil {
		c.sendError(id, "subscription service unavailable")
		return
	}

	unsubscribe := c.handler.hub.SubscribeTelemetry(c.operator.ID, deviceID, func(data interface{}) error {
		c.sendMessage(wsMessage{
			Type:    "next",
			ID:      id,
			Payload: json.RawMessage(`{"data":{"telemetryReceived":` + mustMarshal(data) + `}}`),
		})
		return nil
	})

	c.mu.Lock()
	c.subs[id] = unsubscribe
	c.mu.Unlock()

	c.presenter.AuditSubscribe(c.ctx, c.operator, "telemetry", deviceID)

	c.sendMessage(wsMessage{Type: "next", ID: id, Payload: json.RawMessage(`{"data":{"telemetryReceived":null}}`)})
}

// subscribeCommandStatus subscribes to command status changes.
func (c *Client) subscribeCommandStatus(id string, sub SubscribePayload) {
	dispatchID, _ := sub.Variables["dispatchId"].(string)

	// Check hub availability.
	if c.handler.hub == nil {
		c.sendError(id, "subscription service unavailable")
		return
	}

	unsubscribe := c.handler.hub.SubscribeCommandStatus(c.operator.ID, dispatchID, func(data interface{}) error {
		c.sendMessage(wsMessage{
			Type:    "next",
			ID:      id,
			Payload: json.RawMessage(`{"data":{"commandStatusChanged":` + mustMarshal(data) + `}}`),
		})
		return nil
	})

	c.mu.Lock()
	c.subs[id] = unsubscribe
	c.mu.Unlock()

	c.presenter.AuditSubscribe(c.ctx, c.operator, "command_status", dispatchID)

	c.sendMessage(wsMessage{Type: "next", ID: id, Payload: json.RawMessage(`{"data":{"commandStatusChanged":null}}`)})
}

// subscribeOrganizationEvent subscribes to organization events.
func (c *Client) subscribeOrganizationEvent(id string, sub SubscribePayload) {
	orgID, _ := sub.Variables["orgId"].(string)
	if orgID == "" {
		orgID = c.orgID // Use connection orgID as fallback.
	}

	// Check hub availability.
	if c.handler.hub == nil {
		c.sendError(id, "subscription service unavailable")
		return
	}

	unsubscribe := c.handler.hub.SubscribeOrganizationEvents(c.operator.ID, orgID, func(data interface{}) error {
		c.sendMessage(wsMessage{
			Type:    "next",
			ID:      id,
			Payload: json.RawMessage(`{"data":{"organizationEvent":` + mustMarshal(data) + `}}`),
		})
		return nil
	})

	c.mu.Lock()
	c.subs[id] = unsubscribe
	c.mu.Unlock()

	c.presenter.AuditSubscribe(c.ctx, c.operator, "organization_event", orgID)

	c.sendMessage(wsMessage{Type: "next", ID: id, Payload: json.RawMessage(`{"data":{"organizationEvent":null}}`)})
}

// subscribeMemberEvent subscribes to member events for an organization.
func (c *Client) subscribeMemberEvent(id string, sub SubscribePayload) {
	orgID, _ := sub.Variables["orgId"].(string)
	if orgID == "" {
		orgID = c.orgID // Use connection orgID as fallback.
	}

	// Check hub availability.
	if c.handler.hub == nil {
		c.sendError(id, "subscription service unavailable")
		return
	}

	unsubscribe := c.handler.hub.SubscribeMemberEvents(c.operator.ID, orgID, func(data interface{}) error {
		c.sendMessage(wsMessage{
			Type:    "next",
			ID:      id,
			Payload: json.RawMessage(`{"data":{"memberEvent":` + mustMarshal(data) + `}}`),
		})
		return nil
	})

	c.mu.Lock()
	c.subs[id] = unsubscribe
	c.mu.Unlock()

	c.presenter.AuditSubscribe(c.ctx, c.operator, "member_event", orgID)

	c.sendMessage(wsMessage{Type: "next", ID: id, Payload: json.RawMessage(`{"data":{"memberEvent":null}}`)})
}

// cleanup removes the client and unsubscribes from all subscriptions.
func (c *Client) cleanup() {
	c.mu.Lock()
	for id, unsubscribe := range c.subs {
		unsubscribe()
		delete(c.subs, id)
	}
	c.mu.Unlock()

	c.handler.removeClient(c.conn)

	_ = c.conn.Close()
	c.presenter.LogDisconnect(c.ctx, c.operator)
}

// sendMessage sends a message to the WebSocket client.
func (c *Client) sendMessage(msg wsMessage) {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := json.Marshal(msg)
	if err != nil {
		c.presenter.LogMessageError(c.ctx, err)
		return
	}

	if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
		c.presenter.LogMessageError(c.ctx, err)
	}
}

// sendError sends an error message to the client.
func (c *Client) sendError(id, message string) {
	c.sendMessage(wsMessage{
		Type:    "error",
		ID:      id,
		Payload: json.RawMessage(`{"message":"` + message + `"}`),
	})
}

// mustMarshal marshals data to JSON, panics on error.
func mustMarshal(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}

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
			// cs is uppercase, ct is lowercase - case fold.
		} else if ct >= 'A' && ct <= 'Z' && cs >= 'a' && cs <= 'z' {
			// ct is uppercase, cs is lowercase - case fold.
		} else {
			return false
		}
	}

	return true
}
