// Package hub provides WebSocket functionality.
package hub

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/command"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/telemetry"
)

const (
	writeTimeout = 10 * time.Second
	pongWait     = 70 * time.Second
	pingPeriod   = 30 * time.Second

	// Rate limit exceeded HTTP code.
	RateLimitExceeded = 429
)

// ClientMetrics holds reconnection metrics for a client.
type ClientMetrics struct {
	ConnectAttempts    int32 `json:"connectAttempts"`
	ConnectSuccesses   int32 `json:"connectSuccesses"`
	ConnectFailures    int32 `json:"connectFailures"`
	LastConnectedAt    int64 `json:"lastConnectedAt"`    // Unix timestamp.
	LastDisconnectedAt int64 `json:"lastDisconnectedAt"` // Unix timestamp.
	LastMessageAt      int64 `json:"lastMessageAt"`      // Unix timestamp of last message received.
	MessagesSent       int32 `json:"messagesSent"`
	MessagesReceived   int32 `json:"messagesReceived"`
	PongMissedCount    int32 `json:"pongMissedCount"`
	RateLimitedCount   int32 `json:"rateLimitedCount"`
	LastRateLimitedAt  int64 `json:"lastRateLimitedAt"` // Unix timestamp.
}

// Client represents a WebSocket client connection to a device.
type Client struct {
	Conn        *websocket.Conn
	Send        chan command.CommandFrame
	Hub         *Hub
	log         *slog.Logger
	Done        chan struct{}
	DeviceID    string
	ClientID    string
	metrics     ClientMetrics
	connectedAt int64
	isConnected atomic.Bool
}

// closeConn safely closes a websocket connection, logging any error.
func closeConn(conn *websocket.Conn, log *slog.Logger, ctx string) {
	if conn != nil {
		if err := conn.Close(); err != nil {
			if log != nil {
				log.Warn("websocket close failed", "context", ctx, "err", err)
			}
		}
	}
}

// GetMetrics returns a copy of the client metrics.

func (c *Client) GetMetrics() ClientMetrics {
	return ClientMetrics{
		ConnectAttempts:    atomic.LoadInt32(&c.metrics.ConnectAttempts),
		ConnectSuccesses:   atomic.LoadInt32(&c.metrics.ConnectSuccesses),
		ConnectFailures:    atomic.LoadInt32(&c.metrics.ConnectFailures),
		LastConnectedAt:    atomic.LoadInt64(&c.metrics.LastConnectedAt),
		LastDisconnectedAt: atomic.LoadInt64(&c.metrics.LastDisconnectedAt),
		LastMessageAt:      atomic.LoadInt64(&c.metrics.LastMessageAt),
		MessagesSent:       atomic.LoadInt32(&c.metrics.MessagesSent),
		MessagesReceived:   atomic.LoadInt32(&c.metrics.MessagesReceived),
		PongMissedCount:    atomic.LoadInt32(&c.metrics.PongMissedCount),
		RateLimitedCount:   atomic.LoadInt32(&c.metrics.RateLimitedCount),
		LastRateLimitedAt:  atomic.LoadInt64(&c.metrics.LastRateLimitedAt),
	}
}

// RecordConnectAttempt records a connection attempt.
func (c *Client) RecordConnectAttempt() {
	atomic.AddInt32(&c.metrics.ConnectAttempts, 1)
}

// RecordConnectSuccess records a successful connection.
func (c *Client) RecordConnectSuccess() {
	atomic.AddInt32(&c.metrics.ConnectSuccesses, 1)
	atomic.StoreInt64(&c.metrics.LastConnectedAt, time.Now().Unix())
	atomic.StoreInt64(&c.connectedAt, time.Now().Unix())
	c.isConnected.Store(true)
}

// RecordConnectFailure records a failed connection.
func (c *Client) RecordConnectFailure() {
	atomic.AddInt32(&c.metrics.ConnectFailures, 1)
}

// RecordDisconnect records a disconnection.
func (c *Client) RecordDisconnect() {
	atomic.StoreInt64(&c.metrics.LastDisconnectedAt, time.Now().Unix())
	c.isConnected.Store(false)
}

// RecordMessageSent records a message sent.
func (c *Client) RecordMessageSent() {
	atomic.AddInt32(&c.metrics.MessagesSent, 1)
}

// RecordMessageReceived records a message received and updates LastMessageAt.
func (c *Client) RecordMessageReceived() {
	atomic.AddInt32(&c.metrics.MessagesReceived, 1)
	atomic.StoreInt64(&c.metrics.LastMessageAt, time.Now().Unix())
}

// RecordPongMissed records a missed pong (connection issue).
func (c *Client) RecordPongMissed() {
	atomic.AddInt32(&c.metrics.PongMissedCount, 1)
}

// RecordRateLimited records a rate limiting event.
func (c *Client) RecordRateLimited() {
	atomic.AddInt32(&c.metrics.RateLimitedCount, 1)
	atomic.StoreInt64(&c.metrics.LastRateLimitedAt, time.Now().Unix())
}

// IsConnected returns whether the client is currently connected.
func (c *Client) IsConnected() bool {
	return c.isConnected.Load()
}

// Uptime returns the connection uptime in seconds.
func (c *Client) Uptime() int64 {
	connectedAt := atomic.LoadInt64(&c.connectedAt)
	if connectedAt == 0 {
		return 0
	}

	return time.Now().Unix() - connectedAt
}

