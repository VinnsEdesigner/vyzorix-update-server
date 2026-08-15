package tests

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestPhase2Updates tests public OTA endpoints, tenant update endpoints,
// push-update flow, sync/export, and data integrity. 24 subtests.
//
// Depends on Phase 1 state (org_id, api_keys, session cookies) and the
// static JSON files written by Phase 1's writeStaticFiles.
func TestPhase2Updates(t *testing.T) {
	requireServer(t)

	state := loadState()
	if state.OrgID == "" {
		t.Skip("Phase 1 state not found — run TestPhase1APIKey first")
	}
	orgID := state.OrgID
	ts := timestamp()

	// ── 2.1: Verify static JSON files exist (created by Phase 1) ──
	var versionData map[string]any
	t.Run("version_json_exists", func(t *testing.T) {
		data, err := os.ReadFile(filepath.Join(dataDir, "version.json"))
		if err != nil {
			t.Fatalf("read version.json: %v", err)
		}
		if err := json.Unmarshal(data, &versionData); err != nil {
			t.Fatalf("parse version.json: %v", err)
		}
		v, _ := versionData["version"].(string)
		t.Logf("version.json -> version=%s", v)
	})

	var changelogData map[string]any
	t.Run("changelog_json_exists", func(t *testing.T) {
		data, err := os.ReadFile(filepath.Join(dataDir, "changelog.json"))
		if err != nil {
			t.Fatalf("read changelog.json: %v", err)
		}
		if err := json.Unmarshal(data, &changelogData); err != nil {
			t.Fatalf("parse changelog.json: %v", err)
		}
		versions, _ := changelogData["versions"].([]any)
		t.Logf("changelog.json -> %d versions", len(versions))
	})

	// ── 2.2: Public OTA endpoints ──
	t.Run("GET_api_v1_version", func(t *testing.T) {
		status, body := doRaw(t, httpClient, "GET", "/api/v1/version", nil)
		if status != 200 {
			t.Fatalf("GET /api/v1/version -> %d: %s", status, string(body))
		}
		data := parseJSON(t, body)
		v, _ := data["version"].(string)
		expectedV, _ := versionData["version"].(string)
		if v != expectedV {
			t.Errorf("version mismatch: got %s, want %s", v, expectedV)
		}
		t.Logf("GET /api/v1/version -> version=%s", v)
	})

	t.Run("GET_api_v1_changelog", func(t *testing.T) {
		status, body := doRaw(t, httpClient, "GET", "/api/v1/changelog", nil)
		if status != 200 {
			t.Fatalf("GET /api/v1/changelog -> %d: %s", status, string(body))
		}
		var data map[string]any
		_ = json.Unmarshal(body, &data)
		versions, _ := data["versions"].([]any)
		if versions == nil {
			// Maybe it's a raw array
			var arr []any
			if json.Unmarshal(body, &arr) == nil {
				versions = arr
			}
		}
		if len(versions) < 3 {
			t.Errorf("expected >=3 versions, got %d", len(versions))
		}
		t.Logf("GET /api/v1/changelog -> %d versions", len(versions))
	})

	t.Run("GET_check_update_no_code", func(t *testing.T) {
		status, body := doRaw(t, httpClient, "GET", "/api/v1/check-update", nil)
		if status != 200 {
			t.Fatalf("GET /api/v1/check-update -> %d: %s", status, string(body))
		}
		data := parseJSON(t, body)
		ua, _ := data["update_available"].(bool)
		if !ua {
			t.Error("expected update_available=true with no version_code")
		}
		t.Log("GET /api/v1/check-update (no code) -> update_available=true")
	})

	t.Run("GET_check_update_current_code", func(t *testing.T) {
		code := versionData["version_code"]
		path := fmt.Sprintf("/api/v1/check-update?version_code=%v", code)
		status, body := doRaw(t, httpClient, "GET", path, nil)
		if status != 200 {
			t.Fatalf("GET /api/v1/check-update (current) -> %d: %s", status, string(body))
		}
		data := parseJSON(t, body)
		ua, _ := data["update_available"].(bool)
		if ua {
			t.Error("expected update_available=false with current version_code")
		}
		t.Log("GET /api/v1/check-update (current code) -> update_available=false")
	})

	t.Run("GET_check_update_older_code", func(t *testing.T) {
		code := versionData["version_code"]
		var olderCode int64
		switch v := code.(type) {
		case int64:
			olderCode = v - 1
		case float64:
			olderCode = int64(v) - 1
		default:
			t.Fatalf("unexpected version_code type: %T", code)
		}
		path := fmt.Sprintf("/api/v1/check-update?version_code=%d", olderCode)
		status, body := doRaw(t, httpClient, "GET", path, nil)
		if status != 200 {
			t.Fatalf("GET /api/v1/check-update (older) -> %d: %s", status, string(body))
		}
		data := parseJSON(t, body)
		ua, _ := data["update_available"].(bool)
		if !ua {
			t.Error("expected update_available=true with older version_code")
		}
		t.Log("GET /api/v1/check-update (older code) -> update_available=true")
	})

	// ── 2.3: Insert version records into Turso DB ──
	nowMs := timestampMs()
	dbVersions := []map[string]any{
		{"id": newUUID(), "version": "3.0.0", "apk_filename": "vyzorix-3.0.0.apk", "apk_size": 16777216, "sha256": repeatStr("f", 64), "release_date": nowMs, "release_notes": "Major 3.0 release", "release_type": "major", "is_latest": 1, "created_at": nowMs, "updated_at": nowMs},
		{"id": newUUID(), "version": "2.1.0", "apk_filename": "vyzorix-2.1.0.apk", "apk_size": 15728640, "sha256": repeatStr("e", 64), "release_date": nowMs - 86400000, "release_notes": "Feature release", "release_type": "minor", "is_latest": 0, "created_at": nowMs - 86400000, "updated_at": nowMs - 86400000},
		{"id": newUUID(), "version": "2.0.0", "apk_filename": "vyzorix-2.0.0.apk", "apk_size": 14680064, "sha256": repeatStr("d", 64), "release_date": nowMs - 172800000, "release_notes": "Stable release", "release_type": "minor", "is_latest": 0, "created_at": nowMs - 172800000, "updated_at": nowMs - 172800000},
	}

	for _, v := range dbVersions {
		v := v
		t.Run(fmt.Sprintf("insert_version_%s", v["version"]), func(t *testing.T) {
			tursoExec(t, []map[string]any{{
				"q": "INSERT INTO update_versions (id, version, apk_filename, apk_size, sha256, release_date, release_notes, release_type, is_latest, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
				"params": []any{v["id"], v["version"], v["apk_filename"], fmt.Sprintf("%d", v["apk_size"]), v["sha256"], fmt.Sprintf("%d", v["release_date"]), v["release_notes"], v["release_type"], fmt.Sprintf("%d", v["is_latest"]), fmt.Sprintf("%d", v["created_at"]), fmt.Sprintf("%d", v["updated_at"])},
			}})
			t.Logf("inserted version %s (is_latest=%v)", v["version"], v["is_latest"])
		})
	}

	// ── 2.4: Tenant updates endpoints ──
	time.Sleep(3 * time.Second) // avoid rate limit

	// Create a fresh read key to avoid exhausted rate limit bucket
	readKey := state.APIKeys["read"]
	if rk, ok := state.APIKeys["read"]; ok {
		readKey = rk
	}
	t.Run("create_fresh_read_key", func(t *testing.T) {
		sess, err := newSession()
		if err != nil {
			t.Fatalf("new session: %v", err)
		}
		csrf := getCSRF(t, sess)
		status, body := doJSON(t, sess, "POST", "/v1/auth/login",
			map[string]string{"X-CSRF-Token": csrf},
			map[string]any{"email": state.Email, "password": testPassword})
		if status != 200 {
			t.Fatalf("login -> %d: %s", status, string(body))
		}
		doJSON(t, sess, "POST", "/v1/auth/organizations/select",
			map[string]string{"X-CSRF-Token": csrf},
			map[string]any{"organization_id": orgID})
		key := createAPIKey(t, sess, csrf, orgID, fmt.Sprintf("phase2-key-%d", ts), "read")
		readKey = key
		state.APIKeys["read"] = key
		saveState(state)
		t.Logf("created fresh read key -> %s...", key[:16])
	})

	hdr := map[string]string{"X-API-Key": readKey, "X-Organization-ID": orgID}

	t.Run("GET_v1_updates_versions", func(t *testing.T) {
		time.Sleep(500 * time.Millisecond)
		status, body := doRaw(t, httpClient, "GET", "/v1/updates/versions", hdr)
		if status != 200 {
			t.Fatalf("GET /v1/updates/versions -> %d: %s", status, string(body))
		}
		var data map[string]any
		var versions []any
		if json.Unmarshal(body, &data) == nil {
			versions, _ = data["versions"].([]any)
			if versions == nil {
				versions, _ = data["items"].([]any)
			}
		}
		if json.Unmarshal(body, &versions) == nil && versions == nil {
			// raw array
		}
		if len(versions) < 3 {
			t.Errorf("expected >=3 versions, got %d", len(versions))
		}
		t.Logf("GET /v1/updates/versions -> %d versions", len(versions))
	})

	t.Run("GET_v1_updates_status", func(t *testing.T) {
		time.Sleep(500 * time.Millisecond)
		status, body := doRaw(t, httpClient, "GET", "/v1/updates/status", hdr)
		if status != 200 {
			t.Fatalf("GET /v1/updates/status -> %d: %s", status, string(body))
		}
		data := parseJSON(t, body)
		latest, _ := data["latest_version"].(map[string]any)
		if latest == nil {
			latest, _ = data["latest"].(map[string]any)
		}
		if latest != nil {
			v, _ := latest["version"].(string)
			t.Logf("GET /v1/updates/status -> latest_version=%s", v)
		} else {
			t.Logf("GET /v1/updates/status -> 200 (keys: %v)", mapKeys(data))
		}
	})

	t.Run("GET_v1_updates_changelog", func(t *testing.T) {
		time.Sleep(500 * time.Millisecond)
		status, body := doRaw(t, httpClient, "GET", "/v1/updates/changelog", hdr)
		if status != 200 {
			t.Fatalf("GET /v1/updates/changelog -> %d: %s", status, string(body))
		}
		t.Logf("GET /v1/updates/changelog -> 200")
	})

	t.Run("GET_v1_updates_history_before_push", func(t *testing.T) {
		time.Sleep(500 * time.Millisecond)
		status, body := doRaw(t, httpClient, "GET", "/v1/updates/history", hdr)
		if status != 200 {
			t.Fatalf("GET /v1/updates/history -> %d: %s", status, string(body))
		}
		var data map[string]any
		var history []any
		json.Unmarshal(body, &data)
		history, _ = data["pushes"].([]any)
		if history == nil {
			history, _ = data["items"].([]any)
		}
		if history == nil {
			history, _ = data["history"].([]any)
		}
		t.Logf("GET /v1/updates/history -> %d pushes (before push)", len(history))
	})

	// ── 2.5: Push update + verify history ──
	deviceID := newUUID()
	imei := fmt.Sprintf("TEST%dDEVICE1", ts)

	t.Run("insert_test_device", func(t *testing.T) {
		tursoExec(t, []map[string]any{{
			"q": "INSERT INTO devices (id, imei, organization_id, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
			"params": []any{deviceID, imei, orgID, "active", fmt.Sprintf("%d", nowMs), fmt.Sprintf("%d", nowMs)},
		}})
		t.Logf("inserted test device imei=%s", imei)
	})

	var pushID string
	t.Run("fresh_login_for_push", func(t *testing.T) {
		sess, err := newSession()
		if err != nil {
			t.Fatalf("new session: %v", err)
		}
		csrf := getCSRF(t, sess)
		status, body := doJSON(t, sess, "POST", "/v1/auth/login",
			map[string]string{"X-CSRF-Token": csrf},
			map[string]any{"email": state.Email, "password": testPassword})
		if status != 200 {
			t.Fatalf("login -> %d: %s", status, string(body))
		}
		data := parseJSON(t, body)
		loginEmail, _ := data["email"].(string)
		doJSON(t, sess, "POST", "/v1/auth/organizations/select",
			map[string]string{"X-CSRF-Token": csrf},
			map[string]any{"organization_id": orgID})
		t.Logf("fresh login for push -> %s", loginEmail)
	})

	t.Run("POST_v1_updates_push", func(t *testing.T) {
		sess, err := newSession()
		if err != nil {
			t.Fatalf("new session: %v", err)
		}
		csrf := getCSRF(t, sess)
		status, body := doJSON(t, sess, "POST", "/v1/auth/login",
			map[string]string{"X-CSRF-Token": csrf},
			map[string]any{"email": state.Email, "password": testPassword})
		if status != 200 {
			t.Fatalf("login -> %d: %s", status, string(body))
		}
		doJSON(t, sess, "POST", "/v1/auth/organizations/select",
			map[string]string{"X-CSRF-Token": csrf},
			map[string]any{"organization_id": orgID})

		status, body = doJSON(t, sess, "POST", "/v1/updates/push",
			map[string]string{"X-CSRF-Token": csrf, "X-Organization-ID": orgID},
			map[string]any{"version": "3.0.0", "installType": "immediate", "deviceIds": []string{imei}})
		if status != 200 && status != 201 && status != 202 {
			t.Fatalf("POST /v1/updates/push -> %d: %s", status, string(body))
		}
		data := parseJSON(t, body)
		pushID, _ = data["push_id"].(string)
		if pushID == "" {
			pushID, _ = data["id"].(string)
		}
		if pushID == "" {
			pushID, _ = data["pushId"].(string)
		}
		statusVal, _ := data["status"].(string)
		t.Logf("POST /v1/updates/push -> push_id=%s... status=%s", truncStr(pushID, 16), statusVal)
	})

	t.Run("GET_v1_updates_history_after_push", func(t *testing.T) {
		if pushID == "" {
			t.Skip("no push_id from previous test")
		}
		time.Sleep(2 * time.Second)
		status, body := doRaw(t, httpClient, "GET", "/v1/updates/history", hdr)
		if status != 200 {
			t.Fatalf("GET /v1/updates/history -> %d: %s", status, string(body))
		}
		var data map[string]any
		var history []any
		json.Unmarshal(body, &data)
		history, _ = data["pushes"].([]any)
		if history == nil {
			history, _ = data["items"].([]any)
		}
		if history == nil {
			history, _ = data["history"].([]any)
		}
		if len(history) < 1 {
			t.Errorf("expected >=1 push, got %d", len(history))
		}
		t.Logf("GET /v1/updates/history -> %d pushes (after push)", len(history))
	})

	t.Run("GET_v1_updates_history_detail", func(t *testing.T) {
		if pushID == "" {
			t.Skip("no push_id from previous test")
		}
		time.Sleep(500 * time.Millisecond)
		status, body := doRaw(t, httpClient, "GET", "/v1/updates/history/"+pushID, hdr)
		if status != 200 {
			t.Fatalf("GET /v1/updates/history/%s -> %d: %s", pushID, status, string(body))
		}
		t.Logf("GET /v1/updates/history/%s... -> 200", truncStr(pushID, 12))
	})

	// ── 2.6: Sync status + export ──
	t.Run("GET_v1_updates_sync_status", func(t *testing.T) {
		time.Sleep(3 * time.Second)
		status, body := doRaw(t, httpClient, "GET", "/v1/updates/sync/status", hdr)
		if status != 200 {
			t.Fatalf("GET /v1/updates/sync/status -> %d: %s", status, string(body))
		}
		t.Log("GET /v1/updates/sync/status -> 200")
	})

	t.Run("GET_v1_updates_export", func(t *testing.T) {
		time.Sleep(500 * time.Millisecond)
		status, body := doRaw(t, httpClient, "GET", "/v1/updates/export", hdr)
		if status != 200 {
			t.Fatalf("GET /v1/updates/export -> %d: %s", status, string(body))
		}
		t.Logf("GET /v1/updates/export -> 200 (size=%db)", len(body))
	})

	// ── 2.7: Data integrity ──
	t.Run("data_integrity_versions_present", func(t *testing.T) {
		time.Sleep(3 * time.Second)
		status, body := doRaw(t, httpClient, "GET", "/v1/updates/versions", hdr)
		if status != 200 {
			t.Fatalf("versions integrity check -> %d", status)
		}
		var data map[string]any
		var versions []any
		json.Unmarshal(body, &data)
		versions, _ = data["versions"].([]any)
		if versions == nil {
			versions, _ = data["items"].([]any)
		}
		vNums := map[string]bool{}
		for _, v := range versions {
			if m, ok := v.(map[string]any); ok {
				if vn, ok := m["version"].(string); ok {
					vNums[vn] = true
				}
			}
		}
		if !vNums["3.0.0"] || !vNums["2.1.0"] {
			t.Errorf("missing versions: have %v", vNums)
		}
		t.Log("data integrity: versions include 3.0.0 and 2.1.0")
	})

	t.Run("data_integrity_latest_marked", func(t *testing.T) {
		time.Sleep(500 * time.Millisecond)
		status, body := doRaw(t, httpClient, "GET", "/v1/updates/versions", hdr)
		if status != 200 {
			t.Fatalf("versions latest check -> %d", status)
		}
		var data map[string]any
		var versions []any
		json.Unmarshal(body, &data)
		versions, _ = data["versions"].([]any)
		if versions == nil {
			versions, _ = data["items"].([]any)
		}
		found := false
		for _, v := range versions {
			if m, ok := v.(map[string]any); ok {
				vn, _ := m["version"].(string)
				if vn == "3.0.0" {
					if isLatest, _ := m["is_latest"].(bool); isLatest {
						found = true
					}
					if status, _ := m["status"].(string); status == "latest" {
						found = true
					}
				}
			}
		}
		if !found {
			t.Error("3.0.0 not marked as latest")
		}
		t.Log("data integrity: 3.0.0 marked as latest")
	})

	// ── Update state ──
	state.PushID = pushID
	state.DeviceIMEI = imei
	state.DeviceID = deviceID
	if v, ok := versionData["version"].(string); ok {
		state.Version = v
	}
	saveState(state)
}

// repeatStr repeats s n times.
func repeatStr(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}

// truncStr truncates s to maxLen, appending "..." if truncated.
func truncStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// mapKeys returns the keys of a map[string]any for logging.
func mapKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
