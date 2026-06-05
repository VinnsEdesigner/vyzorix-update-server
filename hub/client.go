package hub

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/VinnsEdesigner/vyzorix-update-server/models"
	"github.com/gorilla/websocket"
)

const (
	writeTimeout = 10 * time.Second
	pongWait     = 70 * time.Second
	pingPeriod   = 30 * time.Second
)

// Client represents a WebSocket client connection to a device.
type Client struct {
	DeviceID string
	Conn     *websocket.Conn
	Send     chan models.CommandFrame
	Hub      *Hub
}

// ReadPump pumps incoming messages from the WebSocket connection.
func (c *Client) ReadPump() {
	defer func() { c.Hub.Unregister(c); _ = c.Conn.Close() }()
	c.Conn.SetReadLimit(1 << 20) // 1MB
	_ = c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error { return c.Conn.SetReadDeadline(time.Now().Add(pongWait)) })
	for {
		_, raw, err := c.Conn.ReadMessage()
		if err != nil {
			return
		}
		var env struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(raw, &env); err != nil {
			if c.Hub.log != nil {
				c.Hub.log.Warn("bad ws frame", "deviceId", c.DeviceID, "err", err)
			}
			continue
		}
		if env.Type == "telemetry" {
			var t models.TelemetryFrame
			if err := json.Unmarshal(raw, &t); err == nil {
				t.Raw = raw
				if t.DeviceID == "" {
					t.DeviceID = c.DeviceID
				}
				if c.Hub.store != nil {
					if err := c.Hub.store.SaveTelemetry(context.Background(), c.DeviceID, raw, struct {
						RiskScore  int
						BufferLevel int
						ThermalTemp float64
					}{RiskScore: t.RiskScore, BufferLevel: t.BufferLevel, ThermalTemp: t.ThermalTemp}); err != nil {
						if c.Hub.log != nil {
							c.Hub.log.Warn("telemetry save failed", "deviceId", c.DeviceID, "err", err)
						}
					}
				}
				c.Hub.BroadcastTelemetry(raw)
			}
		} else {
			if c.Hub.store != nil {
				_ = c.Hub.store.Touch(context.Background(), c.DeviceID)
			}
		}
	}
}

// WritePump pumps outgoing messages to the WebSocket connection.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() { ticker.Stop(); _ = c.Conn.Close() }()
	for {
		select {
		case frame, ok := <-c.Send:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			if !ok {
				_ = c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.Conn.WriteJSON(frame); err != nil {
				return
			}
			if c.Hub.store != nil {
				_ = c.Hub.store.MarkDelivered(context.Background(), frame.DispatchID)
			}
		case <-ticker.C:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
