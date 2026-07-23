// Package subscription provides GraphQL subscription support via WebSocket.
package subscription

import (
	"context"
	"net/http"
	"strings"
	"sync"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/resolver"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/config"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/ws"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// AllowedOrigins returns origin validation function based on config.
func allowedOrigins(cfg config.Config) func(r *http.Request) bool {
	origins := cfg.AllowedOrigins
	return func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true // Allow empty origin (e.g., same-origin requests).
		}
		// Check against allowed origins.
		for _, allowed := range origins {
			if allowed == "*" || allowed == origin {
				return true
			}
			// Support wildcard subdomains (e.g., *.example.com).
			if strings.HasPrefix(allowed, "*.") {
				domain := allowed[2:]
				if strings.HasSuffix(origin, domain) {
					return true
				}
			}
		}
		return false
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     nil, // Set during initialization.
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
	Config      config.Config
}

// Handler manages WebSocket connections for GraphQL subscriptions.
type Handler struct {
	hub       *hub.Hub
	resolver  *resolver.Resolver
	authMw    *middleware.AuthMiddleware
	presenter *Presenter
	clients   map[*websocket.Conn]*Client
	mu        sync.Mutex
}

// NewHandler creates a new subscription handler.
func NewHandler(cfg *Config) *Handler {
	// Set origin check function based on config.
	upgrader.CheckOrigin = allowedOrigins(cfg.Config)

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
	// Extract org from URL parameter.
	orgID := c.Param("org")
	if orgID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "organization ID required"})
		return
	}

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

	client := NewClient(conn, op, orgID, h, h.presenter)

	h.mu.Lock()
	h.clients[conn] = client
	h.mu.Unlock()

	h.presenter.LogConnect(c.Request.Context(), op)

	go client.readPump()
	go client.writePump()
}
