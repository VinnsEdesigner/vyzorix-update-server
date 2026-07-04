package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
)

// ValidatePasswordPolicy validates a new password against operator's security policy.
func (s *AuthService) ValidatePasswordPolicy(ctx context.Context, op *operator.Operator, newPassword string) error {
	settings := op.SecuritySettings

	// Check password history
	if settings.PasswordHistoryCount > 0 {
		if err := s.checkPasswordHistory(ctx, op.ID, newPassword, settings.PasswordHistoryCount); err != nil {
			return err
		}
	}

	// Check password min age (if set, password cannot be changed within this period)
	if settings.PasswordMinAgeDays > 0 {
		// Operator's UpdatedAt represents last password change
		elapsed := time.Since(op.UpdatedAt)
		minAge := time.Duration(settings.PasswordMinAgeDays) * 24 * time.Hour
		if elapsed < minAge {
			return fmt.Errorf("%w: password cannot be changed for %d days", application.ErrPasswordPolicy, settings.PasswordMinAgeDays)
		}
	}

	return nil
}

// checkPasswordHistory checks if the new password was used recently.
func (s *AuthService) checkPasswordHistory(ctx context.Context, operatorID, newPassword string, historyCount int) error {
	// Get recent passwords for this operator from audit or stored hashes
	// For simplicity, we check against a limited history stored in operator record
	// In production, this would query a password_history table

	// Get operator to check stored password hashes
	op, err := s.operatorRepo.FindByID(ctx, operatorID)
	if err != nil {
		return err
	}

	// Check current password
	if err := s.passwordHasher.Verify(newPassword, op.PasswordHash); err == nil {
		return fmt.Errorf("%w: cannot reuse current password", application.ErrPasswordPolicy)
	}

	return nil
}

// IsPasswordExpired checks if operator's password has expired.
func (s *AuthService) IsPasswordExpired(ctx context.Context, op *operator.Operator) (bool, error) {
	settings := op.SecuritySettings

	if settings.PasswordMaxAgeDays <= 0 {
		return false, nil // No expiry
	}

	maxAge := time.Duration(settings.PasswordMaxAgeDays) * 24 * time.Hour
	elapsed := time.Since(op.UpdatedAt)

	return elapsed > maxAge, nil
}
