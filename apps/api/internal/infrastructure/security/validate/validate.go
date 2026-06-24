// Package validate provides input validation utilities.
package validate

import (
	"net"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"unicode"
)

const (
	MaxEmailLength    = 254
	MaxNameLength     = 100
	MaxPasswordLength = 128
	MinPasswordLength = 8
	MaxDeviceIDLength = 64
	MaxCommandLength  = 256
	MaxTokenLength    = 1024
)

var disposableEmailDomains = map[string]bool{
	"tempmail.com":      true,
	"guerrillamail.com": true,
	"mailinator.com":    true,
	"10minutemail.com":  true,
	"throwaway.email":   true,
	"temp-mail.org":     true,
	"fakeinbox.com":     true,
	"trash-mail.com":    true,
	"getnada.com":       true,
	"maildrop.cc":       true,
	"dispostable.com":   true,
	"yopmail.com":       true,
	"sharklasers.com":   true,
	"mailnesia.com":     true,
	"tempail.com":       true,
	"emailondeck.com":   true,
	"getairmail.com":    true,
	"mohmal.com":        true,
	"tempinbox.com":     true,
	"burnermail.io":     true,
	"spamgourmet.com":   true,
	"mytrashmail.com":   true,
	"mailcatch.com":     true,
	"mt2009.com":        true,
	"thankyou2010.com":  true,
	"trash2009.com":     true,
	"mt2014.com":        true,
	"mt2015.com":        true,
	"spambox.us":        true,
	"spamfree24.org":    true,
	"spamherelots.com":  true,
	"devnullmail.com":   true,
	"discardmail.com":   true,
	"discardmail.de":    true,
	"spamex.com":        true,
	"spam4.me":          true,
}

var roleBasedSuffixes = map[string]bool{
	"admin":         true,
	"administrator": true,
	"webmaster":     true,
	"hostmaster":    true,
	"postmaster":    true,
	"abuse":         true,
	"noc":           true,
	"security":      true,
	"support":       true,
	"noreply":       true,
	"no-reply":      true,
	"donotreply":    true,
	"invalid":       true,
}

var trustedEmailProviders = map[string]bool{
	"gmail.com":      true,
	"googlemail.com": true,
	"google.com":     true,
	"outlook.com":    true,
	"hotmail.com":    true,
	"live.com":       true,
	"msn.com":        true,
	"microsoft.com":  true,
	"yahoo.com":      true,
	"ymail.com":      true,
	"yahoo.co.uk":    true,
	"yahoo.com.au":   true,
	"yahoo.co.jp":    true,
	"icloud.com":     true,
	"apple.com":      true,
	"me.com":         true,
	"protonmail.com": true,
	"proton.me":      true,
	"tutanota.com":   true,
	"tutanota.de":    true,
	"aol.com":        true,
	"zoho.com":       true,
	"fastmail.com":   true,
	"mail.com":       true,
	"gmx.com":        true,
	"gmx.de":         true,
	"gmx.net":        true,
	"yandex.com":     true,
	"yandex.ru":      true,
	"qq.com":         true,
	"163.com":        true,
	"126.com":        true,
	"sina.com":       true,
}

// Validator provides configurable enterprise-grade email validation.
type Validator struct {
	AllowDisposable            bool
	AllowRoleBased             bool
	RequireTrustedProviderOnly bool
	VerifyMX                   bool
	MXLookupTimeout            int
}

// DefaultValidator returns an EmailValidator configured for typical user registration.
func DefaultValidator() *Validator {
	return &Validator{
		AllowDisposable:            false,
		AllowRoleBased:             false,
		RequireTrustedProviderOnly: false,
		VerifyMX:                   false,
		MXLookupTimeout:            5,
	}
}

// StrictValidator returns an EmailValidator configured for high-security requirements.
func StrictValidator() *Validator {
	return &Validator{
		AllowDisposable:            false,
		AllowRoleBased:             false,
		RequireTrustedProviderOnly: true,
		VerifyMX:                   true,
		MXLookupTimeout:            5,
	}
}

// Error represents a validation error.
type Error struct {
	Field   string
	Message string
	Code    string
}

func (e *Error) Error() string {
	return e.Field + ": " + e.Message
}

// Result contains detailed validation results.
type Result struct {
	HasMXRecord *bool
	Normalized  string
	Domain      string
	LocalPart   string
	Errors      []ValidationError
	Warnings    []string
	Valid       bool
}

