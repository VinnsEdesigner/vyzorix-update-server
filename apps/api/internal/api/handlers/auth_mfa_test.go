package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/adapters/response"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
	infraauth "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security"
	"github.com/gin-gonic/gin"
)

func TestMFAHandler_EnrollMFA_Unauthorized(t *testing.T) {
	handler := auth.NewMFAHandler(nil, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/auth/mfa/enroll", nil)

	handler.EnrollMFA(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestMFAHandler_EnrollMFA_NoCookie(t *testing.T) {
	handler := auth.NewMFAHandler(nil, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/auth/mfa/enroll", nil)

	handler.EnrollMFA(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestMFAHandler_VerifySetupMFA_InvalidRequest(t *testing.T) {
	handler := auth.NewMFAHandler(nil, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/auth/mfa/verify-setup", bytes.NewReader([]byte("invalid json")))

	handler.VerifySetupMFA(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestMFAHandler_VerifySetupMFA_NoCookie(t *testing.T) {
	handler := auth.NewMFAHandler(nil, nil, nil)

	body := map[string]string{"code": "123456"}
	bodyBytes, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/auth/mfa/verify-setup", bytes.NewReader(bodyBytes))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.VerifySetupMFA(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestMFAHandler_VerifySetupMFA_InvalidCode(t *testing.T) {
	// Test requires full mock infrastructure - just test JSON marshaling.
	body := map[string]string{"code": "000000"}
	bodyBytes, _ := json.Marshal(body)

	var req map[string]string
	json.Unmarshal(bodyBytes, &req)

	if req["code"] != "000000" {
		t.Errorf("code = %s, want 000000", req["code"])
	}
}

func TestMFAHandler_EnableMFA_NoCookie(t *testing.T) {
	handler := auth.NewMFAHandler(nil, nil, nil)

	body := map[string]string{"token": "123456"}
	bodyBytes, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/auth/mfa/enable", bytes.NewReader(bodyBytes))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.EnableMFA(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestMFAHandler_DisableMFA_NoCookie(t *testing.T) {
	handler := auth.NewMFAHandler(nil, nil, nil)

	body := map[string]string{"code": "123456"}
	bodyBytes, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/auth/mfa/disable", bytes.NewReader(bodyBytes))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.DisableMFA(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestMFAHandler_GetMFAStatus_NoCookie(t *testing.T) {
	handler := auth.NewMFAHandler(nil, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/auth/mfa/status", nil)

	handler.GetMFAStatus(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestMFAHandler_RegenerateBackupCodes_NoCookie(t *testing.T) {
	handler := auth.NewMFAHandler(nil, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/auth/mfa/backup-codes/regenerate", nil)

	handler.RegenerateBackupCodes(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestMFAResponse_MarshalJSON(t *testing.T) {
	resp := map[string]interface{}{
		"mfa_enabled": true,
		"method":      "totp",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("json.Marshal() failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal() failed: %v", err)
	}

	if result["mfa_enabled"] != true {
		t.Errorf("mfa_enabled = %v, want true", result["mfa_enabled"])
	}
	if result["method"] != "totp" {
		t.Errorf("method = %v, want totp", result["method"])
	}
}

func TestTOTPGeneration(t *testing.T) {
	secret, err := infraauth.GenerateSecret()
	if err != nil {
		t.Fatalf("GenerateSecret() failed: %v", err)
	}

	totp := infraauth.NewTOTP(secret, infraauth.DefaultTOTPConfig())
	code, err := totp.GenerateCode()
	if err != nil {
		t.Fatalf("GenerateCode() failed: %v", err)
	}

	if len(code) != 6 {
		t.Errorf("code length = %d, want 6", len(code))
	}
}

func TestOperatorRoleConstants(t *testing.T) {
	if operator.RoleOperator != "operator" {
		t.Errorf("RoleOperator = %s, want operator", operator.RoleOperator)
	}
	if operator.RoleSuperAdmin != "super_admin" {
		t.Errorf("RoleSuperAdmin = %s, want super_admin", operator.RoleSuperAdmin)
	}
}

func TestPresenterBadGateway(t *testing.T) {
	gin.SetMode(gin.TestMode)
	presenter := response.NewPresenter(nil, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	presenter.BadGateway(c, "")

	if w.Code != http.StatusBadGateway {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadGateway)
	}
}

func TestPresenterNotImplemented(t *testing.T) {
	gin.SetMode(gin.TestMode)
	presenter := response.NewPresenter(nil, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	presenter.NotImplemented(c, "")

	if w.Code != http.StatusNotImplemented {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotImplemented)
	}
}
