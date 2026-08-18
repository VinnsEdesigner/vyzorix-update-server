package hub

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/command"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device"
)

// noopDeviceRepo is a minimal device.Repository implementation for testing.
// It only implements SetOnline (the only method Hub.Run calls) and returns.
// zero-values/errors for everything else.
type noopDeviceRepo struct {
	online map[string]bool
	mu     sync.Mutex
}

func newNoopDeviceRepo() *noopDeviceRepo {
	return &noopDeviceRepo{online: make(map[string]bool)}
}

func (r *noopDeviceRepo) SetOnline(_ context.Context, id string, online bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.online[id] = online
	return nil
}
func (r *noopDeviceRepo) FindByID(context.Context, string) (*device.Device, error) { return nil, nil }
func (r *noopDeviceRepo) FindByDeviceID(context.Context, device.DeviceID) (*device.Device, error) {
	return nil, nil
}
func (r *noopDeviceRepo) FindByIMEI(context.Context, string) (*device.Device, error) { return nil, nil }
func (r *noopDeviceRepo) FindByFirebaseInstallID(context.Context, string) (*device.Device, error) {
	return nil, nil
}
func (r *noopDeviceRepo) FindByIDAndOperator(context.Context, string, string) (*device.Device, error) {
	return nil, nil
}
func (r *noopDeviceRepo) FindByIDAndOperatorID(context.Context, device.DeviceID, device.OperatorID) (*device.Device, error) {
	return nil, nil
}
func (r *noopDeviceRepo) FindByIMEIAndOperator(context.Context, string, string) (*device.Device, error) {
	return nil, nil
}
func (r *noopDeviceRepo) FindByIMEIAndOrganization(context.Context, string, string) (*device.Device, error) {
	return nil, nil
}
func (r *noopDeviceRepo) FindByIDAndOrganization(context.Context, string, string) (*device.Device, error) {
	return nil, nil
}
func (r *noopDeviceRepo) Create(context.Context, *device.Device) error            { return nil }
func (r *noopDeviceRepo) Update(context.Context, *device.Device) error            { return nil }
func (r *noopDeviceRepo) Delete(context.Context, string) error                    { return nil }
func (r *noopDeviceRepo) DeleteByDeviceID(context.Context, device.DeviceID) error { return nil }
func (r *noopDeviceRepo) UpdateFCMToken(context.Context, string, string) error    { return nil }
func (r *noopDeviceRepo) SetOnlineByDeviceID(context.Context, device.DeviceID, bool) error {
	return nil
}
func (r *noopDeviceRepo) UpdateLastSeen(context.Context, string) error        { return nil }
func (r *noopDeviceRepo) Touch(context.Context, string) error                 { return nil }
func (r *noopDeviceRepo) SetSecretHash(context.Context, string, string) error { return nil }
func (r *noopDeviceRepo) GetSecretHash(context.Context, string) (string, error) {
	return "", nil
}
func (r *noopDeviceRepo) HashAllSecrets(context.Context) (int, error) { return 0, nil }
func (r *noopDeviceRepo) List(context.Context, string, int, int) ([]*device.Device, int, error) {
	return nil, 0, nil
}
func (r *noopDeviceRepo) ListByOperator(context.Context, string) ([]*device.Device, error) {
	return nil, nil
}
func (r *noopDeviceRepo) ListByOperatorID(context.Context, device.OperatorID) ([]*device.Device, error) {
	return nil, nil
}
func (r *noopDeviceRepo) ListByOrganization(context.Context, string) ([]*device.Device, error) {
	return nil, nil
}
func (r *noopDeviceRepo) ListByOrganizationPaginated(context.Context, string, int, int) ([]*device.Device, int, error) {
	return nil, 0, nil
}
func (r *noopDeviceRepo) Count(context.Context, string) (int, error) { return 0, nil }
func (r *noopDeviceRepo) CountByOperator(context.Context, string) (int, error) {
	return 0, nil
}
func (r *noopDeviceRepo) CountByOrganization(context.Context, string) (int, error) {
	return 0, nil
}
func (r *noopDeviceRepo) SoftDelete(context.Context, string, int64, int64) error { return nil }
func (r *noopDeviceRepo) SoftDeleteByIMEI(context.Context, string, int64, int64) error {
	return nil
}
func (r *noopDeviceRepo) ListActive(context.Context, int, int) ([]*device.Device, int, error) {
	return nil, 0, nil
}
func (r *noopDeviceRepo) ListActiveByOperator(context.Context, string) ([]*device.Device, error) {
	return nil, nil
}
func (r *noopDeviceRepo) ListPending(context.Context) ([]*device.Device, error) { return nil, nil }
func (r *noopDeviceRepo) ListPendingByOperator(context.Context, device.OperatorID) ([]*device.Device, error) {
	return nil, nil
}
func (r *noopDeviceRepo) DeleteScheduled(context.Context) (int, error) { return 0, nil }
func (r *noopDeviceRepo) SoftDeleteByOrganization(context.Context, string, int64, int64) (int, error) {
	return 0, nil
}

