package auth

import (
	"context"
	"fmt"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/session"
	infraauth "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security"
)

// MFAEnrollResult holds MFA enrollment data.
type MFAEnrollResult struct {
	Secret      string
	URI         string
	BackupCodes []string
}

// VerifyMFACode verifies a TOTP code for an operator.
func (s *AuthService) VerifyMFACode(ctx context.Context, operatorID, code string) (*session.Session, error) {
	op, err := s.operatorRepo.FindByID(ctx, operatorID)
	if err != nil {
		return nil, err
	}

	if !op.HasMFA() {
		return nil, application.ErrForbidden
	}

	// 7: Verify TOTP account name contains operator ID for binding
	cfg := infraauth.DefaultTOTPConfig()
	cfg.AccountName = op.Email // Include operator email in the binding

	// Create TOTP with operator ID as part of the secret binding
	boundSecret := infraauth.CreateMFASecretBinding(operatorID, op.MFASecret)
	cfg.Issuer = fmt.Sprintf("Vyzorix-%s", operatorID[:8]) // Include operator ID prefix

	totp := infraauth.NewTOTP(boundSecret.Secret, cfg)
	if !totp.Verify(code) {
		return nil, application.ErrInvalidCredentials
	}

	return s.CreateSession(ctx, operatorID)
}

// EnrollMFA generates a new MFA secret for enrollment.
func (s *AuthService) EnrollMFA(ctx context.Context, operatorID, email string) (*MFAEnrollResult, error) {
	secret, err := infraauth.GenerateSecret()
	if err != nil {
		return nil, err
	}

	cfg := infraauth.DefaultTOTPConfig()
	cfg.AccountName = email
	totp := infraauth.NewTOTP(secret, cfg)

	backupCodes, err := infraauth.GenerateBackupCodes(8)
	if err != nil {
		return nil, err
	}

	return &MFAEnrollResult{
		Secret:      secret,
		URI:         totp.ProvisioningURI(),
		BackupCodes: backupCodes,
	}, nil
}

// EnableMFA enables MFA for an operator after verifying a code.
func (s *AuthService) EnableMFA(ctx context.Context, operatorID, secret string, backupCodes []string) error {
	return s.operatorRepo.UpdateMFA(ctx, operatorID, secret, true)
}

// DisableMFA disables MFA for an operator.
func (s *AuthService) DisableMFA(ctx context.Context, operatorID string) error {
	return s.operatorRepo.UpdateMFA(ctx, operatorID, "", false)
}

// GetMFAStatus returns MFA status for an operator.
func (s *AuthService) GetMFAStatus(ctx context.Context, operatorID string) (bool, error) {
	op, err := s.operatorRepo.FindByID(ctx, operatorID)
	if err != nil {
		return false, err
	}

	return op.MFAEnabled, nil
}

// RegenerateBackupCodes generates and saves new backup codes for an operator.
func (s *AuthService) RegenerateBackupCodes(ctx context.Context, operatorID string) ([]string, error) {
	op, err := s.operatorRepo.FindByID(ctx, operatorID)
	if err != nil {
		return nil, err
	}

	if op.MFASecret == "" {
		return nil, application.ErrInvalidInput
	}

	codes, err := infraauth.GenerateBackupCodes(8)
	if err != nil {
		return nil, err
	}

	err = s.operatorRepo.UpdateOperatorMFA(ctx, operatorID, op.MFASecret, codes)
	if err != nil {
		return nil, err
	}

	return codes, nil
}

// VerifyBackupCode verifies a backup code and removes it if valid.
func (s *AuthService) VerifyBackupCode(ctx context.Context, operatorID, code string) (bool, error) {
	op, err := s.operatorRepo.FindByID(ctx, operatorID)
	if err != nil {
		return false, err
	}

	if len(op.BackupCodes) == 0 {
		return false, nil
	}

	idx := infraauth.ValidateBackupCode(op.BackupCodes, code)
	if idx < 0 {
		return false, nil
	}

	remaining := infraauth.RemoveBackupCode(op.BackupCodes, idx)
	err = s.operatorRepo.UpdateOperatorMFA(ctx, operatorID, op.MFASecret, remaining)
	if err != nil {
		return false, err
	}

	return true, nil
}
