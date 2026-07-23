package dto

import (
	"errors"
	"strings"
	"unicode/utf8"
)

// Validation constants for string length bounds.
const (
	// General string length limits.
	MaxEmailLength       = 255
	MaxNameLength        = 256
	MaxPasswordLength    = 128
	MinPasswordLength    = 8
	MaxTokenLength       = 512
	MaxCommandLength     = 256
	MaxNonceLength       = 128
	MaxDeviceIDLength    = 64
	MaxIMEILength        = 20
	MaxVersionLength     = 32
	MaxDeviceClassLength = 64
	MaxFCMTokenLength    = 4096
	MaxOrgNameLength     = 256
	MaxRoleLength        = 64
	MaxOrgIDLength       = 64
)

// ErrStringTooLong is returned when a string exceeds maximum length.
var ErrStringTooLong = errors.New("string exceeds maximum length")

// ErrStringTooShort is returned when a string is below minimum length.
var ErrStringTooShort = errors.New("string below minimum length")

// ErrEmptyString is returned when a required string is empty.
var ErrEmptyString = errors.New("string is empty")

// ValidateEmail validates email length and format.
func ValidateEmail(email string) error {
	if email == "" {
		return ErrEmptyString
	}
	if utf8.RuneCountInString(email) > MaxEmailLength {
		return ErrStringTooLong
	}
	// Basic email format check.
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return errors.New("invalid email format")
	}
	return nil
}

// ValidateName validates a name string.
func ValidateName(name string) error {
	if name == "" {
		return ErrEmptyString
	}
	if utf8.RuneCountInString(name) > MaxNameLength {
		return ErrStringTooLong
	}
	return nil
}

// ValidatePassword validates password length.
func ValidatePassword(password string) error {
	if password == "" {
		return ErrEmptyString
	}
	length := utf8.RuneCountInString(password)
	if length < MinPasswordLength {
		return ErrStringTooShort
	}
	if length > MaxPasswordLength {
		return ErrStringTooLong
	}
	return nil
}

// ValidateToken validates a token string.
func ValidateToken(token string) error {
	if token == "" {
		return ErrEmptyString
	}
	if utf8.RuneCountInString(token) > MaxTokenLength {
		return ErrStringTooLong
	}
	return nil
}

// ValidateCommand validates a command string.
func ValidateCommand(command string) error {
	if command == "" {
		return ErrEmptyString
	}
	if utf8.RuneCountInString(command) > MaxCommandLength {
		return ErrStringTooLong
	}
	return nil
}

// ValidateNonce validates a nonce string.
func ValidateNonce(nonce string) error {
	if utf8.RuneCountInString(nonce) > MaxNonceLength {
		return ErrStringTooLong
	}
	return nil
}

// ValidateDeviceID validates a device ID string.
func ValidateDeviceID(deviceID string) error {
	if deviceID == "" {
		return ErrEmptyString
	}
	if utf8.RuneCountInString(deviceID) > MaxDeviceIDLength {
		return ErrStringTooLong
	}
	return nil
}

// ValidateIMEI validates an IMEI string.
func ValidateIMEI(imei string) error {
	if imei == "" {
		return ErrEmptyString
	}
	if utf8.RuneCountInString(imei) > MaxIMEILength {
		return ErrStringTooLong
	}
	return nil
}

// ValidateVersion validates a version string.
func ValidateVersion(version string) error {
	if utf8.RuneCountInString(version) > MaxVersionLength {
		return ErrStringTooLong
	}
	return nil
}

// ValidateDeviceClass validates a device class string.
func ValidateDeviceClass(deviceClass string) error {
	if utf8.RuneCountInString(deviceClass) > MaxDeviceClassLength {
		return ErrStringTooLong
	}
	return nil
}

// ValidateFCMToken validates an FCM token string.
func ValidateFCMToken(token string) error {
	if utf8.RuneCountInString(token) > MaxFCMTokenLength {
		return ErrStringTooLong
	}
	return nil
}

// ValidateOrgName validates an organization name string.
func ValidateOrgName(name string) error {
	if name == "" {
		return ErrEmptyString
	}
	if utf8.RuneCountInString(name) > MaxOrgNameLength {
		return ErrStringTooLong
	}
	return nil
}

// ValidateRole validates a role string.
func ValidateRole(role string) error {
	if utf8.RuneCountInString(role) > MaxRoleLength {
		return ErrStringTooLong
	}
	return nil
}

// ValidateOrgID validates an organization ID string.
func ValidateOrgID(orgID string) error {
	if orgID == "" {
		return ErrEmptyString
	}
	if utf8.RuneCountInString(orgID) > MaxOrgIDLength {
		return ErrStringTooLong
	}
	return nil
}

// ValidateLoginRequest validates a login request.
func ValidateLoginRequest(req *LoginRequest) error {
	if err := ValidateEmail(req.Email); err != nil {
		return err
	}
	if err := ValidatePassword(req.Password); err != nil {
		return err
	}
	return nil
}

// ValidateRegisterRequest validates a registration request.
func ValidateRegisterRequest(req *RegisterRequest) error {
	if err := ValidateEmail(req.Email); err != nil {
		return err
	}
	if err := ValidatePassword(req.Password); err != nil {
		return err
	}
	if err := ValidateName(req.Name); err != nil {
		return err
	}
	if req.Role != "" {
		if err := ValidateRole(req.Role); err != nil {
			return err
		}
	}
	return nil
}

// ValidateDeviceRegisterRequest validates a device registration request.
func ValidateDeviceRegisterRequest(req *RegisterDeviceRequest) error {
	if err := ValidateDeviceID(req.DeviceID); err != nil {
		return err
	}
	if err := ValidateVersion(req.AppVersion); err != nil {
		return err
	}
	if err := ValidateDeviceClass(req.DeviceClass); err != nil {
		return err
	}
	// FCMToken and FirebaseInstallID are optional but have length limits if provided.
	if req.FCMToken != "" {
		if err := ValidateFCMToken(req.FCMToken); err != nil {
			return err
		}
	}
	return nil
}

// ValidateSendCommandRequest validates a send command request.
func ValidateSendCommandRequest(req *SendCommandRequest) error {
	if err := ValidateDeviceID(req.DeviceID); err != nil {
		return err
	}
	if err := ValidateCommand(req.Command); err != nil {
		return err
	}
	return nil
}

// ValidateUpdateNameRequest validates an update name request.
func ValidateUpdateNameRequest(req *UpdateNameRequest) error {
	// Name is required for update.
	return ValidateName(req.Name)
}

// ValidateForgotPasswordRequest validates a forgot password request.
func ValidateForgotPasswordRequest(req *ForgotPasswordRequest) error {
	return ValidateEmail(req.Email)
}

// ValidateResetPasswordRequest validates a reset password request.
func ValidateResetPasswordRequest(req *ResetPasswordRequest) error {
	if err := ValidateToken(req.Token); err != nil {
		return err
	}
	if err := ValidatePassword(req.NewPassword); err != nil {
		return err
	}
	return nil
}

// ValidateChangePasswordRequest validates a change password request.
func ValidateChangePasswordRequest(req *ChangePasswordRequest) error {
	if err := ValidatePassword(req.OldPassword); err != nil {
		return err
	}
	if err := ValidatePassword(req.NewPassword); err != nil {
		return err
	}
	return nil
}

// ValidateSelectOrganizationRequest validates a select organization request.
func ValidateSelectOrganizationRequest(req *SelectOrganizationRequest) error {
	return ValidateOrgID(req.OrganizationID)
}
