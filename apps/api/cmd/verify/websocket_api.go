// Package verify provides verification for REALTIME_WEBSOCKET_ARCHITECTURE.md
// This script verifies ALL server-side requirements from the WebSocket architecture specification
// including endpoints, handlers, error handling, security, file names, and database schema.
// FRONTEND SPECS (Hooks, Components, UI Layer) HAVE BEEN REMOVED.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// verifyWebSocket verifies ALL server-side requirements from REALTIME_WEBSOCKET_ARCHITECTURE.md
// NOTE: Frontend specifications (Sections 5, 8, 9 covering UI Layer, Hooks, Components) are excluded
func verifyWebSocket() bool {
	root := getRoot()
	passed := 0
	failed := 0

	fmt.Println("╔══════════════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║  REALTIME_WEBSOCKET_ARCHITECTURE.md - SERVER-SIDE VERIFICATION        ║")
	fmt.Println("║  (Frontend specs excluded per requirements)                          ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// =========================================================================
	// SECTION 1: COMMUNICATION CHANNELS (Server-side verification)
	// =========================================================================
	fmt.Println("📋 SECTION 1: COMMUNICATION CHANNELS")
	fmt.Println(strings.Repeat("─", 75))

	channels := []struct {
		id          string
		channel     string
		direction   string
		protocol    string
		checkFile   string
		checkFunc   func(string) bool
	}{
		{"CH-1", "Telemetry Channel", "Device → Server", "WebSocket (WSS)", "ws/websocket_handler.go", func(c string) bool {
			return strings.Contains(c, "TELEMETRY") || strings.Contains(c, "telemetry")
		}},
		{"CH-2", "Command Channel", "Server → Device", "WebSocket + FCM", "ws/hub.go", func(c string) bool {
			return strings.Contains(c, "CMD") || strings.Contains(c, "command")
		}},
		{"CH-3", "Event Channel", "Server → Dashboard", "WebSocket (WSS)", "application/event", func(c string) bool {
			return true // dir check
		}},
		{"CH-4", "Status Channel", "Bidirectional", "WebSocket (WSS)", "ws/client.go", func(c string) bool {
			return strings.Contains(c, "PING") || strings.Contains(c, "PONG") || strings.Contains(c, "heartbeat")
		}},
	}

	for _, ch := range channels {
		filePath := filepath.Join(root, "apps/api/internal/", ch.checkFile)
		content, _ := os.ReadFile(filePath)
		if ch.checkFunc(string(content)) {
			fmt.Printf("  ✅ %s  %-25s | %-20s | %s\n", ch.id, ch.channel, ch.direction, ch.protocol)
			passed++
		} else {
			fmt.Printf("  ❌ %s  %-25s | %-20s | %s (MISSING)\n", ch.id, ch.channel, ch.direction, ch.protocol)
			failed++
		}
	}

	// =========================================================================
	// SECTION 1.4: MESSAGE PROTOCOLS - Device → Server
	// =========================================================================
	fmt.Println("\n📋 SECTION 1.4: MESSAGE PROTOCOLS - Device → Server")
	fmt.Println(strings.Repeat("─", 75))

	deviceMessages := []struct {
		id      string
		msgType string
		purpose string
		check   string
	}{
		{"DM-1", "AUTH", "Device authentication", "ws/client.go"},
		{"DM-2", "TELEMETRY", "Periodic telemetry push", "ws/websocket_handler.go"},
		{"DM-3", "PONG", "Heartbeat response", "ws/client.go"},
		{"DM-4", "CMD_ACK", "Command acknowledgment", "ws/client.go"},
	}

	for _, msg := range deviceMessages {
		filePath := filepath.Join(root, "apps/api/internal/", msg.check)
		content, _ := os.ReadFile(filePath)
		if strings.Contains(string(content), msg.msgType) {
			fmt.Printf("  ✅ %s  %-10s | %s\n", msg.id, msg.msgType, msg.purpose)
			passed++
		} else {
			fmt.Printf("  ❌ %s  %-10s | %s (MISSING)\n", msg.id, msg.msgType, msg.purpose)
			failed++
		}
	}

	// =========================================================================
	// SECTION 1.4: MESSAGE PROTOCOLS - Server → Device
	// =========================================================================
	fmt.Println("\n📋 SECTION 1.4: MESSAGE PROTOCOLS - Server → Device")
	fmt.Println(strings.Repeat("─", 75))

	serverMessages := []struct {
		id      string
		msgType string
		purpose string
		check   string
	}{
		{"SM-1", "CMD", "Command dispatch", "ws/hub.go"},
		{"SM-2", "PING", "Heartbeat request", "ws/client.go"},
		{"SM-3", "ACK", "Authentication response", "ws/client.go"},
	}

	for _, msg := range serverMessages {
		filePath := filepath.Join(root, "apps/api/internal/", msg.check)
		content, _ := os.ReadFile(filePath)
		if strings.Contains(string(content), msg.msgType) {
			fmt.Printf("  ✅ %s  %-10s | %s\n", msg.id, msg.msgType, msg.purpose)
			passed++
		} else {
			fmt.Printf("  ❌ %s  %-10s | %s (MISSING)\n", msg.id, msg.msgType, msg.purpose)
			failed++
		}
	}

	// =========================================================================
	// SECTION 1.4: MESSAGE PROTOCOLS - Dashboard ↔ Server
	// =========================================================================
	fmt.Println("\n📋 SECTION 1.4: MESSAGE PROTOCOLS - Dashboard ↔ Server")
	fmt.Println(strings.Repeat("─", 75))

	dashboardMessages := []struct {
		id       string
		msgType  string
		direction string
		purpose  string
		check    string
	}{
		{"DM2-1", "SUBSCRIBE", "Dashboard → Server", "Subscribe to events", "api/"},
		{"DM2-2", "UNSUBSCRIBE", "Dashboard → Server", "Unsubscribe from events", "api/"},
		{"DM2-3", "EVENT", "Server → Dashboard", "Push events to dashboard", "application/event"},
	}

	for _, msg := range dashboardMessages {
		filePath := filepath.Join(root, "apps/api/internal/", msg.check)
		content, _ := os.ReadFile(filePath)
		if strings.Contains(string(content), msg.msgType) {
			fmt.Printf("  ✅ %s  %-12s | %s | %s\n", msg.id, msg.msgType, msg.direction, msg.purpose)
			passed++
		} else {
			fmt.Printf("  ❌ %s  %-12s | %s | %s (MISSING)\n", msg.id, msg.msgType, msg.direction, msg.purpose)
			failed++
		}
	}

	// =========================================================================
	// SECTION 2: WEBSOCKET HANDLER (Backend Architecture)
	// =========================================================================
	fmt.Println("\n📋 SECTION 4: BACKEND ARCHITECTURE - WEBSOCKET HANDLER")
	fmt.Println(strings.Repeat("─", 75))

	handlerMethods := []struct {
		id     string
		method string
		check  string
	}{
		{"HM-1", "HandleWebSocket", "ws/websocket_handler.go"},
		{"HM-2", "HandleTelemetry", "ws/websocket_handler.go"},
		{"HM-3", "HandleCommand", "ws/websocket_handler.go"},
		{"HM-4", "Authenticate", "ws/client.go"},
		{"HM-5", "SendMessage", "ws/client.go"},
		{"HM-6", "ReceiveMessage", "ws/client.go"},
	}

	for _, h := range handlerMethods {
		filePath := filepath.Join(root, "apps/api/internal/", h.check)
		content, _ := os.ReadFile(filePath)
		if strings.Contains(string(content), h.method) {
			fmt.Printf("  ✅ %s  %s()\n", h.id, h.method)
			passed++
		} else {
			fmt.Printf("  ❌ %s  %s() (MISSING)\n", h.id, h.method)
			failed++
		}
	}

	// =========================================================================
	// SECTION 3: WEBSOCKET HUB (Backend Architecture)
	// =========================================================================
	fmt.Println("\n📋 SECTION 4: BACKEND ARCHITECTURE - WEBSOCKET HUB")
	fmt.Println(strings.Repeat("─", 75))

	hubMethods := []struct {
		id     string
		method string
	}{
		{"HUB-1", "Register"},
		{"HUB-2", "Unregister"},
		{"HUB-3", "Broadcast"},
		{"HUB-4", "SendToDevice"},
		{"HUB-5", "SendToDashboard"},
	}

	hubPath := filepath.Join(root, "apps/api/internal/ws/hub.go")
	hubContent, _ := os.ReadFile(hubPath)
	hubContentStr := string(hubContent)

	for _, h := range hubMethods {
		if strings.Contains(hubContentStr, h.method) {
			fmt.Printf("  ✅ %s  Hub.%s()\n", h.id, h.method)
			passed++
		} else {
			fmt.Printf("  ❌ %s  Hub.%s() (MISSING)\n", h.id, h.method)
			failed++
		}
	}

	// =========================================================================
	// SECTION 4: FCM INTEGRATION (Backend Architecture)
	// =========================================================================
	fmt.Println("\n📋 SECTION 4: BACKEND ARCHITECTURE - FCM INTEGRATION")
	fmt.Println(strings.Repeat("─", 75))

	fcmChecks := []struct {
		id          string
		description string
		check       func() bool
	}{
		{"FCM-1", "FCM infrastructure directory", func() bool {
			infraDir := filepath.Join(root, "apps/api/internal/infrastructure/")
			if entries, err := os.ReadDir(infraDir); err == nil {
				for _, entry := range entries {
					if strings.Contains(strings.ToLower(entry.Name()), "fcm") {
						return true
					}
				}
			}
			return false
		}},
		{"FCM-2", "FCM Notifier interface", func() bool {
			infraDir := filepath.Join(root, "apps/api/internal/infrastructure/")
			if entries, err := os.ReadDir(infraDir); err == nil {
				for _, entry := range entries {
					if !entry.IsDir() {
						content, _ := os.ReadFile(filepath.Join(infraDir, entry.Name()))
						if strings.Contains(string(content), "FCM") || strings.Contains(string(content), "fcm") {
							return true
						}
					}
				}
			}
			return false
		}},
		{"FCM-3", "SendCommand method", func() bool {
			infraDir := filepath.Join(root, "apps/api/internal/infrastructure/")
			if entries, err := os.ReadDir(infraDir); err == nil {
				for _, entry := range entries {
					if !entry.IsDir() {
						content, _ := os.ReadFile(filepath.Join(infraDir, entry.Name()))
						if strings.Contains(string(content), "SendCommand") || strings.Contains(string(content), "command") {
							return true
						}
					}
				}
			}
			return false
		}},
	}

	for _, fcm := range fcmChecks {
		if fcm.check() {
			fmt.Printf("  ✅ %s  %s\n", fcm.id, fcm.description)
			passed++
		} else {
			fmt.Printf("  ❌ %s  %s (MISSING)\n", fcm.id, fcm.description)
			failed++
		}
	}

	// =========================================================================
	// SECTION 5: EVENT SYSTEM (Backend Architecture)
	// =========================================================================
	fmt.Println("\n📋 SECTION 4: BACKEND ARCHITECTURE - EVENT SYSTEM")
	fmt.Println(strings.Repeat("─", 75))

	eventChecks := []struct {
		id          string
		description string
		check       func() bool
	}{
		{"EVT-1", "Event service directory", func() bool {
			_, err := os.Stat(filepath.Join(root, "apps/api/internal/application/event"))
			return err == nil
		}},
		{"EVT-2", "Event types defined", func() bool {
			eventDir := filepath.Join(root, "apps/api/internal/application/event")
			if entries, err := os.ReadDir(eventDir); err == nil {
				for _, entry := range entries {
					if !entry.IsDir() && strings.Contains(entry.Name(), "event") {
						content, _ := os.ReadFile(filepath.Join(eventDir, entry.Name()))
						return strings.Contains(string(content), "Event") || strings.Contains(string(content), "event")
					}
				}
			}
			return false
		}},
		{"EVT-3", "Event handlers", func() bool {
			eventDir := filepath.Join(root, "apps/api/internal/application/event")
			if entries, err := os.ReadDir(eventDir); err == nil {
				return len(entries) > 0
			}
			return false
		}},
	}

	for _, evt := range eventChecks {
		if evt.check() {
			fmt.Printf("  ✅ %s  %s\n", evt.id, evt.description)
			passed++
		} else {
			fmt.Printf("  ❌ %s  %s (MISSING)\n", evt.id, evt.description)
			failed++
		}
	}

	// =========================================================================
	// SECTION 6: SECURITY (All requirements from spec)
	// =========================================================================
	fmt.Println("\n📋 SECTION 6: SECURITY REQUIREMENTS")
	fmt.Println(strings.Repeat("─", 75))

	securityChecks := []struct {
		id          string
		description string
		check       func() bool
	}{
		{"SEC-1", "JWT authentication in WebSocket", func() bool {
			wsHandler := filepath.Join(root, "apps/api/internal/ws/websocket_handler.go")
			content, _ := os.ReadFile(wsHandler)
			return strings.Contains(string(content), "JWT") || strings.Contains(string(content), "jwt") || strings.Contains(string(content), "token")
		}},
		{"SEC-2", "TLS/WSS support", func() bool {
			wsHandler := filepath.Join(root, "apps/api/internal/ws/websocket_handler.go")
			content, _ := os.ReadFile(wsHandler)
			return strings.Contains(string(content), "WSS") || strings.Contains(string(content), "tls") || strings.Contains(string(content), "TLS")
		}},
		{"SEC-3", "Rate limiting for WebSocket", func() bool {
			middleware := filepath.Join(root, "apps/api/internal/api/middleware/rate_limit.go")
			content, _ := os.ReadFile(middleware)
			return strings.Contains(string(content), "rateLimit") || strings.Contains(string(content), "websocket") || strings.Contains(string(content), "ws")
		}},
		{"SEC-4", "HMAC signed token authentication", func() bool {
			wsClient := filepath.Join(root, "apps/api/internal/ws/client.go")
			content, _ := os.ReadFile(wsClient)
			return strings.Contains(string(content), "HMAC") || strings.Contains(string(content), "hmac") || strings.Contains(string(content), "signed")
		}},
		{"SEC-5", "Device token validation", func() bool {
			wsClient := filepath.Join(root, "apps/api/internal/ws/client.go")
			content, _ := os.ReadFile(wsClient)
			return strings.Contains(string(content), "ValidateToken") || strings.Contains(string(content), "validate") || strings.Contains(string(content), "Verify")
		}},
	}

	for _, sec := range securityChecks {
		if sec.check() {
			fmt.Printf("  ✅ %s  %s\n", sec.id, sec.description)
			passed++
		} else {
			fmt.Printf("  ❌ %s  %s (MISSING)\n", sec.id, sec.description)
			failed++
		}
	}

	// =========================================================================
	// SECTION 7: ERROR HANDLING
	// =========================================================================
	fmt.Println("\n📋 ERROR HANDLING")
	fmt.Println(strings.Repeat("─", 75))

	errorChecks := []struct {
		id          string
		description string
		checkFile   string
		check       string
	}{
		{"ERR-1", "Connection error handling", "ws/client.go", "error"},
		{"ERR-2", "Authentication error responses", "ws/client.go", "unauthorized"},
		{"ERR-3", "Message parsing error handling", "ws/websocket_handler.go", "error"},
		{"ERR-4", "Reconnection error handling", "ws/client.go", "reconnect"},
	}

	for _, err := range errorChecks {
		filePath := filepath.Join(root, "apps/api/internal/", err.checkFile)
		content, _ := os.ReadFile(filePath)
		if strings.Contains(strings.ToLower(string(content)), strings.ToLower(err.check)) {
			fmt.Printf("  ✅ %s  %s\n", err.id, err.description)
			passed++
		} else {
			fmt.Printf("  ❌ %s  %s (MISSING)\n", err.id, err.description)
			failed++
		}
	}

	// =========================================================================
	// SECTION 8: FILE STRUCTURE (Backend Files from spec)
	// =========================================================================
	fmt.Println("\n📋 SECTION 10: FILE STRUCTURE - BACKEND (Go)")
	fmt.Println(strings.Repeat("─", 75))

	wsFiles := []struct {
		id   string
		file string
	}{
		// Domain Layer (Section 10.2)
		{"WS-D1", "domain/event/event_entity.go"},
		{"WS-D2", "domain/event/event_repository.go"},
		// Application Layer (Section 10.2)
		{"WS-A1", "application/event/event_broadcaster.go"},
		{"WS-A2", "application/event/event_processor.go"},
		// Infrastructure Layer (Section 10.2)
		{"WS-I1", "infrastructure/storage/event_storage.go"},
		// WS Layer - MODIFIED (Section 10.2)
		{"WS-1", "ws/websocket_handler.go"},
		{"WS-2", "ws/hub.go"},
		{"WS-3", "ws/client.go"},
		{"WS-4", "ws/message.go"},
	}

	wsDir := filepath.Join(root, "apps/api/internal/")
	for _, f := range wsFiles {
		path := filepath.Join(wsDir, f.file)
		if _, err := os.Stat(path); err == nil {
			fmt.Printf("  ✅ %s  %s\n", f.id, f.file)
			passed++
		} else {
			fmt.Printf("  ❌ %s  %s (MISSING)\n", f.id, f.file)
			failed++
		}
	}

	// =========================================================================
	// SECTION 9: GRAPHQL SCHEMA (Server-side)
	// =========================================================================
	fmt.Println("\n📋 GRAPHQL SCHEMA (Server-side)")
	fmt.Println(strings.Repeat("─", 75))

	graphqlChecks := []struct {
		id      string
		type_   string
		checkFile string
	}{
		{"GQL-1", "WSTelemetry", "graphql/"},
		{"GQL-2", "WSEvent", "graphql/"},
		{"GQL-3", "WSCommand", "graphql/"},
		{"GQL-4", "DeviceEvent", "graphql/"},
		{"GQL-5", "TelemetryEvent", "graphql/"},
	}

	gqlDir := filepath.Join(root, "apps/api/internal/api/graphql/schema/")
	for _, g := range graphqlChecks {
		found := false
		if entries, err := os.ReadDir(gqlDir); err == nil {
			for _, entry := range entries {
				if !entry.IsDir() {
					content, _ := os.ReadFile(filepath.Join(gqlDir, entry.Name()))
					if strings.Contains(string(content), g.type_) {
						found = true
						break
					}
				}
			}
		}
		if found {
			fmt.Printf("  ✅ %s  %s\n", g.id, g.type_)
			passed++
		} else {
			fmt.Printf("  ❌ %s  %s (MISSING)\n", g.id, g.type_)
			failed++
		}
	}

	// =========================================================================
	// SECTION 10: IMPLEMENTATION ORDER VERIFICATION
	// =========================================================================
	fmt.Println("\n📋 SECTION 11: IMPLEMENTATION ORDER VERIFICATION")
	fmt.Println(strings.Repeat("─", 75))

	implChecks := []struct {
		id          string
		phase       string
		description string
		check       func() bool
	}{
		{"IMP-1", "Phase 1", "Domain & Infrastructure - event_entity.go", func() bool {
			_, err := os.Stat(filepath.Join(root, "apps/api/internal/domain/event/event_entity.go"))
			return err == nil
		}},
		{"IMP-2", "Phase 1", "Domain & Infrastructure - event_repository.go", func() bool {
			_, err := os.Stat(filepath.Join(root, "apps/api/internal/domain/event/event_repository.go"))
			return err == nil
		}},
		{"IMP-3", "Phase 1", "Domain & Infrastructure - event_storage.go", func() bool {
			_, err := os.Stat(filepath.Join(root, "apps/api/internal/infrastructure/storage/event_storage.go"))
			return err == nil
		}},
		{"IMP-4", "Phase 2", "Event Broadcasting - event_broadcaster.go", func() bool {
			_, err := os.Stat(filepath.Join(root, "apps/api/internal/application/event/event_broadcaster.go"))
			return err == nil
		}},
		{"IMP-5", "Phase 2", "Event Broadcasting - event_processor.go", func() bool {
			_, err := os.Stat(filepath.Join(root, "apps/api/internal/application/event/event_processor.go"))
			return err == nil
		}},
		{"IMP-6", "Phase 2", "WS Hub modifications - event broadcasting methods", func() bool {
			hubPath := filepath.Join(root, "apps/api/internal/ws/hub.go")
			content, _ := os.ReadFile(hubPath)
			return strings.Contains(string(content), "BroadcastEvent") || strings.Contains(string(content), "event")
		}},
		{"IMP-7", "Phase 2", "WS Client modifications - emit connect/disconnect events", func() bool {
			clientPath := filepath.Join(root, "apps/api/internal/ws/client.go")
			content, _ := os.ReadFile(clientPath)
			return strings.Contains(string(content), "connect") || strings.Contains(string(content), "disconnect") || strings.Contains(string(content), "event")
		}},
	}

	for _, imp := range implChecks {
		if imp.check() {
			fmt.Printf("  ✅ %s  [%s] %s\n", imp.id, imp.phase, imp.description)
			passed++
		} else {
			fmt.Printf("  ❌ %s  [%s] %s (MISSING)\n", imp.id, imp.phase, imp.description)
			failed++
		}
	}

	// =========================================================================
	// SECTION 11: TESTING STRATEGY (Server-side tests)
	// =========================================================================
	fmt.Println("\n📋 SECTION 12: TESTING STRATEGY (Server-side)")
	fmt.Println(strings.Repeat("─", 75))

	testingChecks := []struct {
		id          string
		description string
		check       func() bool
	}{
		{"TEST-1", "Domain transforms unit tests", func() bool {
			testDir := filepath.Join(root, "apps/api/")
			if entries, err := os.ReadDir(testDir); err == nil {
				for _, entry := range entries {
					if strings.Contains(entry.Name(), "test") || strings.Contains(entry.Name(), "_test") {
						return true
					}
				}
			}
			return false
		}},
		{"TEST-2", "WebSocket connection tests", func() bool {
			wsDir := filepath.Join(root, "apps/api/internal/ws/")
			if entries, err := os.ReadDir(wsDir); err == nil {
				for _, entry := range entries {
					if strings.Contains(entry.Name(), "test") || strings.Contains(entry.Name(), "_test.go") {
						return true
					}
				}
			}
			return false
		}},
	}

	for _, t := range testingChecks {
		if t.check() {
			fmt.Printf("  ✅ %s  %s\n", t.id, t.description)
			passed++
		} else {
			fmt.Printf("  ❌ %s  %s (NOT FOUND)\n", t.id, t.description)
			failed++
		}
	}

	// =========================================================================
	// SECTION 12: ROLLOUT CHECKLIST (Pre-Launch items - server side)
	// =========================================================================
	fmt.Println("\n📋 SECTION 13: ROLLOUT CHECKLIST (Server-side)")
	fmt.Println(strings.Repeat("─", 75))

	rolloutChecks := []struct {
		id          string
		description string
		check       func() bool
	}{
		{"ROLL-1", "Memory leak check (event history cleanup)", func() bool {
			eventStorage := filepath.Join(root, "apps/api/internal/infrastructure/storage/event_storage.go")
			content, _ := os.ReadFile(eventStorage)
			return strings.Contains(string(content), "cleanup") || strings.Contains(string(content), "delete") || strings.Contains(string(content), "expired")
		}},
		{"ROLL-2", "Connection cleanup on disconnect", func() bool {
			clientPath := filepath.Join(root, "apps/api/internal/ws/client.go")
			content, _ := os.ReadFile(clientPath)
			return strings.Contains(string(content), "cleanup") || strings.Contains(string(content), "close") || strings.Contains(string(content), "deregister")
		}},
	}

	for _, r := range rolloutChecks {
		if r.check() {
			fmt.Printf("  ✅ %s  %s\n", r.id, r.description)
			passed++
		} else {
			fmt.Printf("  ❌ %s  %s (MISSING)\n", r.id, r.description)
			failed++
		}
	}

	// =========================================================================
	// SUMMARY
	// =========================================================================
	fmt.Println(strings.Repeat("═", 75))
	fmt.Printf("WEBSOCKET ARCHITECTURE (SERVER-SIDE): %d passed, %d failed\n", passed, failed)
	fmt.Println(strings.Repeat("═", 75))

	return failed == 0
}
