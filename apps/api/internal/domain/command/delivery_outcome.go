package command

import (
	"errors"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/outcome"
)

var (
	ErrNoTransport   = errors.New("no transport available for command delivery")
	ErrSigningFailed = errors.New("command frame signing failed")
	ErrAckTimeout    = errors.New("device did not acknowledge within the delivery window")
	ErrSendFailed    = errors.New("transport reported a send failure")
)

type TransportResult struct {
	Err   error
	Sent  bool
	Acked bool
}

func ClassifyDelivery(result TransportResult) *outcome.Failure {
	switch {
	case result.Err != nil && !result.Sent:
		return outcome.Unavailable("command_transport_failed", "The command could not be sent to the device.").
			WithCause(outcome.Fault(result.Err))
	case result.Sent && !result.Acked:
		return outcome.Timeout("command_ack_timeout", "The device did not acknowledge the command in time.")
	case !result.Sent && result.Err == nil:
		return outcome.New(outcome.ClassUnavailable, outcome.BlameEnvironment, outcome.EffectNone,
			"command_not_sent", "The command was not sent.")
	case result.Err != nil:
		return outcome.Unavailable("command_transport_failed", "The command could not be sent to the device.").
			WithCause(outcome.Fault(result.Err))
	default:
		return nil
	}
}

type DeliveryPlan struct {
	Reason     string
	Retry      bool
	Reconcile  bool
	MarkFailed bool
}

func PlanDelivery(failure *outcome.Failure, tier RiskTier, retriesUsed, maxRetries int) DeliveryPlan {
	if failure == nil {
		return DeliveryPlan{}
	}

	remediation := outcome.Remediate(failure)

	if remediation.Action == outcome.ActionEscalate || remediation.Action == outcome.ActionReconcile {
		if tier.AtLeast(RiskTierHigh) {
			return DeliveryPlan{
				Reconcile:  true,
				MarkFailed: true,
				Reason:     remediation.Guidance,
			}
		}
	}

	if remediation.Action == outcome.ActionRetry && retriesUsed < maxRetries {
		return DeliveryPlan{Retry: true, Reason: remediation.Guidance}
	}
	if remediation.Action == outcome.ActionWait && retriesUsed < maxRetries {
		return DeliveryPlan{Retry: true, Reason: remediation.Guidance}
	}

	if len(failure.Owned()) == 0 &&
		outcome.DispositionFor(failure).NeedsReconciliation() &&
		!tier.AtLeast(RiskTierHigh) &&
		retriesUsed < maxRetries {
		return DeliveryPlan{Retry: true, Reason: "effect uncertain; retry is idempotent for this tier"}
	}

	return DeliveryPlan{MarkFailed: true, Reason: remediation.Guidance}
}
