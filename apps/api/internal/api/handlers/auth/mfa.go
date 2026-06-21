package auth

import (
	"net/http"

	infraauth "github.com/VinnsEdesigner/vyzorix/apps/api/internal/auth"
	appauth "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"

	"github.com/gin-gonic/gin"
)

// MFAHandler handles MFA endpoints.
type MFAHandler struct {
	authService  *appauth.AuthService
	operatorRepo operator.Repository
}

// NewMFAHandler creates a new MFAHandler.
func NewMFAHandler(authService *appauth.AuthService, operatorRepo operator.Repository) *MFAHandler {
	return &MFAHandler{authService: authService, operatorRepo: operatorRepo}
}

// getOperatorFromSession extracts operator ID from session cookie.
func (h *MFAHandler) getOperatorFromSession(c *gin.Context) (string, error) {
	sessionID, err := c.Cookie("session_id")
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	mfaEnabled, err := h.authService.GetMFAStatus(c.Request.Context(), opID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"mfa_enabled": mfaEnabled})
}

// EnrollMFA handles POST /v1/auth/mfa/enroll.
func (h *MFAHandler) EnrollMFA(c *gin.Context) {
	opID, err := h.getOperatorFromSession(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Get operator for email
	op, err := h.operatorRepo.FindByID(c.Request.Context(), opID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	result, err := h.authService.EnrollMFA(c.Request.Context(), opID, op.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request"})
		return
	}

	opID, err := h.getOperatorFromSession(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Verify TOTP code
	_, err = h.authService.VerifyMFACode(c.Request.Context(), opID, req.Token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "bad_request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"verified": true})
}

// EnableMFA handles POST /v1/auth/mfa/enable.
func (h *MFAHandler) EnableMFA(c *gin.Context) {
	var req struct {
		Code  string `json:"code"`
		Token string `json:"token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request"})
		return
	}

	opID, err := h.getOperatorFromSession(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Verify TOTP code first
	_, err = h.authService.VerifyMFACode(c.Request.Context(), opID, req.Token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "bad_request"})
		return
	}

	// Generate backup codes
	backupCodes, err := infraauth.GenerateBackupCodes(8)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	// Enable MFA and save backup codes using UpdateOperatorMFA
	err = h.operatorRepo.UpdateOperatorMFA(c.Request.Context(), opID, req.Token, backupCodes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "backup_codes": backupCodes})
}

// DisableMFA handles POST /v1/auth/mfa/disable.
func (h *MFAHandler) DisableMFA(c *gin.Context) {
	var req struct {
		Code string `json:"code"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request"})
		return
	}

	opID, err := h.getOperatorFromSession(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	err = h.authService.DisableMFA(c.Request.Context(), opID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// VerifyBackupCode handles POST /v1/auth/mfa/verify-backup.
func (h *MFAHandler) VerifyBackupCode(c *gin.Context) {
	var req struct {
		Code string `json:"code"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request"})
		return
	}

	opID, err := h.getOperatorFromSession(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Verify the operator has MFA enabled
	mfaEnabled, err := h.authService.GetMFAStatus(c.Request.Context(), opID)
	if err != nil || !mfaEnabled {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "forbidden"})
		return
	}

	// Verify as backup code using service method
	valid, err := h.authService.VerifyBackupCode(c.Request.Context(), opID, req.Code)
	if err != nil || !valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "bad_request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"valid": true})
}

// RegenerateBackupCodes handles POST /v1/auth/mfa/regenerate-backup-codes.
func (h *MFAHandler) RegenerateBackupCodes(c *gin.Context) {
	opID, err := h.getOperatorFromSession(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Generate and persist new backup codes via service
	backupCodes, err := h.authService.RegenerateBackupCodes(c.Request.Context(), opID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"backup_codes": backupCodes})
}
