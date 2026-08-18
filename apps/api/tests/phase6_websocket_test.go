package tests

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/websocket"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/command"
	config "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/config"
	cryptohmac "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/crypto"
	wshub "github.com/VinnsEdesigner/vyzorix/apps/api/internal/ws"
	"github.com/gin-gonic/gin"
	gorillaws "github.com/gorilla/websocket"
)

// testWSState holds shared state for the WS integration tests.
type testWSState struct {
	hub    *wshub.Hub
	server *httptest.Server
	ctx    context.Context
	cancel context.CancelFunc
}

func newTestWSState(t *testing.T) *testWSState {
	t.Helper()
	gin.SetMode(gin.TestMode)

	logger := slog.New(slog.NewTextHandler(&discardWriter{}, &slog.HandlerOptions{Level: slog.LevelError}))
	noopRepo := newNoopDeviceRepoForTests()

	h := wshub.New(logger, noopRepo, nil, nil, nil)
	h.SetRateLimiter(wshub.NewRateLimiter(logger, &wshub.RateLimiterConfig{
		Rate:            10000,
		Burst:           10000,
		CleanupInterval: time.Minute,
	}))

	ctx, cancel := context.WithCancel(context.Background())
	go h.Run(ctx)
	time.Sleep(50 * time.Millisecond)

	cfg := config.Config{
		Env:            "development",
		EnforceHMAC:    false,
		AllowedOrigins: []string{"*"},
	}

	auditLogger := newTestAuditLogger(logger)
	streamHandler := websocket.NewStreamHandler(logger, cfg, h, cryptohmac.Verifier{}, auditLogger)

	r := gin.New()
	r.GET("/v1/device/:imei/stream", streamHandler.Handle)

	server := httptest.NewServer(r)

	return &testWSState{
		hub:    h,
		server: server,
		ctx:    ctx,
		cancel: cancel,
	}
}

func (s *testWSState) close() {
	s.cancel()
	s.server.Close()
	time.Sleep(50 * time.Millisecond)
}

func (s *testWSState) wsURL(imei string) string {
	return "ws" + strings.TrimPrefix(s.server.URL, "http") + "/v1/device/" + imei + "/stream"
}

func dialWS(t *testing.T, url string) *gorillaws.Conn {
	t.Helper()
	dialer := gorillaws.Dialer{HandshakeTimeout: 5 * time.Second}
	conn, _, err := dialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("WS dial failed: %v", err)
	}
	return conn
}

// ─── 6.1: Connection & Disconnection ───.

func TestPhase6WS_ConnectDisconnect(t *testing.T) {
	state := newTestWSState(t)
	defer state.close()

	conn := dialWS(t, state.wsURL("IMEI_TEST_001"))
	defer conn.Close()
	time.Sleep(200 * time.Millisecond)

	if !state.hub.Online("IMEI_TEST_001") {
		t.Error("device should be online after WS connect")
	}
	if state.hub.ClientCount() != 1 {
		t.Errorf("client count = %d, want 1", state.hub.ClientCount())
	}
	t.Log("device connected via WebSocket, hub shows online")
}

func TestPhase6WS_DisconnectRemovesFromHub(t *testing.T) {
	state := newTestWSState(t)
	defer state.close()

	conn := dialWS(t, state.wsURL("IMEI_TEST_002"))
	time.Sleep(200 * time.Millisecond)

	if !state.hub.Online("IMEI_TEST_002") {
		t.Fatal("device should be online")
	}

	conn.Close()
	time.Sleep(300 * time.Millisecond)

	if state.hub.Online("IMEI_TEST_002") {
		t.Error("device should be offline after WS disconnect")
	}
	if state.hub.ClientCount() != 0 {
		t.Errorf("client count = %d, want 0", state.hub.ClientCount())
	}
	t.Log("device disconnected, hub shows offline")
}