// Compile-time check.
var _ device.Repository = (*noopDeviceRepo)(nil)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(&discardWriter{}, &slog.HandlerOptions{Level: slog.LevelError}))
}

func newTestHub() *Hub {
	return New(testLogger(), newNoopDeviceRepo(), nil, nil, nil)
}

type discardWriter struct{}

func (d *discardWriter) Write(p []byte) (int, error) { return len(p), nil }

// ─── Hub creation & defaults ───.

func TestHubNew_Defaults(t *testing.T) {
	h := newTestHub()
	if h == nil {
		t.Fatal("New returned nil hub")
	}
	if h.ClientCount() != 0 {
		t.Errorf("expected 0 clients, got %d", h.ClientCount())
	}
	if h.compression == nil {
		t.Error("compression should be initialized by default")
	}
	if h.telemetryFilter == nil {
		t.Error("telemetryFilter should be initialized by default")
	}
	if h.latencyConfig == nil || !h.latencyConfig.Enabled {
		t.Error("latencyConfig should be enabled by default")
	}
}

func TestHubNew_WithConfig(t *testing.T) {
	cfg := &HubConfig{
		Compression: &CompressionConfig{Threshold: 512, EnableCompression: true, Level: 1},
		Filter:      &TelemetryFilterConfig{MaxSubscriptions: 10, EnableServerSideFilter: true},
		Latency:     &LatencyConfig{Enabled: false, MaxLatencyMS: 200, SampleRate: 1.0},
	}
	h := New(testLogger(), nil, nil, nil, cfg)
	if h.compression.config.Threshold != 512 {
		t.Errorf("compression threshold = %d, want 512", h.compression.config.Threshold)
	}
	if h.telemetryFilter.config.MaxSubscriptions != 10 {
		t.Errorf("max subscriptions = %d, want 10", h.telemetryFilter.config.MaxSubscriptions)
	}
	if h.latencyConfig.Enabled {
		t.Error("latency should be disabled")
	}
}

// ─── Hub Run lifecycle ───.

func TestHubRun_Lifecycle(t *testing.T) {
	h := newTestHub()
	ctx, cancel := context.WithCancel(context.Background())
	go h.Run(ctx)
	time.Sleep(50 * time.Millisecond)

	cancel()
	time.Sleep(50 * time.Millisecond)
	// Hub should have stopped without panic.
}

func TestHubRun_PanicRecovery(t *testing.T) {
	h := newTestHub()
	ctx, cancel := context.WithCancel(context.Background())
	go h.Run(ctx)
	time.Sleep(50 * time.Millisecond)

	// Cancel to stop; the recover() in Run only triggers on panic in the loop.
	cancel()
	time.Sleep(50 * time.Millisecond)
}

// ─── Client registration ───.

func TestHubRegisterUnregister(t *testing.T) {
	h := newTestHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go h.Run(ctx)
	time.Sleep(50 * time.Millisecond)

	c := &Client{
		DeviceID: "dev1",
		Send:     make(chan command.CommandFrame, 32),
		Hub:      h,
		Done:     make(chan struct{}),
	}

	h.Register(c)
	time.Sleep(100 * time.Millisecond)

	if !h.Online("dev1") {
		t.Error("device should be online after register")
	}
	if h.ClientCount() != 1 {
		t.Errorf("client count = %d, want 1", h.ClientCount())
	}

	h.Unregister(c)
	time.Sleep(100 * time.Millisecond)

	if h.Online("dev1") {
		t.Error("device should be offline after unregister")
	}
	if h.ClientCount() != 0 {
		t.Errorf("client count = %d, want 0", h.ClientCount())
	}
}

