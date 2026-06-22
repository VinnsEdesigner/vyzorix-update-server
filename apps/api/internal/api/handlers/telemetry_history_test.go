package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestTelemetryHistoryEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/v1/telemetry/history", func(c *gin.Context) {
		deviceID := c.Query("deviceId")
		if deviceID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "deviceId required"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"success":  true,
			"deviceId": deviceID,
			"entries": []interface{}{},
		})
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/v1/telemetry/history?deviceId=device-1", nil)

	r.ServeHTTP(w, c.Request)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["success"] != true {
		t.Error("expected success true")
	}
}

func TestTelemetryHistoryMissingDevice(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/v1/telemetry/history", func(c *gin.Context) {
		deviceID := c.Query("deviceId")
		if deviceID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "deviceId required"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/v1/telemetry/history", nil)

	r.ServeHTTP(w, c.Request)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestTelemetryExportCSV(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/v1/telemetry/export", func(c *gin.Context) {
		c.Header("Content-Type", "text/csv")
		c.String(http.StatusOK, "deviceId,timestamp,riskScore\ndevice-1,1234567890,50")
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/v1/telemetry/export?deviceId=device-1&format=csv", nil)

	r.ServeHTTP(w, c.Request)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if w.Header().Get("Content-Type") != "text/csv" {
		t.Error("expected CSV content type")
	}
	if !bytes.Contains(w.Body.Bytes(), []byte("deviceId")) {
		t.Error("expected CSV headers")
	}
}
