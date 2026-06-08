package security

import (
	"strings"
	"testing"
)

func TestValidatePassword_Valid(t *testing.T) {
	tests := []struct {
		name     string
		password string
		policy   PasswordPolicy
	}{
		// DefaultPasswordPolicy tests
		{"minimum valid default", "Password1!", DefaultPasswordPolicy},
		{"complex password default", "MyP@ssw0rd!2024", DefaultPasswordPolicy},
		{"with special chars default", "Test@123Abc!", DefaultPasswordPolicy},
		{"maximum length default", strings.Repeat("A", 100) + "a1!", DefaultPasswordPolicy},

		// UserPasswordPolicy tests (no special char required, min 12 chars)
		{"minimum valid user", "Password1234", UserPasswordPolicy}, // exactly 12 chars
		{"alphanumeric user", "MySecurePassword99", UserPasswordPolicy},
		{"long password user", strings.Repeat("A", 50) + "a1", UserPasswordPolicy},
		{"with special user", "MyP@ssw0rd!2024", UserPasswordPolicy}, // special chars allowed
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password, tt.policy)
			if err != nil {
				t.Errorf("expected valid password, got error: %v", err)
			}
		})
	}
}

func TestValidatePassword_TooShort(t *testing.T) {
	tests := []struct {
		name     string
		password string
		policy   PasswordPolicy
	}{
		// DefaultPasswordPolicy (8 char min)
		{"7 chars default", "Pass1!", DefaultPasswordPolicy},
		{"6 chars default", "Ab1!", DefaultPasswordPolicy},
		{"empty default", "", DefaultPasswordPolicy},
		{"7 chars user", "Pass1!", UserPasswordPolicy},

		// UserPasswordPolicy (12 char min)
		{"11 chars user", "Password1!", UserPasswordPolicy}, // missing 1 char
		{"8 chars user", "PassWord1", UserPasswordPolicy},   // too short
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password, tt.policy)
			if err == nil {
				t.Errorf("expected error for short password %q", tt.password)
			}
			pe, ok := err.(*PasswordError)
			if !ok {
				t.Errorf("expected PasswordError, got %T", err)
			}
			if pe != nil && !strings.Contains(pe.Error(), "minimum") {
				t.Errorf("expected 'minimum' in error, got: %v", pe.Error())
			}
		})
	}
}

func TestValidatePassword_NoUppercase(t *testing.T) {
	err := ValidatePassword("password1!", DefaultPasswordPolicy)
	if err == nil {
		t.Error("expected error for missing uppercase (default policy)")
	}

	// User policy also requires uppercase
	err = ValidatePassword("password123456", UserPasswordPolicy)
	if err == nil {
		t.Error("expected error for missing uppercase (user policy)")
	}
}

func TestValidatePassword_NoLowercase(t *testing.T) {
	err := ValidatePassword("PASSWORD1!", DefaultPasswordPolicy)
	if err == nil {
		t.Error("expected error for missing lowercase (default policy)")
	}

	err = ValidatePassword("PASSWORD1234", UserPasswordPolicy)
	if err == nil {
		t.Error("expected error for missing lowercase (user policy)")
	}
}

func TestValidatePassword_NoDigit(t *testing.T) {
	err := ValidatePassword("Password!", DefaultPasswordPolicy)
	if err == nil {
		t.Error("expected error for missing digit (default policy)")
	}

	err = ValidatePassword("Password!", UserPasswordPolicy)
	if err == nil {
		t.Error("expected error for missing digit (user policy)")
	}
}

func TestValidatePassword_NoSpecial(t *testing.T) {
	// DefaultPasswordPolicy requires special chars
	err := ValidatePassword("Password1", DefaultPasswordPolicy)
	if err == nil {
		t.Error("expected error for missing special character (default policy)")
	}

	// UserPasswordPolicy does NOT require special chars (but needs 12+ chars)
	err = ValidatePassword("Password1234", UserPasswordPolicy) // 12 chars, no special
	if err != nil {
		t.Errorf("UserPasswordPolicy should NOT require special chars, got error: %v", err)
	}
}

func TestValidatePassword_TooLong(t *testing.T) {
	longPassword := strings.Repeat("A", 130) + "a1!"
	err := ValidatePassword(longPassword, DefaultPasswordPolicy)
	if err == nil {
		t.Error("expected error for too long password (default policy)")
	}

	err = ValidatePassword(longPassword, UserPasswordPolicy)
	if err == nil {
		t.Error("expected error for too long password (user policy)")
	}
}

