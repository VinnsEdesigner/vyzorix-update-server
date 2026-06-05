package security

import (
	"strings"
	"testing"
)

func TestValidatePassword_Valid(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{"minimum valid", "Password1!"},
		{"complex password", "MyP@ssw0rd!2024"},
		{"with special chars", "Test@123Abc!"},
		{"maximum length", strings.Repeat("A", 100) + "a1!"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password, DefaultPasswordPolicy)
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
	}{
		{"7 chars", "Pass1!"},
		{"6 chars", "Ab1!"},
		{"empty", ""},
		{"just 8", "Passwor"}, // 7 chars
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password, DefaultPasswordPolicy)
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
		t.Error("expected error for missing uppercase")
	}
	pe, ok := err.(*PasswordError)
	if !ok {
		t.Fatalf("expected PasswordError, got %T", err)
	}
	found := false
	for _, m := range pe.Missing {
		if strings.Contains(m, "uppercase") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'uppercase' in missing list, got: %v", pe.Missing)
	}
}

func TestValidatePassword_NoLowercase(t *testing.T) {
	err := ValidatePassword("PASSWORD1!", DefaultPasswordPolicy)
	if err == nil {
		t.Error("expected error for missing lowercase")
	}
}

func TestValidatePassword_NoDigit(t *testing.T) {
	err := ValidatePassword("Password!", DefaultPasswordPolicy)
	if err == nil {
		t.Error("expected error for missing digit")
	}
}

func TestValidatePassword_NoSpecial(t *testing.T) {
	err := ValidatePassword("Password1", DefaultPasswordPolicy)
	if err == nil {
		t.Error("expected error for missing special character")
	}
}

func TestValidatePassword_TooLong(t *testing.T) {
	longPassword := strings.Repeat("A", 130) + "a1!"
	err := ValidatePassword(longPassword, DefaultPasswordPolicy)
	if err == nil {
		t.Error("expected error for too long password")
	}
}

func TestValidatePassword_MultipleFailures(t *testing.T) {
	err := ValidatePassword("short", DefaultPasswordPolicy) // no upper, no digit, no special, too short
	if err == nil {
		t.Error("expected error for multiple failures")
	}
	pe, ok := err.(*PasswordError)
	if !ok {
		t.Fatalf("expected PasswordError, got %T", err)
	}
	if len(pe.Missing) < 4 {
		t.Errorf("expected at least 4 failures, got %d: %v", len(pe.Missing), pe.Missing)
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

func TestPasswordStrength(t *testing.T) {
	tests := []struct {
		password string
		minScore int
		maxScore int
	}{
		{"short", 0, 0},
		{"password", 1, 1},
		{"Password1!", 3, 3},
		{"MyVerySecureP@ssw0rd!", 4, 4},
		{"abc", 0, 0},
		{"ABC123!", 2, 2},
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
		{"PASSWORD123", true},
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