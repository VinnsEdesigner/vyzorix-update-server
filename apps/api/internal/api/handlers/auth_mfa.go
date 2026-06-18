// Package handlers provides HTTP handlers for MFA operations.
package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"

	security "github.com/VinnsEdesigner/vyzorix/apps/api/internal/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/storage"
)

// MFAHandler handles MFA-related HTTP requests.
type MFAHandler struct {
	Store      *storage.Store
	TOTPConfig security.TOTPConfig
}

// NewMFAHandler creates a new MFA handler.
func NewMFAHandler(store *storage.Store) *MFAHandler {
	return &MFAHandler{
		Store:      store,
		TOTPConfig: security.DefaultTOTPConfig(),
	}
}

// EnableMFARequest represents the request to enable MFA.
type EnableMFARequest struct {
	Code string `json:"code" binding:"required"`
}

// EnableMFAResponse represents the response when enabling MFA.
type EnableMFAResponse struct {
	Success     bool     `json:"success"`
	BackupCodes []string `json:"backup_codes,omitempty"`
	MFASecret   string   `json:"mfa_secret,omitempty"`
}

// EnrollMFA generates a new MFA secret for enrollment.
// The secret is returned to the user along with a QR code URI.
// POST /v1/auth/mfa/enroll.
func (h *MFAHandler) EnrollMFA(c *gin.Context) {
	operator := GetOperatorFromContext(c)
	if operator == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Generate a new TOTP secret.
	secret, err := security.GenerateSecret()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate secret"})
		return
	}

	cfg := security.DefaultTOTPConfig()
	cfg.AccountName = operator.Email
	totp := security.NewTOTP(secret, cfg)

	// Return the secret and provisioning URI (user will scan this with their authenticator app).
	c.JSON(http.StatusOK, gin.H{
		"secret": secret,
		"uri":    totp.ProvisioningURI(),
	})
}

// VerifySetupMFA verifies a TOTP code during MFA enrollment to confirm the user can generate valid codes.
// POST /v1/auth/mfa/verify-setup.
func (h *MFAHandler) VerifySetupMFA(c *gin.Context) {
	var req struct {
		Secret string `json:"secret" binding:"required"`
		Code   string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	operator := GetOperatorFromContext(c)
	if operator == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Verify the code against the provided secret.
	totp := security.NewTOTP(req.Secret, security.DefaultTOTPConfig())
	if !totp.Verify(req.Code) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid code - please check your authenticator app"})
		return
	}

	// Generate backup codes.
	codes, err := security.GenerateBackupCodes(10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate backup codes"})
		return
	}

	// Save MFA secret and backup codes to operator in database.
	if h.Store != nil {
		if err := h.Store.UpdateOperatorMFA(c.Request.Context(), operator.ID, req.Secret, codes); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save MFA configuration"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"backup_codes": codes,
		"message":      "MFA enabled successfully. Save your backup codes in a safe place.",
	})
}

// EnableMFA finalizes MFA enrollment and saves to database.
// POST /v1/auth/mfa/enable.
func (h *MFAHandler) EnableMFA(c *gin.Context) {
	operator := GetOperatorFromContext(c)
	if operator == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Verify the code.
	totp := security.NewTOTP(operator.MFASecret, security.DefaultTOTPConfig())
	if !totp.Verify(req.Code) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid code"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "MFA is now enabled on your account",
	})
}

// DisableMFA disables MFA for the current user.
// POST /v1/auth/mfa/disable.
func (h *MFAHandler) DisableMFA(c *gin.Context) {
	operator := GetOperatorFromContext(c)
	if operator == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if operator.MFASecret == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "MFA not enabled"})
		return
	}

	totp := security.NewTOTP(operator.MFASecret, security.DefaultTOTPConfig())
	if !totp.Verify(req.Code) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid code"})
		return
	}

	// Clear MFA secret and backup codes from operator in database.
	if h.Store != nil {
		if err := h.Store.DisableOperatorMFA(c.Request.Context(), operator.ID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to disable MFA"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "MFA has been disabled on your account",
	})
}

// GetMFAStatus returns the current MFA status for the user.
// GET /v1/auth/mfa/status.
func (h *MFAHandler) GetMFAStatus(c *gin.Context) {
	operator := GetOperatorFromContext(c)
	if operator == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	hasBackupCodes := len(operator.BackupCodes) > 0

	c.JSON(http.StatusOK, gin.H{
		"enabled":             operator.MFAEnabled,
		"has_backup_codes":   hasBackupCodes,
		"backup_codes_count": len(operator.BackupCodes),
	})
}

// VerifyBackupCode verifies a backup code during login.
// POST /v1/auth/mfa/verify-backup.
func (h *MFAHandler) VerifyBackupCode(c *gin.Context) {
	operator := GetOperatorFromContext(c)
	if operator == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if operator.MFASecret == "" || len(operator.BackupCodes) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "MFA not configured"})
		return
	}

	idx := security.ValidateBackupCode(operator.BackupCodes, req.Code)
	if idx < 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid backup code"})
		return
	}

	// Remove the used backup code from the operator's backup codes in database.
	remainingCodes := operator.BackupCodes
	if idx >= 0 && idx < len(remainingCodes) {
		remainingCodes = append(remainingCodes[:idx], remainingCodes[idx+1:]...)
		if h.Store != nil {
			if err := h.Store.UpdateOperatorMFA(c.Request.Context(), operator.ID, operator.MFASecret, remainingCodes); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update backup codes"})
				return
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":          true,
		"backup_code_used": true,
		"codes_remaining": len(remainingCodes),
	})
}

// RegenerateBackupCodes generates new backup codes for a user with MFA enabled.
// POST /v1/auth/mfa/regenerate-backup-codes.
func (h *MFAHandler) RegenerateBackupCodes(c *gin.Context) {
	operator := GetOperatorFromContext(c)
	if operator == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if operator.MFASecret == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "MFA not enabled"})
		return
	}

	// Verify the TOTP code first.
	totp := security.NewTOTP(operator.MFASecret, security.DefaultTOTPConfig())
	if !totp.Verify(req.Code) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid code"})
		return
	}

	// Generate new backup codes.
	codes, err := security.GenerateBackupCodes(10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate backup codes"})
		return
	}

	// Save new backup codes to operator in database, invalidating old ones.
	if h.Store != nil {
		if err := h.Store.UpdateOperatorMFA(c.Request.Context(), operator.ID, operator.MFASecret, codes); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save backup codes"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"backup_codes":  codes,
		"message":      "New backup codes generated. Old codes are now invalid.",
	})
}

// MFAResponse is the response structure for MFA operations.
type MFAResponse struct {
	Success     bool     `json:"success"`
	Secret      string   `json:"secret,omitempty"`
	URI         string   `json:"uri,omitempty"`
	BackupCodes []string `json:"backup_codes,omitempty"`
	Error       string   `json:"error,omitempty"`
}

// MarshalJSON implements json.Marshaler for MFAResponse.
func (r MFAResponse) MarshalJSON() ([]byte, error) {
	type Alias MFAResponse
	if r.Error != "" {
		return json.Marshal(&struct {
			*Alias
			Error string `json:"error"`
		}{
			Alias: (*Alias)(&r),
			Error: r.Error,
		})
	}
	return json.Marshal((*Alias)(&r))
}