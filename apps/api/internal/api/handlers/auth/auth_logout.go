package auth

import (
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/adapters/response"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
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
//
// The logout route lives in the cookie-authenticated group, so the
// CookieAuth middleware has already decrypted the vyz_session cookie and
// placed the validated *session.Session in the gin context. We MUST read the
// session ID from context (middleware.GetCurrentSessionID) — not from
// c.Cookie("vyz_session") — because the raw cookie value is AES-GCM encrypted
// and is NOT the plaintext session ID the database stores. Reading the raw
// cookie passed an opaque ciphertext to sessionRepo.Delete, which matched zero
// rows, returned ErrNotFound, and caused every logout to 500.
func (h *LogoutHandler) Handle(c *gin.Context) {
	sessionID := middleware.GetCurrentSessionID(c)
	if sessionID == "" {
		// No authenticated session in context — nothing to revoke. Clear the
		// cookie anyway so the browser side is clean and return success.
		h.presenter.ClearSessionCookie(c)
		h.presenter.OK(c, gin.H{"message": "logged out"})
		return
	}

	// Get operator ID from the validated session already in context.
	var operatorID string
	if sess := middleware.GetSession(c); sess != nil {
		operatorID = sess.OperatorID
	}

	// Attempt to logout - fail if session deletion fails.
	if err := h.authService.Logout(c.Request.Context(), sessionID); err != nil {
		h.presenter.LogoutFailure(c, operatorID, err.Error())
		h.presenter.InternalError(c, "logout failed")
		return
	}

	// Clear session cookie.
	h.presenter.ClearSessionCookie(c)

	// Log successful logout.
	if operatorID != "" {
		h.presenter.LogoutSuccess(c, operatorID)
	}

	h.presenter.OK(c, gin.H{"message": "logged out"})
}
