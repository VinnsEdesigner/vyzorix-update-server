// Package hub provides WebSocket functionality.
package hub

import (
	"sync"
)

// subscriptionCallback is a function that receives subscription data.
type subscriptionCallback func(interface{})

// callbackWrapper wraps a callback with an ID for tracking.
type callbackWrapper struct {
	callback subscriptionCallback
	id       int
}

var callbackIDCounter int
var callbackIDMu sync.Mutex

// nextCallbackID generates the next unique callback ID.
func nextCallbackID() int {
	callbackIDMu.Lock()
	defer callbackIDMu.Unlock()

	callbackIDCounter++

	return callbackIDCounter
}

// SubscriptionManager manages real-time subscriptions for GraphQL subscriptions.
type SubscriptionManager struct {
	hub           *Hub
	deviceUpdates map[string][]callbackWrapper
	telemetry     map[string][]callbackWrapper
	commandStatus map[string]subscriptionCallback
	mu            sync.RWMutex
}

var (
	subscriptionMgr *SubscriptionManager
	subscriptionOnce sync.Once
)

// InitSubscriptions initializes the subscription manager.
func (h *Hub) InitSubscriptions() {
	subscriptionOnce.Do(func() {
		subscriptionMgr = &SubscriptionManager{
			hub:           h,
			deviceUpdates: make(map[string][]callbackWrapper),
			telemetry:     make(map[string][]callbackWrapper),
			commandStatus: make(map[string]subscriptionCallback),
		}
	})
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

	wrapper := callbackWrapper{
		id:       nextCallbackID(),
		callback: callback,
	}
	subMgr.deviceUpdates[key] = append(subMgr.deviceUpdates[key], wrapper)

	return func() {
		subMgr.mu.Lock()
		defer subMgr.mu.Unlock()

		subs := subMgr.deviceUpdates[key]
		for i, w := range subs {
			if w.id == wrapper.id {
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

	wrapper := callbackWrapper{
		id:       nextCallbackID(),
		callback: callback,
	}
	subMgr.telemetry[key] = append(subMgr.telemetry[key], wrapper)

	return func() {
		subMgr.mu.Lock()
		defer subMgr.mu.Unlock()

		subs := subMgr.telemetry[key]
		for i, w := range subs {
			if w.id == wrapper.id {
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
	for key, wrappers := range subMgr.deviceUpdates {
		if key == operatorID || key == operatorID+":"+deviceID {
			for _, w := range wrappers {
				go w.callback(data)
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
	if wrappers, ok := subMgr.telemetry[key]; ok {
		for _, w := range wrappers {
			go w.callback(data)
		}
	}

	// Also notify operator-wide subscriptions
	if wrappers, ok := subMgr.telemetry[operatorID]; ok {
		for _, w := range wrappers {
			go w.callback(data)
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