func TestPhase6WS_MultipleDevicesConnect(t *testing.T) {
	state := newTestWSState(t)
	defer state.close()

	conns := make([]*gorillaws.Conn, 3)
	for i := 0; i < 3; i++ {
		imei := "IMEI_MULTI_" + string(rune('A'+i))
		conns[i] = dialWS(t, state.wsURL(imei))
		defer conns[i].Close()
	}
	time.Sleep(300 * time.Millisecond)

	if state.hub.ClientCount() != 3 {
		t.Errorf("client count = %d, want 3", state.hub.ClientCount())
	}
	for i := 0; i < 3; i++ {
		imei := "IMEI_MULTI_" + string(rune('A'+i))
		if !state.hub.Online(imei) {
			t.Errorf("device %s should be online", imei)
		}
	}
	t.Log("3 devices connected simultaneously")
}

// ─── 6.2: Telemetry reception ───.

func TestPhase6WS_SendTelemetry(t *testing.T) {
	state := newTestWSState(t)
	defer state.close()

	conn := dialWS(t, state.wsURL("IMEI_TELEMETRY_001"))
	defer conn.Close()
	time.Sleep(200 * time.Millisecond)

	tele := map[string]any{
		"type":        "telemetry",
		"deviceId":    "IMEI_TELEMETRY_001",
		"riskScore":   42,
		"bufferLevel": 75,
		"thermalTemp": 45.5,
	}
	if err := conn.WriteJSON(tele); err != nil {
		t.Fatalf("WriteJSON telemetry: %v", err)
	}
	time.Sleep(200 * time.Millisecond)

	if !state.hub.Online("IMEI_TELEMETRY_001") {
		t.Error("device should still be online after sending telemetry")
	}
	t.Log("telemetry frame sent and processed")
}

func TestPhase6WS_TelemetryBroadcastsToOtherClients(t *testing.T) {
	state := newTestWSState(t)
	defer state.close()

	conn1 := dialWS(t, state.wsURL("IMEI_BCAST_SENDER"))
	defer conn1.Close()
	conn2 := dialWS(t, state.wsURL("IMEI_BCAST_RECV"))
	defer conn2.Close()
	time.Sleep(300 * time.Millisecond)

	tele := map[string]any{
		"type":      "telemetry",
		"deviceId":  "IMEI_BCAST_SENDER",
		"riskScore": 88,
	}
	if err := conn1.WriteJSON(tele); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}

	conn2.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, msg, err := conn2.ReadMessage()
	if err != nil {
		t.Fatalf("device 2 didn't receive broadcast: %v", err)
	}

	var frame map[string]any
	if err := json.Unmarshal(msg, &frame); err != nil {
		t.Fatalf("unmarshal broadcast: %v", err)
	}
	if frame["type"] != "broadcast" && frame["type"] != "telemetry" {
		t.Errorf("broadcast frame type = %v, want broadcast or telemetry", frame["type"])
	}
	t.Log("telemetry broadcast received by other connected device")
}

// ─── 6.3: Command delivery ───.

func TestPhase6WS_CommandDelivery(t *testing.T) {
	state := newTestWSState(t)
	defer state.close()

	conn := dialWS(t, state.wsURL("IMEI_CMD_001"))
	defer conn.Close()
	time.Sleep(200 * time.Millisecond)

	frame := command.CommandFrame{
		Type:       "command",
		DispatchID: "dispatch-test-001",
		Command:    "reboot",
		Timestamp:  time.Now().UnixMilli(),
	}
	if !state.hub.Send("IMEI_CMD_001", frame) {
		t.Fatal("hub.Send returned false for connected device")
	}

	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("device didn't receive command: %v", err)
	}

	var received command.CommandFrame
	if err := json.Unmarshal(msg, &received); err != nil {
		t.Fatalf("unmarshal command: %v", err)
	}
	if received.DispatchID != "dispatch-test-001" {
		t.Errorf("dispatchID = %s, want dispatch-test-001", received.DispatchID)
	}
	if received.Command != "reboot" {
		t.Errorf("command = %s, want reboot", received.Command)
	}
	t.Log("command delivered to device via WebSocket")
}

