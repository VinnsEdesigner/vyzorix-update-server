package operator

import (
	"context"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
)

// ThresholdService handles threshold operations.
type ThresholdService struct {
	operatorRepo operator.Repository
}

// NewThresholdService creates a new ThresholdService.
func NewThresholdService(repo operator.Repository) *ThresholdService {
	return &ThresholdService{operatorRepo: repo}
}

// GetThresholds retrieves thresholds for an operator.
func (s *ThresholdService) GetThresholds(ctx context.Context, operatorID string) (operator.Thresholds, error) {
	return s.operatorRepo.GetThresholds(ctx, operatorID)
}

// UpdateThresholds updates thresholds for an operator.
func (s *ThresholdService) UpdateThresholds(ctx context.Context, operatorID string, input *operator.ThresholdsInput) (operator.Thresholds, error) {
	// Validate input
	if err := input.Validate(); err != nil {
		return operator.Thresholds{}, err
	}

	// Get current thresholds
	thresholds, err := s.operatorRepo.GetThresholds(ctx, operatorID)
	if err != nil {
		return operator.Thresholds{}, err
	}

	// Apply updates
	if input.RiskWarn != nil {
		thresholds.RiskWarn = *input.RiskWarn
	}
	if input.RiskCrit != nil {
		thresholds.RiskCrit = *input.RiskCrit
	}
	if input.ThermalWarn != nil {
		thresholds.ThermalWarn = *input.ThermalWarn
	}
	if input.ThermalCrit != nil {
		thresholds.ThermalCrit = *input.ThermalCrit
	}
	if input.BufferWarn != nil {
		thresholds.BufferWarn = *input.BufferWarn
	}
	if input.BufferCrit != nil {
		thresholds.BufferCrit = *input.BufferCrit
	}

	// Validate final state
	if thresholds.RiskWarn >= thresholds.RiskCrit {
		return operator.Thresholds{}, operator.ErrValidation
	}
	if thresholds.ThermalWarn >= thresholds.ThermalCrit {
		return operator.Thresholds{}, operator.ErrValidation
	}
	if thresholds.BufferCrit >= thresholds.BufferWarn {
		return operator.Thresholds{}, operator.ErrValidation
	}

	// Save
	err = s.operatorRepo.UpdateThresholds(ctx, operatorID, thresholds)
	if err != nil {
		return operator.Thresholds{}, err
	}

	return thresholds, nil
}
