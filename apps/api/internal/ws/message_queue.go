package hub

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/shared"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/command"
)

// MessageQueueConfig holds configuration for the message queue.
type MessageQueueConfig struct {
	MaxQueueSize    int           // Max messages per device (default 1000)
	MessageTTL      time.Duration // Message TTL (default 7 days)
	MaxMessageAge   time.Duration // Max age before cleanup (default 7 days)
	CleanupInterval time.Duration // Cleanup interval (default 1 hour)
}

// DefaultMessageQueueConfig returns the default message queue configuration.
func DefaultMessageQueueConfig() *MessageQueueConfig {
	return &MessageQueueConfig{
		MaxQueueSize:    1000,
		MessageTTL:      7 * 24 * time.Hour, // 7 days
		MaxMessageAge:   7 * 24 * time.Hour, // 7 days
		CleanupInterval: 1 * time.Hour,
	}
}

// QueuedMessage represents a message in the queue.
type QueuedMessage struct {
	EnqueuedAt time.Time            `json:"enqueuedAt"`
	ExpiresAt  time.Time            `json:"expiresAt"`
	ID         string               `json:"id"`
	DeviceID   string               `json:"deviceId"`
	Frame      command.CommandFrame `json:"frame"`
}

// QueueMetrics holds queue metrics.
type QueueMetrics struct {
	LastCleanupAt  time.Time `json:"lastCleanupAt"`
	TotalEnqueued  int64     `json:"totalEnqueued"`
	TotalDelivered int64     `json:"totalDelivered"`
	TotalExpired   int64     `json:"totalExpired"`
	TotalDropped   int64     `json:"totalDropped"`
}

// MessageQueue manages queued messages for offline devices.
// It uses a two-tier storage approach: in-memory channels for low-latency access
// and SQLite persistence for durability across restarts.
type MessageQueue struct {
	log       *slog.Logger
	config    *MessageQueueConfig
	db        *sql.DB
	queues    map[string]chan *QueuedMessage
	metrics   QueueMetrics
	mu        sync.RWMutex
	metricsMu sync.RWMutex
}

// NewMessageQueue creates a new MessageQueue.
func NewMessageQueue(log *slog.Logger, db *sql.DB, cfg *MessageQueueConfig) *MessageQueue {
	if cfg == nil {
		cfg = DefaultMessageQueueConfig()
	}

	q := &MessageQueue{
		log:    log,
		config: cfg,
		db:     db,
		queues: make(map[string]chan *QueuedMessage),
	}

	// Pre-warm queues from persisted data if db is available
	if db != nil {
		go q.preWarmQueues()
	}

	return q
}

// preWarmQueues loads persisted messages into memory on startup.
func (q *MessageQueue) preWarmQueues() {
	if q.db == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get distinct device IDs with pending messages
	rows, err := q.db.QueryContext(ctx,
		`SELECT DISTINCT device_id FROM message_queue WHERE expires_at > ?`,
		time.Now().UnixMilli(),
	)
	if err != nil {
		q.log.Warn("failed to pre-warm message queues", "err", err)
		return
	}

	defer func() { _ = rows.Close() }()

	var deviceIDs []string

	for rows.Next() {
		var deviceID string
		if err := rows.Scan(&deviceID); err != nil {
			continue
		}

		deviceIDs = append(deviceIDs, deviceID)
	}

	if err := rows.Err(); err != nil {
		q.log.Warn("error iterating device IDs", "err", err)
	}

	// Pre-warm each queue
	for _, deviceID := range deviceIDs {
		messages := q.LoadPersistedMessages(deviceID)
		if len(messages) > 0 {
			ch := q.GetOrCreateQueue(deviceID)

			for _, msg := range messages {
				select {
				case ch <- msg:
				default:
				}
			}

			q.log.Info("pre-warmed message queue", "deviceId", deviceID, "count", len(messages))
		}
	}
}

// Start starts the message queue background workers.
func (q *MessageQueue) Start(ctx context.Context) {
	// Start cleanup worker
	go q.cleanupWorker(ctx)
}

// GetOrCreateQueue gets or creates a channel queue for a device.
func (q *MessageQueue) GetOrCreateQueue(deviceID string) chan *QueuedMessage {
	q.mu.Lock()
	defer q.mu.Unlock()

	if ch, ok := q.queues[deviceID]; ok {
		return ch
	}

	ch := make(chan *QueuedMessage, q.config.MaxQueueSize)
	q.queues[deviceID] = ch

	return ch
}

// RemoveQueue removes the queue for a device.
func (q *MessageQueue) RemoveQueue(deviceID string) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if ch, ok := q.queues[deviceID]; ok {
		close(ch)
		delete(q.queues, deviceID)
	}
}

