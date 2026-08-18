package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

// TestPhase3Invite tests the invitation flow: admin creates an invitation,
// mock-resend captures the email, invitee registers + accepts, and we confirm.
// org-scoped access. 15 subtests (Phase 3 + Phase 4 from the Python harness).
func TestPhase3Invite(t *testing.T) {
	requireServer(t)
	requireMockEmail(t)

	state := loadState()
	if state.OrgID == "" || state.Email == "" {
		t.Skip("Phase 1 state not found — run TestPhase1APIKey first")
	}
	orgID := state.OrgID
	adminEmail := state.Email

// ── Phase 3: Admin login + create invitation ──.
	adminSess, err := newSession()
	if err != nil {
		t.Fatalf("new session: %v", err)
	}
	adminCSRF := getCSRF(t, adminSess)

	t.Run("admin_login", func(t *testing.T) {
		status, body := doJSON(t, adminSess, "POST", "/v1/auth/login",
			map[string]string{"X-CSRF-Token": adminCSRF},
			map[string]any{"email": adminEmail, "password": testPassword})
		if status != 200 {
			t.Fatalf("admin login -> %d: %s", status, string(body))
		}
		data := parseJSON(t, body)
		t.Logf("admin login -> %s", data["email"])
	})

	t.Run("admin_select_org", func(t *testing.T) {
		status, body := doJSON(t, adminSess, "POST", "/v1/auth/organizations/select",
			map[string]string{"X-CSRF-Token": adminCSRF},
			map[string]any{"organization_id": orgID})
		if status != 200 {
			t.Fatalf("admin select org -> %d: %s", status, string(body))
		}
		t.Logf("admin select org -> %s", orgID)
	})

	inviteeEmail := fmt.Sprintf("invitee_%d@vyzorix-test.local", timestampMs()%1000000)
	var invToken string

	t.Run("create_invitation", func(t *testing.T) {
		status, body := doJSON(t, adminSess, "POST", "/v1/invitations",
			map[string]string{"X-CSRF-Token": adminCSRF, "X-Organization-ID": orgID},
			map[string]any{"organizationId": orgID, "email": inviteeEmail, "role": "operator", "notes": "Phase 3 test invite"})
		if status != 200 && status != 201 {
			t.Fatalf("create invitation -> %d: %s", status, string(body))
		}
		data := parseJSON(t, body)
		invToken, _ = data["token"].(string)
		if invToken == "" {
			invToken, _ = data["invitationToken"].(string)
		}
		invID, _ := data["id"].(string)
		if invID == "" {
			invID, _ = data["invitationId"].(string)
		}
		t.Logf("create invitation -> email=%s role=operator id=%s...", inviteeEmail, truncStr(invID, 12))
	})

	t.Run("mock_resend_captured_invite_email", func(t *testing.T) {
		time.Sleep(2 * time.Second)
		emailLog := fetchInviteEmail(t, inviteeEmail)
		if emailLog == nil {
			t.Fatalf("no invitation email for %s", inviteeEmail)
		}
		t.Logf("mock-resend captured invitation email -> to=%v subj=%s", emailLog.To, truncStr(emailLog.Subject, 40))
	})

	var extractedToken string
	t.Run("extract_invite_token_from_email", func(t *testing.T) {
		emailLog := fetchInviteEmail(t, inviteeEmail)
		if emailLog == nil {
			t.Fatal("no invitation email found")
		}
		extractedToken = extractInviteToken(emailLog.HTML)
		if extractedToken == "" {
			t.Fatal("invite token not found in email HTML")
		}
		t.Logf("extracted invite token from email HTML -> %s...", truncStr(extractedToken, 24))
	})

	t.Run("api_token_matches_email_token", func(t *testing.T) {
		if invToken == "" || extractedToken == "" {
			t.Skip("missing API token or email token")
		}
		if invToken != extractedToken {
			t.Errorf("token mismatch: api=%s vs email=%s", truncStr(invToken, 16), truncStr(extractedToken, 16))
		}
		t.Logf("API token matches email token -> %s...", truncStr(invToken, 24))
	})

	finalToken := invToken
	if finalToken == "" {
		finalToken = extractedToken
	}

	t.Run("GET_invite_token", func(t *testing.T) {
		if finalToken == "" {
			t.Skip("no invite token available")
		}
		status, body := doJSON(t, adminSess, "GET", "/v1/invite/"+finalToken,
			map[string]string{"X-CSRF-Token": adminCSRF}, nil)
		if status != 200 {
			t.Fatalf("GET /v1/invite/:token -> %d: %s", status, string(body))
		}
		data := parseJSON(t, body)
		email, _ := data["email"].(string)
		statusVal, _ := data["status"].(string)
		t.Logf("GET /v1/invite/:token -> email=%s status=%s", email, statusVal)
	})

// ── Phase 4: Invitee registers + accepts invitation ──.
	var invSess *http.Client
	var invCSRF string

	t.Run("invitee_registered_and_verified", func(t *testing.T) {
		if finalToken == "" {
			t.Skip("no invite token available")
		}
		invSess, invCSRF = registerAndVerify(t, inviteeEmail, "Invitee Test")
		t.Logf("invitee registered + email verified -> %s", inviteeEmail)
	})

	t.Run("invitee_login", func(t *testing.T) {
		if invSess == nil {
			t.Skip("invitee not registered")
		}
		status, body := doJSON(t, invSess, "POST", "/v1/auth/login",
			map[string]string{"X-CSRF-Token": invCSRF},
			map[string]any{"email": inviteeEmail, "password": testPassword})
		if status != 200 {
			t.Fatalf("invitee login -> %d: %s", status, string(body))
		}
		data := parseJSON(t, body)
		t.Logf("invitee login -> %s", data["email"])
	})

	t.Run("invitee_accepts_invitation", func(t *testing.T) {
		if invSess == nil || finalToken == "" {
			t.Skip("missing invitee session or token")
		}
		status, body := doJSON(t, invSess, "POST", "/v1/invite/"+finalToken+"/accept",
			map[string]string{"X-CSRF-Token": invCSRF},
			map[string]any{"notes": "accepted via Phase 4 test"})
		if status != 200 {
			t.Fatalf("invitee accepts invitation -> %d: %s", status, string(body))
		}
		data := parseJSON(t, body)
		msg, _ := data["message"].(string)
		t.Logf("invitee accepts invitation -> message=%s", msg)
	})

	t.Run("invitee_sees_org_in_list", func(t *testing.T) {
		if invSess == nil {
			t.Skip("invitee not registered")
		}
// Refresh CSRF.
		invCSRF = getCSRF(t, invSess)
		status, body := doJSON(t, invSess, "GET", "/v1/organizations", nil, nil)
		if status != 200 {
			t.Fatalf("invitee list organizations -> %d: %s", status, string(body))
		}
		data := parseJSON(t, body)
		orgs, _ := data["organizations"].([]any)
		if orgs == nil {
// maybe it's a raw list.
			var rawList []map[string]any
			json.Unmarshal(body, &rawList)
			for _, o := range rawList {
				orgs = append(orgs, o)
			}
		}
		found := false
		for _, o := range orgs {
			if m, ok := o.(map[string]any); ok {
				if id, _ := m["id"].(string); id == orgID {
					found = true
					break
				}
			}
		}
		if !found {
			t.Errorf("target org %s not in invitee's org list", orgID)
		}
		t.Logf("invitee sees org in their organization list -> %d org(s), target org present", len(orgs))
	})

	t.Run("invitee_select_org", func(t *testing.T) {
		if invSess == nil {
			t.Skip("invitee not registered")
		}
		status, body := doJSON(t, invSess, "POST", "/v1/auth/organizations/select",
			map[string]string{"X-CSRF-Token": invCSRF},
			map[string]any{"organization_id": orgID})
		if status != 200 {
			t.Fatalf("invitee select org -> %d: %s", status, string(body))
		}
		t.Logf("invitee select org -> %s", orgID)
	})

	t.Run("invitee_access_org_scoped_endpoint", func(t *testing.T) {
		if invSess == nil {
			t.Skip("invitee not registered")
		}
		status, body := doJSON(t, invSess, "GET", "/v1/updates/versions",
			map[string]string{"X-Organization-ID": orgID}, nil)
		if status != 200 {
			t.Fatalf("invitee GET /v1/updates/versions -> %d: %s", status, string(body))
		}
		t.Logf("invitee GET /v1/updates/versions (session) -> %d", status)
	})

	t.Run("admin_sees_invitee_in_members", func(t *testing.T) {
		status, body := doJSON(t, adminSess, "GET", "/v1/organizations/"+orgID+"/members",
			map[string]string{"X-CSRF-Token": adminCSRF, "X-Organization-ID": orgID}, nil)
		if status != 200 {
			t.Fatalf("admin list org members -> %d: %s", status, string(body))
		}
		var members []map[string]any
		data := parseJSON(t, body)
		if items, ok := data["members"].([]any); ok {
			for _, item := range items {
				if m, ok := item.(map[string]any); ok {
					members = append(members, m)
				}
			}
		} else if items, ok := data["items"].([]any); ok {
			for _, item := range items {
				if m, ok := item.(map[string]any); ok {
					members = append(members, m)
				}
			}
		} else {
			json.Unmarshal(body, &members)
		}
		found := false
		var invMember map[string]any
		for _, m := range members {
			email := getStr(m, "operator_email", "email", "operatorEmail")
			if email == inviteeEmail {
				found = true
				invMember = m
				break
			}
		}
		if !found {
			t.Errorf("invitee %s not in member list (%d members)", inviteeEmail, len(members))
		}
		role := getStr(invMember, "role", "")
		statusVal := getStr(invMember, "status", "")
		t.Logf("admin sees invitee in org members -> %d members, invitee present role=%s status=%s", len(members), role, statusVal)
	})

	t.Run("invitation_status_accepted", func(t *testing.T) {
		time.Sleep(500 * time.Millisecond)
		status, body := doJSON(t, adminSess, "GET", "/v1/organizations/"+orgID+"/invitations",
			map[string]string{"X-CSRF-Token": adminCSRF, "X-Organization-ID": orgID}, nil)
		if status != 200 {
			t.Fatalf("admin list invitations -> %d: %s", status, string(body))
		}
		var invs []map[string]any
		data := parseJSON(t, body)
		if items, ok := data["invitations"].([]any); ok {
			for _, item := range items {
				if m, ok := item.(map[string]any); ok {
					invs = append(invs, m)
				}
			}
		} else if items, ok := data["items"].([]any); ok {
			for _, item := range items {
				if m, ok := item.(map[string]any); ok {
					invs = append(invs, m)
				}
			}
		} else {
			json.Unmarshal(body, &invs)
		}
		var target map[string]any
		for _, inv := range invs {
			if email, _ := inv["email"].(string); email == inviteeEmail {
				target = inv
				break
			}
		}
		if target == nil {
			t.Fatal("invitation not found in admin list")
		}
		st, _ := target["status"].(string)
		if st != "accepted" && st != "approved" {
			t.Errorf("invitation status = %s, expected accepted/approved", st)
		}
		t.Logf("invitation status = %s -> email=%s", st, inviteeEmail)
	})
}

// getStr retrieves a string value from a map, trying multiple keys.
func getStr(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k].(string); ok && v != "" {
			return v
		}
	}
	return ""
}

// Suppress unused import warning.
var _ = strings.Contains