func TestValidatePassword_MultipleFailures(t *testing.T) {
	// DefaultPasswordPolicy: short, no upper, no digit, no special
	err := ValidatePassword("short", DefaultPasswordPolicy)
	if err == nil {
		t.Error("expected error for multiple failures (default policy)")
	}
	pe, ok := err.(*PasswordError)
	if !ok {
		t.Fatalf("expected PasswordError, got %T", err)
	}
	if len(pe.Missing) < 4 {
		t.Errorf("expected at least 4 failures, got %d: %v", len(pe.Missing), pe.Missing)
	}

	// UserPasswordPolicy: short, no upper, no digit (but no special required)
	err = ValidatePassword("short", UserPasswordPolicy)
	if err == nil {
		t.Error("expected error for multiple failures (user policy)")
	}
	pe, ok = err.(*PasswordError)
	if !ok {
		t.Fatalf("expected PasswordError, got %T", err)
	}
	if len(pe.Missing) < 3 {
		t.Errorf("expected at least 3 failures (no special), got %d: %v", len(pe.Missing), pe.Missing)
	}
}

func TestValidatePassword_CustomPolicy(t *testing.T) {
	// Custom policy with minimal requirements
	policy := PasswordPolicy{
		MinLength:      6,
		MaxLength:      50,
		RequireUpper:   false,
		RequireLower:   false,
		RequireDigit:   false,
		RequireSpecial: false,
	}

	// All passwords should pass
	err := ValidatePassword("simple", policy)
	if err != nil {
		t.Errorf("expected valid with custom policy, got: %v", err)
	}
}

func TestUserPasswordPolicy_NoSpecialRequired(t *testing.T) {
	// These should all pass with UserPasswordPolicy (12+ chars, mixed case, number)
	validPasswords := []string{
		"Password1234",    // basic alphanumeric, 12 chars
		"MySecurePass99",  // longer alphanumeric
		"MixedCase12345",  // mixed case + numbers
		"Another12345678", // different alphanumeric
		"TestPassword99",  // 14 chars
	}

	for _, pwd := range validPasswords {
		err := ValidatePassword(pwd, UserPasswordPolicy)
		if err != nil {
			t.Errorf("UserPasswordPolicy should accept %q, got: %v", pwd, err)
		}
	}
}

func TestPasswordStrength(t *testing.T) {
	tests := []struct {
		password string
		minScore int
		maxScore int
	}{
		{"short", 0, 0},                 // < 8 chars
		{"password", 1, 1},              // 8 chars, no other features
		{"Password1!", 3, 3},            // 12 chars, mixed case, digit, special
		{"MyVerySecureP@ssw0rd!", 5, 5}, // 20+ chars, all features (5 points)
		{"abc", 0, 0},                   // too short
		{"ABC123!", 1, 1},               // 7 chars, only uppercase + special
		{"Password1234", 4, 4},          // 14 chars, mixed case, digit (no special)
		{"MyVeryLongPassword16!", 5, 5}, // 16+ chars, all features (5 points)
		{"LongPassword1234", 5, 5},      // 16 chars, all features (5 points)
	}

	for _, tt := range tests {
		t.Run(tt.password, func(t *testing.T) {
			score := PasswordStrength(tt.password)
			if score < tt.minScore || score > tt.maxScore {
				t.Errorf("PasswordStrength(%q) = %d, want between %d and %d",
					tt.password, score, tt.minScore, tt.maxScore)
			}
		})
	}
}

func TestContainsUpper(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"Password", true},
		{"password", false},
		{"PASSWORD", true},
		{"Pass123", true},
		{"123456!", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := containsUpper(tt.input)
			if result != tt.expected {
				t.Errorf("containsUpper(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestContainsLower(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"password", true},
		{"PASSWORD", false},
		{"Password", true},
		{"PASSWORD123", false}, // No lowercase letters
		{"123456!", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := containsLower(tt.input)
			if result != tt.expected {
				t.Errorf("containsLower(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestContainsDigit(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"password1", true},
		{"password", false},
		{"123456", true},
		{"Pass1word", true},
		{"Password!", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := containsDigit(tt.input)
			if result != tt.expected {
				t.Errorf("containsDigit(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestContainsSpecial(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"password!", true},
		{"password", false},
		{"@#$%", true},
		{"Pass1!", true},
		{"Password1", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := containsSpecial(tt.input)
			if result != tt.expected {
				t.Errorf("containsSpecial(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestPasswordError_Error(t *testing.T) {
	pe := &PasswordError{
		Missing: []string{"minimum 8 characters", "at least 1 uppercase letter"},
	}
	errStr := pe.Error()
	if !strings.Contains(errStr, "minimum 8 characters") {
		t.Errorf("expected 'minimum 8 characters' in error, got: %s", errStr)
	}
	if !strings.Contains(errStr, "uppercase") {
		t.Errorf("expected 'uppercase' in error, got: %s", errStr)
	}
}