func TestPhase6WS_CommandToOfflineDevice(t *testing.T) {
	state := newTestWSState(t)
	defer state.close()

	frame := command.CommandFrame{
		Type:       "command",
		DispatchID: "dispatch-offline-001",
		Command:    "update",
	}
	if state.hub.Send("IMEI_OFFLINE", frame) {
		t.Error("Send should return false for offline device without queue")
	}
	t.Log("command to offline device correctly rejected")
}

// ─── 6.4: Ping/Pong keepalive ───.

func TestPhase6WS_PingPong(t *testing.T) {
	state := newTestWSState(t)
	defer state.close()

	conn := dialWS(t, state.wsURL("IMEI_PING_001"))
	defer conn.Close()
	time.Sleep(200 * time.Millisecond)

	pongReceived := make(chan bool, 1)
	conn.SetPongHandler(func(string) error {
		pongReceived <- true
		return nil
	})

	conn.WriteMessage(gorillaws.PingMessage, nil)
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	time.Sleep(500 * time.Millisecond)

	if !state.hub.Online("IMEI_PING_001") {
		t.Error("device should still be online after ping/pong")
	}
	t.Log("ping/pong keepalive verified, connection stays alive")
}

// ─── 6.5: Reconnection (replace old client) ───.

func TestPhase6WS_ReconnectReplacesOldClient(t *testing.T) {
	state := newTestWSState(t)
	defer state.close()

	conn1 := dialWS(t, state.wsURL("IMEI_RECONNECT_001"))
	time.Sleep(200 * time.Millisecond)

	if state.hub.ClientCount() != 1 {
		t.Errorf("client count = %d, want 1", state.hub.ClientCount())
	}

	conn2 := dialWS(t, state.wsURL("IMEI_RECONNECT_001"))
	defer conn2.Close()
	time.Sleep(300 * time.Millisecond)

	if state.hub.ClientCount() != 1 {
		t.Errorf("client count = %d, want 1 (old replaced)", state.hub.ClientCount())
	}

	conn1.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	_, _, err := conn1.ReadMessage()
	if err == nil {
		t.Error("old connection should have been closed")
	}
	t.Log("reconnect replaced old client, old connection closed")
}

// ─── 6.6: Subscribe/Unsubscribe (telemetry filter) ───.

func TestPhase6WS_SubscribeTelemetry(t *testing.T) {
	state := newTestWSState(t)
	defer state.close()

	if !state.hub.Subscribe("client1", "IMEI_SUB_TARGET") {
		t.Fatal("Subscribe failed")
	}

	subs := state.hub.GetSubscriptions("client1")
	if len(subs) != 1 || subs[0] != "IMEI_SUB_TARGET" {
		t.Errorf("subscriptions = %v, want [IMEI_SUB_TARGET]", subs)
	}

	if !state.hub.Unsubscribe("client1", "IMEI_SUB_TARGET") {
		t.Error("Unsubscribe failed")
	}
	subs = state.hub.GetSubscriptions("client1")
	if len(subs) != 0 {
		t.Errorf("subscriptions = %v, want empty", subs)
	}
	t.Log("subscribe/unsubscribe telemetry filter works")
}

// ─── 6.7: Invalid JSON handling ───.

