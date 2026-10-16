// Package confirmation provides the business logic for issuing and consuming.
// single-use confirmation tokens that gate risky device commands.
package confirmation

import (
	"context"
	"errors"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/shared"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/command"
	domainconfirmation "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/confirmation"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/outcome"
)

// DefaultConfirmationTTL is used when the command's risk profile does not.
// specify a TTL. It mirrors the catalog's DefaultConfirmationTTL.
const DefaultConfirmationTTL = 5 * time.Minute

// Service issues and consumes confirmation tokens. It is the only writer to.
// the confirmation store; the command execution handler consumes tokens via.
// ConsumeForCommand.
type Service struct {
	repo domainconfirmation.Repository
}

// NewService creates a confirmation Service backed by the given repository.
func NewService(repo domainconfirmation.Repository) *Service {
	return &Service{repo: repo}
}

// RequestConfirmation issues a single-use token authorizing the given.
// operator to execute the named command (optionally on a specific device).
// The TTL is taken from the command's risk profile, falling back to.
// DefaultConfirmationTTL. The returned token must be presented back within.
// the TTL via ConsumeForCommand.
func (s *Service) RequestConfirmation(ctx context.Context, operatorID, orgID, commandName, deviceID string) (*domainconfirmation.PendingConfirmation, error) {
	if operatorID == "" || commandName == "" {
		return nil, errors.New("operatorID and commandName are required")
	}

	profile := command.LookupRiskProfile(commandName)
	ttl := profile.ConfirmationTTL
	if ttl <= 0 {
		ttl = DefaultConfirmationTTL
	}

	now := time.Now()
	c := &domainconfirmation.PendingConfirmation{
		Token:      shared.GenerateID(),
		OperatorID: operatorID,
		OrgID:      orgID,
		Command:    commandName,
		DeviceID:   deviceID,
		RiskTier:   string(profile.Tier),
		CreatedAt:  now,
		ExpiresAt:  now.Add(ttl),
	}
	if err := s.repo.Create(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

// ConsumeResult is ConsumeForCommand with the failure expressed as a typed
// outcome value instead of a bare error, so callers can tell a clean rejection
// (the token was already consumed: nothing happened) apart from an uncertain
// consume (the store write failed: the token may or may not be burned).
type ConsumeResult struct {
	Confirmation *domainconfirmation.PendingConfirmation
	Failure      *outcome.Failure
}

// ConsumeForCommandTyped behaves like ConsumeForCommand but classifies every
// failure. A caller that receives EffectUnknown must treat the token as
// possibly consumed and reconcile rather than blindly retrying, since a second
// successful consume of a destructive command's token would double-execute it.
func (s *Service) ConsumeForCommandTyped(ctx context.Context, token, operatorID, commandName, deviceID string) ConsumeResult {
	c, err := s.repo.Get(ctx, token)
	if err != nil {
		if errors.Is(err, domainconfirmation.ErrNotFound) {
			return ConsumeResult{Failure: outcome.NotFound("confirmation_not_found", "No confirmation token matches.")}
		}
		return ConsumeResult{Failure: outcome.Unavailable("confirmation_store_unavailable", "The confirmation store is unreachable.").WithCause(outcome.Fault(err))}
	}
	if c.IsExpired() {
		return ConsumeResult{Confirmation: c, Failure: outcome.Invalid("confirmation_expired", "This confirmation has expired.")}
	}
	if !c.Matches(operatorID, commandName, deviceID) {
		return ConsumeResult{Confirmation: c, Failure: outcome.Denied("confirmation_mismatch", "This confirmation does not match the request.")}
	}
	if c.IsConsumed() {
		// The token was already burned by an earlier request. Nothing changed
		// now, so this is a clean, effect-free rejection.
		return ConsumeResult{Confirmation: c, Failure: outcome.New(
			outcome.ClassConflict, outcome.BlameCaller, outcome.EffectNone,
			"confirmation_already_consumed", "This confirmation was already used.")}
	}

	consumed, err := s.repo.Consume(ctx, token, time.Now())
	if err != nil {
		// The store rejected or failed the write. We cannot tell whether the
		// token was burned before the failure, so the effect is unknown.
		return ConsumeResult{Confirmation: c, Failure: outcome.Conflict(
			"confirmation_consume_failed",
			"The confirmation could not be finalized; verify before retrying.").WithCause(outcome.Fault(err))}
	}
	return ConsumeResult{Confirmation: consumed}
}

// ConsumeForCommand validates and consumes a confirmation token for a specific.
// operator/command/device execution. It enforces ownership, command match,
// device scope, expiry, and single-use semantics, marking the token consumed.
// on success. Returns the consumed confirmation so the caller can include its.
// risk tier / timing in audit records.
func (s *Service) ConsumeForCommand(ctx context.Context, token, operatorID, commandName, deviceID string) (*domainconfirmation.PendingConfirmation, error) {
	result := s.ConsumeForCommandTyped(ctx, token, operatorID, commandName, deviceID)
	if result.Failure == nil {
		return result.Confirmation, nil
	}
	switch result.Failure.Code() {
	case "confirmation_not_found":
		return result.Confirmation, domainconfirmation.ErrNotFound
	case "confirmation_expired":
		return result.Confirmation, domainconfirmation.ErrExpired
	case "confirmation_mismatch":
		return result.Confirmation, domainconfirmation.ErrMismatch
	case "confirmation_already_consumed":
		return result.Confirmation, domainconfirmation.ErrAlreadyConsumed
	}
	return result.Confirmation, errors.New(result.Failure.Message())
}
