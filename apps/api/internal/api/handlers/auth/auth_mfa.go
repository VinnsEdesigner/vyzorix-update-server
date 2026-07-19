package auth

import (
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/adapters/response"
	appauth "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/audit"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/organization"
	infraauth "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security"

	"github.com/gin-gonic/gin"
)

// MFAHandler handles MFA endpoints.
type MFAHandler struct {
	authService  *appauth.AuthService
	operatorRepo operator.Repository
	presenter    *response.Presenter
	auditLogger  *audit.Logger
}

// NewMFAHandler creates a new MFAHandler.
func NewMFAHandler(authService *appauth.AuthService, operatorRepo operator.Repository, presenter *response.Presenter) *MFAHandler {
	return &MFAHandler{authService: authService, operatorRepo: operatorRepo, presenter: presenter}
}

// SetAuditLogger sets the audit logger for MFA operations.
func (h *MFAHandler) SetAuditLogger(logger *audit.Logger) {
	h.auditLogger = logger
}

// getOperatorFromSession extracts operator ID from session cookie.
func (h *MFAHandler) getOperatorFromSession(c *gin.Context) (string, error) {
	sessionID, err := c.Cookie("vyz_session")
	if err != nil {
		return "", err
	}

	_, op, err := h.authService.ValidateSession(c.Request.Context(), sessionID)
	if err != nil {
		return "", err
	}

	return op.ID, nil
}

// GetMFAStatus handles GET /v1/auth/mfa/status.
func (h *MFAHandler) GetMFAStatus(c *gin.Context) {
	opID, err := h.getOperatorFromSession(c)
	if err != nil {
		h.presenter.Unauthorized(c, "")
		return
	}

	mfaEnabled, err := h.authService.GetMFAStatus(c.Request.Context(), opID)
	if err != nil {
		h.presenter.InternalError(c, "")
		return
	}

	h.presenter.OK(c, gin.H{"mfa_enabled": mfaEnabled})
}

// EnrollMFA handles POST /v1/auth/mfa/enroll.
func (h *MFAHandler) EnrollMFA(c *gin.Context) {
	opID, err := h.getOperatorFromSession(c)
	if err != nil {
		h.presenter.Unauthorized(c, "")
		return
	}

	// Get operator for email
	op, err := h.operatorRepo.FindByID(c.Request.Context(), opID)
	if err != nil {
		h.presenter.InternalError(c, "")
		return
	}

	result, err := h.authService.EnrollMFA(c.Request.Context(), opID, op.Email)
	if err != nil {
		h.presenter.InternalError(c, "")
		return
	}

	h.presenter.OK(c, gin.H{
		"secret": result.Secret,
		"uri":    result.URI,
	})
}

// VerifySetupMFA handles POST /v1/auth/mfa/verify-setup.
func (h *MFAHandler) VerifySetupMFA(c *gin.Context) {
	var req struct {
		Code  string `json:"code"`
		Token string `json:"token"` // TOTP code to verify
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.presenter.BadRequest(c, "")
		return
	}

	opID, err := h.getOperatorFromSession(c)
	if err != nil {
		h.presenter.Unauthorized(c, "")
		return
	}

	// Verify TOTP code
	_, err = h.authService.VerifyMFACode(c.Request.Context(), opID, req.Token)
	if err != nil {
		h.presenter.Unauthorized(c, "")
		return
	}

	h.presenter.OK(c, gin.H{"verified": true})
}

// EnableMFA handles POST /v1/auth/mfa/enable.
func (h *MFAHandler) EnableMFA(c *gin.Context) {
	var req struct {
		Code  string `json:"code"`
		Token string `json:"token"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.presenter.BadRequest(c, "")
		return
	}

	opID, err := h.getOperatorFromSession(c)
	if err != nil {
		h.presenter.Unauthorized(c, "")
		return
	}

	// Verify TOTP code first
	_, err = h.authService.VerifyMFACode(c.Request.Context(), opID, req.Token)
	if err != nil {
		h.presenter.Unauthorized(c, "")
		return
	}

	// Generate backup codes
	backupCodes, err := infraauth.GenerateBackupCodes(8)
	if err != nil {
		h.presenter.InternalError(c, "")
		return
	}

	binding := infraauth.CreateMFASecretBinding(opID, req.Token)

	// Enable MFA and save backup codes using UpdateOperatorMFA
	err = h.operatorRepo.UpdateOperatorMFA(c.Request.Context(), opID, req.Token, binding.MAC, backupCodes)
	if err != nil {
		h.presenter.InternalError(c, "")
		return
	}

	h.presenter.OK(c, gin.H{"success": true, "backup_codes": backupCodes})
}

// DisableMFA handles POST /v1/auth/mfa/disable.
func (h *MFAHandler) DisableMFA(c *gin.Context) {
	var req struct {
		Code string `json:"code"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.presenter.BadRequest(c, "code is required")
		return
	}

	if req.Code == "" {
		h.presenter.BadRequest(c, "MFA code is required to disable MFA")
		return
	}

	opID, err := h.getOperatorFromSession(c)
	if err != nil {
		h.presenter.Unauthorized(c, "")
		return
	}

	// Verify TOTP code before disabling MFA
	_, err = h.authService.VerifyMFACode(c.Request.Context(), opID, req.Code)
	if err != nil {
		h.presenter.Unauthorized(c, "Invalid MFA code")
		return
	}

	err = h.authService.DisableMFA(c.Request.Context(), opID)
	if err != nil {
		h.presenter.InternalError(c, "")
		return
	}

	h.presenter.OK(c, gin.H{"success": true})
}

