package command

import (
	"context"

	confirmationapp "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/confirmation"
	domaincommand "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/command"
)

// Authorizer applies the command risk gate uniformly for every entry point
// (REST, GraphQL). It evaluates a command's risk profile against the actor and,
// when a confirmation is required, consumes the caller's confirmation token.
// Centralizing it guarantees no transport can dispatch a command while skipping
// the MFA or confirmation requirements the REST path enforces.
type Authorizer struct {
	evaluator     *domaincommand.RiskEvaluator
	confirmations *confirmationapp.Service
}

// NewAuthorizer builds an Authorizer. confirmations may be nil; when nil,
// confirmation-gated commands are always blocked.
func NewAuthorizer(confirmations *confirmationapp.Service) *Authorizer {
	return &Authorizer{
		evaluator:     domaincommand.NewRiskEvaluator(),
		confirmations: confirmations,
	}
}

// AuthorizeOutcome is the result of authorizing a command.
type AuthorizeOutcome struct {
	// Allowed reports whether the command may proceed to dispatch.
	Allowed bool
	// Profile is the command's risk profile (for audit).
	Profile domaincommand.CommandRiskProfile
	// Reason explains a denial (empty when Allowed).
	Reason string
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
		return AuthorizeOutcome{Allowed: true, Profile: profile}
	case domaincommand.DecisionDeny:
		return AuthorizeOutcome{Profile: profile, Reason: "denied"}
	}

	// RequireConfirmation.
	if profile.Tier == domaincommand.RiskTierCritical && !actor.MFAVerified {
		return AuthorizeOutcome{Profile: profile, Reason: "mfa required", NeedsConfirmation: true}
	}
	if confirmationToken == "" {
		return AuthorizeOutcome{Profile: profile, Reason: "confirmation required", NeedsConfirmation: true}
	}
	if a.confirmations == nil {
		return AuthorizeOutcome{Profile: profile, Reason: "confirmations disabled", NeedsConfirmation: true}
	}
	if _, err := a.confirmations.ConsumeForCommand(ctx, confirmationToken, actor.OperatorID, commandName, deviceID); err != nil {
		return AuthorizeOutcome{Profile: profile, Reason: "confirmation invalid", NeedsConfirmation: true}
	}
	return AuthorizeOutcome{Allowed: true, Profile: profile}
}
