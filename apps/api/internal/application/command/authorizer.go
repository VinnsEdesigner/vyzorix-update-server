package command

import (
	"context"
	"errors"

	confirmationapp "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/confirmation"
	domaincommand "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/command"
	domainconfirmation "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/confirmation"
)

// ConfirmationConsumer consumes a single-use confirmation token for a command,
// returning an error when the token is missing, expired, consumed, or mismatched.
type ConfirmationConsumer interface {
	ConsumeForCommand(ctx context.Context, token, operatorID, commandName, deviceID string) error
}

// Authorizer applies the command risk gate uniformly for every entry point
// (REST, GraphQL). Centralizing it guarantees no transport can dispatch a
// command while skipping the MFA or confirmation requirements.
type Authorizer struct {
	evaluator     *domaincommand.RiskEvaluator
	confirmations ConfirmationConsumer
}

// NewAuthorizer builds an Authorizer. confirmations may be nil; when nil,
// confirmation-gated commands are always blocked.
func NewAuthorizer(confirmations ConfirmationConsumer) *Authorizer {
	return &Authorizer{
		evaluator:     domaincommand.NewRiskEvaluator(),
		confirmations: confirmations,
	}
}

// NewAuthorizerFromService adapts the application confirmation service.
func NewAuthorizerFromService(svc *confirmationapp.Service) *Authorizer {
	if svc == nil {
		return NewAuthorizer(nil)
	}
	return NewAuthorizer(consumerFunc(svc.ConsumeForCommand))
}

// consumerFunc adapts the service method (which returns a PendingConfirmation)
// to the error-only ConfirmationConsumer interface.
type consumerFunc func(ctx context.Context, token, operatorID, commandName, deviceID string) (*domainconfirmation.PendingConfirmation, error)

func (f consumerFunc) ConsumeForCommand(ctx context.Context, token, operatorID, commandName, deviceID string) error {
	_, err := f(ctx, token, operatorID, commandName, deviceID)
	return err
}

// AuthorizeOutcome is the result of authorizing a command.
type AuthorizeOutcome struct {
	// Reason is the audit reason for a denial ("denied" or "confirmation required").
	Reason string
	// Message is the client-facing explanation for a denial.
	Message string
	// Tier is the command's risk tier (for audit).
	Tier domaincommand.RiskTier
	// Allowed reports whether the command may proceed to dispatch.
	Allowed bool
	// NeedsConfirmation is true when the block is a confirmation/MFA gate the
	// client can satisfy and retry (HTTP 425), as opposed to a hard deny (403).
	NeedsConfirmation bool
}

// Authorize evaluates the command's risk against the actor and, when required,
// consumes the confirmation token. Critical-tier commands additionally require
// an MFA-verified session; a token alone cannot authorize them.
func (a *Authorizer) Authorize(ctx context.Context, actor domaincommand.ActorContext, commandName, deviceID, confirmationToken string) AuthorizeOutcome {
	decision, profile := a.evaluator.Evaluate(commandName, actor)

	switch decision {
	case domaincommand.DecisionAllow:
		return AuthorizeOutcome{Allowed: true, Tier: profile.Tier}
	case domaincommand.DecisionDeny:
		return AuthorizeOutcome{Tier: profile.Tier, Reason: "denied", Message: "This command is not permitted."}
	case domaincommand.DecisionRequireConfirmation:
		return a.confirm(ctx, actor, profile, commandName, deviceID, confirmationToken)
	}

	// Unreachable: Decision only has the three values above.
	return AuthorizeOutcome{Tier: profile.Tier, Reason: "denied", Message: "This command is not permitted."}
}

// confirm resolves a RequireConfirmation decision: critical-tier commands need
// an MFA-verified session first, then all confirmation-gated commands need a
// valid, unconsumed, matching token.
func (a *Authorizer) confirm(ctx context.Context, actor domaincommand.ActorContext, profile domaincommand.CommandRiskProfile, commandName, deviceID, token string) AuthorizeOutcome {
	blocked := func(msg string) AuthorizeOutcome {
		return AuthorizeOutcome{Tier: profile.Tier, Reason: "confirmation required", Message: msg, NeedsConfirmation: true}
	}

	if profile.Tier == domaincommand.RiskTierCritical && !actor.MFAVerified {
		return blocked("This critical command requires an MFA-verified session before a confirmation token can be issued.")
	}
	if token == "" {
		return blocked("This command requires a confirmation token. Request one via POST /v1/device/:imei/command/confirm.")
	}
	if a.confirmations == nil {
		return blocked("Confirmations are not enabled on this server.")
	}
	if err := a.confirmations.ConsumeForCommand(ctx, token, actor.OperatorID, commandName, deviceID); err != nil {
		return blocked(confirmationErrorMessage(err))
	}
	return AuthorizeOutcome{Allowed: true, Tier: profile.Tier}
}

// confirmationErrorMessage maps a consume error to a client-facing message.
func confirmationErrorMessage(err error) string {
	switch {
	case errors.Is(err, domainconfirmation.ErrAlreadyConsumed):
		return "Confirmation token already used."
	case errors.Is(err, domainconfirmation.ErrExpired):
		return "Confirmation token expired."
	case errors.Is(err, domainconfirmation.ErrMismatch):
		return "Confirmation token does not match this command or device."
	case errors.Is(err, domainconfirmation.ErrNotFound):
		return "Confirmation token not found."
	default:
		return "Invalid or expired confirmation token."
	}
}