// VerifyBackupCode handles POST /v1/auth/mfa/verify-backup.
func (h *MFAHandler) VerifyBackupCode(c *gin.Context) {
	var req struct {
		Code string `json:"code"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.presenter.BadRequest(c, "")
		return
	}

	opID, err := h.getOperatorFromSession(c)
	if err != nil {
		h.presenter.Unauthorized(c, "")
		return
	}

	// Verify the operator has MFA enabled
	mfaEnabled, err := h.authService.GetMFAStatus(c.Request.Context(), opID)
	if err != nil || !mfaEnabled {
		h.presenter.Forbidden(c, "")
		return
	}

	// Verify as backup code using service method
	valid, err := h.authService.VerifyBackupCode(c.Request.Context(), opID, req.Code)
	if err != nil || !valid {
		h.presenter.Unauthorized(c, "")
		return
	}

	h.presenter.OK(c, gin.H{"valid": true})
}

// RegenerateBackupCodes handles POST /v1/auth/mfa/regenerate-backup-codes.
func (h *MFAHandler) RegenerateBackupCodes(c *gin.Context) {
	opID, err := h.getOperatorFromSession(c)
	if err != nil {
		h.presenter.Unauthorized(c, "")
		return
	}

	// Generate and persist new backup codes via service
	backupCodes, err := h.authService.RegenerateBackupCodes(c.Request.Context(), opID)
	if err != nil {
		h.presenter.InternalError(c, "")
		return
	}

	h.presenter.OK(c, gin.H{"backup_codes": backupCodes})
}

// VerifyMFA handles POST /v1/auth/mfa/verify - Main MFA verification during login.
func (h *MFAHandler) VerifyMFA(c *gin.Context) {
	var req struct {
		OperatorID string `json:"operator_id" binding:"required"`
		Code      string `json:"code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.presenter.BadRequest(c, "operator_id and code are required")
		return
	}

	// 2: Log MFA verify attempt
	if h.auditLogger != nil {
		h.auditLogger.MFAVerifyAttempt(c.Request.Context(), req.OperatorID, c.ClientIP(), c.GetHeader("User-Agent"))
	}

	// Verify the MFA code first
	session, err := h.authService.VerifyMFACode(c.Request.Context(), req.OperatorID, req.Code)
	if err != nil {
		// 2: Log failed MFA attempt
		if h.auditLogger != nil {
			h.auditLogger.MFAVerifyFailed(c.Request.Context(), req.OperatorID, c.ClientIP())
		}
		h.presenter.Unauthorized(c, "Invalid MFA code")
		return
	}

	// 2: Re-validate operator state before session creation
	// This prevents race conditions where operator could be:
	// - Deleted between MFA verify and session creation
	// - MFA disabled
	// - Role changed
	op, err := h.operatorRepo.FindByID(c.Request.Context(), req.OperatorID)
	if err != nil {
		h.presenter.Unauthorized(c, "Operator not found or invalid")
		return
	}

	// Verify operator is still valid (not deleted, has required fields)
	if !op.IsValid() {
		h.presenter.Unauthorized(c, "Operator is invalid")
		return
	}

	// 2 END: Operator state validated, safe to create session

	// Create session cookie with session ID (critical - must not fail silently)
	if h.authService.GetSessionManager() != nil {
		cookie, cookieErr := h.authService.GetSessionManager().CreateCookie(session.ID)
		if cookieErr != nil {
			h.presenter.InternalError(c, "Failed to create session")
			return
		}
		h.presenter.SetSessionCookie(c, cookie)
	}

	// Issue refresh token and access token for API clients
	var refreshToken string
	var accessToken string
	var expiresAt int64

	// Get role from the operator's last organization membership
	var roleStr string
	if m := op.GetMembership(op.LastOrganizationID); m != nil {
		roleStr = string(m.Role)
	} else if len(op.Memberships) > 0 {
		// Fallback to first active membership
		for _, m := range op.Memberships {
			if m.IsActive() {
				roleStr = string(m.Role)
				break
			}
		}
	}
	if roleStr == "" {
		roleStr = string(organization.RoleViewer) // Default role
	}

	if h.authService != nil {
		refreshToken, err = h.authService.IssueRefreshToken(c.Request.Context(), req.OperatorID, session.ID)
		if err != nil {
			h.presenter.InternalError(c, "Failed to issue refresh token")
			return
		}

		// Generate proper JWT access token
		tokenResult, tokenErr := h.authService.GenerateAccessToken(c.Request.Context(), op.ID, op.Email, op.Name, roleStr)
		if tokenErr != nil {
			h.presenter.InternalError(c, "Failed to generate access token")
			return
		}
		accessToken = tokenResult.AccessToken
		expiresAt = tokenResult.ExpiresAt
	}

	// Return success with session info and tokens
	// 2: Log MFA verify success
	if h.auditLogger != nil {
		h.auditLogger.MFAVerifySuccess(c.Request.Context(), req.OperatorID, session.ID, c.ClientIP())
	}

	h.presenter.OK(c, gin.H{
		"success":       true,
		"session_id":    session.ID,
		"access_token":  accessToken, // Proper JWT access token
		"refresh_token": refreshToken,
		"expires_at":    expiresAt,
		"operator": gin.H{
			"id":          op.ID,
			"email":       op.Email,
			"name":        op.Name,
			"role":        roleStr,
			"mfa_enabled": op.MFAEnabled,
		},
	})
}