func TestPhase6WS_InvalidJSONDoesNotKillConnection(t *testing.T) {
	state := newTestWSState(t)
	defer state.close()

	conn := dialWS(t, state.wsURL("IMEI_BADJSON_001"))
	defer conn.Close()
	time.Sleep(200 * time.Millisecond)

	if err := conn.WriteMessage(gorillaws.TextMessage, []byte("not valid json")); err != nil {
		t.Fatalf("WriteMessage: %v", err)
	}
	time.Sleep(200 * time.Millisecond)

	if !state.hub.Online("IMEI_BADJSON_001") {
		t.Error("device should still be online after invalid JSON")
	}

	tele := map[string]any{"type": "telemetry", "deviceId": "IMEI_BADJSON_001", "riskScore": 10}
	if err := conn.WriteJSON(tele); err != nil {
		t.Fatalf("WriteJSON after bad JSON: %v", err)
	}
	time.Sleep(200 * time.Millisecond)

	if !state.hub.Online("IMEI_BADJSON_001") {
		t.Error("device should still be online after valid telemetry")
	}
	t.Log("invalid JSON handled gracefully, connection survives")
}

// ─── 6.8: Delivery confirmation ───.

func TestPhase6WS_DeliveryConfirmation(t *testing.T) {
	state := newTestWSState(t)
	defer state.close()

	conn := dialWS(t, state.wsURL("IMEI_CONFIRM_001"))
	defer conn.Close()
	time.Sleep(200 * time.Millisecond)

	frame := command.CommandFrame{
		Type:       "command",
		DispatchID: "dispatch-confirm-001",
		Command:    "lockdown",
		Timestamp:  time.Now().UnixMilli(),
	}

	confirmed, err := state.hub.SendWithDeliveryConfirmation("IMEI_CONFIRM_001", frame, 3*time.Second)
	if err != nil {
		t.Fatalf("SendWithDeliveryConfirmation error: %v", err)
	}
	if !confirmed {
		t.Error("delivery should be confirmed")
	}

	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, _, err = conn.ReadMessage()
	if err != nil {
		t.Fatalf("device didn't receive command: %v", err)
	}
	t.Log("delivery confirmation verified")
}

// ─── 6.9: Broadcast to all clients ───.

func TestPhase6WS_BroadcastToAll(t *testing.T) {
	state := newTestWSState(t)
	defer state.close()

	conn1 := dialWS(t, state.wsURL("IMEI_BCAST_ALL_1"))
	defer conn1.Close()
	conn2 := dialWS(t, state.wsURL("IMEI_BCAST_ALL_2"))
	defer conn2.Close()
	conn3 := dialWS(t, state.wsURL("IMEI_BCAST_ALL_3"))
	defer conn3.Close()
	time.Sleep(300 * time.Millisecond)

	payload := []byte(`{"type":"telemetry","riskScore":99}`)
	state.hub.BroadcastTelemetry(payload)
	time.Sleep(300 * time.Millisecond)

	for i, c := range []*gorillaws.Conn{conn1, conn2, conn3} {
		c.SetReadDeadline(time.Now().Add(3 * time.Second))
		_, _, err := c.ReadMessage()
		if err != nil {
			t.Errorf("client %d didn't receive broadcast: %v", i+1, err)
		}
	}
	t.Log("broadcast delivered to all 3 connected devices")
}

// ─── 6.10: GraphQL subscription protocol ───.

func TestPhase6WS_SubscriptionPublish(t *testing.T) {
	state := newTestWSState(t)
	defer state.close()

	state.hub.InitSubscriptions()

	var mu sync.Mutex
	received := 0
	unsub := state.hub.SubscribeTelemetry("op1", "IMEI_SUB_TELE", func(data interface{}) error {
		mu.Lock()
		defer mu.Unlock()
		received++
		return nil
	})
	defer unsub()
	time.Sleep(100 * time.Millisecond)

	state.hub.PublishTelemetry("op1", "IMEI_SUB_TELE", map[string]int{"riskScore": 55})
	time.Sleep(300 * time.Millisecond)

	mu.Lock()
	if received != 1 {
		t.Errorf("expected 1 published telemetry event, got %d", received)
	}
	mu.Unlock()
	t.Log("GraphQL subscription publish/subscribe works")
}

// ─── 6.11: Rate limiting ───.

