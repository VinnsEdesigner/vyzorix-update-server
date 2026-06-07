package security

import (
	"testing"
)

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{"valid email", "test@example.com", false},
		{"valid with dots", "user.name@domain.co.uk", false},
		{"valid with plus", "user+tag@example.com", false},
		{"empty email", "", true},
		{"no at sign", "testexample.com", true},
		{"no domain", "test@", true},
		{"no local part", "@example.com", true},
		{"spaces", " test@example.com ", false},  // trimmed
		{"uppercase", "Test@Example.COM", false}, // lowercased
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateEmail(tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEmail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == "" {
				t.Error("ValidateEmail() returned empty string on success")
			}
		})
	}
}

func TestValidateEmailMaxLength(t *testing.T) {
	longEmail := make([]byte, MaxEmailLength+1)
	for i := range longEmail {
		longEmail[i] = 'a'
	}
	longEmail[0] = 't'
	longEmail[1] = '@'
	longEmail[2] = 'a'
	longEmail[3] = '.'

	_, err := ValidateEmail(string(longEmail))
	if err == nil {
		t.Error("ValidateEmail() should reject email exceeding max length")
	}
}

func TestValidateName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid name", "John Doe", false},
		{"single word", "John", false},
		{"empty", "", true},
		{"spaces only", "   ", true},
		{"with trim", "  John  ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateNameMaxLength(t *testing.T) {
	longName := make([]byte, MaxNameLength+1)
	for i := range longName {
		longName[i] = 'a'
	}

	_, err := ValidateName(string(longName))
	if err == nil {
		t.Error("ValidateName() should reject name exceeding max length")
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name    string
		pw      string
		wantErr bool
	}{
		{"valid password", "password123", false},
		{"minimum length", "12345678", false},
		{"too short", "1234567", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePasswordLength(tt.pw)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePasswordLength() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePasswordMaxLength(t *testing.T) {
	longPw := make([]byte, MaxPasswordLength+1)
	for i := range longPw {
		longPw[i] = 'a'
	}

	err := ValidatePasswordLength(string(longPw))
	if err == nil {
		t.Error("ValidatePasswordLength() should reject password exceeding max length")
	}
}

func TestValidateDeviceID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{"valid alphanumeric", "device123", false},
		{"valid with hyphen", "device-123", false},
		{"valid with underscore", "device_123", false},
		{"empty", "", true},
		{"with spaces", "device 123", true},
		{"with special chars", "device@123", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateDeviceID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDeviceID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateCommand(t *testing.T) {
	tests := []struct {
		name    string
		cmd     string
		wantErr bool
	}{
		{"valid command", "FORCE_SPEAKER", false},
		{"empty", "", true},
		{"whitespace", "   ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateCommand(tt.cmd)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		want   string
		maxLen int
	}{
		{"no truncation", "hello", "hello", 10},
		{"with truncation", "hello world", "hello", 5},
		{"exact length", "hello", "hello", 5},
		{"empty string", "", "", 5},
		{"whitespace trimmed", "  hello  ", "hello", 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeString(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("SanitizeString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestValidationError(t *testing.T) {
	err := &ValidationError{Field: "email", Message: "invalid format"}
	if err.Error() != "email: invalid format" {
		t.Errorf("ValidationError.Error() = %q, want %q", err.Error(), "email: invalid format")
	}
}