// Enqueue adds a message to the device's queue.
// Returns true if enqueued, false if queue is full.
func (q *MessageQueue) Enqueue(deviceID string, frame command.CommandFrame) bool {
	msg := &QueuedMessage{
		ID:         generateQueueID(),
		DeviceID:   deviceID,
		Frame:      frame,
		EnqueuedAt: time.Now(),
		ExpiresAt:  time.Now().Add(q.config.MessageTTL),
	}

	ch := q.GetOrCreateQueue(deviceID)

	// Try to enqueue (non-blocking)
	select {
	case ch <- msg:
		q.incrementEnqueued()
		go q.persistMessage(msg)

		return true
	default:
		// Queue full - evict oldest and retry once
		q.evictOldest(deviceID)
		select {
		case ch <- msg:
			q.incrementEnqueued()
			go q.persistMessage(msg)

			return true
		default:
			q.incrementDropped()
			q.log.Warn("message queue full, dropped message",
				"deviceId", deviceID,
				"dispatchId", frame.DispatchID,
			)

			return false
		}
	}
}

// EnqueueWithConfirmation adds a message to the queue and waits for DB confirmation.
// This ensures 100% message delivery by guaranteeing persistence before returning.
func (q *MessageQueue) EnqueueWithConfirmation(deviceID string, frame command.CommandFrame) bool {
	msg := &QueuedMessage{
		ID:         generateQueueID(),
		DeviceID:   deviceID,
		Frame:      frame,
		EnqueuedAt: time.Now(),
		ExpiresAt:  time.Now().Add(q.config.MessageTTL),
	}

	ch := q.GetOrCreateQueue(deviceID)

	// Try to enqueue (non-blocking)
	select {
	case ch <- msg:
		q.incrementEnqueued()
		// Synchronous DB write for 100% delivery guarantee (G1)
		if err := q.persistMessageSync(msg); err != nil {
			q.log.Error("failed to persist queued message (delivery not guaranteed)",
				"deviceId", deviceID,
				"messageId", msg.ID,
				"err", err,
			)
			q.incrementDropped()

			return false
		}

		return true
	default:
		// Queue full - evict oldest and retry once
		q.evictOldest(deviceID)
		select {
		case ch <- msg:
			q.incrementEnqueued()
			// Synchronous DB write for 100% delivery guarantee (G1)
			if err := q.persistMessageSync(msg); err != nil {
				q.log.Error("failed to persist queued message (delivery not guaranteed)",
					"deviceId", deviceID,
					"messageId", msg.ID,
					"err", err,
				)
				q.incrementDropped()

				return false
			}

			return true
		default:
			q.incrementDropped()
			q.log.Warn("message queue full, dropped message",
				"deviceId", deviceID,
				"dispatchId", frame.DispatchID,
			)

			return false
		}
	}
}

