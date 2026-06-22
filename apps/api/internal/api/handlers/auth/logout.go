package auth

import (
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/adapters/response"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"

	"github.com/gin-gonic/gin"
)

// LogoutHandler handles POST /v1/auth/logout.
type LogoutHandler struct {
	authService *auth.AuthService
	presenter   *response.Presenter
}

// NewLogoutHandler creates a new LogoutHandler.
func NewLogoutHandler(authService *auth.AuthService, presenter *response.Presenter) *LogoutHandler {
	return &LogoutHandler{
		authService: authService,
		presenter:   presenter,
	}
}

// Handle processes the logout request.
func (h *LogoutHandler) Handle(c *gin.Context) {
	sessionID, err := c.Cookie("vyz_session")
	if err != nil {
		h.presenter.OK(c, gin.H{"message": "logged out"})
		return
	}

	// Get operator ID from session before logout
	var operatorID string
	session, err := h.authService.GetSession(c.Request.Context(), sessionID)
	if err == nil && session != nil {
		operatorID = session.OperatorID
	}

	_ = h.authService.Logout(c.Request.Context(), sessionID)

	// Clear session cookie
	h.presenter.ClearSessionCookie(c)

	// Log successful logout
	if operatorID != "" {
		h.presenter.LogoutSuccess(c, operatorID)
	}

	h.presenter.OK(c, gin.H{"message": "logged out"})
}