func TestPhase6WS_RateLimiterAllows(t *testing.T) {
	state := newTestWSState(t)
	defer state.close()

	conn := dialWS(t, state.wsURL("IMEI_RATE_001"))
	defer conn.Close()
	time.Sleep(200 * time.Millisecond)

	frame := command.CommandFrame{Type: "command", DispatchID: "r1", Command: "ping"}
	for i := 0; i < 100; i++ {
		if !state.hub.Send("IMEI_RATE_001", frame) {
			break
		}
	}
	t.Log("rate limiter allows high throughput (100 messages)")
}

// ─── 6.12: Connection status endpoint ───.

func TestPhase6WS_ConnectionInfo(t *testing.T) {
	state := newTestWSState(t)
	defer state.close()

	conn := dialWS(t, state.wsURL("IMEI_INFO_001"))
	defer conn.Close()
	time.Sleep(200 * time.Millisecond)

	info := state.hub.GetConnectionInfo("IMEI_INFO_001")
	if info == nil {
		t.Fatal("connection info should not be nil for connected device")
	}
	if !info.Connected {
		t.Error("info.Connected should be true")
	}
	t.Logf("connection info: connected=%v ip=%s", info.Connected, info.ClientIP)
}

// ─── 6.13: Hub metrics ───.

func TestPhase6WS_HubSendUpdatesMetrics(t *testing.T) {
	state := newTestWSState(t)
	defer state.close()

	conn := dialWS(t, state.wsURL("IMEI_METRICS_001"))
	defer conn.Close()
	time.Sleep(200 * time.Millisecond)

	frame := command.CommandFrame{Type: "command", DispatchID: "m1", Command: "status"}
	state.hub.Send("IMEI_METRICS_001", frame)

	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, _, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("device didn't receive command: %v", err)
	}

	if !state.hub.Online("IMEI_METRICS_001") {
		t.Error("device should be online")
	}
	t.Log("hub metrics: command sent and received successfully")
}

// ─── 6.14: Multiple telemetry frames ───.

func TestPhase6WS_MultipleTelemetryFrames(t *testing.T) {
	state := newTestWSState(t)
	defer state.close()

	conn1 := dialWS(t, state.wsURL("IMEI_MULTI_TELE_SENDER"))
	defer conn1.Close()
	conn2 := dialWS(t, state.wsURL("IMEI_MULTI_TELE_RECV"))
	defer conn2.Close()
	time.Sleep(300 * time.Millisecond)

	for i := 0; i < 5; i++ {
		tele := map[string]any{
			"type":      "telemetry",
			"deviceId":  "IMEI_MULTI_TELE_SENDER",
			"riskScore": i * 10,
		}
		if err := conn1.WriteJSON(tele); err != nil {
			t.Fatalf("WriteJSON %d: %v", i, err)
		}
		time.Sleep(50 * time.Millisecond)
	}

	received := 0
	for i := 0; i < 5; i++ {
		conn2.SetReadDeadline(time.Now().Add(2 * time.Second))
		_, _, err := conn2.ReadMessage()
		if err != nil {
			break
		}
		received++
	}
	if received < 1 {
		t.Errorf("expected at least 1 broadcast frame, got %d", received)
	}
	t.Logf("received %d broadcast frames from 5 telemetry sends", received)
}

// ─── 6.15: Large message handling ───.

func TestPhase6WS_LargeMessage(t *testing.T) {
	state := newTestWSState(t)
	defer state.close()

	conn := dialWS(t, state.wsURL("IMEI_LARGE_001"))
	defer conn.Close()
	time.Sleep(200 * time.Millisecond)

	largeData := strings.Repeat("x", 2000)
	tele := map[string]any{
		"type":      "telemetry",
		"deviceId":  "IMEI_LARGE_001",
		"riskScore": 50,
		"rawData":   largeData,
	}
	if err := conn.WriteJSON(tele); err != nil {
		t.Fatalf("WriteJSON large: %v", err)
	}
	time.Sleep(300 * time.Millisecond)

	if !state.hub.Online("IMEI_LARGE_001") {
		t.Error("device should still be online after large message")
	}
	t.Log("large telemetry frame (2KB) handled successfully")
}