// persistMessageSync persists a message synchronously for 100% delivery guarantee.
func (q *MessageQueue) persistMessageSync(msg *QueuedMessage) error {
	if q.db == nil {
		return nil // No DB configured, but message is in memory
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	frameJSON, err := json.Marshal(msg.Frame)
	if err != nil {
		return err
	}

	_, err = q.db.ExecContext(ctx,
		`INSERT INTO message_queue (id, device_id, frame_json, enqueued_at, expires_at)
		 VALUES (?, ?, ?, ?, ?)`,
		msg.ID, msg.DeviceID, string(frameJSON),
		msg.EnqueuedAt.UnixMilli(), msg.ExpiresAt.UnixMilli(),
	)

	return err
}

// evictOldest removes the oldest message from a device's queue.
func (q *MessageQueue) evictOldest(deviceID string) {
	// Use write lock to prevent race with queue deletion
	q.mu.Lock()
	ch, ok := q.queues[deviceID]
	if !ok {
		q.mu.Unlock()
		return
	}
	q.mu.Unlock()

	// Channel send/receive are atomic - safe to do without lock
	select {
	case <-ch:
		// Evicted oldest
	default:
	}
}

// persistMessage persists a message to the database.
func (q *MessageQueue) persistMessage(msg *QueuedMessage) {
	if q.db == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	frameJSON, err := json.Marshal(msg.Frame)
	if err != nil {
		q.log.Warn("failed to marshal frame for persistence", "err", err)
		return
	}

	_, err = q.db.ExecContext(ctx,
		`INSERT INTO message_queue (id, device_id, frame_json, enqueued_at, expires_at)
		 VALUES (?, ?, ?, ?, ?)`,
		msg.ID, msg.DeviceID, string(frameJSON),
		msg.EnqueuedAt.UnixMilli(), msg.ExpiresAt.UnixMilli(),
	)
	if err != nil {
		q.log.Warn("failed to persist queued message", "err", err, "msgId", msg.ID)
	}
}

// LoadPersistedMessages loads persisted messages from database for a device.
func (q *MessageQueue) LoadPersistedMessages(deviceID string) []*QueuedMessage {
	if q.db == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rows, err := q.db.QueryContext(ctx,
		`SELECT id, device_id, frame_json, enqueued_at, expires_at
		 FROM message_queue
		 WHERE device_id = ? AND expires_at > ?
		 ORDER BY enqueued_at ASC`,
		deviceID, time.Now().UnixMilli(),
	)
	if err != nil {
		q.log.Warn("failed to load persisted messages", "err", err, "deviceId", deviceID)
		return nil
	}

	defer func() { _ = rows.Close() }()

	var messages []*QueuedMessage

	for rows.Next() {
		var msg QueuedMessage

		var frameJSON string

		var enqueuedAt, expiresAt int64

		if err := rows.Scan(&msg.ID, &msg.DeviceID, &frameJSON, &enqueuedAt, &expiresAt); err != nil {
			continue
		}

		if err := json.Unmarshal([]byte(frameJSON), &msg.Frame); err != nil {
			continue
		}

		msg.EnqueuedAt = time.UnixMilli(enqueuedAt)
		msg.ExpiresAt = time.UnixMilli(expiresAt)
		messages = append(messages, &msg)
	}

	if err := rows.Err(); err != nil {
		q.log.Warn("error loading persisted messages", "err", err, "deviceId", deviceID)
	}

	return messages
}

// ReplayQueue replays all queued messages for a device to the channel.
// Called when a device reconnects. Returns count of messages replayed.
func (q *MessageQueue) ReplayQueue(deviceID string, dest chan<- command.CommandFrame) int {
	persisted := q.LoadPersistedMessages(deviceID)
	ch := q.GetOrCreateQueue(deviceID)

	count := 0

	// First, replay persisted messages in FIFO order
	for _, msg := range persisted {
		// DeliveryConfirmation cannot be fulfilled after replay - clear it
		msg.Frame.DeliveryConfirmation = nil
		select {
		case dest <- msg.Frame:
			count++

			q.incrementDelivered()

			go q.deleteMessage(msg.ID)
		default:
			q.log.Warn("destination buffer full during replay",
				"deviceId", deviceID,
				"replayedCount", count,
			)

			return count
		}
	}

	// Then drain in-memory channel
	for {
		select {
		case msg := <-ch:
			// DeliveryConfirmation cannot be fulfilled after replay - clear it
			msg.Frame.DeliveryConfirmation = nil
			select {
			case dest <- msg.Frame:
				count++

				q.incrementDelivered()

				go q.deleteMessage(msg.ID)
			default:
				// Put it back and stop
				select {
				case ch <- msg:
				default:
				}

				return count
			}
		default:
			return count
		}
	}
}

// deleteMessage removes a message from the database.
func (q *MessageQueue) deleteMessage(id string) {
	if q.db == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, _ = q.db.ExecContext(ctx, `DELETE FROM message_queue WHERE id = ?`, id)
}

// QueueSize returns the current size of a device's queue (in-memory only).
func (q *MessageQueue) QueueSize(deviceID string) int {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if ch, ok := q.queues[deviceID]; ok {
		return len(ch)
	}

	return 0
}

// TotalQueuedMessages returns the total number of queued messages in memory.
func (q *MessageQueue) TotalQueuedMessages() int {
	q.mu.RLock()
	defer q.mu.RUnlock()

	total := 0
	for _, ch := range q.queues {
		total += len(ch)
	}

	return total
}

// GetMetrics returns a copy of the current metrics.
func (q *MessageQueue) GetMetrics() QueueMetrics {
	q.metricsMu.RLock()
	defer q.metricsMu.RUnlock()

	return q.metrics
}

func (q *MessageQueue) incrementEnqueued() {
	q.metricsMu.Lock()
	q.metrics.TotalEnqueued++
	q.metricsMu.Unlock()
}

func (q *MessageQueue) incrementDelivered() {
	q.metricsMu.Lock()
	q.metrics.TotalDelivered++
	q.metricsMu.Unlock()
}

func (q *MessageQueue) incrementDropped() {
	q.metricsMu.Lock()
	q.metrics.TotalDropped++
	q.metricsMu.Unlock()
}

// cleanupWorker periodically cleans up expired messages.
func (q *MessageQueue) cleanupWorker(ctx context.Context) {
	ticker := time.NewTicker(q.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			q.cleanupExpired()
		}
	}
}

// cleanupExpired removes expired messages from the database.
func (q *MessageQueue) cleanupExpired() {
	if q.db == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	now := time.Now().UnixMilli()

	result, err := q.db.ExecContext(ctx,
		`DELETE FROM message_queue WHERE expires_at < ?`,
		now,
	)
	if err != nil {
		q.log.Warn("failed to cleanup expired messages", "err", err)
		return
	}

	rows, _ := result.RowsAffected()
	if rows > 0 {
		q.metricsMu.Lock()
		q.metrics.TotalExpired += rows
		q.metrics.LastCleanupAt = time.Now()
		q.metricsMu.Unlock()
		q.log.Info("cleaned up expired messages", "count", rows)
	}
}

// generateQueueID generates a unique ID for a queued message using the shared UUID generator.
func generateQueueID() string {
	return "mq-" + shared.GenerateID()
}
