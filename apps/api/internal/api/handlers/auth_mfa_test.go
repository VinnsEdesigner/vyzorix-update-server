package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	infraauth "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/dto"
	"github.com/gin-gonic/gin"
)

func TestMFAHandler_EnrollMFA_Unauthorized(t *testing.T) {
	handler := NewMFAHandler(nil)

	w := httptest.NewRecorder()
	c, _ := createTestContext(w, nil)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/auth/mfa/enroll", nil)

	handler.EnrollMFA(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestMFAHandler_EnrollMFA_Success(t *testing.T) {
	handler := NewMFAHandler(nil)

	operator := &dto.Operator{
		ID:    "op-test-123",
		Email: "test@example.com",
		Role:  dto.RoleOperator,
	}

	w := httptest.NewRecorder()
	c, _ := createTestContextWithOperator(w, operator)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/auth/mfa/enroll", nil)

	handler.EnrollMFA(c)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json.Unmarshal() failed: %v", err)
	}

	if resp["secret"] == "" {
		t.Error("secret should not be empty")
	}
	if resp["uri"] == "" {
		t.Error("uri should not be empty")
	}
}

func TestMFAHandler_VerifySetupMFA_InvalidRequest(t *testing.T) {
	handler := NewMFAHandler(nil)

	operator := &dto.Operator{
		ID:    "op-test-123",
		Email: "test@example.com",
		Role:  dto.RoleOperator,
	}

	// Empty body.
	w := httptest.NewRecorder()
	c, _ := createTestContextWithOperator(w, operator)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/auth/mfa/verify-setup", nil)

	handler.VerifySetupMFA(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestMFAHandler_VerifySetupMFA_Success(t *testing.T) {
	handler := NewMFAHandler(nil)

	operator := &dto.Operator{
		ID:    "op-test-123",
		Email: "test@example.com",
		Role:  dto.RoleOperator,
	}

	// Generate a valid TOTP secret and code.
	secret, err := infraauth.GenerateSecret()
	if err != nil {
		t.Fatalf("GenerateSecret() failed: %v", err)
	}

	totp := infraauth.NewTOTP(secret, infraauth.DefaultTOTPConfig())
	code, err := totp.GenerateCode()
	if err != nil {
		t.Fatalf("GenerateCode() failed: %v", err)
	}

	reqBody := map[string]string{
		"secret": secret,
		"code":   code,
	}
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	c, _ := createTestContextWithOperator(w, operator)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/auth/mfa/verify-setup", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.VerifySetupMFA(c)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json.Unmarshal() failed: %v", err)
	}

	if resp["success"] != true {
		t.Errorf("success = %v, want true", resp["success"])
	}

	backupCodes, ok := resp["backup_codes"].([]interface{})
	if !ok || len(backupCodes) == 0 {
		t.Error("backup_codes should be returned")
	}
}

func TestMFAHandler_VerifySetupMFA_InvalidCode(t *testing.T) {
	handler := NewMFAHandler(nil)

	operator := &dto.Operator{
		ID:    "op-test-123",
		Email: "test@example.com",
		Role:  dto.RoleOperator,
	}

	// Generate a valid TOTP secret but use wrong code.
	secret, err := infraauth.GenerateSecret()
	if err != nil {
		t.Fatalf("GenerateSecret() failed: %v", err)
	}

	reqBody := map[string]string{
		"secret": secret,
		"code":   "000000", // Invalid code
	}
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	c, _ := createTestContextWithOperator(w, operator)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/auth/mfa/verify-setup", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.VerifySetupMFA(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestMFAHandler_EnableMFA_MFAEnabled(t *testing.T) {
	handler := NewMFAHandler(nil)

	// Operator already has MFA enabled.
	operator := &dto.Operator{
		ID:         "op-test-123",
		Email:      "test@example.com",
		Role:       dto.RoleOperator,
		MFASecret:  "JBSWY3DPEHPK3PXP", // Valid base32 secret
		MFAEnabled: true,
	}

	// Generate a valid code.
	totp := infraauth.NewTOTP(operator.MFASecret, infraauth.DefaultTOTPConfig())
	code, _ := totp.GenerateCode()

	reqBody := map[string]string{
		"code": code,
	}
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	c, _ := createTestContextWithOperator(w, operator)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/auth/mfa/enable", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.EnableMFA(c)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d: %s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestMFAHandler_DisableMFA_Success(t *testing.T) {
	handler := NewMFAHandler(nil)

	operator := &dto.Operator{
		ID:         "op-test-123",
		Email:      "test@example.com",
		Role:       dto.RoleOperator,
		MFASecret:  "JBSWY3DPEHPK3PXP",
		MFAEnabled: true,
		BackupCodes: []string{"ABCD-EFGH-IJKL", "MNPQ-RSTU-VWXY"},
	}

	// Generate a valid code.
	totp := infraauth.NewTOTP(operator.MFASecret, infraauth.DefaultTOTPConfig())
	code, _ := totp.GenerateCode()

	reqBody := map[string]string{
		"code": code,
	}
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	c, _ := createTestContextWithOperator(w, operator)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/auth/mfa/disable", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.DisableMFA(c)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d: %s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestMFAHandler_DisableMFA_NotEnabled(t *testing.T) {
	handler := NewMFAHandler(nil)

	operator := &dto.Operator{
		ID:    "op-test-123",
		Email: "test@example.com",
		Role:  dto.RoleOperator,
		// MFASecret is empty - MFA not enabled.
	}

	reqBody := map[string]string{
		"code": "123456",
	}
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	c, _ := createTestContextWithOperator(w, operator)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/auth/mfa/disable", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.DisableMFA(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestMFAHandler_GetMFAStatus_Enabled(t *testing.T) {
	handler := NewMFAHandler(nil)

	operator := &dto.Operator{
		ID:           "op-test-123",
		Email:        "test@example.com",
		Role:         dto.RoleOperator,
		MFASecret:    "JBSWY3DPEHPK3PXP",
		MFAEnabled:   true,
		BackupCodes:  []string{"ABCD-EFGH-IJKL", "MNPQ-RSTU-VWXY"},
	}

	w := httptest.NewRecorder()
	c, _ := createTestContextWithOperator(w, operator)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/auth/mfa/status", nil)

	handler.GetMFAStatus(c)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json.Unmarshal() failed: %v", err)
	}

	if resp["enabled"] != true {
		t.Errorf("enabled = %v, want true", resp["enabled"])
	}
	if resp["has_backup_codes"] != true {
		t.Errorf("has_backup_codes = %v, want true", resp["has_backup_codes"])
	}
	if int(resp["backup_codes_count"].(float64)) != 2 {
		t.Errorf("backup_codes_count = %v, want 2", resp["backup_codes_count"])
	}
}

