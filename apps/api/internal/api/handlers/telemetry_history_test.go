package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/telemetry"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/storage"
	"github.com/gin-gonic/gin"
)

type mockTelemetryRepo struct {
	data []telemetry.TelemetryFrame
}

func (m *mockTelemetryRepo) QueryByDeviceAndTime(ctx context.Context, deviceID string, startTime, endTime int64, limit int) ([]telemetry.TelemetryFrame, error) {
	var result []telemetry.TelemetryFrame
	for _, t := range m.data {
		if t.DeviceID == deviceID && t.ReceivedAt >= startTime && t.ReceivedAt <= endTime {
			result = append(result, t)
			if len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}

func (m *mockTelemetryRepo) GetLatestForDevice(ctx context.Context, deviceID string) (*telemetry.TelemetryFrame, error) {
	for i := len(m.data) - 1; i >= 0; i-- {
		if m.data[i].DeviceID == deviceID {
			return &m.data[i], nil
		}
	}
	return nil, nil
}

func (m *mockTelemetryRepo) GetStatsForDevice(ctx context.Context, deviceID string, startTime, endTime int64) (map[string]interface{}, error) {
	return map[string]interface{}{
		"count": 10,
		"avg":   50.0,
	}, nil
}

func (m *mockTelemetryRepo) CleanupOld(ctx context.Context, olderThan int64) (int64, error) {
	return 0, nil
}

func TestTelemetryHistoryQuery(t *testing.T) {
	log := testLogger()
	repo := &mockTelemetryRepo{
		data: []telemetry.TelemetryFrame{
			{DeviceID: "device-1", ReceivedAt: time.Now().Unix() - 3600, RiskScore: 50},
			{DeviceID: "device-1", ReceivedAt: time.Now().Unix() - 1800, RiskScore: 60},
			{DeviceID: "device-2", ReceivedAt: time.Now().Unix() - 900, RiskScore: 70},
		},
	}

	handler := NewTelemetryHistoryHandler(log, repo, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/v1/telemetry/history?deviceId=device-1&hours=2", nil)

	handler.Query(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["success"] != true {
		t.Error("expected success true")
	}
}

func TestTelemetryHistoryExportJSON(t *testing.T) {
	log := testLogger()
	repo := &mockTelemetryRepo{
		data: []telemetry.TelemetryFrame{
			{DeviceID: "device-1", ReceivedAt: time.Now().Unix() - 3600, RiskScore: 50},
		},
	}

	handler := NewTelemetryHistoryHandler(log, repo, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/v1/telemetry/history/export?deviceId=device-1&hours=2&format=json", nil)

	handler.ExportJSON(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Error("expected JSON content type")
	}
}

func TestTelemetryHistoryExportCSV(t *testing.T) {
	log := testLogger()
	repo := &mockTelemetryRepo{
		data: []telemetry.TelemetryFrame{
			{DeviceID: "device-1", ReceivedAt: time.Now().Unix() - 3600, RiskScore: 50},
		},
	}

	handler := NewTelemetryHistoryHandler(log, repo, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/v1/telemetry/history/export?deviceId=device-1&hours=2&format=csv", nil)

	handler.ExportCSV(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if w.Header().Get("Content-Type") != "text/csv" {
		t.Error("expected CSV content type")
	}

	if !bytes.Contains(w.Body.Bytes(), []byte("deviceId")) {
		t.Error("expected CSV to contain headers")
	}
}

func TestTelemetryHistoryGetLatest(t *testing.T) {
	log := testLogger()
	repo := &mockTelemetryRepo{
		data: []telemetry.TelemetryFrame{
			{DeviceID: "device-1", ReceivedAt: time.Now().Unix() - 3600, RiskScore: 50},
			{DeviceID: "device-1", ReceivedAt: time.Now().Unix() - 1800, RiskScore: 60},
		},
	}

	handler := NewTelemetryHistoryHandler(log, repo, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "deviceId", Value: "device-1"}}
	c.Request, _ = http.NewRequest("GET", "/api/v1/telemetry/latest/device-1", nil)

	handler.GetLatest(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["success"] != true {
		t.Error("expected success true")
	}
}

func TestTelemetryHistoryGetStats(t *testing.T) {
	log := testLogger()
	repo := &mockTelemetryRepo{}

	handler := NewTelemetryHistoryHandler(log, repo, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "deviceId", Value: "device-1"}}
	c.Request, _ = http.NewRequest("GET", "/api/v1/telemetry/stats/device-1?hours=24", nil)

	handler.GetStats(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["success"] != true {
		t.Error("expected success true")
	}
}

func TestTelemetryHistoryCleanupOld(t *testing.T) {
	log := testLogger()
	repo := &mockTelemetryRepo{}

	handler := NewTelemetryHistoryHandler(log, repo, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("DELETE", "/api/v1/telemetry/cleanup?days=30", nil)

	handler.CleanupOld(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["success"] != true {
		t.Error("expected success true")
	}
}

func TestTelemetryHistoryInvalidDevice(t *testing.T) {
	log := testLogger()
	repo := &mockTelemetryRepo{}

	handler := NewTelemetryHistoryHandler(log, repo, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/v1/telemetry/history?hours=2", nil)

	handler.Query(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestTelemetryHistoryInvalidTimeRange(t *testing.T) {
	log := testLogger()
	repo := &mockTelemetryRepo{}

	handler := NewTelemetryHistoryHandler(log, repo, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/v1/telemetry/history?deviceId=device-1&hours=-1", nil)

	handler.Query(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestTelemetryHistoryLimit(t *testing.T) {
	log := testLogger()
	repo := &mockTelemetryRepo{
		data: []telemetry.TelemetryFrame{
			{DeviceID: "device-1", ReceivedAt: time.Now().Unix() - 3600, RiskScore: 50},
			{DeviceID: "device-1", ReceivedAt: time.Now().Unix() - 1800, RiskScore: 60},
			{DeviceID: "device-1", ReceivedAt: time.Now().Unix() - 900, RiskScore: 70},
		},
	}

	handler := NewTelemetryHistoryHandler(log, repo, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/v1/telemetry/history?deviceId=device-1&hours=2&limit=2", nil)

	handler.Query(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	results := response["data"].([]interface{})
	if len(results) > 2 {
		t.Error("expected limit to be respected")
	}
}

func testLogger() *slog.Logger {
	return slog.Default()
}
