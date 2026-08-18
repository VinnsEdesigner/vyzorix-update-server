<<<<<<< HEAD

=======
>>>>>>> 34b853d (feat: production hardening — structured errors, risk/audit, confirmation flow, validation, security hardening)
package errors

import (
	stderrors "errors"
	"fmt"
	"testing"
)

func TestValidationError_Error(t *testing.T) {
	tests := []struct {
		name string
		e    *ValidationError
		want string
	}{
		{"nil", nil, "validation failed"},
		{"empty details", &ValidationError{}, "validation failed"},
		{"one detail", &ValidationError{Details: []ValidationDetail{{Field: "email", Message: "invalid"}}},
			"validation failed: email: invalid"},
		{"many details", &ValidationError{Details: []ValidationDetail{{Field: "a", Message: "x"}, {Field: "b", Message: "y"}}},
			"validation failed: 2 field error(s)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.Error(); got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNewValidationError(t *testing.T) {
	details := []ValidationDetail{{Field: "f", Message: "m"}}
	ve := NewValidationError(details)
	if ve == nil || len(ve.Details) != 1 || ve.Details[0].Field != "f" {
		t.Fatalf("unexpected validation error: %+v", ve)
	}
}

func TestValidationDetailsOf_DirectAndWrapped(t *testing.T) {
	ve := NewValidationError([]ValidationDetail{{Field: "email", Message: "bad format"}})

	// Direct.
	details, ok := ValidationDetailsOf(ve)
	if !ok {
		t.Fatal("expected ok for a ValidationError")
	}
	if len(details) != 1 || details[0].Field != "email" {
		t.Errorf("unexpected details: %+v", details)
	}

	// Wrapped — the error middleware wraps recorded errors; ensure As still finds them.
	wrapped := fmt.Errorf("handler rejected: %w", ve)
	details, ok = ValidationDetailsOf(wrapped)
	if !ok {
		t.Fatal("expected ok for a wrapped ValidationError")
	}
	if len(details) != 1 || details[0].Field != "email" {
		t.Errorf("unexpected wrapped details: %+v", details)
	}
}

func TestValidationDetailsOf_NonValidationError(t *testing.T) {
	if _, ok := ValidationDetailsOf(stderrors.New("some other error")); ok {
		t.Error("expected ok=false for a non-validation error")
	}
	if _, ok := ValidationDetailsOf(nil); ok {
		t.Error("expected ok=false for nil error")
	}
<<<<<<< HEAD
}
=======
}
>>>>>>> 34b853d (feat: production hardening — structured errors, risk/audit, confirmation flow, validation, security hardening)
