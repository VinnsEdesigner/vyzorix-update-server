// Package hub provides WebSocket functionality.
package hub

import (
	"sync"
)

// subscriptionCallback is a function that receives subscription data.
type subscriptionCallback func(interface{})

// SubscriptionManager manages real-time subscriptions for GraphQL subscriptions.
type SubscriptionManager struct {
	hub            *Hub
	deviceUpdates  map[string][]subscriptionCallback
	telemetry      map[string][]subscriptionCallback
	commandStatus  map[string]subscriptionCallback
	mu             sync.RWMutex
}

var subscriptionMgr *SubscriptionManager

// InitSubscriptions initializes the subscription manager.
func (h *Hub) InitSubscriptions() {
	if subscriptionMgr == nil {
		subscriptionMgr = &SubscriptionManager{
			hub:           h,
			deviceUpdates: make(map[string][]subscriptionCallback),
			telemetry:     make(map[string][]subscriptionCallback),
			commandStatus: make(map[string]subscriptionCallback),
		}
	}
}

// SubscribeDeviceUpdates subscribes to device update events.
func (h *Hub) SubscribeDeviceUpdates(operatorID, deviceID string, callback subscriptionCallback) func() {
	h.InitSubscriptions()

	subMgr := subscriptionMgr
	subMgr.mu.Lock()
	defer subMgr.mu.Unlock()

	key := operatorID
	if deviceID != "" {
		key = operatorID + ":" + deviceID
	}

	subMgr.deviceUpdates[key] = append(subMgr.deviceUpdates[key], callback)

	return func() {
		subMgr.mu.Lock()
		defer subMgr.mu.Unlock()
		subs := subMgr.deviceUpdates[key]
		for i, cb := range subs {
			if cb == callback {
				subMgr.deviceUpdates[key] = append(subs[:i], subs[i+1:]...)
				break
			}
		}
	}
}

// SubscribeTelemetry subscribes to real-time telemetry events.
func (h *Hub) SubscribeTelemetry(operatorID, deviceID string, callback subscriptionCallback) func() {
	h.InitSubscriptions()

	subMgr := subscriptionMgr
	subMgr.mu.Lock()
	defer subMgr.mu.Unlock()

	key := operatorID + ":" + deviceID

	subMgr.telemetry[key] = append(subMgr.telemetry[key], callback)

	return func() {
		subMgr.mu.Lock()
		defer subMgr.mu.Unlock()
		subs := subMgr.telemetry[key]
		for i, cb := range subs {
			if cb == callback {
				subMgr.telemetry[key] = append(subs[:i], subs[i+1:]...)
				break
			}
		}
	}
}

// SubscribeCommandStatus subscribes to command status change events.
func (h *Hub) SubscribeCommandStatus(operatorID, dispatchID string, callback subscriptionCallback) func() {
	h.InitSubscriptions()

	subMgr := subscriptionMgr
	subMgr.mu.Lock()
	defer subMgr.mu.Unlock()

	key := operatorID + ":" + dispatchID

	subMgr.commandStatus[key] = callback

	return func() {
		subMgr.mu.Lock()
		defer subMgr.mu.Unlock()
		delete(subMgr.commandStatus, key)
	}
}

// PublishDeviceUpdate publishes a device update event.
func (h *Hub) PublishDeviceUpdate(operatorID, deviceID string, data interface{}) {
	h.InitSubscriptions()

	subMgr := subscriptionMgr
	subMgr.mu.RLock()
	defer subMgr.mu.RUnlock()

	// Notify all subscriptions for this operator
	for key, cbs := range subMgr.deviceUpdates {
		if key == operatorID || key == operatorID+":"+deviceID {
			for _, cb := range cbs {
				go cb(data)
			}
		}
	}
}

// PublishTelemetry publishes a telemetry event.
func (h *Hub) PublishTelemetry(operatorID, deviceID string, data interface{}) {
	h.InitSubscriptions()

	subMgr := subscriptionMgr
	subMgr.mu.RLock()
	defer subMgr.mu.RUnlock()

	key := operatorID + ":" + deviceID
	if cbs, ok := subMgr.telemetry[key]; ok {
		for _, cb := range cbs {
			go cb(data)
		}
	}

	// Also notify operator-wide subscriptions
	if cbs, ok := subMgr.telemetry[operatorID]; ok {
		for _, cb := range cbs {
			go cb(data)
		}
	}
}

// PublishCommandStatus publishes a command status change event.
func (h *Hub) PublishCommandStatus(operatorID, dispatchID string, data interface{}) {
	h.InitSubscriptions()

	subMgr := subscriptionMgr
	subMgr.mu.RLock()
	defer subMgr.mu.RUnlock()

	key := operatorID + ":" + dispatchID
	if cb, ok := subMgr.commandStatus[key]; ok {
		go cb(data)
	}
}
