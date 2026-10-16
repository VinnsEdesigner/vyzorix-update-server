package apierrors

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application"
	apporganization "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/organization"
	devicedomain "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device"
	apperrors "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/errors"
	orgdomain "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/organization"

	"github.com/gin-gonic/gin"
)

func newCtx() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	return c, w
}

// TestMapError_Sentinels pins the status and code for representative sentinels
// from every category, so a change to the classification table is caught.
func TestMapError_Sentinels(t *testing.T) {
	cases := []struct {
		name   string
		err    error
		status int
		code   apperrors.ErrorCode
	}{
		{"device not found", devicedomain.ErrNotFound, http.StatusNotFound, apperrors.CodeResourceNotFound},
		{"org not found", orgdomain.ErrNotFound, http.StatusNotFound, apperrors.CodeOrgNotFound},
		{"member not found", orgdomain.ErrMemberNotFound, http.StatusNotFound, apperrors.CodeResourceNotFound},
		{"invalid credentials", application.ErrInvalidCredentials, http.StatusUnauthorized, apperrors.CodeAuthInvalidCredentials},
		{"mfa required", application.ErrMFARequired, http.StatusUnauthorized, apperrors.CodeAuthMFARequired},
		{"forbidden", application.ErrForbidden, http.StatusForbidden, apperrors.CodeAuthzInsufficientPermissions},
		{"not org member", application.ErrNotOrgMember, http.StatusForbidden, apperrors.CodeAuthzOrgMembershipRequired},
		{"higher role", apporganization.ErrCannotModifyHigherRole, http.StatusForbidden, apperrors.CodeAuthzInsufficientPermissions},
		{"invalid input", application.ErrInvalidInput, http.StatusBadRequest, apperrors.CodeValidationFailed},
		{"password breached", application.ErrPasswordBreached, http.StatusBadRequest, apperrors.CodeValidationPasswordBreached},
		{"invitation expired", orgdomain.ErrInvitationExpired, http.StatusBadRequest, apperrors.CodeOrgInvitationExpired},
		{"rate limited", application.ErrRateLimited, http.StatusTooManyRequests, apperrors.CodeRateLimitExceeded},
		{"device online", application.ErrDeviceOnline, http.StatusBadRequest, apperrors.CodeDeviceNotOnline},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := MapError(tc.err)
			if m.Status != tc.status {
				t.Fatalf("status = %d, want %d", m.Status, tc.status)
			}
			if m.Code != tc.code {
				t.Fatalf("code = %s, want %s", m.Code, tc.code)
			}
		})
	}
}

// TestMapError_WrappedSentinels proves classification survives error wrapping,
// since application services routinely add context with fmt.Errorf("%w").
func TestMapError_WrappedSentinels(t *testing.T) {
	wrapped := fmt.Errorf("transfer device: %w", devicedomain.ErrNotFound)
	if m := MapError(wrapped); m.Status != http.StatusNotFound {
		t.Fatalf("wrapped sentinel status = %d, want 404", m.Status)
	}
}

// TestMapError_ServerErrorHonoured proves a structured ServerError keeps its own
// code, status and details rather than being re-derived from the status.
func TestMapError_ServerErrorHonoured(t *testing.T) {
	se := apperrors.NewServerError(apperrors.CodeDeviceCommandTimeout, "command timed out")
	m := MapError(se)
	if m.Status != http.StatusGatewayTimeout {
		t.Fatalf("status = %d, want 504", m.Status)
	}
	if m.Code != apperrors.CodeDeviceCommandTimeout {
		t.Fatalf("code = %s", m.Code)
	}
}

// TestMapError_UnknownIsInternal proves an unclassified error does not leak its
// message to the client.
func TestMapError_UnknownIsInternal(t *testing.T) {
	m := MapError(errors.New("pq: relation \"secret_table\" does not exist"))
	if m.Status != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", m.Status)
	}
	if m.Message != "an unexpected error occurred" {
		t.Fatalf("raw error leaked to client: %q", m.Message)
	}
}

// TestRespondError_EnvelopeShape proves the written body matches the shared
// structured envelope clients already parse.
func TestRespondError_EnvelopeShape(t *testing.T) {
	c, w := newCtx()
	RespondError(c, application.ErrForbidden)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d", w.Code)
	}

	var env struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
		t.Fatalf("body is not the structured envelope: %v (%s)", err, w.Body.String())
	}
	if env.Error.Code != string(apperrors.CodeAuthzInsufficientPermissions) {
		t.Fatalf("code = %s", env.Error.Code)
	}
}

// TestResponse_JSONValue proves the value type encodes and writes correctly.
func TestResponse_JSONValue(t *testing.T) {
	r := OK(map[string]string{"status": "ok"})
	if r.Status != http.StatusOK {
		t.Fatalf("status = %d", r.Status)
	}

	c, w := newCtx()
	r.WriteTo(c)

	if w.Code != http.StatusOK {
		t.Fatalf("written status = %d", w.Code)
	}
	if w.Body.String() != `{"status":"ok"}` {
		t.Fatalf("body = %s", w.Body.String())
	}
}

// TestResponse_ErrorCarriesTrace proves an error Response gains a trace id when
// written (it is request-scoped, so it is filled at write time, not build time).
func TestResponse_ErrorCarriesTrace(t *testing.T) {
	c, w := newCtx()
	c.Set("trace_id", "trace-abc")
	Error(devicedomain.ErrNotFound).WriteTo(c)

	var env struct {
		Error struct {
			TraceID string `json:"trace_id"`
			DocsURL string `json:"docs_url"`
		} `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if env.Error.TraceID != "trace-abc" {
		t.Fatalf("trace_id = %q, want trace-abc", env.Error.TraceID)
	}
	if env.Error.DocsURL == "" {
		t.Fatal("docs_url should be populated")
	}
}

// TestResponse_UnencodableValueBecomes500 proves a value that cannot be encoded
// does not produce an empty 2xx.
func TestResponse_UnencodableValueBecomes500(t *testing.T) {
	r := OK(func() {}) // funcs are not JSON-encodable
	if r.Status != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", r.Status)
	}
}