// setReadDeadline safely sets read deadline, logging any error.
func setReadDeadline(conn *websocket.Conn, t time.Time, log *slog.Logger) {
	if err := conn.SetReadDeadline(t); err != nil {
		if log != nil {
			log.Warn("set read deadline failed", "err", err)
		}
	}
}

// setWriteDeadline safely sets write deadline, logging any error.
func setWriteDeadline(conn *websocket.Conn, t time.Time, log *slog.Logger) {
	if err := conn.SetWriteDeadline(t); err != nil {
		if log != nil {
			log.Warn("set write deadline failed", "err", err)
		}
	}
}

// ReadPump pumps incoming messages from the WebSocket connection.
func (c *Client) ReadPump() {
	defer func() {
		c.RecordDisconnect()
		c.Hub.Unregister(c)
		closeConn(c.Conn, c.log, "readPump")
	}()

	// Mark as connected.
	c.RecordConnectSuccess()

	c.Conn.SetReadLimit(1 << 20) // 1MB.
	setReadDeadline(c.Conn, time.Now().Add(pongWait), c.log)
	c.Conn.SetPongHandler(func(string) error {
		setReadDeadline(c.Conn, time.Now().Add(pongWait), c.log)
		return nil
	})

	for {
		_, raw, err := c.Conn.ReadMessage()
		if err != nil {
			c.log.Debug("ws read error", "deviceId", c.DeviceID, "err", err)
			return
		}

		c.RecordMessageReceived()

		var env struct {
			Type string `json:"type"`
		}

		if err := json.Unmarshal(raw, &env); err != nil {
			if c.log != nil {
				c.log.Warn("bad ws frame", "deviceId", c.DeviceID, "err", err)
			}

			continue
		}

		if env.Type == "telemetry" {
			c.processTelemetry(raw)
		}
	}
}

func (c *Client) processTelemetry(raw []byte) {

	defer func() {
		if r := recover(); r != nil {
			if c.log != nil {
				c.log.Error("panic in processTelemetry recovered", "deviceId", c.DeviceID, "panic", r)
			}
		}
	}()

	var t telemetry.TelemetryFrame
	if err := json.Unmarshal(raw, &t); err != nil {
		return
	}
	t.Raw = raw
	if t.DeviceID == "" {
		t.DeviceID = c.DeviceID
	}
	if c.Hub.telemetryRepo != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := c.Hub.telemetryRepo.Save(ctx, c.DeviceID, raw, t); err != nil {
			if c.log != nil {
				c.log.Warn("telemetry save failed", "deviceId", c.DeviceID, "err", err)
			}
		}
	}
	c.Hub.BroadcastTelemetry(raw)

	// Process telemetry for threshold breach events.
	if c.Hub.eventProcessor != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		telemetryData := map[string]any{
			"riskScore":   t.RiskScore,
			"thermalTemp": t.ThermalTemp,
			"bufferLevel": t.BufferLevel,
		}
		if err := c.Hub.eventProcessor.ProcessTelemetry(ctx, c.DeviceID, telemetryData); err != nil {
			if c.log != nil {
				c.log.Warn("telemetry event processing failed", "deviceId", c.DeviceID, "err", err)
			}
		}
	}
}

func (c *Client) writeCompressed(frame command.CommandFrame) error {
	data, err := json.Marshal(frame)
	if err != nil {
		return c.Conn.WriteJSON(frame)
	}
	compressed, didCompress, _ := c.Hub.compression.CompressMessage(data)
	if didCompress {
		c.Hub.compression.RecordCompression(len(data), len(compressed))
		compFrame := CompressedFrame{
			Type:         frame.Type,
			Compressed:   true,
			OriginalSize: len(data),
			Data:         compressed,
		}
		return c.Conn.WriteJSON(compFrame)
	}
	return c.Conn.WriteJSON(frame)
}

// WritePump pumps outgoing messages to the WebSocket connection.
// Applies rate limiting and compression as configured.
// Non-telemetry messages are handled by hub registration/onDisconnect.

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		closeConn(c.Conn, c.log, "writePump")

		close(c.Done)
	}()

	for {
		select {
		case frame, ok := <-c.Send:
			setWriteDeadline(c.Conn, time.Now().Add(writeTimeout), c.log)

			if !ok {
				if err := c.Conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					if c.log != nil {
						c.log.Warn("close message failed", "err", err)
					}
				}

				return
			}

			// Apply rate limiting before sending.
			if c.Hub.rateLimiter != nil {
				if !c.Hub.rateLimiter.Allow(c.DeviceID) {
					c.RecordRateLimited()
					c.log.Debug("client rate limited, dropping message",
						"deviceId", c.DeviceID,
						"dispatchId", frame.DispatchID,
					)

					continue // Skip this message but don't close connection.
				}
			}

			// Compress message if needed and configured (G4: 60% bandwidth reduction).
			if c.Hub.compression != nil {
				if err := c.writeCompressed(frame); err != nil {
					return
				}
			} else if err := c.Conn.WriteJSON(frame); err != nil {
				return
			}

			c.RecordMessageSent()

			// Send delivery confirmation if requested (G1: 100% delivery guarantee).
			// Set to nil after close to prevent double-close panic.
			if frame.DeliveryConfirmation != nil {
				frame.DeliveryConfirmation <- true
				close(frame.DeliveryConfirmation)
				frame.DeliveryConfirmation = nil
			}
		case <-ticker.C:
			setWriteDeadline(c.Conn, time.Now().Add(writeTimeout), c.log)

			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
