// Package hub provides WebSocket functionality.
package hub

import (
	"context"
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
	maxConcurrentCallbacks = 100              // Maximum concurrent callback goroutines.
	callbackTimeout        = 30 * time.Second // Timeout for callback execution.
	// droppedLogInterval is how often a saturated worker pool logs a drop, so a
	// flood is visible without one log line per event.
	droppedLogInterval = 100
)

// subscriptionWorker handles bounded callback execution.
type subscriptionWorker struct {
	sem     chan struct{}
	log     *slog.Logger
	wg      sync.WaitGroup
	dropped atomic.Int64
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
		if dropped%droppedLogInterval == 0 {
			w.log.Warn("subscription callback dropped (worker pool saturated)",
				"droppedCount", dropped,
				"maxConcurrent", maxConcurrentCallbacks,
			)
		}
	}
}

// SubscriptionManager manages real-time subscriptions for GraphQL subscriptions.
type SubscriptionManager struct {
	hub           *Hub
	deviceUpdates map[string][]callbackWrapper
	telemetry     map[string][]callbackWrapper
	commandStatus map[string][]callbackWrapper
	orgEvents     map[string][]callbackWrapper
	memberEvents  map[string][]callbackWrapper
	worker        *subscriptionWorker
	mu            sync.RWMutex
}

var (
	subscriptionMgr  *SubscriptionManager
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

// SubscribeDeviceUpdates subscribes to device update events for an organization.
// Keyed by organization (not operator): a device event is an org-level fact, so
// every operator in the org should observe it.
func (h *Hub) SubscribeDeviceUpdates(orgID, deviceID string, callback subscriptionCallback) func() {
	h.InitSubscriptions()

	subMgr := subscriptionMgr
	subMgr.mu.Lock()
	defer subMgr.mu.Unlock()

	key := orgID
	if deviceID != "" {
		key = orgID + ":" + deviceID
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
				subs = append(subs[:i], subs[i+1:]...)
				break
			}
		}
		// Drop the key once its last subscriber leaves; the subscription
		// manager is a process-lifetime singleton, so otherwise it accumulates
		// a key per device/org ever subscribed to.
		if len(subs) == 0 {
			delete(subMgr.deviceUpdates, key)
		} else {
			subMgr.deviceUpdates[key] = subs
		}
	}
}

// SubscribeTelemetry subscribes to real-time telemetry events for an organization.
// Keyed by organization so any operator in the org receives the org's telemetry.
func (h *Hub) SubscribeTelemetry(orgID, deviceID string, callback subscriptionCallback) func() {
	h.InitSubscriptions()

	subMgr := subscriptionMgr
	subMgr.mu.Lock()
	defer subMgr.mu.Unlock()

	// If deviceID is empty, subscribe to all devices in this organization.
	key := orgID
	if deviceID != "" {
		key = orgID + ":" + deviceID
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
				subs = append(subs[:i], subs[i+1:]...)
				break
			}
		}
		// Drop the key once its last subscriber leaves; the subscription
		// manager is a process-lifetime singleton, so otherwise it accumulates
		// a key per device/org ever subscribed to.
		if len(subs) == 0 {
			delete(subMgr.telemetry, key)
		} else {
			subMgr.telemetry[key] = subs
		}
	}
}

// SubscribeCommandStatus subscribes to command status change events for an organization.
// Keyed by organization so any operator in the org observes command progress.
func (h *Hub) SubscribeCommandStatus(orgID, dispatchID string, callback subscriptionCallback) func() {
	h.InitSubscriptions()

	subMgr := subscriptionMgr
	subMgr.mu.Lock()
	defer subMgr.mu.Unlock()

	key := orgID + ":" + dispatchID

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
				subs = append(subs[:i], subs[i+1:]...)
				break
			}
		}
		// Drop the key once its last subscriber leaves; the subscription
		// manager is a process-lifetime singleton, so otherwise it accumulates
		// a key per device/org ever subscribed to.
		if len(subs) == 0 {
			delete(subMgr.commandStatus, key)
		} else {
			subMgr.commandStatus[key] = subs
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
				subs = append(subs[:i], subs[i+1:]...)
				break
			}
		}
		// Drop the key once its last subscriber leaves; the subscription
		// manager is a process-lifetime singleton, so otherwise it accumulates
		// a key per device/org ever subscribed to.
		if len(subs) == 0 {
			delete(subMgr.orgEvents, key)
		} else {
			subMgr.orgEvents[key] = subs
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
				subs = append(subs[:i], subs[i+1:]...)
				break
			}
		}
		// Drop the key once its last subscriber leaves; the subscription
		// manager is a process-lifetime singleton, so otherwise it accumulates
		// a key per device/org ever subscribed to.
		if len(subs) == 0 {
			delete(subMgr.memberEvents, key)
		} else {
			subMgr.memberEvents[key] = subs
		}
	}
}

// PublishCommandStatusForDevice resolves the device's organization and publishes
// a command status change to that org's subscribers. It is a no-op when the org
// cannot be resolved, so callers do not need to guard.
func (h *Hub) PublishCommandStatusForDevice(ctx context.Context, dispatchID, deviceID, status string) {
	if dispatchID == "" {
		return
	}
	orgID := h.OrganizationForDevice(ctx, deviceID)
	if orgID == "" {
		return
	}

	h.PublishCommandStatus(orgID, dispatchID, map[string]any{
		"dispatchId": dispatchID,
		"deviceId":   deviceID,
		"status":     status,
		"timestamp":  time.Now().UnixMilli(),
	})
}

// OrganizationForDevice returns the organization ID a device belongs to, or ""
// if it cannot be resolved. Used by the realtime fan-out so device, telemetry
// and command events are delivered to every operator in the device's org.
func (h *Hub) OrganizationForDevice(ctx context.Context, deviceID string) string {
	if h.deviceRepo == nil || deviceID == "" {
		return ""
	}

	dev, err := h.deviceRepo.FindByIMEI(ctx, deviceID)
	if err != nil || dev == nil {
		return ""
	}

	return dev.OrganizationID
}

// PublishDeviceUpdate publishes a device update event.
func (h *Hub) PublishDeviceUpdate(orgID, deviceID string, data interface{}) {
	h.InitSubscriptions()

	subMgr := subscriptionMgr
	subMgr.mu.RLock()
	defer subMgr.mu.RUnlock()

	// Notify all subscriptions for this organization.
	for key, wrappers := range subMgr.deviceUpdates {
		if key == orgID || key == orgID+":"+deviceID {
			for _, w := range wrappers {

				subMgr.worker.execute(w.callback, data)
			}
		}
	}
}

// PublishTelemetry publishes a telemetry event.
func (h *Hub) PublishTelemetry(orgID, deviceID string, data interface{}) {
	h.InitSubscriptions()

	subMgr := subscriptionMgr
	subMgr.mu.RLock()
	defer subMgr.mu.RUnlock()

	key := orgID + ":" + deviceID
	if wrappers, ok := subMgr.telemetry[key]; ok {
		for _, w := range wrappers {

			subMgr.worker.execute(w.callback, data)
		}
	}

	// Also notify organization-wide subscriptions.
	if wrappers, ok := subMgr.telemetry[orgID]; ok {
		for _, w := range wrappers {

			subMgr.worker.execute(w.callback, data)
		}
	}
}

// PublishCommandStatus publishes a command status change event.
func (h *Hub) PublishCommandStatus(orgID, dispatchID string, data interface{}) {
	h.InitSubscriptions()

	subMgr := subscriptionMgr
	subMgr.mu.RLock()
	defer subMgr.mu.RUnlock()

	key := orgID + ":" + dispatchID
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