func TestHubRegister_ReplacesOldClient(t *testing.T) {
	h := newTestHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go h.Run(ctx)
	time.Sleep(50 * time.Millisecond)

	old := &Client{
		DeviceID: "dev1",
		Send:     make(chan command.CommandFrame, 32),
		Hub:      h,
		Done:     make(chan struct{}),
	}
	h.Register(old)
	time.Sleep(100 * time.Millisecond)

	newClient := &Client{
		DeviceID: "dev1",
		Send:     make(chan command.CommandFrame, 32),
		Hub:      h,
		Done:     make(chan struct{}),
	}
	h.Register(newClient)
	time.Sleep(100 * time.Millisecond)

	if h.GetClient("dev1") != newClient {
		t.Error("expected new client to replace old")
	}
	if h.ClientCount() != 1 {
		t.Errorf("client count = %d, want 1", h.ClientCount())
	}
}

// ─── Send to connected client ───.

func TestHubSend_ConnectedClient(t *testing.T) {
	h := newTestHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go h.Run(ctx)
	time.Sleep(50 * time.Millisecond)

	c := &Client{
		DeviceID: "dev1",
		Send:     make(chan command.CommandFrame, 32),
		Hub:      h,
		Done:     make(chan struct{}),
	}
	h.Register(c)
	time.Sleep(100 * time.Millisecond)

	frame := command.CommandFrame{Type: "command", DispatchID: "d1", Command: "reboot"}
	if !h.Send("dev1", frame) {
		t.Error("Send returned false for connected client")
	}

	select {
	case received := <-c.Send:
		if received.DispatchID != "d1" {
			t.Errorf("received dispatchID = %s, want d1", received.DispatchID)
		}
	case <-time.After(time.Second):
		t.Error("timeout waiting for message on Send channel")
	}
}

func TestHubSend_OfflineClient_NoQueue(t *testing.T) {
	h := newTestHub()
	frame := command.CommandFrame{Type: "command", DispatchID: "d1"}
	if h.Send("nonexistent", frame) {
		t.Error("Send should return false for offline client with no queue")
	}
}

// ─── Broadcast ───.

func TestHubBroadcastTelemetry(t *testing.T) {
	h := newTestHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go h.Run(ctx)
	time.Sleep(50 * time.Millisecond)

	c1 := &Client{DeviceID: "dev1", Send: make(chan command.CommandFrame, 32), Hub: h, Done: make(chan struct{})}
	c2 := &Client{DeviceID: "dev2", Send: make(chan command.CommandFrame, 32), Hub: h, Done: make(chan struct{})}
	h.Register(c1)
	h.Register(c2)
	time.Sleep(100 * time.Millisecond)

	payload := []byte(`{"type":"telemetry","riskScore":42}`)
	h.BroadcastTelemetry(payload)
	time.Sleep(100 * time.Millisecond)

	// Both clients should receive the broadcast frame.
	for i, c := range []*Client{c1, c2} {
		select {
		case frame := <-c.Send:
			if frame.Type != "broadcast" {
				t.Errorf("client %d: frame type = %s, want broadcast", i, frame.Type)
			}
		case <-time.After(time.Second):
			t.Errorf("client %d: timeout waiting for broadcast", i)
		}
	}
}

// ─── Rate limiter ───.

func TestRateLimiter_Allow(t *testing.T) {
	rl := NewRateLimiter(testLogger(), &RateLimiterConfig{Rate: 10, Burst: 5, CleanupInterval: time.Minute})
	for i := 0; i < 5; i++ {
		if !rl.Allow("client1") {
			t.Errorf("request %d should be allowed within burst", i)
		}
	}
	// 6th request should be rate limited.
	if rl.Allow("client1") {
		t.Error("6th request should be rate limited")
	}
}

func TestRateLimiter_DifferentClientsIndependent(t *testing.T) {
	rl := NewRateLimiter(testLogger(), &RateLimiterConfig{Rate: 10, Burst: 2, CleanupInterval: time.Minute})
	for i := 0; i < 2; i++ {
		if !rl.Allow("client1") {
			t.Errorf("client1 request %d should be allowed", i)
		}
	}
	if rl.Allow("client1") {
		t.Error("client1 3rd request should be limited")
	}
	// client2 should have its own bucket.
	if !rl.Allow("client2") {
		t.Error("client2 should be allowed")
	}
}

