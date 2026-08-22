package channel

import (
	"context"
	"encoding/json"
	"time"

	ws "github.com/VinnsEdesigner/vyzorix/apps/api/internal/ws"
)

// HubBridge bridges the channel manager to the websocket hub: channel
// publications fan out to hub clients subscribed via SubscribeChannel.
type HubBridge struct {
	mgr *Manager
	hub *ws.Hub
}

// NewHubBridge wires the manager to the hub. The manager must already be
// wired with an authorizer; subscriptions also flow through the same gate.
func NewHubBridge(mgr *Manager, hub *ws.Hub) *HubBridge {
	return &HubBridge{mgr: mgr, hub: hub}
}

// Subscribe registers a hub client callback for a channel. Callback is the
// raw frame delivered on the websocket; operator authorization flows through
// the manager.
func (b *HubBridge) Subscribe(ctx context.Context, operatorID, channelAddr string, onMessage func(*Message)) (*Stream, error) {
	ch, err := Parse(channelAddr)
	if err != nil {
		return nil, err
	}
	stream, err := b.mgr.Subscribe(ctx, SubscribeEvent{SubjectID: operatorID, Channel: ch})
	if err != nil {
		return nil, err
	}
	stream.Subscribe(operatorID, onMessage)
	return stream, nil
}

// Unsubscribe detaches and closes streams with no remaining subscribers.
func (b *HubBridge) Unsubscribe(channelAddr, subscriberID string) {
	b.mgr.Unsubscribe(channelAddr, subscriberID)
}

// Publish routes a typed message both to the manager's stream and to hub
// clients that subscribed via the callback bridge.
func (b *HubBridge) Publish(channelAddr, evtType string, data map[string]interface{}) {
	msg := &Message{At: time.Now(), Channel: channelAddr, Type: evtType, Data: data}
	b.mgr.Publish(channelAddr, msg)
	b.forward(channelAddr, msg)
}

// forward dispatches the message to hub subscribers keyed by channel address.
func (b *HubBridge) forward(channelAddr string, msg *Message) {
	if b.hub == nil {
		return
	}
	frame, err := json.Marshal(map[string]interface{}{
		"type":    "channel_message",
		"channel": channelAddr,
		"message": msg.Type,
		"data":    msg.Data,
	})
	if err != nil {
		return
	}
	_ = b.hub.BroadcastEvent("channel:"+channelAddr, frame)
}

// Manager returns the underlying channel manager.
func (b *HubBridge) Manager() *Manager { return b.mgr }