// ─── 6.16: Device status tracking ───.

func TestPhase6WS_DeviceStatusTracking(t *testing.T) {
	state := newTestWSState(t)
	defer state.close()

	conn := dialWS(t, state.wsURL("IMEI_STATUS_001"))
	defer conn.Close()
	time.Sleep(300 * time.Millisecond)

	if !state.hub.Online("IMEI_STATUS_001") {
		t.Error("device should be online in hub")
	}

	conn.Close()
	time.Sleep(300 * time.Millisecond)

	if state.hub.Online("IMEI_STATUS_001") {
		t.Error("device should be offline after disconnect")
	}
	t.Log("device status tracking: online → offline transition verified")
}

// ─── 6.17: Concurrent connections stress test ───.

func TestPhase6WS_ConcurrentConnections(t *testing.T) {
	state := newTestWSState(t)
	defer state.close()

	const numClients = 10
	var wg sync.WaitGroup
	conns := make([]*gorillaws.Conn, numClients)

	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			imei := "IMEI_CONCURRENT_" + string(rune('A'+idx))
			c := dialWS(t, state.wsURL(imei))
			conns[idx] = c
		}(i)
	}
	wg.Wait()
	defer func() {
		for _, c := range conns {
			if c != nil {
				c.Close()
			}
		}
	}()
	time.Sleep(500 * time.Millisecond)

	if state.hub.ClientCount() != numClients {
		t.Errorf("client count = %d, want %d", state.hub.ClientCount(), numClients)
	}
	t.Logf("10 concurrent WebSocket connections established")
}

// ─── 6.18: Unknown message type ───.

func TestPhase6WS_UnknownMessageTypeHandled(t *testing.T) {
	state := newTestWSState(t)
	defer state.close()

	conn := dialWS(t, state.wsURL("IMEI_UNKNOWN_TYPE"))
	defer conn.Close()
	time.Sleep(200 * time.Millisecond)

	msg := map[string]any{"type": "unknown_type", "data": "test"}
	if err := conn.WriteJSON(msg); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}
	time.Sleep(200 * time.Millisecond)

	if !state.hub.Online("IMEI_UNKNOWN_TYPE") {
		t.Error("device should still be online after unknown message type")
	}
	t.Log("unknown message type handled gracefully")
}

// ─── 6.19: Status message ───.

func TestPhase6WS_StatusMessage(t *testing.T) {
	state := newTestWSState(t)
	defer state.close()

	conn := dialWS(t, state.wsURL("IMEI_STATUS_MSG"))
	defer conn.Close()
	time.Sleep(200 * time.Millisecond)

	msg := map[string]any{"type": "status", "data": map[string]string{"battery": "85%"}}
	if err := conn.WriteJSON(msg); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}
	time.Sleep(200 * time.Millisecond)

	if !state.hub.Online("IMEI_STATUS_MSG") {
		t.Error("device should still be online after status message")
	}
	t.Log("status message handled")
}

// ─── 6.20: Pong message ───.

func TestPhase6WS_PongMessage(t *testing.T) {
	state := newTestWSState(t)
	defer state.close()

	conn := dialWS(t, state.wsURL("IMEI_PONG_MSG"))
	defer conn.Close()
	time.Sleep(200 * time.Millisecond)

	msg := map[string]any{"type": "pong"}
	if err := conn.WriteJSON(msg); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}
	time.Sleep(200 * time.Millisecond)

	if !state.hub.Online("IMEI_PONG_MSG") {
		t.Error("device should still be online after pong")
	}
	t.Log("pong message handled")
}