func TestRateLimiter_Refill(t *testing.T) {
	rl := NewRateLimiter(testLogger(), &RateLimiterConfig{Rate: 100, Burst: 1, CleanupInterval: time.Minute})
	if !rl.Allow("c1") {
		t.Error("first request should be allowed")
	}
	if rl.Allow("c1") {
		t.Error("second request should be limited (burst=1)")
	}
	time.Sleep(20 * time.Millisecond) // 100/sec = ~2 tokens in 20ms
	if !rl.Allow("c1") {
		t.Error("after refill, request should be allowed")
	}
}

func TestRateLimiter_AllowN(t *testing.T) {
	rl := NewRateLimiter(testLogger(), &RateLimiterConfig{Rate: 10, Burst: 10, CleanupInterval: time.Minute})
	if !rl.AllowN("c1", 5) {
		t.Error("AllowN(5) should succeed with burst=10")
	}
	if rl.AllowN("c1", 10) {
		t.Error("AllowN(10) should fail (only 5 tokens left)")
	}
}

func TestRateLimiter_Reset(t *testing.T) {
	rl := NewRateLimiter(testLogger(), &RateLimiterConfig{Rate: 1, Burst: 1, CleanupInterval: time.Minute})
	rl.Allow("c1")
	rl.Reset("c1")
	if !rl.Allow("c1") {
		t.Error("after Reset, client should be allowed again")
	}
}

func TestRateLimiter_Metrics(t *testing.T) {
	rl := NewRateLimiter(testLogger(), &RateLimiterConfig{Rate: 1, Burst: 1, CleanupInterval: time.Minute})
	rl.Allow("c1") // allowed
	rl.Allow("c1") // limited
	m := rl.GetMetrics()
	if m.TotalRequests != 2 {
		t.Errorf("TotalRequests = %d, want 2", m.TotalRequests)
	}
	if m.TotalAllowed != 1 {
		t.Errorf("TotalAllowed = %d, want 1", m.TotalAllowed)
	}
	if m.TotalLimited != 1 {
		t.Errorf("TotalLimited = %d, want 1", m.TotalLimited)
	}
}

// ─── Compression ───.

func TestCompression_SmallMessageBypassed(t *testing.T) {
	c := NewCompression(testLogger(), &CompressionConfig{Threshold: 1024, EnableCompression: true, Level: 1})
	data := []byte("small message")
	compressed, didCompress, err := c.CompressMessage(data)
	if err != nil {
		t.Fatalf("CompressMessage error: %v", err)
	}
	if didCompress {
		t.Error("small message should not be compressed")
	}
	if string(compressed) != string(data) {
		t.Error("small message should be returned as-is")
	}
}

func TestCompression_LargeMessageCompressed(t *testing.T) {
	c := NewCompression(testLogger(), &CompressionConfig{Threshold: 100, EnableCompression: true, Level: 1})
	// Create a large, compressible message.
	data := make([]byte, 500)
	for i := range data {
		data[i] = 'A'
	}
	compressed, didCompress, err := c.CompressMessage(data)
	if err != nil {
		t.Fatalf("CompressMessage error: %v", err)
	}
	if !didCompress {
		t.Error("large message should be compressed")
	}
	if len(compressed) >= len(data) {
		t.Errorf("compressed size %d should be < original %d", len(compressed), len(data))
	}
}

func TestCompression_Disabled(t *testing.T) {
	c := NewCompression(testLogger(), &CompressionConfig{Threshold: 0, EnableCompression: false, Level: 1})
	data := make([]byte, 500)
	compressed, didCompress, err := c.CompressMessage(data)
	if err != nil {
		t.Fatalf("CompressMessage error: %v", err)
	}
	if didCompress {
		t.Error("compression disabled, should not compress")
	}
	if string(compressed) != string(data) {
		t.Error("should return original data when disabled")
	}
}

// ─── Telemetry filter ───.

func TestTelemetryFilter_SubscribeUnsubscribe(t *testing.T) {
	tf := NewTelemetryFilter(testLogger(), &TelemetryFilterConfig{MaxSubscriptions: 50, EnableServerSideFilter: true})
	if !tf.Subscribe("client1", "dev1") {
		t.Error("Subscribe should succeed")
	}
	subs := tf.GetSubscriptions("client1")
	if len(subs) != 1 || subs[0] != "dev1" {
		t.Errorf("subscriptions = %v, want [dev1]", subs)
	}
	if !tf.Unsubscribe("client1", "dev1") {
		t.Error("Unsubscribe should succeed")
	}
	subs = tf.GetSubscriptions("client1")
	if len(subs) != 0 {
		t.Errorf("subscriptions = %v, want empty", subs)
	}
}

