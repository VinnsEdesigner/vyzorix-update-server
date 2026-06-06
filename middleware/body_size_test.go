package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestBodySizeLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		limit      int64
		bodySize   int
		wantStatus int
	}{
		{
			name:       "body within limit",
			limit:      100,
			bodySize:   50,
			wantStatus: http.StatusOK,
		},
		{
			name:       "body at exact limit",
			limit:      100,
			bodySize:   100,
			wantStatus: http.StatusOK,
		},
		{
			name:       "body exceeds limit",
			limit:      100,
			bodySize:   150,
			wantStatus: http.StatusRequestEntityTooLarge,
		},
		{
			name:       "Content-Length exceeds limit",
			limit:      100,
			bodySize:   50,
			wantStatus: http.StatusRequestEntityTooLarge,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(BodySizeLimit(tt.limit))
			router.POST("/test", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			body := make([]byte, tt.bodySize)
			for i := range body {
				body[i] = 'x'
			}

			req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(body))
			req.ContentLength = int64(tt.bodySize)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("BodySizeLimit() status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestBodySizeLimitConstants(t *testing.T) {
	if DefaultBodySizeLimit != 1<<20 {
		t.Errorf("DefaultBodySizeLimit = %d, want %d", DefaultBodySizeLimit, 1<<20)
	}
	if LargeBodySizeLimit != 8<<20 {
		t.Errorf("LargeBodySizeLimit = %d, want %d", LargeBodySizeLimit, 8<<20)
	}
}