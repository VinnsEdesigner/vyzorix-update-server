package hub

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/models"
	"github.com/gorilla/websocket"
)

const (
	writeTimeout = 10 * time.Second
	pongWait     = 70 * time.Second
	pingPeriod   = 30 * time.Second
)

// Client represents a WebSocket client connection to a device.
type Client struct {
	Conn     *websocket.Conn
	Send     chan models.CommandFrame
	Hub      *Hub
	log      *slog.Logger
	DeviceID string
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
		c.Hub.Unregister(c)
		closeConn(c.Conn, c.log, "readPump")
	}()
	c.Conn.SetReadLimit(1 << 20) // 1MB
	setReadDeadline(c.Conn, time.Now().Add(pongWait), c.log)
	c.Conn.SetPongHandler(func(string) error {
		setReadDeadline(c.Conn, time.Now().Add(pongWait), c.log)
		return nil
	})
	for {
		_, raw, err := c.Conn.ReadMessage()
		if err != nil {
			return
		}
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
			var t models.TelemetryFrame
			if err := json.Unmarshal(raw, &t); err == nil {
				t.Raw = raw
				if t.DeviceID == "" {
					t.DeviceID = c.DeviceID
				}
				if c.Hub.store != nil {
					if err := c.Hub.store.SaveTelemetry(context.Background(), c.DeviceID, raw, t); err != nil {
						if c.log != nil {
							c.log.Warn("telemetry save failed", "deviceId", c.DeviceID, "err", err)
						}
					}
				}
				c.Hub.BroadcastTelemetry(raw)
			}
		} else {
			if c.Hub.store != nil {
				if err := c.Hub.store.Touch(context.Background(), c.DeviceID); err != nil {
					if c.log != nil {
						c.log.Warn("touch failed", "deviceId", c.DeviceID, "err", err)
					}
				}
			}
		}
	}
}

// WritePump pumps outgoing messages to the WebSocket connection.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		closeConn(c.Conn, c.log, "writePump")
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
			if err := c.Conn.WriteJSON(frame); err != nil {
				return
			}
			if c.Hub.store != nil {
				if err := c.Hub.store.MarkDelivered(context.Background(), frame.DispatchID); err != nil {
					if c.log != nil {
						c.log.Warn("mark delivered failed", "dispatchId", frame.DispatchID, "err", err)
					}
				}
			}
		case <-ticker.C:
			setWriteDeadline(c.Conn, time.Now().Add(writeTimeout), c.log)
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
