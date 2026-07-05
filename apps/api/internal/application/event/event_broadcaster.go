// Package event provides event broadcasting for real-time dashboard updates.
package event

import (
	"encoding/json"
	"log/slog"
	"sync"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/event"
)

// DashboardEventMessage represents an event message sent to dashboard clients.
type DashboardEventMessage struct {
	Event *event.Event `json:"event"`
	Type  string       `json:"type"`
}

// DashboardClient represents a connected dashboard WebSocket client.
type DashboardClient struct {
	Send       chan []byte
	Subscribed map[string]bool
	ID         string
	OperatorID string
	mu         sync.RWMutex
}

// NewDashboardClient creates a new dashboard client.
func NewDashboardClient(id, operatorID string) *DashboardClient {
	return &DashboardClient{
		ID:         id,
		OperatorID: operatorID,
		Send:       make(chan []byte, 256),
		Subscribed: make(map[string]bool),
	}
}

// Subscribe adds a device subscription for this dashboard client.
func (c *DashboardClient) Subscribe(deviceID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Subscribed[deviceID] = true
}

// Unsubscribe removes a device subscription for this dashboard client.
func (c *DashboardClient) Unsubscribe(deviceID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.Subscribed, deviceID)
}

// IsSubscribed checks if this client is subscribed to a device.
func (c *DashboardClient) IsSubscribed(deviceID string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Subscribed[deviceID]
}

// GetSubscriptions returns a copy of subscribed device IDs.
func (c *DashboardClient) GetSubscriptions() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	subs := make([]string, 0, len(c.Subscribed))
	for deviceID := range c.Subscribed {
		subs = append(subs, deviceID)
	}
	return subs
}

// Broadcaster manages dashboard client connections and broadcasts events.
type Broadcaster struct {
	clients         map[string]*DashboardClient // clientID -> client
	operatorClients map[string][]string         // operatorID -> []clientID
	log             *slog.Logger
	mu              sync.RWMutex
}

// NewBroadcaster creates a new event broadcaster.
func NewBroadcaster(log *slog.Logger) *Broadcaster {
	return &Broadcaster{
		clients:         make(map[string]*DashboardClient),
		operatorClients: make(map[string][]string),
		log:             log,
	}
}

// Register adds a new dashboard client.
func (b *Broadcaster) Register(client *DashboardClient) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.clients[client.ID] = client

	// Track by operator
	b.operatorClients[client.OperatorID] = append(b.operatorClients[client.OperatorID], client.ID)

	b.log.Info("dashboard client registered", "clientId", client.ID, "operatorId", client.OperatorID)
}

// Unregister removes a dashboard client.
func (b *Broadcaster) Unregister(clientID string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	client, ok := b.clients[clientID]
	if !ok {
		return
	}

	// Remove from operator tracking
	if clientIDs, exists := b.operatorClients[client.OperatorID]; exists {
		newIDs := make([]string, 0, len(clientIDs))
		for _, id := range clientIDs {
			if id != clientID {
				newIDs = append(newIDs, id)
			}
		}
		if len(newIDs) == 0 {
			delete(b.operatorClients, client.OperatorID)
		} else {
			b.operatorClients[client.OperatorID] = newIDs
		}
	}

	delete(b.clients, clientID)
	b.log.Info("dashboard client unregistered", "clientId", clientID)
}

// BroadcastToDashboard sends an event to all dashboard clients subscribed to the device.
func (b *Broadcaster) BroadcastToDashboard(deviceID, operatorID, eventType string, data []byte) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	// Find all clients subscribed to this device
	sent := 0
	for _, client := range b.clients {
		if client.IsSubscribed(deviceID) {
			msg := DashboardEventMessage{
				Type:  eventType,
				Event: nil,
			}
			if data != nil {
				_ = json.Unmarshal(data, &msg.Event)
			}

			msgBytes, err := json.Marshal(msg)
			if err != nil {
				b.log.Warn("failed to marshal dashboard event", "clientId", client.ID, "err", err)
				continue
			}

			select {
			case client.Send <- msgBytes:
				sent++
			default:
				b.log.Warn("dashboard client buffer full, dropping event", "clientId", client.ID)
			}
		}
	}

	if sent > 0 {
		b.log.Debug("broadcast event to dashboards", "deviceId", deviceID, "sentTo", sent)
	}

	return nil
}

// BroadcastToOperator sends an event to all dashboard clients for an operator.
func (b *Broadcaster) BroadcastToOperator(operatorID string, eventType string, evt *event.Event) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	clientIDs, ok := b.operatorClients[operatorID]
	if !ok {
		return nil
	}

	msg := DashboardEventMessage{
		Type:  eventType,
		Event: evt,
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	sent := 0
	for _, clientID := range clientIDs {
		client, ok := b.clients[clientID]
		if !ok {
			continue
		}

		select {
		case client.Send <- msgBytes:
			sent++
		default:
			b.log.Warn("dashboard client buffer full, dropping event", "clientId", client.ID)
		}
	}

	if sent > 0 {
		b.log.Debug("broadcast event to operator", "operatorId", operatorID, "sentTo", sent)
	}

	return nil
}

// BroadcastToAll sends an event to all connected dashboard clients.
func (b *Broadcaster) BroadcastToAll(eventType string, evt *event.Event) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	msg := DashboardEventMessage{
		Type:  eventType,
		Event: evt,
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	for _, client := range b.clients {
		select {
		case client.Send <- msgBytes:
		default:
			b.log.Warn("dashboard client buffer full, dropping event", "clientId", client.ID)
		}
	}

	return nil
}

// GetClientCount returns the number of connected dashboard clients.
func (b *Broadcaster) GetClientCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.clients)
}

// GetOperatorCount returns the number of connected operators.
func (b *Broadcaster) GetOperatorCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.operatorClients)
}

// BroadcastDeviceEvent implements the hub.DashboardBroadcaster interface.
func (b *Broadcaster) BroadcastDeviceEvent(deviceID string, evt *event.Event) error {
	return b.BroadcastToDashboard(deviceID, evt.OperatorID, string(evt.Type), nil)
}

// BroadcastOperatorEvent implements the hub.DashboardBroadcaster interface.
func (b *Broadcaster) BroadcastOperatorEvent(deviceID string, operatorID string, evt *event.Event) error {
	return b.BroadcastToOperator(operatorID, string(evt.Type), evt)
}
