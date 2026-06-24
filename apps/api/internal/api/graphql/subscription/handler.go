// Package subscription provides GraphQL subscription support via WebSocket.
package subscription

import (
	"context"
	"net/http"
	"sync"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/resolver"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/ws"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

// AuditLogger interface for audit logging.
type AuditLogger interface {
	LogAction(ctx context.Context, operatorID, action, resourceType, resourceID string)
}

// Config holds handler configuration.
type Config struct {
	Hub         *hub.Hub
	Resolver    *resolver.Resolver
	AuthMw      *middleware.AuthMiddleware
	Logger      Logger
	AuditLogger AuditLogger
}

// Handler manages WebSocket connections for GraphQL subscriptions.
type Handler struct {
	hub       *hub.Hub
	resolver  *resolver.Resolver
	authMw    *middleware.AuthMiddleware
	presenter *Presenter
	mu        sync.Mutex
	clients   map[*websocket.Conn]*Client
}

// NewHandler creates a new subscription handler.
func NewHandler(cfg *Config) *Handler {
	return &Handler{
		hub:       cfg.Hub,
		resolver:  cfg.Resolver,
		authMw:    cfg.AuthMw,
		presenter: NewPresenter(cfg.Logger, cfg.AuditLogger),
		clients:   make(map[*websocket.Conn]*Client),
	}
}

// removeClient removes a client from the handler's client map.
func (h *Handler) removeClient(conn *websocket.Conn) {
	h.mu.Lock()
	delete(h.clients, conn)
	h.mu.Unlock()
}

// HandleWebSocket upgrades HTTP to WebSocket and handles subscription connections.
func (h *Handler) HandleWebSocket(c *gin.Context) {
	headers := map[string]string{
		"Cookie": c.GetHeader("Cookie"),
	}

	op, err := h.authMw.Authenticate(c.Request.Context(), headers)
	if err != nil {
		h.presenter.LogAuthFail(c.Request.Context(), err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.presenter.LogWebSocketError(c.Request.Context(), "upgrade", err)
		return
	}

	client := NewClient(conn, op, h, h.presenter)

	h.mu.Lock()
	h.clients[conn] = client
	h.mu.Unlock()

	h.presenter.LogConnect(c.Request.Context(), op)

	go client.readPump()
	go client.writePump()
}