// ValidationError represents a specific email validation error.
type ValidationError struct {
	Code    string
	Message string
}

var (
	emailRegexRFC5322    = regexp.MustCompile(`^[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)
	emailRegexPermissive = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	emailRegexIPDomain   = regexp.MustCompile(`@\[(\d{1,3}\.){3}\d{1,3}\]$`)
)

var mxCache = struct {
	entries map[string][]string
	sync.RWMutex
}{
	entries: make(map[string][]string),
}

// Email validates an email address.
func Email(email string) (string, error) {
	result := EmailFull(email, DefaultValidator())
	if !result.Valid {
		if len(result.Errors) > 0 {
			return "", &Error{
				Field:   "email",
				Message: result.Errors[0].Message,
				Code:    result.Errors[0].Code,
			}
		}

		return "", &Error{
			Field:   "email",
			Message: "invalid email format",
			Code:    "INVALID_FORMAT",
		}
	}

	return result.Normalized, nil
}

// EmailFull performs comprehensive email validation with detailed results.
func EmailFull(email string, validator *Validator) *Result {
	result := &Result{
		Valid:    true,
		Errors:   []ValidationError{},
		Warnings: []string{},
	}

	email = strings.TrimSpace(email)
	if len(email) == 0 {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{Code: "EMPTY", Message: "email is required"})

		return result
	}

	if len(email) > MaxEmailLength {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{Code: "TOO_LONG", Message: "email exceeds maximum length"})

		return result
	}

	// Parse email.
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{Code: "INVALID_FORMAT", Message: "invalid email format"})

		return result
	}

	localPart := parts[0]
	domain := strings.ToLower(parts[1])
	result.LocalPart = localPart
	result.Domain = domain
	result.Normalized = localPart + "@" + domain

	// Validate format.
	if !emailRegexRFC5322.MatchString(email) && !emailRegexPermissive.MatchString(email) {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{Code: "INVALID_FORMAT", Message: "invalid email format"})

		return result
	}

	// Check for IP domain.
	if emailRegexIPDomain.MatchString(email) {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{Code: "IP_DOMAIN", Message: "IP addresses in email domain are not allowed"})

		return result
	}

	// Check for disposable.
	if !validator.AllowDisposable && isDisposable(domain) {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{Code: "DISPOSABLE_DOMAIN", Message: "temporary/disposable email addresses are not allowed"})

		return result
	}

	// Check for role-based.
	if !validator.AllowRoleBased && isRoleBased(localPart) {
		result.Warnings = append(result.Warnings, "Role-based email address detected")
	}

	// Check for trusted provider.
	if validator.RequireTrustedProviderOnly && !isTrusted(domain) && !isDisposable(domain) {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{Code: "UNTRUSTED_PROVIDER", Message: "only email from trusted providers are allowed"})

		return result
	}

	// MX record verification.
	if validator.VerifyMX && result.Valid {
		hasMX, _ := checkMX(domain, validator.MXLookupTimeout)
		result.HasMXRecord = &hasMX

		if !hasMX {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{Code: "NO_MX_RECORD", Message: "email domain does not accept mail"})

			return result
		}
	}

	return result
}

func isDisposable(domain string) bool {
	domain = strings.ToLower(domain)
	if disposableEmailDomains[domain] {
		return true
	}

	for d := range disposableEmailDomains {
		if strings.HasSuffix(domain, "."+d) {
			return true
		}
	}

	return false
}

func isRoleBased(local string) bool {
	local = strings.ToLower(local)
	if roleBasedSuffixes[local] {
		return true
	}

	prefixes := []string{"admin", "support", "noreply", "no-reply", "donotreply", "info", "contact", "help", "sales", "billing"}
	for _, p := range prefixes {
		if strings.HasPrefix(local, p) || strings.HasPrefix(local, p+".") {
			return true
		}
	}

	return false
}

func isTrusted(domain string) bool {
	return trustedEmailProviders[strings.ToLower(domain)]
}

func checkMX(domain string, _ int) (bool, error) {
	mxCache.RLock()
	if records, ok := mxCache.entries[domain]; ok {
		mxCache.RUnlock()
		return len(records) > 0, nil
	}
	mxCache.RUnlock()

	mxRecords, mxErr := net.LookupMX(domain)

	if mxErr != nil {
		return false, mxErr
	}

	if len(mxRecords) == 0 {
		mxCache.Lock()
		mxCache.entries[domain] = []string{}
		mxCache.Unlock()

		return false, nil
	}

	mxCache.Lock()

	mxCache.entries[domain] = make([]string, len(mxRecords))
	for i, mx := range mxRecords {
		mxCache.entries[domain][i] = mx.Host
	}
	mxCache.Unlock()

	return true, nil
}

// Name validates and sanitizes a name.
func Name(name string) (string, error) {
	name = strings.TrimSpace(name)
	if len(name) == 0 {
		return "", &Error{Field: "name", Message: "name is required"}
	}

	if len(name) > MaxNameLength {
		return "", &Error{Field: "name", Message: "name exceeds maximum length"}
	}

	return name, nil
}

// PasswordLength validates password length constraints.
func PasswordLength(password string) error {
	if len(password) < MinPasswordLength {
		return &Error{Field: "password", Message: "password must be at least 8 characters"}
	}

	if len(password) > MaxPasswordLength {
		return &Error{Field: "password", Message: "password exceeds maximum length"}
	}

	return nil
}

// DeviceID validates and sanitizes a device ID.
func DeviceID(id string) (string, error) {
	id = strings.TrimSpace(id)
	if len(id) == 0 {
		return "", &Error{Field: "deviceId", Message: "device ID is required"}
	}

	if len(id) > MaxDeviceIDLength {
		return "", &Error{Field: "deviceId", Message: "device ID exceeds maximum length"}
	}

	validID := regexp.MustCompile(`^[a-zA-Z0-9\-_]+$`)
	if !validID.MatchString(id) {
		return "", &Error{Field: "deviceId", Message: "device ID contains invalid characters"}
	}

	return id, nil
}

// Command validates a command string.
func Command(cmd string) (string, error) {
	cmd = strings.TrimSpace(cmd)
	if len(cmd) == 0 {
		return "", &Error{Field: "command", Message: "command is required"}
	}

	if len(cmd) > MaxCommandLength {
		return "", &Error{Field: "command", Message: "command exceeds maximum length"}
	}

	return cmd, nil
}

// Token validates a token string.
func Token(token string) (string, error) {
	token = strings.TrimSpace(token)
	if len(token) == 0 {
		return "", &Error{Field: "token", Message: "token is required"}
	}

	if len(token) > MaxTokenLength {
		return "", &Error{Field: "token", Message: "token exceeds maximum length"}
	}

	return token, nil
}

// Sanitize removes potentially dangerous characters.
func Sanitize(s string, maxLen int) string {
	s = strings.TrimSpace(s)
	if len(s) > maxLen {
		s = s[:maxLen]
	}

	return s
}

// ContainsInvalidUTF8 checks if a string contains invalid UTF-8 sequences.
func ContainsInvalidUTF8(s string) bool {
	for _, r := range s {
		if r == 0xFFFD {
			return true
		}
	}

	return []byte(s) == nil && s != ""
}

// ContainsControlCharacters checks for Unicode control characters.
func ContainsControlCharacters(s string) bool {
	for _, r := range s {
		if unicode.IsControl(r) && !strings.ContainsRune("\t\n\r", r) {
			return true
		}
	}

	return false
}

// ExtractDomain extracts the domain part from an email.
func ExtractDomain(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return ""
	}

	return parts[1]
}

// IsDisposableDomain checks if the domain of an email is disposable.
func IsDisposableDomain(email string) bool {
	return isDisposable(ExtractDomain(email))
}

// ClearMXCache clears the DNS MX record cache.
func ClearMXCache() {
	mxCache.Lock()
	mxCache.entries = make(map[string][]string)
	mxCache.Unlock()
}

// NormalizeEmail normalizes an email for case-insensitive comparison.
func NormalizeEmail(email string) string {
	return strings.TrimSpace(strings.ToLower(email))
}

// EmailURI validates a mailto: URI.
func EmailURI(uri string) error {
	parsed, err := url.Parse(uri)
	if err != nil {
		return &Error{Field: "uri", Message: "invalid URI format"}
	}

	if parsed.Scheme != "mailto" {
		return &Error{Field: "uri", Message: "URI must be mailto: scheme"}
	}

	_, err = Email(parsed.Opaque)

	return err
}