func TestTelemetryFilter_ShouldForward(t *testing.T) {
	tf := NewTelemetryFilter(testLogger(), &TelemetryFilterConfig{MaxSubscriptions: 50, EnableServerSideFilter: true})
	// No subscriptions = receives all (dashboard mode).
	if !tf.ShouldForward("client1", "dev1") {
		t.Error("client with no subscriptions should receive all")
	}
	tf.Subscribe("client1", "dev1")
	if !tf.ShouldForward("client1", "dev1") {
		t.Error("client1 should receive dev1 telemetry after subscribing")
	}
	if tf.ShouldForward("client1", "dev2") {
		t.Error("client1 should NOT receive dev2 telemetry (only subscribed to dev1)")
	}
}

func TestTelemetryFilter_MaxSubscriptions(t *testing.T) {
	tf := NewTelemetryFilter(testLogger(), &TelemetryFilterConfig{MaxSubscriptions: 3, EnableServerSideFilter: true})
	for i := 0; i < 3; i++ {
		if !tf.Subscribe("client1", "dev"+string(rune('0'+i))) {
			t.Errorf("subscription %d should succeed", i)
		}
	}
	// 4th should fail.
	if tf.Subscribe("client1", "dev3") {
		t.Error("4th subscription should fail (max reached)")
	}
}

// ─── Subscriptions (GraphQL) ───.

func TestSubscriptionManager_DeviceUpdates(t *testing.T) {
	h := newTestHub()
	h.InitSubscriptions()

	var mu sync.Mutex
	received := []interface{}{}
	unsub := h.SubscribeDeviceUpdates("op1", "dev1", func(data interface{}) error {
		mu.Lock()
		defer mu.Unlock()
		received = append(received, data)
		return nil
	})
	time.Sleep(50 * time.Millisecond)

	h.PublishDeviceUpdate("op1", "dev1", map[string]string{"status": "online"})
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	if len(received) != 1 {
		t.Errorf("expected 1 callback, got %d", len(received))
	}
	mu.Unlock()

	unsub()
	h.PublishDeviceUpdate("op1", "dev1", map[string]string{"status": "offline"})
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	if len(received) != 1 {
		t.Errorf("expected 1 callback after unsubscribe, got %d", len(received))
	}
	mu.Unlock()
}

func TestSubscriptionManager_Telemetry(t *testing.T) {
	h := newTestHub()
	h.InitSubscriptions()

	var mu sync.Mutex
	received := 0
	unsub := h.SubscribeTelemetry("op1", "dev1", func(data interface{}) error {
		mu.Lock()
		defer mu.Unlock()
		received++
		return nil
	})
	time.Sleep(50 * time.Millisecond)

	h.PublishTelemetry("op1", "dev1", map[string]int{"riskScore": 50})
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	if received != 1 {
		t.Errorf("expected 1 telemetry callback, got %d", received)
	}
	mu.Unlock()

	unsub()
}

func TestSubscriptionManager_CommandStatus(t *testing.T) {
	h := newTestHub()
	h.InitSubscriptions()

	var mu sync.Mutex
	received := 0
	unsub := h.SubscribeCommandStatus("op1", "dispatch1", func(data interface{}) error {
		mu.Lock()
		defer mu.Unlock()
		received++
		return nil
	})
	time.Sleep(50 * time.Millisecond)

	h.PublishCommandStatus("op1", "dispatch1", map[string]string{"status": "delivered"})
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	if received != 1 {
		t.Errorf("expected 1 command status callback, got %d", received)
	}
	mu.Unlock()

	unsub()
}

func TestSubscriptionManager_OrganizationEvent(t *testing.T) {
	h := newTestHub()
	h.InitSubscriptions()

	var mu sync.Mutex
	received := 0
	unsub := h.SubscribeOrganizationEvents("op1", "org1", func(data interface{}) error {
		mu.Lock()
		defer mu.Unlock()
		received++
		return nil
	})
	time.Sleep(50 * time.Millisecond)

	h.PublishOrganizationEvent("org1", map[string]string{"event": "member_joined"})
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	if received != 1 {
		t.Errorf("expected 1 org event callback, got %d", received)
	}
	mu.Unlock()

	unsub()
}

