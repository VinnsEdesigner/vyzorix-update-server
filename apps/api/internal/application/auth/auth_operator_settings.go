package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
)

// UpdateSettingsRequest represents a request to update operator settings.
type UpdateSettingsRequest struct {
	Thresholds *operator.Thresholds     `json:"thresholds,omitempty"`
	Client     *operator.ClientSettings `json:"client,omitempty"`
	Name       *string                  `json:"name,omitempty"`
	Reset      bool                     `json:"reset,omitempty"`
}

// UpdateOperatorName updates an operator's name.
func (s *AuthService) UpdateOperatorName(ctx context.Context, operatorID, name string) (*operator.Operator, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, application.ErrInvalidInput
	}

	op, err := s.operatorRepo.FindByID(ctx, operatorID)
	if err != nil {
		return nil, err
	}

	op.Name = name
	op.UpdatedAt = time.Now()

	if err := s.operatorRepo.Update(ctx, op); err != nil {
		return nil, err
	}

	return op, nil
}

// UpdateSettings updates an operator's settings.
func (s *AuthService) UpdateSettings(ctx context.Context, operatorID string, req *UpdateSettingsRequest) (*operator.Operator, error) {
	op, err := s.operatorRepo.FindByID(ctx, operatorID)
	if err != nil {
		if err == operator.ErrNotFound {
			return nil, application.ErrUnauthorized
		}
		return nil, err
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, application.ErrInvalidInput
		}

		op.Name = name
		if err := s.operatorRepo.Update(ctx, op); err != nil {
			return nil, err
		}
	}

	// Update thresholds in operator_settings table
	if req.Thresholds != nil {
		if err := s.operatorRepo.UpdateThresholds(ctx, operatorID, *req.Thresholds); err != nil {
			return nil, err
		}
		op.Thresholds = *req.Thresholds
	}

	// Update client settings in operator_settings table with validation
	if req.Client != nil {
		// Validate client settings
		if req.Client.RequestTimeoutMs < 500 || req.Client.RequestTimeoutMs > 60000 {
			return nil, fmt.Errorf("requestTimeoutMs must be between 500 and 60000")
		}
		if req.Client.LogBufferLimit < 50 || req.Client.LogBufferLimit > 5000 {
			return nil, fmt.Errorf("logBufferLimit must be between 50 and 5000")
		}
		if req.Client.SignalHistoryLimit < 30 || req.Client.SignalHistoryLimit > 2000 {
			return nil, fmt.Errorf("signalHistoryLimit must be between 30 and 2000")
		}

		if err := s.operatorRepo.UpdateClientSettings(ctx, operatorID, *req.Client); err != nil {
			return nil, err
		}
		op.ClientSettings = *req.Client
	}

	op.UpdatedAt = time.Now()

	return op, nil
}

// ResetSettings resets operator settings to defaults.
func (s *AuthService) ResetSettings(ctx context.Context, operatorID string) (*operator.Operator, error) {
	_, err := s.operatorRepo.FindByID(ctx, operatorID)
	if err != nil {
		if err == operator.ErrNotFound {
			return nil, application.ErrUnauthorized
		}
		return nil, err
	}

	if resetErr := s.operatorRepo.ResetSettings(ctx, operatorID); resetErr != nil {
		return nil, resetErr
	}

	// Reload operator to get fresh data
	op, err := s.operatorRepo.FindByID(ctx, operatorID)
	if err != nil {
		return nil, err
	}

	return op, nil
}
