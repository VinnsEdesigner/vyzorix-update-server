package auth

import (

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/adapters/response"
	appauth "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	infraauth "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security"

	"github.com/gin-gonic/gin"
)

// SessionsHandler handles session management endpoints.
type SessionsHandler struct {
	authService    *appauth.AuthService
	sessionManager *infraauth.SessionManager
	presenter      *response.Presenter
}

// NewSessionsHandler creates a new SessionsHandler.
func NewSessionsHandler(authService *appauth.AuthService, sessionManager *infraauth.SessionManager, presenter *response.Presenter) *SessionsHandler {
	return &SessionsHandler{
		authService:    authService,
		sessionManager: sessionManager,
		presenter:     presenter,
	}
}

// getOperatorID extracts operator ID from session.
func (h *SessionsHandler) getOperatorID(c *gin.Context) (string, error) {
	sessionID, err := c.Cookie("vyz_session")
	if err != nil {
		return "", err
	}

	sess, _, err := h.authService.ValidateSession(c.Request.Context(), sessionID)
	if err != nil {
		return "", err
	}

	return sess.OperatorID, nil
}

// ListSessions handles GET /v1/auth/sessions - List all active sessions for the operator.
func (h *SessionsHandler) ListSessions(c *gin.Context) {
	operatorID, err := h.getOperatorID(c)
	if err != nil {
		h.presenter.Unauthorized(c, "")
		return
	}

	sessions, err := h.sessionManager.ListActiveSessions(c.Request.Context(), operatorID)
	if err != nil {
		h.presenter.InternalError(c, "Failed to list sessions")
		return
	}

	// Build response with session metadata (excluding sensitive data).
	sessionList := make([]gin.H, 0, len(sessions))
	currentSessionID, _ := c.Cookie("vyz_session")

	for _, sess := range sessions {
		sessionList = append(sessionList, gin.H{
			"id":          sess.ID,
			"ip_address":  sess.IPAddress,
			"user_agent":  sess.UserAgent,
			"created_at":   sess.CreatedAt,
			"expires_at":   sess.ExpiresAt,
			"is_current":  sess.ID == currentSessionID,
		})
	}

	h.presenter.OK(c, gin.H{
		"sessions": sessionList,
		"total":    len(sessionList),
	})
}

// CheckConcurrent handles GET /v1/auth/sessions/concurrent - Check for concurrent logins.
func (h *SessionsHandler) CheckConcurrent(c *gin.Context) {
	operatorID, err := h.getOperatorID(c)
	if err != nil {
		h.presenter.Unauthorized(c, "")
		return
	}

	sessions, err := h.sessionManager.ListActiveSessions(c.Request.Context(), operatorID)
	if err != nil {
		h.presenter.InternalError(c, "Failed to check concurrent sessions")
		return
	}

	// Check for different IPs or user agents.
	var concurrentLogins []gin.H
	currentSessionID, _ := c.Cookie("vyz_session")

	for _, sess := range sessions {
		if sess.ID != currentSessionID {
			concurrentLogins = append(concurrentLogins, gin.H{
				"session_id": sess.ID,
				"ip_address":  sess.IPAddress,
				"user_agent":  sess.UserAgent,
				"created_at":  sess.CreatedAt,
			})
		}
	}

	h.presenter.OK(c, gin.H{
		"has_concurrent": len(concurrentLogins) > 0,
		"count":          len(concurrentLogins),
		"sessions":       concurrentLogins,
	})
}