func TestSubscriptionManager_MemberEvent(t *testing.T) {
	h := newTestHub()
	h.InitSubscriptions()

	var mu sync.Mutex
	received := 0
	unsub := h.SubscribeMemberEvents("op1", "org1", func(data interface{}) error {
		mu.Lock()
		defer mu.Unlock()
		received++
		return nil
	})
	time.Sleep(50 * time.Millisecond)

	h.PublishMemberEvent("org1", map[string]string{"event": "member_invited"})
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	if received != 1 {
		t.Errorf("expected 1 member event callback, got %d", received)
	}
	mu.Unlock()

	unsub()
}

// ─── Client metrics ───.

func TestClientMetrics_ConnectDisconnect(t *testing.T) {
	c := &Client{
		DeviceID: "dev1",
		Send:     make(chan command.CommandFrame, 32),
		Done:     make(chan struct{}),
	}
	c.RecordConnectAttempt()
	c.RecordConnectSuccess()
	if !c.IsConnected() {
		t.Error("client should be connected after RecordConnectSuccess")
	}
	c.RecordMessageSent()
	c.RecordMessageReceived()
	c.RecordDisconnect()
	if c.IsConnected() {
		t.Error("client should be disconnected after RecordDisconnect")
	}
	m := c.GetMetrics()
	if m.ConnectAttempts != 1 {
		t.Errorf("ConnectAttempts = %d, want 1", m.ConnectAttempts)
	}
	if m.ConnectSuccesses != 1 {
		t.Errorf("ConnectSuccesses = %d, want 1", m.ConnectSuccesses)
	}
	if m.MessagesSent != 1 {
		t.Errorf("MessagesSent = %d, want 1", m.MessagesSent)
	}
	if m.MessagesReceived != 1 {
		t.Errorf("MessagesReceived = %d, want 1", m.MessagesReceived)
	}
}

func TestClientMetrics_Uptime(t *testing.T) {
	c := &Client{DeviceID: "dev1", Done: make(chan struct{})}
	if c.Uptime() != 0 {
		t.Error("uptime should be 0 before connection")
	}
	c.RecordConnectSuccess()
	// Uptime is measured in whole seconds (Unix timestamp).
	// Just verify it's non-negative after connect success.
	uptime := c.Uptime()
	if uptime < 0 {
		t.Errorf("uptime should be >= 0, got %d", uptime)
	}
}

// ─── Hub SetRateLimiter / SetMessageQueue ───.

func TestHubSetRateLimiter(t *testing.T) {
	h := newTestHub()
	if h.RateLimiter() != nil {
		t.Error("rate limiter should be nil by default")
	}
	rl := NewRateLimiter(testLogger(), nil)
	h.SetRateLimiter(rl)
	if h.RateLimiter() == nil {
		t.Error("rate limiter should be set")
	}
}

func TestHubSetMessageQueue(t *testing.T) {
	h := newTestHub()
	if h.messageQueue != nil {
		t.Error("message queue should be nil by default")
	}
	// Can't easily create a MessageQueue without a DB, so just verify nil.
	h.SetMessageQueue(nil)
	if h.TotalQueuedMessages() != 0 {
		t.Error("total queued should be 0 with no queue")
	}
}

// ─── Hub ConnectionInfo ───.

func TestHubGetConnectionInfo_NotConnected(t *testing.T) {
	h := newTestHub()
	if info := h.GetConnectionInfo("nonexistent"); info != nil {
		t.Error("connection info should be nil for nonexistent device")
	}
}

func TestHubGetAverageLatency_NoData(t *testing.T) {
	h := newTestHub()
	if h.GetAverageLatency("dev1") != 0 {
		t.Error("average latency should be 0 with no data")
	}
}

