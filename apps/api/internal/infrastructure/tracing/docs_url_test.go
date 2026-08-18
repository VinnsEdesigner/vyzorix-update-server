<<<<<<< HEAD

=======
>>>>>>> 34b853d (feat: production hardening — structured errors, risk/audit, confirmation flow, validation, security hardening)
package tracing

import "testing"

func TestBuildErrorDocsURLUsesDefaultBase(t *testing.T) {
	got := BuildErrorDocsURL("AUTH_INVALID_CREDENTIALS")
	want := DefaultDocsBaseURL + "/AUTH_INVALID_CREDENTIALS"
	if got != want {
		t.Errorf("BuildErrorDocsURL = %q, want %q", got, want)
	}
}

func TestBuildErrorDocsURLRespectsOverride(t *testing.T) {
	original := DefaultDocsURLBuilder.GetBaseURL()
	defer DefaultDocsURLBuilder.SetBaseURL(original)

	SetDocsBaseURL("https://staging.docs.example.com/errors")
	if got := BuildErrorDocsURL("RESOURCE_NOT_FOUND"); got != "https://staging.docs.example.com/errors/RESOURCE_NOT_FOUND" {
		t.Errorf("BuildErrorDocsURL = %q, want overridden base", got)
	}
}

func TestBuildErrorDocsURLFallsBackToDefaultWhenCleared(t *testing.T) {
	original := DefaultDocsURLBuilder.GetBaseURL()
	defer DefaultDocsURLBuilder.SetBaseURL(original)

	DefaultDocsURLBuilder.SetBaseURL("")
	if got := BuildErrorDocsURL("INTERNAL_SERVER_ERROR"); got != DefaultDocsBaseURL+"/INTERNAL_SERVER_ERROR" {
		t.Errorf("BuildErrorDocsURL with empty base = %q, want default %q", got, DefaultDocsBaseURL+"/INTERNAL_SERVER_ERROR")
	}
<<<<<<< HEAD
}
=======
}
>>>>>>> 34b853d (feat: production hardening — structured errors, risk/audit, confirmation flow, validation, security hardening)
