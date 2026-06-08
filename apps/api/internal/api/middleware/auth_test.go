package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestAuthenticator_ValidBearerToken(t *testing.T) {
	auth := Authenticator{
		TokenSecret:       "secret-token-123",
		DevelopmentBypass: false,
	}

	handler := auth.Middleware()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer secret-token-123")

	handler(c)

	if c.IsAborted() {
		t.Error("request should not be aborted")
	}
}

func TestAuthenticator_ValidVyzorixToken(t *testing.T) {
	auth := Authenticator{
		TokenSecret:       "secret-token-123",
		DevelopmentBypass: false,
	}

	handler := auth.Middleware()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("X-Vyzorix-Token", "secret-token-123")

	handler(c)

	if c.IsAborted() {
		t.Error("request should not be aborted")
	}
}

func TestAuthenticator_InvalidBearerToken(t *testing.T) {
	auth := Authenticator{
		TokenSecret:       "secret-token-123",
		DevelopmentBypass: false,
	}

	handler := auth.Middleware()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer wrong-token")

	handler(c)

	if !c.IsAborted() {
		t.Error("request should be aborted for invalid token")
	}
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuthenticator_InvalidVyzorixToken(t *testing.T) {
	auth := Authenticator{
		TokenSecret:       "secret-token-123",
		DevelopmentBypass: false,
	}

	handler := auth.Middleware()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("X-Vyzorix-Token", "wrong-token")

	handler(c)

	if !c.IsAborted() {
		t.Error("request should be aborted for invalid token")
	}
}

func TestAuthenticator_NoToken(t *testing.T) {
	auth := Authenticator{
		TokenSecret:       "secret-token-123",
		DevelopmentBypass: false,
	}

	handler := auth.Middleware()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	// No auth headers

	handler(c)

	if !c.IsAborted() {
		t.Error("request should be aborted for missing token")
	}
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuthenticator_DevelopmentBypass(t *testing.T) {
	auth := Authenticator{
		TokenSecret:       "", // No token secret
		DevelopmentBypass: true,
	}

	handler := auth.Middleware()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	// No auth headers

	handler(c)

	if c.IsAborted() {
		t.Error("request should not be aborted when DevelopmentBypass is true")
	}
}

func TestAuthenticator_MissingBearerPrefix(t *testing.T) {
	auth := Authenticator{
		TokenSecret:       "secret-token-123",
		DevelopmentBypass: false,
	}

	handler := auth.Middleware()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("Authorization", "secret-token-123") // Missing "Bearer "

	handler(c)

	if !c.IsAborted() {
		t.Error("request should be aborted for missing Bearer prefix")
	}
}

func TestAuthenticator_BearerWithDifferentToken(t *testing.T) {
	auth := Authenticator{
		TokenSecret:       "secret-token-123",
		DevelopmentBypass: false,
	}

	handler := auth.Middleware()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer different-token-456")

	handler(c)

	if !c.IsAborted() {
		t.Error("request should be aborted for non-matching token")
	}
}

func TestAuthenticator_ResponseBody(t *testing.T) {
	auth := Authenticator{
		TokenSecret:       "secret-token-123",
		DevelopmentBypass: false,
	}

	handler := auth.Middleware()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	handler(c)

	body := w.Body.String()
	if body == "" {
		t.Error("expected error message in response body")
	}
}

func TestAuthenticator_EmptyTokenSecret(t *testing.T) {
	auth := Authenticator{
		TokenSecret:       "",
		DevelopmentBypass: false,
	}

	handler := auth.Middleware()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer any-token")

	handler(c)

	// With empty token secret and no bypass, should abort
	if !c.IsAborted() {
		t.Error("request should be aborted for empty token secret")
	}
}

func TestAuthenticator_CaseSensitiveToken(t *testing.T) {
	auth := Authenticator{
		TokenSecret:       "SecretToken123",
		DevelopmentBypass: false,
	}

	handler := auth.Middleware()

	tests := []struct {
		name  string
		token string
		valid bool
	}{
		{"exact match", "SecretToken123", true},
		{"lowercase", "secretToken123", false},
		{"uppercase", "SECRETTOKEN123", false},
		{"missing letter", "SecretToken12", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/test", nil)
			c.Request.Header.Set("Authorization", "Bearer "+tt.token)

			handler(c)

			if tt.valid && c.IsAborted() {
				t.Error("valid token should not be aborted")
			}
			if !tt.valid && !c.IsAborted() {
				t.Error("invalid token should be aborted")
			}
		})
	}
}

func TestAuthenticator_BothHeadersSet(t *testing.T) {
	auth := Authenticator{
		TokenSecret:       "secret-token-123",
		DevelopmentBypass: false,
	}

	handler := auth.Middleware()

	// Both headers set with valid token in Authorization
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer secret-token-123")
	c.Request.Header.Set("X-Vyzorix-Token", "wrong-token")

	handler(c)

	// Should pass because Authorization header is valid
	if c.IsAborted() {
		t.Error("request should not be aborted when Authorization header is valid")
	}
}

func TestAuthenticator_XVyzorixTokenOnly(t *testing.T) {
	auth := Authenticator{
		TokenSecret:       "secret-token-123",
		DevelopmentBypass: false,
	}

	handler := auth.Middleware()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("X-Vyzorix-Token", "secret-token-123")
	// No Authorization header

	handler(c)

	if c.IsAborted() {
		t.Error("request should not be aborted when X-Vyzorix-Token is valid")
	}
}