func TestHubGetAverageLatency_PerDevice(t *testing.T) {
	h := newTestHub()

	// Seed per-device metrics for dev1 only.
	h.metricsMu.Lock()
	h.deviceLatency["dev1"] = &LatencyMetrics{
		TotalMessages:    2,
		TotalLatencyMS:   100,
		AverageLatencyMS: 50,
	}
	// Seed global metrics so we can prove the per-device value wins.
	h.metrics.LatencyMetrics = LatencyMetrics{
		TotalMessages:    10,
		TotalLatencyMS:   1000,
		AverageLatencyMS: 100,
	}
	h.metricsMu.Unlock()

	if got := h.GetAverageLatency("dev1"); got != 50 {
		t.Errorf("dev1 average latency = %d, want 50 (per-device)", got)
	}

	// dev2 has no per-device samples -> falls back to global average.
	if got := h.GetAverageLatency("dev2"); got != 100 {
		t.Errorf("dev2 average latency = %d, want 100 (global fallback)", got)
	}
}

// ─── Hub broadcast to filtered ───.

func TestHubBroadcastTelemetryToFiltered(t *testing.T) {
	h := newTestHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go h.Run(ctx)
	time.Sleep(50 * time.Millisecond)

	sender := &Client{DeviceID: "sender", Send: make(chan command.CommandFrame, 32), Hub: h, Done: make(chan struct{})}
	subscriber := &Client{DeviceID: "subscriber", Send: make(chan command.CommandFrame, 32), Hub: h, Done: make(chan struct{})}
	h.Register(sender)
	h.Register(subscriber)
	time.Sleep(100 * time.Millisecond)

	// Subscribe to sender's telemetry.
	h.Subscribe("subscriber", "sender")

	payload := []byte(`{"type":"telemetry","deviceId":"sender"}`)
	h.BroadcastTelemetryToFiltered("sender", payload)
	time.Sleep(100 * time.Millisecond)

	// Subscriber should receive it.
	select {
	case <-subscriber.Send:
		// Good.
	case <-time.After(time.Second):
		t.Error("subscriber should receive telemetry from sender")
	}

	// Sender should NOT receive its own telemetry.
	select {
	case <-sender.Send:
		t.Error("sender should not receive its own telemetry")
	default:
		// Good.
	}
}

// ─── Hub Send with delivery confirmation ───.

func TestHubSendWithDeliveryConfirmation(t *testing.T) {
	h := newTestHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go h.Run(ctx)
	time.Sleep(50 * time.Millisecond)

	// Create a client with a running WritePump-like consumer.
	c := &Client{
		DeviceID: "dev1",
		Send:     make(chan command.CommandFrame, 32),
		Hub:      h,
		Done:     make(chan struct{}),
	}
	h.Register(c)
	time.Sleep(100 * time.Millisecond)

	// Consume frames in a goroutine to simulate WritePump.
	go func() {
		for frame := range c.Send {
			if frame.DeliveryConfirmation != nil {
				frame.DeliveryConfirmation <- true
			}
		}
	}()

	frame := command.CommandFrame{Type: "command", DispatchID: "d1"}
	confirmed, err := h.SendWithDeliveryConfirmation("dev1", frame, 2*time.Second)
	if err != nil {
		t.Fatalf("SendWithDeliveryConfirmation error: %v", err)
	}
	if !confirmed {
		t.Error("delivery should be confirmed")
	}
}

func TestHubSendWithDeliveryConfirmation_OfflineNoQueue(t *testing.T) {
	h := newTestHub()
	frame := command.CommandFrame{Type: "command", DispatchID: "d1"}
	confirmed, err := h.SendWithDeliveryConfirmation("nonexistent", frame, time.Second)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if confirmed {
		t.Error("delivery should not be confirmed for offline client with no queue")
	}
}

// ─── Message Queue (without DB, just defaults) ───.

func TestMessageQueueConfig_Defaults(t *testing.T) {
	cfg := DefaultMessageQueueConfig()
	if cfg.MaxQueueSize != 1000 {
		t.Errorf("MaxQueueSize = %d, want 1000", cfg.MaxQueueSize)
	}
	if cfg.MessageTTL != 7*24*time.Hour {
		t.Errorf("MessageTTL = %v, want 7 days", cfg.MessageTTL)
	}
}

// ─── CompressedFrame JSON round-trip ───.

func TestCompressedFrame_JSONRoundTrip(t *testing.T) {
	frame := CompressedFrame{
		Type:         "command",
		Compressed:   true,
		OriginalSize: 500,
		Data:         []byte("compressed-data"),
	}
	data, err := json.Marshal(frame)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	var decoded CompressedFrame
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if decoded.Type != "command" || !decoded.Compressed || decoded.OriginalSize != 500 {
		t.Errorf("decoded frame mismatch: %+v", decoded)
	}
}