func TestMFAHandler_GetMFAStatus_Disabled(t *testing.T) {
	handler := NewMFAHandler(nil)

	operator := &dto.Operator{
		ID:    "op-test-123",
		Email: "test@example.com",
		Role:  dto.RoleOperator,
		// MFA not enabled.
	}

	w := httptest.NewRecorder()
	c, _ := createTestContextWithOperator(w, operator)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/auth/mfa/status", nil)

	handler.GetMFAStatus(c)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json.Unmarshal() failed: %v", err)
	}

	if resp["enabled"] != false {
		t.Errorf("enabled = %v, want false", resp["enabled"])
	}
	if resp["has_backup_codes"] != false {
		t.Errorf("has_backup_codes = %v, want false", resp["has_backup_codes"])
	}
}

func TestMFAHandler_VerifyBackupCode_Success(t *testing.T) {
	handler := NewMFAHandler(nil)

	backupCodes := []string{"ABCD-EFGH-IJKL", "MNPQ-RSTU-VWXY", "YYYY-ZZZZ-WWWW"}
	operator := &dto.Operator{
		ID:           "op-test-123",
		Email:        "test@example.com",
		Role:         dto.RoleOperator,
		MFASecret:    "JBSWY3DPEHPK3PXP",
		MFAEnabled:   true,
		BackupCodes:  backupCodes,
	}

	reqBody := map[string]string{
		"code": "ABCD-EFGH-IJKL",
	}
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	c, _ := createTestContextWithOperator(w, operator)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/auth/mfa/verify-backup", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.VerifyBackupCode(c)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json.Unmarshal() failed: %v", err)
	}

	if resp["success"] != true {
		t.Errorf("success = %v, want true", resp["success"])
	}
	if int(resp["codes_remaining"].(float64)) != 2 {
		t.Errorf("codes_remaining = %v, want 2", resp["codes_remaining"])
	}
}