// RevokeSession handles DELETE /v1/auth/sessions/:id - Revoke a specific session.
func (h *SessionsHandler) RevokeSession(c *gin.Context) {
	operatorID, err := h.getOperatorID(c)
	if err != nil {
		h.presenter.Unauthorized(c, "")
		return
	}

	sessionID := c.Param("id")
	if sessionID == "" {
		h.presenter.BadRequest(c, "Session ID is required")
		return
	}

	// Prevent revoking current session via this endpoint.
	currentSessionID, _ := c.Cookie("vyz_session")
	if sessionID == currentSessionID {
		h.presenter.BadRequest(c, "Cannot revoke current session. Use logout instead.")
		return
	}

	// Verify session belongs to this operator.
	sessions, err := h.sessionManager.ListActiveSessions(c.Request.Context(), operatorID)
	if err != nil {
		h.presenter.InternalError(c, "Failed to verify session")
		return
	}

	found := false
	for _, sess := range sessions {
		if sess.ID == sessionID {
			found = true
			break
		}
	}

	if !found {
		h.presenter.NotFound(c, "Session not found")
		return
	}

	// Revoke the session.
	if err := h.sessionManager.RevokeSession(c.Request.Context(), sessionID); err != nil {
		h.presenter.InternalError(c, "Failed to revoke session")
		return
	}

	h.presenter.OK(c, gin.H{"success": true, "message": "Session revoked"})
}

// RevokeAllExceptCurrent handles DELETE /v1/auth/sessions - Revoke all sessions except current.
func (h *SessionsHandler) RevokeAllExceptCurrent(c *gin.Context) {
	operatorID, err := h.getOperatorID(c)
	if err != nil {
		h.presenter.Unauthorized(c, "")
		return
	}

	currentSessionID, _ := c.Cookie("vyz_session")

	// Get all sessions.
	sessions, err := h.sessionManager.ListActiveSessions(c.Request.Context(), operatorID)
	if err != nil {
		h.presenter.InternalError(c, "Failed to list sessions")
		return
	}

	// Revoke all except current.
	count := 0
	for _, sess := range sessions {
		if sess.ID != currentSessionID {
			if err := h.sessionManager.RevokeSession(c.Request.Context(), sess.ID); err == nil {
				count++
			}
		}
	}

	h.presenter.OK(c, gin.H{
		"success":        true,
		"revoked_count":  count,
		"message":         "All other sessions revoked",
	})
}

// RevokeAllDevices handles POST /v1/auth/sessions/revoke-all - Logout from all devices.
func (h *SessionsHandler) RevokeAllDevices(c *gin.Context) {
	operatorID, err := h.getOperatorID(c)
	if err != nil {
		h.presenter.Unauthorized(c, "")
		return
	}

	// Get all sessions.
	sessions, err := h.sessionManager.ListActiveSessions(c.Request.Context(), operatorID)
	if err != nil {
		h.presenter.InternalError(c, "Failed to list sessions")
		return
	}

	// Revoke all sessions (including current).
	count := 0
	for _, sess := range sessions {
		if err := h.sessionManager.RevokeSession(c.Request.Context(), sess.ID); err == nil {
			count++
		}
	}

	// Clear the session cookie.
	c.SetCookie("vyz_session", "", -1, "/", "", false, true)

	h.presenter.OK(c, gin.H{
		"success":       true,
		"revoked_count": count,
		"message":        "All sessions revoked. Please login again.",
	})
}

// GetSession handles GET /v1/auth/sessions/:id - Get a specific session by ID.
func (h *SessionsHandler) GetSession(c *gin.Context) {
	operatorID, err := h.getOperatorID(c)
	if err != nil {
		h.presenter.Unauthorized(c, "")
		return
	}

	sessionID := c.Param("id")
	if sessionID == "" {
		h.presenter.BadRequest(c, "Session ID is required")
		return
	}

	// Get all sessions and find the one with matching ID.
	sessions, err := h.sessionManager.ListActiveSessions(c.Request.Context(), operatorID)
	if err != nil {
		h.presenter.InternalError(c, "Failed to list sessions")
		return
	}

	currentSessionID, _ := c.Cookie("vyz_session")

	for _, sess := range sessions {
		if sess.ID == sessionID {
			h.presenter.OK(c, gin.H{
				"id":          sess.ID,
				"ip_address":  sess.IPAddress,
				"user_agent":  sess.UserAgent,
				"created_at":   sess.CreatedAt,
				"expires_at":   sess.ExpiresAt,
				"is_current":  sess.ID == currentSessionID,
			})
			return
		}
	}

	h.presenter.NotFound(c, "Session not found")
}
