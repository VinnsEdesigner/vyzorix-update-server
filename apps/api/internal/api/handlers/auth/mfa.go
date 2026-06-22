package auth

import (
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/adapters/response"
	infraauth "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security"
	appauth "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"

	"github.com/gin-gonic/gin"
)

// MFAHandler handles MFA endpoints.
type MFAHandler struct {
	authService  *appauth.AuthService
	operatorRepo operator.Repository
	presenter  *response.Presenter
}

// NewMFAHandler creates a new MFAHandler.
func NewMFAHandler(authService *appauth.AuthService, operatorRepo operator.Repository, presenter *response.Presenter) *MFAHandler {
	return &MFAHandler{authService: authService, operatorRepo: operatorRepo, presenter: presenter}
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

	// Enable MFA and save backup codes using UpdateOperatorMFA
	err = h.operatorRepo.UpdateOperatorMFA(c.Request.Context(), opID, req.Token, backupCodes)
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
		h.presenter.BadRequest(c, "")
		return
	}

	opID, err := h.getOperatorFromSession(c)
	if err != nil {
		h.presenter.Unauthorized(c, "")
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
