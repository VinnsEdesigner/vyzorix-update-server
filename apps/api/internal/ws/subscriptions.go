// Package hub provides WebSocket functionality.
package hub

import (
	"log/slog"
	"sync"
	"sync/atomic"
	"time"
)

// subscriptionCallback is a function that receives subscription data and returns an error.

type subscriptionCallback func(interface{}) error

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

const (
	maxConcurrentCallbacks = 100      // Maximum concurrent callback goroutines.
	callbackTimeout        = 30 * time.Second // Timeout for callback execution.
	droppedCounterMax      = 1000      // Max dropped events before logging.
)

// subscriptionWorker handles bounded callback execution.
type subscriptionWorker struct {
	sem     chan struct{}   // Semaphore for limiting concurrency.
	dropped atomic.Int64    // Counter for dropped events.
	log     *slog.Logger
	wg      sync.WaitGroup
}

// newSubscriptionWorker creates a new subscription worker pool.
func newSubscriptionWorker(log *slog.Logger) *subscriptionWorker {
	return &subscriptionWorker{
		sem: make(chan struct{}, maxConcurrentCallbacks),
		log: log,
	}
}

// execute runs the callback with bounded concurrency and timeout.
func (w *subscriptionWorker) execute(callback subscriptionCallback, data interface{}) {
	select {
	case w.sem <- struct{}{}:

		w.wg.Add(1)
		go func() {
			defer w.wg.Done()
			defer func() { <-w.sem }() // Release semaphore.

			// Execute with panic recovery.
			done := make(chan error, 1)
			go func() {
				defer func() {
					if r := recover(); r != nil {
						w.log.Error("subscription callback panicked", "panic", r)
					}
				}()
				done <- callback(data)
			}()

			select {
			case err := <-done:
				
				if err != nil {
					w.log.Warn("subscription callback returned error", "error", err)
				}
				return
			case <-time.After(callbackTimeout):
				w.log.Warn("subscription callback timed out", "timeout", callbackTimeout)
			}
		}()
	default:
		// Semaphore full - drop the event.
		dropped := w.dropped.Add(1)
		if dropped%100 == 0 {
			w.log.Warn("subscription callback dropped (worker pool saturated)",
				"droppedCount", dropped,
				"maxConcurrent", maxConcurrentCallbacks,
			)
		}
	}
}

// SubscriptionManager manages real-time subscriptions for GraphQL subscriptions.
type SubscriptionManager struct {
	hub             *Hub
	deviceUpdates   map[string][]callbackWrapper
	telemetry       map[string][]callbackWrapper
	commandStatus   map[string][]callbackWrapper
	orgEvents       map[string][]callbackWrapper
	memberEvents    map[string][]callbackWrapper
	mu              sync.RWMutex
	worker          *subscriptionWorker
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
			commandStatus: make(map[string][]callbackWrapper),
			orgEvents:     make(map[string][]callbackWrapper),
			memberEvents:  make(map[string][]callbackWrapper),
			
			worker: newSubscriptionWorker(h.log),
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

	// If deviceID is empty, subscribe to all devices for this operator.
	key := operatorID
	if deviceID != "" {
		key = operatorID + ":" + deviceID
	}

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

	wrapper := callbackWrapper{
		id:       nextCallbackID(),
		callback: callback,
	}
	subMgr.commandStatus[key] = append(subMgr.commandStatus[key], wrapper)

	return func() {
		subMgr.mu.Lock()
		defer subMgr.mu.Unlock()

		subs := subMgr.commandStatus[key]
		for i, w := range subs {
			if w.id == wrapper.id {
				subMgr.commandStatus[key] = append(subs[:i], subs[i+1:]...)
				break
			}
		}
	}
}

// SubscribeOrganizationEvents subscribes to organization events.
func (h *Hub) SubscribeOrganizationEvents(operatorID, orgID string, callback subscriptionCallback) func() {
	h.InitSubscriptions()

	subMgr := subscriptionMgr
	subMgr.mu.Lock()
	defer subMgr.mu.Unlock()

	key := orgID

	wrapper := callbackWrapper{
		id:       nextCallbackID(),
		callback: callback,
	}
	subMgr.orgEvents[key] = append(subMgr.orgEvents[key], wrapper)

	return func() {
		subMgr.mu.Lock()
		defer subMgr.mu.Unlock()

		subs := subMgr.orgEvents[key]
		for i, w := range subs {
			if w.id == wrapper.id {
				subMgr.orgEvents[key] = append(subs[:i], subs[i+1:]...)
				break
			}
		}
	}
}

// SubscribeMemberEvents subscribes to member events for an organization.
func (h *Hub) SubscribeMemberEvents(operatorID, orgID string, callback subscriptionCallback) func() {
	h.InitSubscriptions()

	subMgr := subscriptionMgr
	subMgr.mu.Lock()
	defer subMgr.mu.Unlock()

	key := orgID

	wrapper := callbackWrapper{
		id:       nextCallbackID(),
		callback: callback,
	}
	subMgr.memberEvents[key] = append(subMgr.memberEvents[key], wrapper)

	return func() {
		subMgr.mu.Lock()
		defer subMgr.mu.Unlock()

		subs := subMgr.memberEvents[key]
		for i, w := range subs {
			if w.id == wrapper.id {
				subMgr.memberEvents[key] = append(subs[:i], subs[i+1:]...)
				break
			}
		}
	}
}

// PublishDeviceUpdate publishes a device update event.
func (h *Hub) PublishDeviceUpdate(operatorID, deviceID string, data interface{}) {
	h.InitSubscriptions()

	subMgr := subscriptionMgr
	subMgr.mu.RLock()
	defer subMgr.mu.RUnlock()

	// Notify all subscriptions for this operator.
	for key, wrappers := range subMgr.deviceUpdates {
		if key == operatorID || key == operatorID+":"+deviceID {
			for _, w := range wrappers {
				
				subMgr.worker.execute(w.callback, data)
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
			
			subMgr.worker.execute(w.callback, data)
		}
	}

	// Also notify operator-wide subscriptions.
	if wrappers, ok := subMgr.telemetry[operatorID]; ok {
		for _, w := range wrappers {
			
			subMgr.worker.execute(w.callback, data)
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
	if wrappers, ok := subMgr.commandStatus[key]; ok {
		for _, w := range wrappers {
			
			subMgr.worker.execute(w.callback, data)
		}
	}
}

// PublishOrganizationEvent publishes an organization event.
func (h *Hub) PublishOrganizationEvent(orgID string, data interface{}) {
	h.InitSubscriptions()

	subMgr := subscriptionMgr
	subMgr.mu.RLock()
	defer subMgr.mu.RUnlock()

	if wrappers, ok := subMgr.orgEvents[orgID]; ok {
		for _, w := range wrappers {
			
			subMgr.worker.execute(w.callback, data)
		}
	}
}

// PublishMemberEvent publishes a member event.
func (h *Hub) PublishMemberEvent(orgID string, data interface{}) {
	h.InitSubscriptions()

	subMgr := subscriptionMgr
	subMgr.mu.RLock()
	defer subMgr.mu.RUnlock()

	if wrappers, ok := subMgr.memberEvents[orgID]; ok {
		for _, w := range wrappers {
			
			subMgr.worker.execute(w.callback, data)
		}
	}
}