func TestMFAHandler_VerifyBackupCode_Invalid(t *testing.T) {
	handler := NewMFAHandler(nil)

	operator := &dto.Operator{
		ID:           "op-test-123",
		Email:        "test@example.com",
		Role:         dto.RoleOperator,
		MFASecret:    "JBSWY3DPEHPK3PXP",
		MFAEnabled:   true,
		BackupCodes:  []string{"ABCD-EFGH-IJKL"},
	}

	reqBody := map[string]string{
		"code": "XXXX-YYYY-ZZZZ", // Invalid code
	}
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	c, _ := createTestContextWithOperator(w, operator)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/auth/mfa/verify-backup", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.VerifyBackupCode(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestMFAHandler_VerifyBackupCode_NotConfigured(t *testing.T) {
	handler := NewMFAHandler(nil)

	operator := &dto.Operator{
		ID:    "op-test-123",
		Email: "test@example.com",
		Role:  dto.RoleOperator,
		// MFA not configured.
	}

	reqBody := map[string]string{
		"code": "ABCD-EFGH-IJKL",
	}
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	c, _ := createTestContextWithOperator(w, operator)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/auth/mfa/verify-backup", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.VerifyBackupCode(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestMFAHandler_RegenerateBackupCodes_Success(t *testing.T) {
	handler := NewMFAHandler(nil)

	operator := &dto.Operator{
		ID:          "op-test-123",
		Email:       "test@example.com",
		Role:        dto.RoleOperator,
		MFASecret:   "JBSWY3DPEHPK3PXP",
		MFAEnabled:  true,
		BackupCodes: []string{"OLD1-OLD2-OLD3"},
	}

	// Generate a valid TOTP code.
	totp := infraauth.NewTOTP(operator.MFASecret, infraauth.DefaultTOTPConfig())
	code, _ := totp.GenerateCode()

	reqBody := map[string]string{
		"code": code,
	}
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	c, _ := createTestContextWithOperator(w, operator)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/auth/mfa/regenerate-backup-codes", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.RegenerateBackupCodes(c)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json.Unmarshal() failed: %v", err)
	}

	if resp["success"] != true {
		t.Errorf("success = %v, want true", resp["success"])
	}

	backupCodes, ok := resp["backup_codes"].([]interface{})
	if !ok || len(backupCodes) == 0 {
		t.Error("backup_codes should be returned")
	}
}

func TestMFAHandler_RegenerateBackupCodes_MFAEnabled(t *testing.T) {
	handler := NewMFAHandler(nil)

	operator := &dto.Operator{
		ID:    "op-test-123",
		Email: "test@example.com",
		Role:  dto.RoleOperator,
		// MFASecret is empty - MFA not enabled.
	}

	reqBody := map[string]string{
		"code": "123456",
	}
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	c, _ := createTestContextWithOperator(w, operator)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/auth/mfa/regenerate-backup-codes", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.RegenerateBackupCodes(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestMFAResponse_MarshalJSON(t *testing.T) {
	resp := MFAResponse{
		Success:     true,
		Secret:      "JBSWY3DPEHPK3PXP",
		URI:         "otpauth://totp/Vyzorix:test@example.com",
		BackupCodes: []string{"ABCD-EFGH-IJKL"},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("json.Marshal() failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal() failed: %v", err)
	}

	if result["success"] != true {
		t.Errorf("success = %v, want true", result["success"])
	}
	if result["secret"] != "JBSWY3DPEHPK3PXP" {
		t.Errorf("secret = %v, want JBSWY3DPEHPK3PXP", result["secret"])
	}
}

func TestMFAResponse_MarshalJSON_WithError(t *testing.T) {
	resp := MFAResponse{
		Success: false,
		Error:   "bad_request",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("json.Marshal() failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal() failed: %v", err)
	}

	if result["success"] != false {
		t.Errorf("success = %v, want false", result["success"])
	}
	if result["error"] != "bad_request" {
		t.Errorf("error = %v, want 'invalid code'", result["error"])
	}
}

// Helper functions.

func createTestContext(w *httptest.ResponseRecorder, operator *dto.Operator) (*gin.Context, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	e := gin.New()
	c, _ := gin.CreateTestContext(w)
	if operator != nil {
		c.Set("operator", operator)
	}
	return c, e
}

func createTestContextWithOperator(w *httptest.ResponseRecorder, operator *dto.Operator) (*gin.Context, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	e := gin.New()
	c, _ := gin.CreateTestContext(w)
	if operator != nil {
		c.Set("operator", operator)
	}
	return c, e
}
